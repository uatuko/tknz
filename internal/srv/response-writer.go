package srv

import (
	"net/http"
)

type responseWriter struct {
	statusCode int
	size       int

	http.Flusher
	http.ResponseWriter
}

func (w *responseWriter) Size() uint64 {
	if w.size < 0 {
		return 0
	}

	return uint64(w.size)
}

func (w *responseWriter) Status() uint32 {
	return uint32(w.statusCode)
}

func (w *responseWriter) Write(b []byte) (int, error) {
	size, err := w.ResponseWriter.Write(b)
	w.size += size

	return size, err
}

func (w *responseWriter) WriteHeader(statusCode int) {
	w.statusCode = statusCode
	w.ResponseWriter.WriteHeader(statusCode)
}

func newResponseWriter(w http.ResponseWriter) *responseWriter {
	return &responseWriter{
		Flusher:        w.(http.Flusher),
		ResponseWriter: w,
		statusCode:     http.StatusOK,
	}
}
