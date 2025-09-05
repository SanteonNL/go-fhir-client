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
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestAddHeaderIfNotPresent(t *testing.T) {
	t.Run("adds value if not present", func(t *testing.T) {
		header := http.Header{}
		addHeaderValueIfNotPresent(&header, "X-Custom", "value")
		assert.Equal(t, "value", header.Get("X-Custom"))
	})

	t.Run("does not add duplicate value", func(t *testing.T) {
		header := http.Header{}
		header.Add("X-Custom", "value")
		addHeaderValueIfNotPresent(&header, "X-Custom", "value")
		assert.Equal(t, []string{"value"}, header["X-Custom"])
	})

	t.Run("adds multiple different values", func(t *testing.T) {
		header := http.Header{}
		addHeaderValueIfNotPresent(&header, "X-Custom", "value1")
		addHeaderValueIfNotPresent(&header, "X-Custom", "value2")
		assert.Equal(t, []string{"value1", "value2"}, header["X-Custom"])
	})
}

func TestSetHeaderIfNotPresent(t *testing.T) {
	t.Run("sets value if not present", func(t *testing.T) {
		header := http.Header{}
		setHeaderValueIfNotPresent(&header, "X-Custom", "value")
		assert.Equal(t, "value", header.Get("X-Custom"))
	})

	t.Run("does not overwrite existing value", func(t *testing.T) {
		header := http.Header{}
		header.Set("X-Custom", "existing")
		setHeaderValueIfNotPresent(&header, "X-Custom", "new")
		assert.Equal(t, "existing", header.Get("X-Custom"))
	})
}
