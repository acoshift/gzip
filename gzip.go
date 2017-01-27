package gzip

import (
	"compress/gzip"
	"io/ioutil"
	"net/http"
	"strings"
	"sync"
)

// Config is the gzip middleware config
type Config struct {
	Level int
}

// New creates new gzip middleware
func New(config Config) func(http.Handler) http.Handler {
	pool := sync.Pool{
		New: func() interface{} {
			gz, err := gzip.NewWriterLevel(ioutil.Discard, config.Level)
			if err != nil {
				panic(err)
			}
			return gz
		},
	}

	return func(h http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if !strings.Contains(r.Header.Get(headerAcceptEncoding), encodingGzip) {
				h.ServeHTTP(w, r)
				return
			}

			if len(r.Header.Get(headerSecWebSocketKey)) > 0 {
				h.ServeHTTP(w, r)
				return
			}

			header := w.Header()

			if header.Get(headerContentEncoding) == encodingGzip {
				h.ServeHTTP(w, r)
				return
			}

			g := pool.Get().(*gzip.Writer)
			defer g.Close()
			defer pool.Put(g)
			g.Reset(w)

			header.Set(headerVary, headerAcceptEncoding)
			header.Set(headerContentEncoding, encodingGzip)

			h.ServeHTTP(&responseWriter{w, g}, r)
			header.Del(headerContentLength)
		})
	}
}