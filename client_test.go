//go:generate mockgen -destination=client_mock_test.go -package=fhirclient_test -source=client.go
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

package fhirclient_test

import (
	"bytes"
	"encoding/json"
	"io"
	"net/http"
	"net/url"
	"testing"

	fhirclient "github.com/SanteonNL/go-fhir-client"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestDefaultClient_Read(t *testing.T) {
	t.Run("by ID", func(t *testing.T) {
		stub := &requestResponder{
			response: okResponse(Resource{Id: "123"}),
		}
		client := fhirclient.New(baseURL, stub, nil)
		var result Resource

		err := client.Read("Resource/123", &result)

		require.NoError(t, err)
		require.NotNil(t, result)
		assert.Equal(t, "http://example.com/fhir/Resource/123", stub.request.URL.String())
		t.Run("unmarshals properly", func(t *testing.T) {
			require.Equal(t, "123", result.Id)
		})
		t.Run("right accept header", func(t *testing.T) {
			assert.Equal(t, fhirclient.FhirJsonMediaType, stub.request.Header.Get("Accept"))
		})
	})
	t.Run("with query params", func(t *testing.T) {
		stub := &requestResponder{
			response: okResponse(Resource{Id: "123"}),
		}
		client := fhirclient.New(baseURL, stub, nil)
		var result Resource

		err := client.Read("Resource", &result, fhirclient.QueryParam("_id", "123"), fhirclient.QueryParam("_count", "1"))

		require.NoError(t, err)
		require.NotNil(t, result)
		assert.Equal(t, "http://example.com/fhir/Resource?_count=1&_id=123", stub.request.URL.String())
	})
	t.Run("at path", func(t *testing.T) {
		stub := &requestResponder{
			response: okResponse(Resource{Id: "123"}),
		}
		client := fhirclient.New(baseURL, stub, nil)
		var result Resource

		err := client.Read("Resource", &result, fhirclient.AtPath("123"))

		require.NoError(t, err)
		require.NotNil(t, result)
		assert.Equal(t, "http://example.com/fhir/123", stub.request.URL.String())
	})
}

func TestBaseClient_Create(t *testing.T) {
	t.Run("derive path from resource type", func(t *testing.T) {
		stub := &requestResponder{
			response: okResponse(Resource{Id: "123"}),
		}
		client := fhirclient.New(baseURL, stub, nil)
		var result Resource

		err := client.Create(Resource{Id: "123"}, &result)

		require.NoError(t, err)
		require.NotNil(t, result)
		assert.Equal(t, "http://example.com/fhir/Resource", stub.request.URL.String())
	})
	t.Run("path is set using option", func(t *testing.T) {
		stub := &requestResponder{
			response: okResponse(Resource{Id: "123"}),
		}
		client := fhirclient.New(baseURL, stub, nil)
		var result Resource

		err := client.Create(Resource{Id: "123"}, &result, fhirclient.AtPath("123"))

		require.NoError(t, err)
		require.NotNil(t, result)
		assert.Equal(t, "http://example.com/fhir/123", stub.request.URL.String())
	})
}

func TestDefaultClient_doRequest(t *testing.T) {
	t.Run("non-2xx status code", func(t *testing.T) {
		stub := &requestResponder{
			response: &http.Response{
				StatusCode: http.StatusNotFound,
				Header: map[string][]string{
					"Content-Type": {fhirclient.FhirJsonMediaType},
				},
				Body: io.NopCloser(bytes.NewReader([]byte(`{}`))),
			},
		}
		var result Resource
		t.Run("with handler", func(t *testing.T) {
			var capturedResponseBody []byte
			client := fhirclient.New(baseURL, stub, &fhirclient.Config{
				Non2xxStatusHandler: func(response *http.Response, responseBody []byte) {
					capturedResponseBody = responseBody
				},
			})

			err := client.Read("Resource/123", &result)

			require.Error(t, err)
			assert.Equal(t, []byte(`{}`), capturedResponseBody)
		})
		t.Run("without handler", func(t *testing.T) {
			client := fhirclient.New(baseURL, stub, nil)

			err := client.Read("Resource/123", &result)

			require.Error(t, err)
		})
	})
	t.Run("200 status code & OperationOutcomeError", func(t *testing.T) {
		stub := &requestResponder{
			response: &http.Response{
				StatusCode: http.StatusOK,
				Header: map[string][]string{
					"Content-Type": {fhirclient.FhirJsonMediaType},
				},
				Body: io.NopCloser(bytes.NewReader([]byte(`{"resourceType":"OperationOutcomeError","issue":[{"severity":"error","code":"processing","diagnostics":"some error message"}]}`))),
			},
		}
		var result Resource
		t.Run("without handler", func(t *testing.T) {
			client := fhirclient.New(baseURL, stub, nil)

			err := client.Read("Resource/123", &result)
			assert.IsType(t, fhirclient.OperationOutcomeError{}, err)
			assert.Equal(t, http.StatusOK, err.(fhirclient.OperationOutcomeError).HttpStatusCode)
			assert.EqualError(t, err, "OperationOutcomeError, issues: [processing error] some error message")
		})
	})
	t.Run("non-2xx status code & OperationOutcomeError", func(t *testing.T) {
		stub := &requestResponder{
			response: &http.Response{
				StatusCode: http.StatusNotFound,
				Header: map[string][]string{
					"Content-Type": {fhirclient.FhirJsonMediaType},
				},
				Body: io.NopCloser(bytes.NewReader([]byte(`{"resourceType":"OperationOutcomeError","issue":[{"severity":"error","code":"processing","diagnostics":"some error message"}]}`))),
			},
		}
		var result Resource
		t.Run("without handler", func(t *testing.T) {
			client := fhirclient.New(baseURL, stub, nil)

			err := client.Read("Resource/123", &result)
			assert.IsType(t, fhirclient.OperationOutcomeError{}, err)
			assert.Equal(t, http.StatusNotFound, err.(fhirclient.OperationOutcomeError).HttpStatusCode)
			assert.EqualError(t, err, "OperationOutcomeError, issues: [processing error] some error message")
		})
	})
	t.Run("unmarshal as []byte", func(t *testing.T) {
		stub := &requestResponder{
			response: &http.Response{
				StatusCode: http.StatusOK,
				Header: map[string][]string{
					"Content-Type": {fhirclient.FhirJsonMediaType},
				},
				Body: io.NopCloser(bytes.NewReader([]byte(`{"key":"value"}`))),
			},
		}
		client := fhirclient.New(baseURL, stub, nil)
		var result []byte

		err := client.Read("Resource/123", &result)

		require.NoError(t, err)
		assert.Equal(t, []byte(`{"key":"value"}`), result)
	})
	t.Run("max. response size exceeded", func(t *testing.T) {
		stub := &requestResponder{
			response: &http.Response{
				StatusCode: http.StatusOK,
				Header: map[string][]string{
					"Content-Type": {fhirclient.FhirJsonMediaType},
				},
				Body: io.NopCloser(bytes.NewReader([]byte("Hello, World!"))),
			},
		}
		client := fhirclient.New(baseURL, stub, &fhirclient.Config{MaxResponseSize: 2})
		var result Resource

		err := client.Read("Resource/123", &result)

		require.EqualError(t, err, "FHIR response exceeds max. safety limit of 2 bytes (GET http://example.com/fhir/Resource/123, status=200)")
	})
	t.Run("caller passes an absolute URL", func(t *testing.T) {
		stub := &requestResponder{
			response: okResponse(Resource{Id: "123"}),
		}
		client := fhirclient.New(baseURL, stub, nil)
		var result Resource

		err := client.Read("http://example.com/fhir/Resource/123", &result)

		require.NoError(t, err)
		require.NotNil(t, result)
		assert.Equal(t, "http://example.com/fhir/Resource/123", stub.request.URL.String())
	})
}

func TestResponseHeaders(t *testing.T) {
	t.Run("response headers are copied", func(t *testing.T) {
		stub := &requestResponder{
			response: &http.Response{
				StatusCode: http.StatusOK,
				Header: map[string][]string{
					"Content-Type": {fhirclient.FhirJsonMediaType},
					"X-Custom":     {"value"},
					"Date":         {"Mon, 02 Jan 2006 15:04:05 GMT"},
					"LastModified": {"Mon, 02 Jan 2020 15:04:05 GMT"},
					"ETag":         {"123456789"},
				},
				Body: io.NopCloser(bytes.NewReader([]byte(`{}`))),
			},
		}
		client := fhirclient.New(baseURL, stub, nil)
		var result Resource

		var actual fhirclient.Headers
		err := client.Read("Resource/123", &result, fhirclient.ResponseHeaders(&actual))

		require.NoError(t, err)
		assert.Equal(t, "value", actual.Get("X-Custom"))
		assert.Equal(t, "2006-01-02 15:04:05 +0000 UTC", actual.Date.String())
		assert.Equal(t, "2020-01-02 15:04:05 +0000 UTC", actual.LastModified.String())
		assert.Equal(t, "123456789", actual.ETag)
	})
}

var _ json.Marshaler = &Resource{}

type Resource struct {
	Id string `json:"id"`
}

func (r Resource) MarshalJSON() ([]byte, error) {
	return json.Marshal(map[string]interface{}{
		"id":           r.Id,
		"resourceType": "Resource",
	})
}

func okResponse(resource interface{}) *http.Response {
	data, _ := json.Marshal(resource)
	return &http.Response{
		StatusCode: http.StatusOK,
		Header: map[string][]string{
			"Content-Type": {fhirclient.FhirJsonMediaType},
		},
		Body: io.NopCloser(bytes.NewReader(data)),
	}
}

type requestResponder struct {
	request  *http.Request
	response *http.Response
}

func (s *requestResponder) Do(req *http.Request) (*http.Response, error) {
	s.request = req
	return s.response, nil
}

type requestsResponder struct {
	requests  []*http.Request
	responses []*http.Response
}

func (s *requestsResponder) Do(req *http.Request) (*http.Response, error) {
	s.requests = append(s.requests, req)
	return s.responses[len(s.requests)-1], nil
}

var baseURL, _ = url.Parse("http://example.com/fhir")
