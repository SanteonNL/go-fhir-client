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
	"fmt"
	fhirclient "github.com/SanteonNL/go-fhir-client"
	"net/http"
)

// ExampleResolveRef demonstrates how to resolve a reference in a FHIR resource.
// To resolve a list of references, pass a pointer to a slice (e.g. &[]Resource{} instead of &Resource{}).
func ExampleResolveRef() {
	// Stub the HTTP client
	httpClient := &requestsResponder{
		responses: []*http.Response{
			okResponse(RefResource{
				Id: "123",
				OneToOne: map[string]interface{}{
					"reference": "Resource/456",
					"type":      "Resource",
				},
			}),
			okResponse(Resource{Id: "456"}),
		},
	}

	// Create the FHIR client
	client := fhirclient.New(baseURL, httpClient)

	// Read a Resource/123 (e.g. Patient)
	var reffingResource RefResource
	var referencedResource Resource
	_ = client.Read("Resource/123", &reffingResource, fhirclient.ResolveRef("oneToOne", &referencedResource))
	fmt.Println(referencedResource.Id)

	// Output: 456
}
