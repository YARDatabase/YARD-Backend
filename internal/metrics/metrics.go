package metrics

import (
	"net/http"
	"strconv"
	"strings"

	"yard-backend/internal/config"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

var (
	enabled bool

	httpRequestsTotal = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "yard_http_requests_total",
			Help: "Total number of HTTP requests",
		},
		[]string{"method", "endpoint", "status_code"},
	)

	httpRequestErrors = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "yard_http_request_errors_total",
			Help: "Total number of HTTP request errors",
		},
		[]string{"method", "endpoint", "status_code"},
	)

	httpRequestsByCountry = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "yard_http_requests_by_country_total",
			Help: "Total number of HTTP requests by country",
		},
		[]string{"country", "endpoint"},
	)
)

func Init(enable bool) {
	enabled = enable
}

func IsEnabled() bool {
	return enabled
}

func RecordRequest(method, endpoint string, statusCode int) {
	if !enabled {
		return
	}
	httpRequestsTotal.WithLabelValues(method, endpoint, http.StatusText(statusCode)).Inc()
}

func RecordError(method, endpoint string, statusCode int) {
	if !enabled {
		return
	}
	if statusCode >= 400 {
		httpRequestErrors.WithLabelValues(method, endpoint, http.StatusText(statusCode)).Inc()
	}
}

func RecordRequestByCountry(country, endpoint string) {
	if !enabled {
		return
	}
	if country == "" {
		country = "unknown"
	}
	httpRequestsByCountry.WithLabelValues(country, endpoint).Inc()
}

// returns the prometheus metrics handler with optional ip whitelisting
func GetHandler() http.Handler {
	handler := promhttp.Handler()

	if config.MetricsIPWhitelist == "" {
		return handler
	}

	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		clientIP := getClientIPFromRequest(r)

		if !isIPAllowed(clientIP, config.MetricsIPWhitelist) {
			http.Error(w, "Forbidden", http.StatusForbidden)
			return
		}

		handler.ServeHTTP(w, r)
	})
}

func getClientIPFromRequest(r *http.Request) string {
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

func isIPAllowed(ip, whitelist string) bool {
	if whitelist == "" {
		return true
	}

	allowedIPs := strings.Split(whitelist, ",")
	for _, allowed := range allowedIPs {
		allowed = strings.TrimSpace(allowed)
		if allowed == "*" {
			return true
		}
		if ip == allowed {
			return true
		}

		// handle CIDR notation
		if strings.Contains(allowed, "/") {
			parts := strings.Split(allowed, "/")
			if len(parts) != 2 {
				continue
			}

			networkIP := parts[0]
			maskStr := parts[1]

			maskBits, err := strconv.Atoi(maskStr)
			if err != nil || maskBits < 0 || maskBits > 32 {
				continue
			}

			// parse IPs
			networkParts := strings.Split(networkIP, ".")
			ipParts := strings.Split(ip, ".")

			if len(networkParts) != 4 || len(ipParts) != 4 {
				continue
			}

			// calculate how many full octets to check
			fullOctets := maskBits / 8
			remainingBits := maskBits % 8

			// check full octets
			match := true
			for i := 0; i < fullOctets && i < 4; i++ {
				netOctet, err1 := strconv.Atoi(networkParts[i])
				ipOctet, err2 := strconv.Atoi(ipParts[i])
				if err1 != nil || err2 != nil || netOctet != ipOctet {
					match = false
					break
				}
			}

			if !match {
				continue
			}

			// check partial octet if needed
			if remainingBits > 0 && fullOctets < 4 {
				netOctet, err1 := strconv.Atoi(networkParts[fullOctets])
				ipOctet, err2 := strconv.Atoi(ipParts[fullOctets])
				if err1 != nil || err2 != nil {
					continue
				}

				mask := (0xFF << (8 - remainingBits)) & 0xFF
				if (netOctet & mask) != (ipOctet & mask) {
					continue
				}
			}

			return true
		}
	}

	return false
}

// cleans up endpoint paths so metrics are consistent
func NormalizeEndpoint(path string) string {
	path = strings.TrimPrefix(path, "/")
	if path == "" {
		return "root"
	}

	parts := strings.Split(path, "/")
	if len(parts) > 0 && strings.HasPrefix(parts[0], "api") {
		if len(parts) > 1 {
			return "/" + parts[0] + "/" + parts[1]
		}
		return "/" + parts[0]
	}

	return "/" + parts[0]
}
