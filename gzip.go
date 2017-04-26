package gzip

import (
	"compress/gzip"
	"io/ioutil"
	"net/http"
	"strings"
	"sync"
)

// Copy from compress/gzip
const (
	NoCompression      = gzip.NoCompression
	BestSpeed          = gzip.BestSpeed
	BestCompression    = gzip.BestCompression
	DefaultCompression = gzip.DefaultCompression
	HuffmanOnly        = gzip.HuffmanOnly
)

// Config is the gzip middleware config
type Config struct {
	Level int
}

// DefaultConfig use default compression level
var DefaultConfig = Config{
	Level: DefaultCompression,
}

// New creates new gzip middleware
func New(config Config) func(http.Handler) http.Handler {
	pool := &sync.Pool{
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

			header.Set(headerVary, headerAcceptEncoding)

			gw := &responseWriter{
				ResponseWriter: w,
				pool:           pool,
			}
			defer gw.Close()

			h.ServeHTTP(gw, r)
		})
	}
}
