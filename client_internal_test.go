package fhirclient

import (
	"github.com/stretchr/testify/assert"
	"net/http"
	"testing"
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
