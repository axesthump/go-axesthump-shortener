package service

import (
	"github.com/stretchr/testify/assert"
	"go-axesthump-shortener/internal/app/repository"
	"testing"
)

func Test_getURLsFromArr(t *testing.T) {
	data := `["http://localhost:8080/1", "http://localhost:8080/2"]`
	expected := []repository.DeleteURL{
		{
			URL:    "1",
			UserID: 0,
		},
		{
			URL:    "2",
			UserID: 0,
		},
	}

	actual := getURLsFromArr(data, 0, "http://localhost:8080")
	assert.Equalf(t, expected, actual, "")
}
