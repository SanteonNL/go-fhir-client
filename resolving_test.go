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
	fhirclient "github.com/SanteonNL/go-fhir-client"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/mock/gomock"
	"testing"
)

type RefResource struct {
	Id        string                   `json:"id"`
	OneToOne  map[string]interface{}   `json:"oneToOne"`
	OneToMany []map[string]interface{} `json:"oneToMany"`
}

func TestResolveRef(t *testing.T) {
	t.Run("resolve single reference", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		client := NewMockClient(ctrl)
		client.EXPECT().Read("Resource/123", gomock.Any()).DoAndReturn(func(_ string, r *Resource, _ ...fhirclient.Option) error {
			*r = Resource{
				Id: "123",
			}
			return nil
		})

		var resolved Resource
		err := fhirclient.ResolveRef("oneToOne", &resolved)(client, RefResource{
			OneToOne: map[string]interface{}{
				"reference": "Resource/123",
				"type":      "Resource",
			},
		})

		require.NoError(t, err)
		require.Equal(t, "123", resolved.Id)
	})
	t.Run("resolve multiple references", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		client := NewMockClient(ctrl)

		client.EXPECT().Read("Resource/123", gomock.Any()).DoAndReturn(func(_ string, r *Resource, _ ...fhirclient.Option) error {
			*r = Resource{
				Id: "123",
			}
			return nil
		})
		client.EXPECT().Read("Resource/789", gomock.Any()).DoAndReturn(func(_ string, r *Resource, _ ...fhirclient.Option) error {
			*r = Resource{
				Id: "789",
			}
			return nil
		})

		var resolved []Resource
		err := fhirclient.ResolveRef("oneToMany", &resolved)(client, RefResource{
			OneToMany: []map[string]interface{}{
				{
					"reference": "Resource/123",
					"type":      "Resource",
				},
				{
					"reference": "Resource/789",
					"type":      "Resource",
				},
			},
		})

		require.NoError(t, err)
		assert.Len(t, resolved, 2)
		require.Equal(t, "123", resolved[0].Id)
		require.Equal(t, "789", resolved[1].Id)
	})
}
