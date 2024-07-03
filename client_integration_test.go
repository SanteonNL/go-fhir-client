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
	"encoding/json"
	fhirclient "github.com/SanteonNL/go-fhir-client"
	"github.com/stretchr/testify/require"
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"
)

func Test_Integration_DefaultClient_Read(t *testing.T) {
	mux := http.NewServeMux()
	mux.HandleFunc("GET /Resource/1", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		data, _ := json.Marshal(Resource{Id: "123"})
		_, _ = w.Write(data)
	})
	httpServer := httptest.NewServer(mux)
	baseURL, _ := url.Parse(httpServer.URL)
	client := fhirclient.New(baseURL, httpServer.Client(), nil)

	t.Run("at path", func(t *testing.T) {
		var result Resource

		err := client.Read("Foo", &result, fhirclient.AtPath("Resource/1"))

		require.NoError(t, err)
		require.NotNil(t, result)
	})
}
