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
	"context"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"net/url"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/zorgbijjou/golang-fhir-models/fhir-models/fhir"
)

func TestPaginate(t *testing.T) {
	baseURL, _ := url.Parse("http://example.com/fhir")

	t.Run("successful pagination through multiple pages", func(t *testing.T) {
		// Create test bundles
		firstBundle := createBundleWithNextLink("http://example.com/fhir/page2")
		secondBundle := createBundleWithNextLink("http://example.com/fhir/page3")
		thirdBundle := createBundleWithoutNextLink()

		// Create responses for subsequent pages
		page2Response := createBundleResponse(secondBundle)
		page3Response := createBundleResponse(thirdBundle)

		stub := &requestsResponder{
			responses: []*http.Response{page2Response, page3Response},
		}

		client := New(baseURL, stub, nil)
		ctx := context.Background()

		// Track consumed bundles
		consumedBundles := []*fhir.Bundle{}
		consumeFunc := func(bundle *fhir.Bundle) (bool, error) {
			consumedBundles = append(consumedBundles, bundle)
			return true, nil // Continue pagination
		}

		err := Paginate(ctx, client, firstBundle, consumeFunc)

		require.NoError(t, err)
		assert.Len(t, consumedBundles, 3)

		// Verify the correct URLs were called
		require.Len(t, stub.requests, 2)
		assert.Equal(t, "http://example.com/fhir/page2", stub.requests[0].URL.String())
		assert.Equal(t, "http://example.com/fhir/page3", stub.requests[1].URL.String())
	})

	t.Run("early termination when consume function returns false", func(t *testing.T) {
		bundle := createBundleWithNextLink("http://example.com/fhir/page2")

		stub := &requestsResponder{
			responses: []*http.Response{}, // No responses needed since we'll stop early
		}

		client := New(baseURL, stub, nil)
		ctx := context.Background()

		callCount := 0
		consumeFunc := func(bundle *fhir.Bundle) (bool, error) {
			callCount++
			return false, nil // Stop pagination immediately
		}

		err := Paginate(ctx, client, bundle, consumeFunc)

		require.NoError(t, err)
		assert.Equal(t, 1, callCount)
		assert.Len(t, stub.requests, 0) // No additional requests should be made
	})

	t.Run("error from consume function", func(t *testing.T) {
		bundle := createBundleWithNextLink("http://example.com/fhir/page2")
		expectedError := errors.New("consume function error")

		stub := &requestsResponder{
			responses: []*http.Response{},
		}

		client := New(baseURL, stub, nil)
		ctx := context.Background()

		consumeFunc := func(bundle *fhir.Bundle) (bool, error) {
			return false, expectedError
		}

		err := Paginate(ctx, client, bundle, consumeFunc)

		require.Error(t, err)
		assert.Equal(t, expectedError, err)
	})

	t.Run("error from SearchWithContext", func(t *testing.T) {
		bundle := createBundleWithNextLink("http://example.com/fhir/page2")

		// Return a 500 error response
		errorResponse := &http.Response{
			StatusCode: http.StatusInternalServerError,
			Header: map[string][]string{
				"Content-Type": {FhirJsonMediaType},
			},
			Body: io.NopCloser(bytes.NewReader([]byte(`{"resourceType":"OperationOutcome","issue":[{"severity":"error","code":"processing","diagnostics":"Server error"}]}`))),
		}

		stub := &requestsResponder{
			responses: []*http.Response{errorResponse},
		}

		client := New(baseURL, stub, nil)
		ctx := context.Background()

		consumeFunc := func(bundle *fhir.Bundle) (bool, error) {
			return true, nil
		}

		err := Paginate(ctx, client, bundle, consumeFunc)

		require.Error(t, err)
		assert.Contains(t, err.Error(), "pagintate: query next page failed")
	})

	t.Run("invalid next URL", func(t *testing.T) {
		// Create bundle with invalid URL
		bundle := fhir.Bundle{
			Link: []fhir.BundleLink{
				{
					Relation: "next",
					Url:      "://invalid-url",
				},
			},
		}

		stub := &requestsResponder{
			responses: []*http.Response{},
		}

		client := New(baseURL, stub, nil)
		ctx := context.Background()

		consumeFunc := func(bundle *fhir.Bundle) (bool, error) {
			return true, nil
		}

		err := Paginate(ctx, client, bundle, consumeFunc)

		require.Error(t, err)
		assert.Contains(t, err.Error(), "paginate: invalid 'next' link for search set")
	})

	t.Run("max iterations reached", func(t *testing.T) {
		// Create bundle that always has a next link
		bundle := createBundleWithNextLink("http://example.com/fhir/page2")
		nextPageBundle := createBundleWithNextLink("http://example.com/fhir/page2") // Always same next link

		// Create multiple responses that all point to the same next page
		responses := []*http.Response{}
		for i := 0; i < 5; i++ {
			responses = append(responses, createBundleResponse(nextPageBundle))
		}

		stub := &requestsResponder{
			responses: responses,
		}

		client := New(baseURL, stub, nil)
		ctx := context.Background()

		consumeFunc := func(bundle *fhir.Bundle) (bool, error) {
			return true, nil
		}

		err := Paginate(ctx, client, bundle, consumeFunc, WithMaxIterations(3))

		require.Error(t, err)
		assert.Contains(t, err.Error(), "paginate: max. search iterations reached (3), possible bug")
	})

	t.Run("single page without next link", func(t *testing.T) {
		bundle := createBundleWithoutNextLink()

		stub := &requestsResponder{
			responses: []*http.Response{},
		}

		client := New(baseURL, stub, nil)
		ctx := context.Background()

		callCount := 0
		consumeFunc := func(bundle *fhir.Bundle) (bool, error) {
			callCount++
			return true, nil
		}

		err := Paginate(ctx, client, bundle, consumeFunc)

		require.NoError(t, err)
		assert.Equal(t, 1, callCount)
		assert.Len(t, stub.requests, 0) // No additional requests should be made
	})

	t.Run("custom max iterations option", func(t *testing.T) {
		bundle := createBundleWithNextLink("http://example.com/fhir/page2")
		nextPageBundle := createBundleWithNextLink("http://example.com/fhir/page2")

		// Create enough responses to exceed the custom max iterations
		responses := []*http.Response{}
		for i := 0; i < 10; i++ {
			responses = append(responses, createBundleResponse(nextPageBundle))
		}

		stub := &requestsResponder{
			responses: responses,
		}

		client := New(baseURL, stub, nil)
		ctx := context.Background()

		consumeFunc := func(bundle *fhir.Bundle) (bool, error) {
			return true, nil
		}

		err := Paginate(ctx, client, bundle, consumeFunc, WithMaxIterations(5))

		require.Error(t, err)
		assert.Contains(t, err.Error(), "paginate: max. search iterations reached (5), possible bug")
	})

	t.Run("bundle with multiple links, only next is used", func(t *testing.T) {
		// Create bundle with multiple link types
		bundle := fhir.Bundle{
			Link: []fhir.BundleLink{
				{
					Relation: "self",
					Url:      "http://example.com/fhir/current",
				},
				{
					Relation: "next",
					Url:      "http://example.com/fhir/page2",
				},
				{
					Relation: "prev",
					Url:      "http://example.com/fhir/prev",
				},
			},
		}

		secondBundle := createBundleWithoutNextLink()
		page2Response := createBundleResponse(secondBundle)

		stub := &requestsResponder{
			responses: []*http.Response{page2Response},
		}

		client := New(baseURL, stub, nil)
		ctx := context.Background()

		callCount := 0
		consumeFunc := func(bundle *fhir.Bundle) (bool, error) {
			callCount++
			return true, nil
		}

		err := Paginate(ctx, client, bundle, consumeFunc)

		require.NoError(t, err)
		assert.Equal(t, 2, callCount)
		require.Len(t, stub.requests, 1)
		assert.Equal(t, "http://example.com/fhir/page2", stub.requests[0].URL.String())
	})
}

func createBundleWithNextLink(nextURL string) fhir.Bundle {
	return fhir.Bundle{
		Link: []fhir.BundleLink{
			{
				Relation: "next",
				Url:      nextURL,
			},
		},
	}
}

func createBundleWithoutNextLink() fhir.Bundle {
	return fhir.Bundle{
		Link: []fhir.BundleLink{
			{
				Relation: "self",
				Url:      "http://example.com/fhir/current",
			},
		},
	}
}

func createBundleResponse(bundle fhir.Bundle) *http.Response {
	data, _ := json.Marshal(bundle)
	return &http.Response{
		StatusCode: http.StatusOK,
		Header: map[string][]string{
			"Content-Type": {FhirJsonMediaType},
		},
		Body: io.NopCloser(bytes.NewReader(data)),
	}
}

// requestsResponder handles multiple sequential requests
type requestsResponder struct {
	requests  []*http.Request
	responses []*http.Response
}

func (s *requestsResponder) Do(req *http.Request) (*http.Response, error) {
	s.requests = append(s.requests, req)
	return s.responses[len(s.requests)-1], nil
}
