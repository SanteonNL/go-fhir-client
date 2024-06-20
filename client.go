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
	Read(path string, target any, opts ...Option) error
	Create(path string, resource any, result any) error
	Update(path string, resource any, result any) error
}

type HttpRequestDoer interface {
	Do(req *http.Request) (*http.Response, error)
}

func New(fhirBaseURL *url.URL, httpClient HttpRequestDoer) *BaseClient {
	return &BaseClient{
		baseURL:    fhirBaseURL,
		httpClient: httpClient,
	}
}

var _ Client = &BaseClient{}

type BaseClient struct {
	baseURL    *url.URL
	httpClient HttpRequestDoer
}

func (d BaseClient) Read(path string, target any, opts ...Option) error {
	httpRequest, err := http.NewRequest(http.MethodGet, d.resourceURL(path).String(), nil)
	if err != nil {
		return err
	}
	httpRequest.Header.Add("Cache-Control", "no-cache")
	return d.doRequest(httpRequest, target, opts...)
}

func (d BaseClient) Create(path string, resource any, result any) error {
	data, err := json.Marshal(resource)
	if err != nil {
		return err
	}
	httpRequest, err := http.NewRequest(http.MethodPost, d.resourceURL(path).String(), io.NopCloser(bytes.NewReader(data)))
	if err != nil {
		return err
	}

	httpRequest.Header.Add("Content-Type", FhirJsonMediaType)
	return d.doRequest(httpRequest, result)
}

func (d BaseClient) Update(path string, resource any, result any) error {
	data, err := json.Marshal(resource)
	if err != nil {
		return err
	}
	httpRequest, err := http.NewRequest(http.MethodPut, d.resourceURL(path).String(), io.NopCloser(bytes.NewReader(data)))
	if err != nil {
		return err
	}
	httpRequest.Header.Add("Content-Type", FhirJsonMediaType)
	return d.doRequest(httpRequest, result)
}

func (d BaseClient) resourceURL(path string) *url.URL {
	return d.baseURL.JoinPath(path)
}

func (d BaseClient) doRequest(httpRequest *http.Request, target any, opts ...Option) error {
	for _, opt := range opts {
		opt(httpRequest)
	}
	httpRequest.Header.Add("Accept", FhirJsonMediaType)
	httpResponse, err := d.httpClient.Do(httpRequest)
	if err != nil {
		return fmt.Errorf("FHIR request failed (url=%s): %w", httpRequest.URL.String(), err)
	}
	defer httpResponse.Body.Close()
	if httpResponse.StatusCode < 200 || httpResponse.StatusCode >= 300 {
		return fmt.Errorf("FHIR request failed (url=%s, status=%d)", httpRequest.URL.String(), httpResponse.StatusCode)
	}
	data, err := io.ReadAll(httpResponse.Body)
	if err != nil {
		return fmt.Errorf("FHIR response read failed (url=%s): %w", httpRequest.URL.String(), err)
	}
	// TODO: Handle errornous responses (OperationOutcome?)
	err = json.Unmarshal(data, target)
	if err != nil {
		return fmt.Errorf("FHIR response unmarshal failed (url=%s): %w", httpRequest.URL.String(), err)
	}
	return nil
}

type Option func(r *http.Request)

func QueryParam(key, value string) Option {
	return func(r *http.Request) {
		q := r.URL.Query()
		q.Add(key, value)
		r.URL.RawQuery = q.Encode()
	}
}
