package repository

import (
	"context"
	"github.com/stretchr/testify/assert"
	"sync"
	"testing"
)

func TestStorage_CreateShortURL(t *testing.T) {
	type fields struct {
		mx     *sync.RWMutex
		urls   map[int64]*StorageURL
		lastID int64
	}
	type args struct {
		beginURL string
		url      string
	}
	tests := []struct {
		name   string
		fields fields
		args   args
		want   string
	}{
		{
			name: "check success create",
			fields: fields{
				mx:     &sync.RWMutex{},
				urls:   map[int64]*StorageURL{},
				lastID: 0,
			},
			args: args{
				beginURL: "http://begin:8080/",
				url:      "http://begin:8080/some/path",
			},
			want: "http://begin:8080/0",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := &InMemoryStorage{
				userURLs: tt.fields.urls,
				lastID:   tt.fields.lastID,
			}
			got, _ := s.CreateShortURL(context.Background(), tt.args.beginURL, tt.args.url, 0)
			if got != tt.want {
				t.Errorf("CreateShortURL() got = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestStorage_GetFullURL(t *testing.T) {
	type fields struct {
		mx     *sync.RWMutex
		urls   map[int64]*StorageURL
		lastID int64
	}
	type args struct {
		shortURL int64
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		want    string
		wantErr bool
	}{
		{
			name: "check success",
			fields: fields{
				mx: &sync.RWMutex{},
				urls: map[int64]*StorageURL{
					0: {
						url:    "fullURL",
						userID: 0,
					},
				},
				lastID: 1,
			},
			args: args{
				shortURL: int64(0),
			},
			want:    "fullURL",
			wantErr: false,
		},
		{
			name: "check fail",
			fields: fields{
				mx:     &sync.RWMutex{},
				urls:   map[int64]*StorageURL{},
				lastID: 1,
			},
			args: args{
				shortURL: int64(1),
			},
			want:    "",
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := &InMemoryStorage{
				userURLs: tt.fields.urls,
				lastID:   tt.fields.lastID,
			}
			got, err := s.GetFullURL(context.Background(), tt.args.shortURL)
			if (err != nil) != tt.wantErr {
				t.Errorf("GetFullURL() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("GetFullURL() got = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestStorage_CreateShortURLDoubleCheck(t *testing.T) {
	s := &InMemoryStorage{
		userURLs: map[int64]*StorageURL{},
		lastID:   0,
	}
	beginURL := "http://begin:8080/"
	fullURL := beginURL + "some/path"
	fullURL2 := beginURL + "some/path/path"
	got, _ := s.CreateShortURL(context.Background(), beginURL, fullURL, 0)
	assert.Equal(t, beginURL+"0", got)
	got, _ = s.CreateShortURL(context.Background(), beginURL, fullURL2, 0)
	assert.Equal(t, beginURL+"1", got)

}
