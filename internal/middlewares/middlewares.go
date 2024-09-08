package middlewares

import (
	"io"
	"log"
	"net/http"
	"time"
)

type responseWriter struct {
	http.ResponseWriter
	statusCode int
}

func (rw *responseWriter) WriteHeader(code int) {
	rw.statusCode = code
	rw.ResponseWriter.WriteHeader(code)
}

func Logging(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		log.Printf("[] Received request: Method=%s, URL=%s, RemoteAddr=%s", r.Method, r.URL.Path, r.RemoteAddr)
		start := time.Now()
		wrappedWriter := &responseWriter{w, http.StatusOK}
		body, err := io.ReadAll(r.Body)
		if err != nil {
			log.Printf("Error reading request body: %v", err)
		}
		r.Body = io.NopCloser(io.NopCloser(r.Body).(io.ReadCloser))
		next.ServeHTTP(w, r)
		log.Printf(
			"Method: %s, URL: %s, ClientIP: %s, User-Agent: %s, Status: %d, Latency: %v, Request Body: %s",
			r.Method,
			r.URL.String(),
			r.RemoteAddr,
			r.UserAgent(),
			wrappedWriter.statusCode,
			time.Since(start),
			string(body),
		)
		log.Printf("Request processed: Method=%s, URL=%s, Duration=%s", r.Method, r.URL.Path, time.Since(start))
	})
}
