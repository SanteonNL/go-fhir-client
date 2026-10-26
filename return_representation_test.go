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
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"

	fhirclient "github.com/SanteonNL/go-fhir-client"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// representationServer is a FHIR server whose create and update either return the stored resource or
// answer without a body, as a server ignoring Prefer: return=representation would.
type representationServer struct {
	withBody     bool
	withLocation bool
	reads        []string
	prefer       string
}

func (s *representationServer) start(t *testing.T) *url.URL {
	t.Helper()
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", fhirclient.FhirJsonMediaType)
		switch r.Method {
		case http.MethodGet:
			s.reads = append(s.reads, r.URL.Path)
			w.Header().Set("ETag", `W/"1"`)
			_, _ = w.Write([]byte(`{"resourceType":"Resource","id":"123"}`))
		case http.MethodPost, http.MethodPut:
			s.prefer = r.Header.Get("Prefer")
			if s.withLocation {
				w.Header().Set("Location", "http://"+r.Host+"/fhir/Resource/123/_history/1")
			}
			w.Header().Set("ETag", `W/"1"`)
			w.Header().Set("Last-Modified", "Mon, 02 Jan 2020 15:04:05 GMT")
			if r.Method == http.MethodPost {
				w.WriteHeader(http.StatusCreated)
			}
			if s.withBody {
				_, _ = w.Write([]byte(`{"resourceType":"Resource","id":"123"}`))
			}
		}
	}))
	t.Cleanup(server.Close)
	base, err := url.Parse(server.URL + "/fhir")
	require.NoError(t, err)
	return base
}

func TestReturnRepresentation(t *testing.T) {
	t.Run("the returned resource is used, and asked for", func(t *testing.T) {
		server := &representationServer{withBody: true, withLocation: true}
		client := fhirclient.New(server.start(t), http.DefaultClient, nil)
		var result Resource

		err := client.Create(Resource{}, &result, fhirclient.ReturnRepresentation())

		require.NoError(t, err)
		assert.Equal(t, "123", result.Id)
		assert.Equal(t, "return=representation", server.prefer)
		assert.Empty(t, server.reads, "no extra read")
	})

	t.Run("without a body, a create reads the Location", func(t *testing.T) {
		server := &representationServer{withLocation: true}
		client := fhirclient.New(server.start(t), http.DefaultClient, nil)
		var result Resource
		var headers fhirclient.Headers
		var status int

		err := client.Create(Resource{}, &result, fhirclient.ReturnRepresentation(),
			fhirclient.ResponseHeaders(&headers), fhirclient.ResponseStatusCode(&status))

		require.NoError(t, err)
		assert.Equal(t, "123", result.Id)
		assert.Equal(t, []string{"/fhir/Resource/123/_history/1"}, server.reads)
		assert.Equal(t, http.StatusCreated, status, "response options describe the create, not the read")
		assert.Contains(t, headers.Get("Location"), "/fhir/Resource/123/_history/1")
	})

	t.Run("without a body or Location, an update reads the URL it wrote to", func(t *testing.T) {
		server := &representationServer{}
		client := fhirclient.New(server.start(t), http.DefaultClient, nil)
		var result Resource

		err := client.Update("Resource/123", Resource{Id: "123"}, &result, fhirclient.ReturnRepresentation())

		require.NoError(t, err)
		assert.Equal(t, "123", result.Id)
		assert.Equal(t, []string{"/fhir/Resource/123"}, server.reads)
	})

	t.Run("without a body or Location, a create fails", func(t *testing.T) {
		server := &representationServer{}
		client := fhirclient.New(server.start(t), http.DefaultClient, nil)

		err := client.Create(Resource{}, new(Resource), fhirclient.ReturnRepresentation())

		assert.ErrorContains(t, err, "neither the stored resource nor its location")
	})

	t.Run("a raw target gets the read resource too", func(t *testing.T) {
		server := &representationServer{withLocation: true}
		client := fhirclient.New(server.start(t), http.DefaultClient, nil)
		var raw []byte

		err := client.Create(Resource{}, &raw, fhirclient.ReturnRepresentation())

		require.NoError(t, err)
		assert.JSONEq(t, `{"resourceType":"Resource","id":"123"}`, string(raw))
	})
}

func TestEmptyWriteResponse(t *testing.T) {
	t.Run("a create without a body leaves the target untouched", func(t *testing.T) {
		server := &representationServer{withLocation: true}
		client := fhirclient.New(server.start(t), http.DefaultClient, nil)
		result := Resource{Id: "unchanged"}

		err := client.Create(Resource{}, &result)

		require.NoError(t, err)
		assert.Equal(t, "unchanged", result.Id)
		assert.Empty(t, server.reads, "without ReturnRepresentation, nothing is read")
	})
}

func TestResponseHeaders_CanonicalHeaders(t *testing.T) {
	server := &representationServer{withBody: true}
	client := fhirclient.New(server.start(t), http.DefaultClient, nil)
	var headers fhirclient.Headers

	err := client.Update("Resource/123", Resource{Id: "123"}, new(Resource), fhirclient.ResponseHeaders(&headers))

	require.NoError(t, err)
	assert.Equal(t, `W/"1"`, headers.ETag, "ETag arrives as Etag over HTTP")
	assert.Equal(t, "2020-01-02 15:04:05 +0000 UTC", headers.LastModified.String(), "from the Last-Modified header")
}
