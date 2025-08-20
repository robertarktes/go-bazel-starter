package httpx_test

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/robertarktes/go-bazel-starter/pkg/httpx"

	"github.com/stretchr/testify/assert"
)

func TestClient_Get(t *testing.T) {
	tests := []struct {
		name       string
		serverFunc func(w http.ResponseWriter, r *http.Request)
		opts       []httpx.Option
		wantErr    bool
	}{
		{
			name: "successful request",
			serverFunc: func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(http.StatusOK)
			},
			opts:    []httpx.Option{httpx.WithTimeout(1 * time.Second)},
			wantErr: false,
		},
		// Add more table-driven tests as needed
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			srv := httptest.NewServer(http.HandlerFunc(tt.serverFunc))
			defer srv.Close()

			c := httpx.NewClient(tt.opts...)
			_, err := c.Get(context.Background(), srv.URL)
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func BenchmarkClient_Get(b *testing.B) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))
	defer srv.Close()

	c := httpx.NewClient()
	for i := 0; i < b.N; i++ {
		_, err := c.Get(context.Background(), srv.URL)
		if err != nil {
			b.Fatal(err)
		}
	}
}
