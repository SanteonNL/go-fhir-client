/*
 * This program is free software: you can redistribute it and/or modify
 * it under the terms of the GNU General Public License as published by
 * the Free Software Foundation, either version 3 of the License, or
 * (at your option) any later version.
 *
 * This program is distributed in the hope that it will be useful,
 * but WITHOUT ANY WARRANTY; without even the implied warranty of
 * MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
 * GNU General Public License for more details.
 *
 * You should have received a copy of the GNU General Public License
 * along with this program.  If not, see <https://www.gnu.org/licenses/>.
 */

package fhirclient

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
)

const FhirJsonMediaType = "application/fhir+json"

type Client interface {
	// Read reads a resource at the given path from the FHIR server and unmarshals it into the target.
	// Options can be used to, e.g., add query parameters to the request.
	Read(path string, target any, opts ...Option) error
	// Create creates a new resource on the FHIR server.
	// The path is derived from the resource's resourceType.
	// The response is unmarshaled into the result.
	Create(resource any, result any, opts ...Option) error
	// Update updates the resource at the given path on the FHIR server.
	// The response is unmarshaled into the result.
	Update(path string, resource any, result any, opts ...Option) error
	// Path returns the full URL for the given path.
	Path(path ...string) *url.URL
}

type HttpRequestDoer interface {
	Do(req *http.Request) (*http.Response, error)
}

// New creates a new FHIR client with the given base URL and HTTP client.
// The base URL should point to the FHIR server's base URL, e.g. https://example.com/fhir
// If no config is passed, the default configuration is used.
func New(fhirBaseURL *url.URL, httpClient HttpRequestDoer, config *Config) *BaseClient {
	var cfg Config
	if config != nil {
		cfg = *config
		if cfg.MaxResponseSize == 0 {
			// In case people supply a config but forget to set max. response size.
			cfg.MaxResponseSize = DefaultConfig().MaxResponseSize
		}
	} else {
		cfg = DefaultConfig()
	}
	return &BaseClient{
		baseURL:    fhirBaseURL,
		httpClient: httpClient,
		config:     cfg,
	}
}

type Config struct {
	// Non2xxStatusHandler is called when a non-2xx status code is returned by the FHIR server.
	// Its primary use is logging.
	Non2xxStatusHandler func(response *http.Response, responseBody []byte)
	// MaxResponseSize is the maximum size of a response body in bytes that will be read.
	MaxResponseSize int
}

func DefaultConfig() Config {
	return Config{
		// 10mb
		MaxResponseSize: 10 * 1024 * 1024,
	}
}

var _ Client = &BaseClient{}

// BaseClient is a basic FHIR client that can read, create and update resources.
type BaseClient struct {
	baseURL    *url.URL
	httpClient HttpRequestDoer
	config     Config
}

func (d BaseClient) Path(path ...string) *url.URL {
	return d.baseURL.JoinPath(path...)
}

func (d BaseClient) Read(path string, target any, opts ...Option) error {
	opts = append([]Option{AtPath(path)}, opts...)
	httpRequest, err := http.NewRequest(http.MethodGet, d.baseURL.String(), nil)
	if err != nil {
		return err
	}
	httpRequest.Header.Add("Cache-Control", "no-cache")
	return d.doRequest(httpRequest, target, opts...)
}

func (d BaseClient) Create(resource any, result any, opts ...Option) error {
	desc, err := DescribeResource(resource)
	if err != nil {
		return err
	}
	opts = append([]Option{AtPath(desc.Type)}, opts...)
	httpRequest, err := http.NewRequest(http.MethodPost, d.baseURL.String(), io.NopCloser(bytes.NewReader(desc.Data)))
	if err != nil {
		return err
	}

	httpRequest.Header.Add("Content-Type", FhirJsonMediaType)
	return d.doRequest(httpRequest, result, opts...)
}

func (d BaseClient) Update(path string, resource any, result any, opts ...Option) error {
	data, err := json.Marshal(resource)
	if err != nil {
		return err
	}
	opts = append([]Option{AtPath(path)}, opts...)
	httpRequest, err := http.NewRequest(http.MethodPut, d.baseURL.String(), io.NopCloser(bytes.NewReader(data)))
	if err != nil {
		return err
	}
	httpRequest.Header.Add("Content-Type", FhirJsonMediaType)
	return d.doRequest(httpRequest, result, opts...)
}

