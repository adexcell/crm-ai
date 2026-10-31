package middleware

import (
	"bufio"
	"errors"
	"net"
	"net/http"
)

type Writer struct {
	http.ResponseWriter
	isSetCode bool
	code      int
}

func WriterWrapper(w http.ResponseWriter) *Writer {
	return &Writer{ResponseWriter: w}
}

func (w *Writer) WriteHeader(code int) {
	if !w.isSetCode {
		w.setCode(code)
		w.ResponseWriter.WriteHeader(code)
	}
}

func (w *Writer) Code() int {
	if !w.isSetCode {
		return http.StatusOK
	}
	return w.code
}

func (w *Writer) setCode(code int) {
	w.isSetCode = true
	w.code = code
}

func (w *Writer) Write(b []byte) (int, error) {
	if !w.isSetCode {
		w.WriteHeader(http.StatusOK)
	}
	return w.ResponseWriter.Write(b)
}

func (w *Writer) Hijack() (net.Conn, *bufio.ReadWriter, error) {
	if hijacker, ok := w.ResponseWriter.(http.Hijacker); ok {
		return hijacker.Hijack()
	}
	return nil, nil, errors.New("underlying ResponseWriter does not implement http.Hijacker")
}

func (w *Writer) Flush() {
	if flusher, ok := w.ResponseWriter.(http.Flusher); ok {
		flusher.Flush()
	}
}
