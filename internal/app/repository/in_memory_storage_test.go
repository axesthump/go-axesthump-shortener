package repository

import (
	"context"
	"github.com/stretchr/testify/assert"
	"go-axesthump-shortener/internal/app/generator"
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
				userURLs:    tt.fields.urls,
				idGenerator: generator.NewIDGenerator(tt.fields.lastID),
			}
			defer s.Close()
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
		{
			name: "check deleted",
			fields: fields{
				mx: &sync.RWMutex{},
				urls: map[int64]*StorageURL{
					0: {
						url:       "fullURL",
						userID:    0,
						isDeleted: true,
					},
				},
				lastID: 1,
			},
			args: args{
				shortURL: int64(0),
			},
			want:    "",
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := &InMemoryStorage{
				userURLs:    tt.fields.urls,
				idGenerator: generator.NewIDGenerator(tt.fields.lastID),
			}
			defer s.Close()
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
		userURLs:    map[int64]*StorageURL{},
		idGenerator: generator.NewIDGenerator(0),
	}
	defer s.Close()
	beginURL := "http://begin:8080/"
	fullURL := beginURL + "some/path"
	fullURL2 := beginURL + "some/path/path"
	got, _ := s.CreateShortURL(context.Background(), beginURL, fullURL, 0)
	assert.Equal(t, beginURL+"0", got)
	got, _ = s.CreateShortURL(context.Background(), beginURL, fullURL2, 0)
	assert.Equal(t, beginURL+"1", got)

}

func TestStorage_NewInMemoryStorage(t *testing.T) {
	s := NewInMemoryStorage()
	assert.Equal(t, 0, len(s.userURLs))
	assert.NotNil(t, s.idGenerator)
}

func TestInMemoryStorage_DeleteURLs(t *testing.T) {
	type fields struct {
		userURLs    map[int64]*StorageURL
		idGenerator *generator.IDGenerator
	}
	type args struct {
		urlsForDelete []DeleteURL
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		want    bool
		wantErr bool
	}{
		{
			name: "Test success delete",
			fields: fields{
				userURLs: map[int64]*StorageURL{
					0: {
						url:       "fullURL",
						userID:    0,
						isDeleted: false,
					},
					1: {
						url:       "fullURL2",
						userID:    0,
						isDeleted: false,
					},
					2: {
						url:       "fullURL3",
						userID:    0,
						isDeleted: false,
					},
				},
				idGenerator: generator.NewIDGenerator(3),
			},
			args: args{
				urlsForDelete: []DeleteURL{
					{URL: "0", UserID: 0},
					{URL: "1", UserID: 0},
					{URL: "2", UserID: 0},
				},
			},
			want:    true,
			wantErr: false,
		},
		{
			name: "Test delete from another user",
			fields: fields{
				userURLs: map[int64]*StorageURL{
					0: {
						url:       "fullURL",
						userID:    0,
						isDeleted: false,
					},
					1: {
						url:       "fullURL2",
						userID:    0,
						isDeleted: false,
					},
					2: {
						url:       "fullURL3",
						userID:    0,
						isDeleted: false,
					},
				},
				idGenerator: generator.NewIDGenerator(3),
			},
			args: args{
				urlsForDelete: []DeleteURL{
					{URL: "0", UserID: 1},
					{URL: "1", UserID: 1},
					{URL: "2", UserID: 1},
				},
			},
			want:    false,
			wantErr: false,
		},
		{
			name: "Test delete with bad short url",
			fields: fields{
				userURLs: map[int64]*StorageURL{
					0: {
						url:       "fullURL",
						userID:    0,
						isDeleted: false,
					},
					1: {
						url:       "fullURL2",
						userID:    0,
						isDeleted: false,
					},
					2: {
						url:       "fullURL3",
						userID:    0,
						isDeleted: false,
					},
				},
				idGenerator: generator.NewIDGenerator(3),
			},
			args: args{
				urlsForDelete: []DeleteURL{
					{URL: "asd", UserID: 0},
					{URL: "dsa", UserID: 0},
					{URL: "zxc", UserID: 0},
				},
			},
			want:    false,
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := &InMemoryStorage{
				userURLs:    tt.fields.userURLs,
				idGenerator: tt.fields.idGenerator,
			}
			err := s.DeleteURLs(tt.args.urlsForDelete)
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
			for _, url := range s.userURLs {
				assert.Equal(t, tt.want, url.isDeleted)
			}
		})
	}
}

