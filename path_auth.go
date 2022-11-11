package echo_middleware_path_auth

import (
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	"net/http"
)

type (
	// PathAuthConfig defines the config for PathAuth middleware.
	PathAuthConfig struct {
		// Skipper defines a function to skip middleware.
		Skipper middleware.Skipper

		// Validator is a function to validate key.
		// Required.
		Validator PathAuthValidator

		Param string
	}

	// PathAuthValidator defines a function to validate PathAuth credentials.
	PathAuthValidator func(auth string, c echo.Context) (bool, error)
)

var (
	// DefaultKeyAuthConfig is the default PathAuth middleware config.
	DefaultKeyAuthConfig = PathAuthConfig{
		Skipper: middleware.DefaultSkipper,
	}
)

// ErrKeyAuthMissing is error type when PathAuth middleware is unable to extract value from lookups
var ErrKeyAuthMissing = echo.NewHTTPError(http.StatusBadRequest, "Missing key in the request")

// PathAuth returns an PathAuth middleware.
//
// For valid key it calls the next handler.
// For invalid key, it sends "401 - Unauthorized" response.
// For missing key, it sends "400 - Bad Request" response.
func PathAuth(param string, fn PathAuthValidator) echo.MiddlewareFunc {
	c := DefaultKeyAuthConfig
	c.Validator = fn
	c.Param = param
	return PathAuthWithConfig(c)
}

func PathAuthWithConfig(config PathAuthConfig) echo.MiddlewareFunc {
	if config.Skipper == nil {
		config.Skipper = DefaultKeyAuthConfig.Skipper
	}
	if config.Validator == nil {
		panic("PathAuth: requires a validator function")
	}

	if config.Param == "" {
		panic("PathAuth: requires a param")
	}

	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			if config.Skipper(c) {
				return next(c)
			}

			if !extract(config.Param, c.ParamNames()) {
				return &echo.HTTPError{
					Code:     http.StatusBadRequest,
					Message:  http.StatusText(http.StatusBadRequest),
					Internal: ErrKeyAuthMissing,
				}
			}

			valid, err := config.Validator(c.Param(config.Param), c)
			if err == nil && valid {
				return next(c)
			}

			return &echo.HTTPError{
				Code:     http.StatusUnauthorized,
				Message:  http.StatusText(http.StatusUnauthorized),
				Internal: err,
			}
		}
	}
}

func extract(cParam string, params []string) bool {
	for _, param := range params {
		if cParam == param {
			return true
		}
	}

	return false
}
