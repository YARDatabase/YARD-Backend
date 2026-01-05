package metrics

import (
	"net/http"
	"strings"
)

type responseWriter struct {
	http.ResponseWriter
	statusCode int
}

func (rw *responseWriter) WriteHeader(code int) {
	rw.statusCode = code
	rw.ResponseWriter.WriteHeader(code)
}

func MetricsMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if !IsEnabled() {
			next.ServeHTTP(w, r)
			return
		}

		rw := &responseWriter{
			ResponseWriter: w,
			statusCode:     http.StatusOK,
		}

		next.ServeHTTP(rw, r)

		method := r.Method
		endpoint := NormalizeEndpoint(r.URL.Path)
		statusCode := rw.statusCode

		RecordRequest(method, endpoint, statusCode)
		RecordError(method, endpoint, statusCode)

		country := getCountryFromRequest(r)
		RecordRequestByCountry(country, endpoint)
	})
}

// tries to figure out the country from request headers
// looks for cloudflare header first then falls back to other headers
func getCountryFromRequest(r *http.Request) string {
	country := r.Header.Get("CF-IPCountry")
	if country != "" && country != "XX" {
		return strings.ToUpper(country)
	}

	country = r.Header.Get("X-Country-Code")
	if country != "" {
		return strings.ToUpper(country)
	}

	ip := getClientIP(r)
	if ip != "" {
		return getCountryFromIP(ip)
	}

	return "unknown"
}

func getClientIP(r *http.Request) string {
	ip := r.Header.Get("CF-Connecting-IP")
	if ip != "" {
		return ip
	}

	ip = r.Header.Get("X-Real-IP")
	if ip != "" {
		return ip
	}

	ip = r.Header.Get("X-Forwarded-For")
	if ip != "" {
		ips := strings.Split(ip, ",")
		if len(ips) > 0 {
			return strings.TrimSpace(ips[0])
		}
	}

	return strings.Split(r.RemoteAddr, ":")[0]
}

func getCountryFromIP(ip string) string {
	if ip == "" || ip == "::1" || ip == "127.0.0.1" {
		return "localhost"
	}

	return "unknown"
}
