package middleware

import (
	"net/http"
	"time"

	"github.com/go-chi/chi/v5/middleware"
	"github.com/rs/zerolog/log"
)

// StructuredLogger is a Chi middleware for logging requests using zerolog.
func StructuredLogger(next http.Handler) http.Handler {
	fn := func(w http.ResponseWriter, r *http.Request) {
		ww := middleware.NewWrapResponseWriter(w, r.ProtoMajor)
		t1 := time.Now()
		defer func() {
			log.Info().
				Str("proto", r.Proto).
				Str("method", r.Method).
				Str("path", r.URL.Path).
				Str("remote_addr", r.RemoteAddr).
				Str("user_agent", r.UserAgent()).
				Int("status", ww.Status()).
				Int("bytes_written", ww.BytesWritten()).
				Dur("latency_ms", time.Since(t1)).
				Msg("request handled")
		}()
		next.ServeHTTP(ww, r)
	}
	return http.HandlerFunc(fn)
}
