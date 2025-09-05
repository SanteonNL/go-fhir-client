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
	"context"
	"fmt"
	"net/http"
	"net/url"

	fhirclient "github.com/SanteonNL/go-fhir-client"
	"github.com/zorgbijjou/golang-fhir-models/fhir-models/fhir"
)

// ExamplePaginate demonstrates how to use the Paginate function to iterate through
// all pages of a FHIR search result. This is useful when a search returns many results
// that are split across multiple pages.
func ExamplePaginate() {
	// Stub the HTTP client
	page1ID := "page1"
	page2ID := "page2"
	httpClient := &requestsResponder{
		responses: []*http.Response{
			okResponse(fhir.Bundle{
				ID: &page1ID,
				Link: []fhir.BundleLink{
					{
						Relation: "next",
						Url:      "http://example.com/fhir/page2",
					},
				},
			}),
			okResponse(fhir.Bundle{
				ID: &page2ID,
			}),
		},
	}

	// Create the FHIR client
	client := fhirclient.New(baseURL, httpClient, nil)

	// Perform initial search
	var searchSet fhir.Bundle
	_ = client.Search("SomeResourceType", url.Values{}, &searchSet)

	// Paginate to process the result and all subsequent pages
	_ = fhirclient.Paginate(context.Background(), client, searchSet, func(searchSet *fhir.Bundle) (bool, error) {
		fmt.Println(*searchSet.ID)
		return true, nil
	})

	// Output:
	// page1
	// page2
}