func TestInMemoryStorage_GetAllURLs(t *testing.T) {
	type fields struct {
		userURLs    map[int64]*StorageURL
		idGenerator *generator.IDGenerator
	}
	type args struct {
		ctx      context.Context
		beginURL string
		userID   uint32
	}
	tests := []struct {
		name   string
		fields fields
		args   args
		want   []URLInfo
	}{
		{
			name: "Test success all URLs",
			fields: fields{
				userURLs: map[int64]*StorageURL{
					0: {
						url:       "fullURL",
						userID:    0,
						isDeleted: false,
					},
					1: {
						url:       "fullURL2",
						userID:    0,
						isDeleted: false,
					},
					2: {
						url:       "fullURL3",
						userID:    0,
						isDeleted: false,
					},
				},
				idGenerator: generator.NewIDGenerator(3),
			},
			args: args{
				ctx:      context.TODO(),
				beginURL: "http://localhost:8080/",
				userID:   0,
			},
			want: []URLInfo{
				{
					ShortURL:    "http://localhost:8080/0",
					OriginalURL: "fullURL",
				},
				{
					ShortURL:    "http://localhost:8080/1",
					OriginalURL: "fullURL2",
				},
				{
					ShortURL:    "http://localhost:8080/2",
					OriginalURL: "fullURL3",
				},
			},
		},
		{
			name: "Test all URLs with different userID",
			fields: fields{
				userURLs: map[int64]*StorageURL{
					0: {
						url:       "fullURL",
						userID:    0,
						isDeleted: false,
					},
					1: {
						url:       "fullURL2",
						userID:    0,
						isDeleted: false,
					},
					2: {
						url:       "fullURL3",
						userID:    1,
						isDeleted: false,
					},
				},
				idGenerator: generator.NewIDGenerator(3),
			},
			args: args{
				ctx:      context.TODO(),
				beginURL: "http://localhost:8080/",
				userID:   0,
			},
			want: []URLInfo{
				{
					ShortURL:    "http://localhost:8080/0",
					OriginalURL: "fullURL",
				},
				{
					ShortURL:    "http://localhost:8080/1",
					OriginalURL: "fullURL2",
				},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := &InMemoryStorage{
				userURLs:    tt.fields.userURLs,
				idGenerator: tt.fields.idGenerator,
			}
			actual := s.GetAllURLs(tt.args.ctx, tt.args.beginURL, tt.args.userID)
			for _, url := range actual {
				assert.True(t, contains(tt.want, url))
			}
		})
	}
}

func TestInMemoryStorage_CreateShortURLs(t *testing.T) {
	type fields struct {
		userURLs    map[int64]*StorageURL
		idGenerator *generator.IDGenerator
	}
	type args struct {
		ctx      context.Context
		beginURL string
		urls     []URLWithID
		userID   uint32
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		want    []URLWithID
		wantErr bool
	}{
		{
			name: "Test success add URLs",
			fields: fields{
				userURLs:    map[int64]*StorageURL{},
				idGenerator: generator.NewIDGenerator(0),
			},
			args: args{
				ctx:      context.TODO(),
				beginURL: "http://localhost:8080/",
				urls: []URLWithID{
					{
						CorrelationID: "0",
						URL:           "fullURL0",
					},
					{
						CorrelationID: "1",
						URL:           "fullURL1",
					},
					{
						CorrelationID: "2",
						URL:           "fullURL2",
					},
				},
				userID: 0,
			},
			want: []URLWithID{
				{
					CorrelationID: "0",
					URL:           "http://localhost:8080/0",
				},
				{
					CorrelationID: "1",
					URL:           "http://localhost:8080/1",
				},
				{
					CorrelationID: "2",
					URL:           "http://localhost:8080/2",
				},
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := &InMemoryStorage{
				userURLs:    tt.fields.userURLs,
				idGenerator: tt.fields.idGenerator,
			}
			got, err := s.CreateShortURLs(tt.args.ctx, tt.args.beginURL, tt.args.urls, tt.args.userID)
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
			for _, url := range got {
				assert.True(t, containsURLWithID(tt.want, url))
			}
		})
	}
}

func contains(urls []URLInfo, url URLInfo) bool {
	for _, tURL := range urls {
		if tURL.OriginalURL == url.OriginalURL && tURL.ShortURL == url.ShortURL {
			return true
		}
	}
	return false
}

func containsURLWithID(urls []URLWithID, url URLWithID) bool {
	for _, tURL := range urls {
		if tURL.CorrelationID == url.CorrelationID && tURL.URL == url.URL {
			return true
		}
	}
	return false
}
