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
)

func ExampleBaseClient_Read() {
	// Stub the HTTP client
	httpClient := &requestResponder{
		response: okResponse(Resource{Id: "123"}),
	}

	// Create the FHIR client
	client := fhirclient.New(baseURL, httpClient, nil)

	// Read a Resource/123 (e.g. Patient)
	var result Resource
	_ = client.Read("Resource/123", &result)
	fmt.Println(result.Id)

	// Output: 123
}

func ExampleBaseClient_Create() {
	// Stub the HTTP client
	httpClient := &requestResponder{
		response: okResponse(Resource{Id: "123"}),
	}

	// Create the FHIR client
	client := fhirclient.New(baseURL, httpClient, nil)

	// Create a new Resource
	var createdResource Resource
	_ = client.Create(Resource{}, &createdResource)
	fmt.Println(createdResource.Id)

	// Output: 123
}
