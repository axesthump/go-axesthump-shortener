package repository

import (
	"context"
	"github.com/stretchr/testify/assert"
	"go-axesthump-shortener/internal/app/generator"
	"os"
	"strconv"
	"testing"
)

func Test_createRow(t *testing.T) {
	expected := "1~s~e~c~short~s~e~c~fullURL~s~e~c~false"
	actual := createRow(1, "short", "fullURL", "false")
	assert.Equal(t, expected, actual)
}

func Test_getLastID(t *testing.T) {
	tests := []struct {
		name     string
		fileData string
		expected int64
	}{
		{
			name:     "Test getLastID with empty file",
			fileData: "",
			expected: 1,
		},
		{
			name:     "Test getLastID with not empty file",
			fileData: "1~s~e~c~1~s~e~c~fullURL~s~e~c~false\n1~s~e~c~2~s~e~c~fullURL~s~e~c~false",
			expected: 3,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := os.WriteFile("test", []byte(tt.fileData), 0665)
			assert.NoError(t, err)
			f, err := os.Open("test")
			assert.NoError(t, err)
			id := getLastID(f)
			assert.Equal(t, tt.expected, id)
			err = os.Remove("test")
			assert.NoError(t, err)
			err = f.Close()
			assert.NoError(t, err)
		})
	}
}

func TestLocalStorage_GetUserLastID(t *testing.T) {
	tests := []struct {
		name     string
		fileData string
		expected uint32
	}{
		{
			name:     "Test GetUserLastID with empty file",
			fileData: "",
			expected: 1,
		},
		{
			name:     "Test GetUserLastID with not empty file",
			fileData: "1~s~e~c~1~s~e~c~fullURL~s~e~c~false\n2~s~e~c~2~s~e~c~fullURL~s~e~c~false",
			expected: 3,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := os.WriteFile("test", []byte(tt.fileData), 0665)
			assert.NoError(t, err)
			f, err := os.Open("test")
			assert.NoError(t, err)
			ls := LocalStorage{
				file:        f,
				idGenerator: generator.NewIDGenerator(0),
			}
			id := ls.GetUserLastID()
			assert.Equal(t, tt.expected, id)
			err = os.Remove("test")
			assert.NoError(t, err)
			err = ls.Close()
			assert.NoError(t, err)
		})
	}
}

func TestLocalStorage_CreateShortURL(t *testing.T) {
	ls, err := NewLocalStorage("test")
	assert.NoError(t, err)

	shortURL, err := ls.CreateShortURL(
		context.TODO(),
		"http://localhost:8080/",
		"http://google.com/some/url",
		12,
	)

	assert.Equal(t, "http://localhost:8080/1", shortURL)

	err = os.Remove("test")
	assert.NoError(t, err)
	err = ls.Close()
	assert.NoError(t, err)
}

func TestLocalStorage_CreateShortURLs(t *testing.T) {
	ls, err := NewLocalStorage("test")
	assert.NoError(t, err)
	urls := []URLWithID{
		{
			CorrelationID: "1",
			URL:           "http://google.com/some/url",
		},
		{
			CorrelationID: "2",
			URL:           "http://google.com/some/url/another",
		},
		{
			CorrelationID: "3",
			URL:           "http://google.com/some/url/maybe/sin",
		},
	}
	expected := []URLWithID{
		{
			CorrelationID: "1",
			URL:           "http://localhost:8080/1",
		},
		{
			CorrelationID: "2",
			URL:           "http://localhost:8080/2",
		},
		{
			CorrelationID: "3",
			URL:           "http://localhost:8080/3",
		},
	}

	shortURLs, err := ls.CreateShortURLs(
		context.TODO(),
		"http://localhost:8080/",
		urls,
		12,
	)
	assert.NoError(t, err)

	assert.Equalf(t, expected, shortURLs, "")

	err = os.Remove("test")
	assert.NoError(t, err)
	err = ls.Close()
	assert.NoError(t, err)
}

func TestLocalStorage_GetFullURL(t *testing.T) {
	ls, err := NewLocalStorage("test")
	assert.NoError(t, err)
	urls := []URLWithID{
		{
			CorrelationID: "1",
			URL:           "http://google.com/some/url",
		},
		{
			CorrelationID: "2",
			URL:           "http://google.com/some/url/another",
		},
		{
			CorrelationID: "3",
			URL:           "http://google.com/some/url/maybe/sin",
		},
	}
	shortURLs, err := ls.CreateShortURLs(
		context.TODO(),
		"http://localhost:8080/",
		urls,
		12,
	)
	assert.NoError(t, err)

	for i, shortURL := range shortURLs {
		shortURLInt, err := strconv.ParseInt(shortURL.URL[len(shortURL.URL)-1:], 10, 64)
		assert.NoError(t, err)
		fullURL, err := ls.GetFullURL(context.TODO(), shortURLInt)
		assert.NoError(t, err)
		assert.Equal(t, urls[i].URL, fullURL)
	}

	err = os.Remove("test")
	assert.NoError(t, err)
	err = ls.Close()
	assert.NoError(t, err)
}
