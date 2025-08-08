package middleware

import (
	"net/http"
	"regexp"
	"strings"

	"github.com/labstack/echo/v4"
)

type CorsMiddleware struct {
	handler *echo.Echo
}

type DomainPattern struct {
	Pattern     string
	IsRegex     bool
	AllowHTTP   bool
	AllowHTTPS  bool
	IsLocalhost bool
}

func ModuleCorsMiddleware(handler *echo.Echo) *CorsMiddleware {
	return &CorsMiddleware{
		handler: handler,
	}
}

func (m *CorsMiddleware) Setup() {
	m.handler.Use(func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			req := c.Request()
			res := c.Response()

			allowedPatterns := []DomainPattern{
				{
					Pattern:     "localhost",
					IsRegex:     false,
					AllowHTTP:   true,
					AllowHTTPS:  true,
					IsLocalhost: true,
				},
				{
					Pattern:    "chrome-extension://amknoiejhlmhancpahfcfcfhllgkpbld",
					IsRegex:    false,
					AllowHTTP:  true,
					AllowHTTPS: true,
				},
				{
					Pattern:    "",
					IsRegex:    false,
					AllowHTTP:  true,
					AllowHTTPS: true,
				},
			}

			origin := req.Header.Get("Origin")
			if isOriginAllowed(origin, allowedPatterns) {
				res.Header().Add("Access-Control-Allow-Origin", origin)
				res.Header().Add("Access-Control-Allow-Methods", "*")
				res.Header().Add("Access-Control-Allow-Headers", "*")
				res.Header().Add("Content-Type", "application/json")

				if req.Method != "OPTIONS" {
					err := next(c)
					if err != nil {
						c.Error(err)
					}
				}
			} else {
				return echo.NewHTTPError(http.StatusUnauthorized, "Invalid Cors Origin")
			}

			return nil
		}
	})
}

// isOriginAllowed checks if the origin matches any of the allowed patterns
func isOriginAllowed(origin string, patterns []DomainPattern) bool {
	for _, pattern := range patterns {
		if pattern.Pattern == "" && origin == "" {
			return true
		}

		if strings.HasPrefix(origin, "chrome-extension://") {
			if strings.HasPrefix(pattern.Pattern, "chrome-extension://") {
				if origin == pattern.Pattern {
					return true
				}
			}
			continue
		}

		parts := strings.SplitN(origin, "://", 2)
		if len(parts) != 2 {
			continue
		}
		protocol := parts[0]
		domain := parts[1]

		if protocol == "http" && !pattern.AllowHTTP {
			continue
		}
		if protocol == "https" && !pattern.AllowHTTPS {
			continue
		}

		if pattern.IsLocalhost {
			// Match localhost with any port
			localhostPattern := regexp.MustCompile(`^localhost(:\d+)?$`)
			if localhostPattern.MatchString(domain) {
				return true
			}
			continue
		}

		if !pattern.IsRegex {
			if pattern.Pattern == domain {
				return true
			}
			continue
		}

		// Handle wildcard pattern
		if strings.HasPrefix(pattern.Pattern, "*.") {
			suffix := strings.TrimPrefix(pattern.Pattern, "*")
			if strings.HasSuffix(domain, suffix) {
				return true
			}
		}
	}

	return false
}
