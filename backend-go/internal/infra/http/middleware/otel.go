package middleware

import (
	"net/http"

	"github.com/riandyrn/otelchi"
)

func OTel(serviceName string) func(http.Handler) http.Handler {
	return otelchi.Middleware(serviceName)
}
