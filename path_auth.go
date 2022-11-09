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

		// ErrorHandler defines a function which is executed for an invalid key.
		// It may be used to define a custom error.
		ErrorHandler PathAuthErrorHandler

		Param string
	}

	// PathAuthValidator defines a function to validate PathAuth credentials.
	PathAuthValidator func(auth string, c echo.Context) (bool, error)

	// PathAuthErrorHandler defines a function which is executed for an invalid key.
	PathAuthErrorHandler func(err error, c echo.Context) error
)

var (
	// DefaultKeyAuthConfig is the default PathAuth middleware config.
	DefaultKeyAuthConfig = PathAuthConfig{
		Skipper: middleware.DefaultSkipper,
	}
)

// ErrKeyAuthMissing is error type when PathAuth middleware is unable to extract value from lookups
type ErrKeyAuthMissing struct {
	Err error
}

// Error returns errors text
func (e *ErrKeyAuthMissing) Error() string {
	return e.Err.Error()
}

// Unwrap unwraps error
func (e *ErrKeyAuthMissing) Unwrap() error {
	return e.Err
}

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

	if len(config.Param) == 0 {
		panic("PathAuth: requires a param")
	}

	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			if config.Skipper(c) {
				return next(c)
			}
			valid, err := config.Validator(c.Param(config.Param), c)
			if err != nil {
				return &echo.HTTPError{
					Code:     http.StatusUnauthorized,
					Message:  "Unauthorized",
					Internal: err,
				}
			}

			if valid {
				return next(c)
			}

			return echo.NewHTTPError(http.StatusBadRequest, err.Error())
		}
	}
}
