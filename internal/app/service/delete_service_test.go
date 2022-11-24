package service

import (
	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"
	"go-axesthump-shortener/internal/app/mocks"
	"go-axesthump-shortener/internal/app/repository"
	"testing"
)

func TestDeleteService_getURLsFromArr(t *testing.T) {
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

func TestDeleteService_Close(t *testing.T) {
	ctrl := gomock.NewController(t)
	repo := mocks.NewMockRepository(ctrl)
	defer ctrl.Finish()

	ds := NewDeleteService(repo, "http://localhost:8080")

	ds.Close()
	_, ok := <-ds.urlsForDelete
	assert.False(t, ok)
}
