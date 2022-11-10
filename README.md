
# echo_middleware_path_auth

middleware for path-based authentication of labstack echo. Best when using apikey for path.

example) https://example.com/api/this_is_api_key

This `this_is_api_key` part can be dynamically submitted to authentication.
For example, whether apikey is active, RateLimit is not exceeded, etc.


Much of this code is based on [key_auth.go in labstack/echo and its test code](https://github.com/labstack/echo/blob/01d7d01bbc1948cd308b2ae93a131654e6dba195/middleware/key_auth.go).

## Badges


[![MIT License](https://img.shields.io/badge/License-MIT-green.svg)](https://choosealicense.com/licenses/mit/)
[![Test](https://github.com/bootjp/echo_middleware_path_auth/actions/workflows/test.yml/badge.svg)](https://github.com/bootjp/echo_middleware_path_auth/actions/workflows/test.yml)


## Usage/Examples

```go
package main

import (
	pa "github.com/bootjp/echo_middleware_path_auth"
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
)

func main() {

	e := echo.New()
	// group route
	e.Group("/api", pa.PathAuth("apikey", func(auth string, c echo.Context) (bool, error) {
		// add your logic
		return true, nil
	}))

	// single route
	yourHttpHandler := func(c echo.Context) error { return c.String(200, "OK") }
	yourPathAuthLogic := func(auth string, c echo.Context) (bool, error) {
		return true, nil
	}

	e.GET("/api/:apikey", yourHttpHandler, pa.PathAuth("apikey", yourPathAuthLogic))

	// with config
	config := pa.PathAuthConfig{}
	config.Skipper = middleware.DefaultSkipper
	config.Param = "apikey"
	config.Validator = yourPathAuthLogic
	e.GET("/api/:apikey", yourHttpHandler, pa.PathAuthWithConfig(config))
}
```