func (d BaseClient) doRequest(httpRequest *http.Request, target any, opts ...Option) error {
	httpRequest.Header.Add("Accept", FhirJsonMediaType)
	// Execute pre-request options
	for _, opt := range opts {
		if fn, ok := opt.(PreRequestOption); ok {
			fn(d, httpRequest)
		}
	}
	// recreate HTTP request in case URL, body or method was edited by one of the options
	newHttpRequest, err := http.NewRequestWithContext(httpRequest.Context(), httpRequest.Method, httpRequest.URL.String(), httpRequest.Body)
	if err != nil {
		return err
	}
	newHttpRequest.Header = httpRequest.Header
	*httpRequest = *newHttpRequest

	httpResponse, err := d.httpClient.Do(httpRequest)
	if err != nil {
		return fmt.Errorf("FHIR request failed (%s %s): %w", httpRequest.Method, httpRequest.URL.String(), err)
	}
	defer httpResponse.Body.Close()
	data, err := io.ReadAll(io.LimitReader(httpResponse.Body, int64(d.config.MaxResponseSize+1)))
	if err != nil {
		return fmt.Errorf("FHIR response read failed (%s %s): %w", httpRequest.Method, httpRequest.URL.String(), err)
	}
	if httpResponse.StatusCode < 200 || httpResponse.StatusCode >= 300 {
		if d.config.Non2xxStatusHandler != nil {
			d.config.Non2xxStatusHandler(httpResponse, data)
		}
		return fmt.Errorf("FHIR request failed (%s %s, status=%d)", httpRequest.Method, httpRequest.URL.String(), httpResponse.StatusCode)
	}
	if len(data) > d.config.MaxResponseSize {
		return fmt.Errorf("FHIR response exceeds max. safety limit of %d bytes (%s %s, status=%d)", d.config.MaxResponseSize, httpRequest.Method, httpRequest.URL.String(), httpResponse.StatusCode)
	}
	// TODO: Handle errornous responses (OperationOutcome?)
	err = json.Unmarshal(data, target)
	if err != nil {
		return fmt.Errorf("FHIR response unmarshal failed (%s %s, status=%d): %w", httpRequest.Method, httpRequest.URL.String(), httpResponse.StatusCode, err)
	}
	for _, opt := range opts {
		if fn, ok := opt.(PostParseOption); ok {
			if err := fn(d, target); err != nil {
				return err
			}
		}
	}
	return nil
}

// DescribeResource is used to extract often-used information from a resource.
func DescribeResource(resource any) (*ResourceDescription, error) {
	data, err := json.Marshal(resource)
	if err != nil {
		return nil, fmt.Errorf("invalid resource of type %T: %w", resource, err)
	}
	var desc ResourceDescription
	if err := json.Unmarshal(data, &desc); err != nil {
		return nil, fmt.Errorf("invalid resource of type %T: %w", resource, err)
	}
	if desc.Type == "" {
		return nil, fmt.Errorf("resourceType not present in resource of type %T", resource)
	}
	desc.Data = data
	return &desc, nil
}

// ResourceDescription contains information about a resource.
type ResourceDescription struct {
	// Type is the resource type, e.g. "Patient".
	Type string `json:"resourceType"`
	// Data is the JSON representation of the resource, so that callers don't need to marshal it again.
	Data []byte `json:"-"`
}

type Option any

type PreRequestOption func(client Client, r *http.Request)

type PostParseOption func(client Client, result any) error

func QueryParam(key, value string) PreRequestOption {
	return func(_ Client, r *http.Request) {
		q := r.URL.Query()
		q.Add(key, value)
		r.URL.RawQuery = q.Encode()
	}
}

// AtPath sets the path of the request. The path is appended to the base URL.
func AtPath(path string) PreRequestOption {
	return func(client Client, r *http.Request) {
		r.URL = client.Path(path)
	}
}
