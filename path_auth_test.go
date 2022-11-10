package echo_middleware_path_auth

import (
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/labstack/echo/v4"
	"github.com/stretchr/testify/assert"
)

func testKeyValidator(key string, _ echo.Context) (bool, error) {
	switch key {
	case "valid-key":
		return true, nil
	case "error-key":
		return false, errors.New("some user defined error")
	default:
		return false, nil
	}
}

func TestKeyAuth(t *testing.T) {
	t.Run("auth ok", func(t *testing.T) {
		handlerCalled := false
		handler := func(c echo.Context) error {
			handlerCalled = true
			return c.String(http.StatusOK, "test")
		}
		middlewareChain := PathAuth("apikey", testKeyValidator)(handler)

		e := echo.New()
		e.GET("/:apikey", middlewareChain)

		req := httptest.NewRequest(http.MethodGet, "/", nil)
		rec := httptest.NewRecorder()

		c := e.NewContext(req, rec)
		e.Router().Find(http.MethodGet, "/valid-key", c)
		err := middlewareChain(c)

		assert.NoError(t, err)
		assert.True(t, handlerCalled)
	})

	t.Run("auth nok", func(t *testing.T) {
		handlerCalled := false
		handler := func(c echo.Context) error {
			handlerCalled = true
			return c.String(http.StatusOK, "test")
		}
		middlewareChain := PathAuth("apikey", testKeyValidator)(handler)

		e := echo.New()
		e.GET("/:apikey", middlewareChain)

		req := httptest.NewRequest(http.MethodGet, "/", nil)
		rec := httptest.NewRecorder()

		c := e.NewContext(req, rec)
		e.Router().Find(http.MethodGet, "/error-key", c)
		err := middlewareChain(c)

		assert.Error(t, err)
		assert.False(t, handlerCalled)
	})

}

func TestPathAuthWithConfig(t *testing.T) {
	var testCases = []struct {
		name                string
		givenRequestFunc    func() *http.Request
		givenRequest        func(req *http.Request)
		whenConfig          func(conf *PathAuthConfig)
		pathName            string
		expectHandlerCalled bool
		expectError         string
	}{
		{
			name: "ok success",
			givenRequestFunc: func() *http.Request {
				req := httptest.NewRequest(http.MethodGet, "/valid-key", nil)
				return req
			},
			expectHandlerCalled: true,
			expectError:         "",
		},
		{
			name: "ng user error",
			givenRequestFunc: func() *http.Request {
				req := httptest.NewRequest(http.MethodGet, "/error-key", nil)
				return req
			},
			expectHandlerCalled: false,
			expectError:         "code=401, message=Unauthorized, internal=some user defined error",
		},
		{
			name: "ng no valid no error",
			givenRequestFunc: func() *http.Request {
				req := httptest.NewRequest(http.MethodGet, "/bad", nil)
				return req
			},
			expectHandlerCalled: false,
			expectError:         "code=400, message=Bad Request",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			handlerCalled := false
			handler := func(c echo.Context) error {
				handlerCalled = true
				return c.String(http.StatusOK, "test")
			}
			config := PathAuthConfig{
				Validator: testKeyValidator,
				Param:     "apikey",
			}
			if tc.whenConfig != nil {
				tc.whenConfig(&config)
			}
			middlewareChain := PathAuthWithConfig(config)(handler)

			e := echo.New()
			req := httptest.NewRequest(http.MethodGet, "/", nil)
			if tc.givenRequestFunc != nil {
				req = tc.givenRequestFunc()
			}
			if tc.givenRequest != nil {
				tc.givenRequest(req)
			}
			e.GET("/:apikey", middlewareChain)
			rec := httptest.NewRecorder()
			c := e.NewContext(req, rec)
			// use params
			e.Router().Find(http.MethodGet, req.URL.Path, c)
			err := middlewareChain(c)

			assert.Equal(t, tc.expectHandlerCalled, handlerCalled)
			if tc.expectError != "" {
				assert.EqualError(t, err, tc.expectError)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestPathAuthWithConfig_panicsOnEmptyValidator(t *testing.T) {
	assert.PanicsWithValue(
		t,
		"PathAuth: requires a validator function",
		func() {
			handler := func(c echo.Context) error {
				return c.String(http.StatusOK, "test")
			}
			PathAuthWithConfig(PathAuthConfig{
				Validator: nil,
			})(handler)
		},
	)
}

func TestPathAuthWithConfig_panicsOnEmptyParam(t *testing.T) {
	assert.PanicsWithValue(
		t,
		"PathAuth: requires a param",
		func() {
			handler := func(c echo.Context) error {
				return c.String(http.StatusOK, "test")
			}
			PathAuthWithConfig(PathAuthConfig{
				Validator: func(auth string, c echo.Context) (bool, error) {
					return true, nil
				},
				Param: "",
			})(handler)
		},
	)

	assert.PanicsWithValue(
		t,
		"PathAuth: requires a param",
		func() {
			handler := func(c echo.Context) error {
				return c.String(http.StatusOK, "test")
			}
			PathAuth("", func(auth string, c echo.Context) (bool, error) {
				return true, nil
			})(handler)
		},
	)
}
