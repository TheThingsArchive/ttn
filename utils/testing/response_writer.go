package testing

import "net/http"

// ResponseWriter mocks http.ResponseWriter
type responseWriter struct {
	TheHeaders *http.Header
	TheStatus  int
	TheBody    []byte
}

// Header implements http.ResponseWriter
func (rw *responseWriter) Header() http.Header {
	return *rw.TheHeaders
}

// Write implements http.ResponseWriter
func (rw *responseWriter) Write(m []byte) (int, error) {
	rw.TheBody = m
	return len(m), nil
}

// WriteHeader implements http.ResponseWriter
func (rw *responseWriter) WriteHeader(h int) {
	rw.TheStatus = h
}

func NewResponseWriter() responseWriter {
	h := http.Header{}
	return responseWriter{
		TheHeaders: &h,
	}
}
