package handler

import (
	"bytes"
	"context"
	"io"
	"sync"
	"testing"

	"github.com/getkin/kin-openapi/openapi3"
	"github.com/getkin/kin-openapi/openapi3filter"
	"github.com/getkin/kin-openapi/routers"
	"github.com/getkin/kin-openapi/routers/gorillamux"
	"github.com/gin-gonic/gin"
)

var (
	specOnce   sync.Once
	specDoc    *openapi3.T
	specRouter routers.Router
)

func loadSpec(t *testing.T) (*openapi3.T, routers.Router) {
	t.Helper()
	specOnce.Do(func() {
		loader := openapi3.NewLoader()
		doc, err := loader.LoadFromFile("../../api/openapi.yaml")
		if err != nil {
			panic("failed to load OpenAPI spec: " + err.Error())
		}
		if err := doc.Validate(loader.Context); err != nil {
			panic("invalid OpenAPI spec: " + err.Error())
		}
		specDoc = doc
		specRouter, _ = gorillamux.NewRouter(doc)
	})
	return specDoc, specRouter
}

type responseCapture struct {
	gin.ResponseWriter
	body *bytes.Buffer
}

func (rc *responseCapture) Write(b []byte) (int, error) {
	rc.body.Write(b)
	return rc.ResponseWriter.Write(b)
}

func contractValidationMiddleware(t *testing.T) gin.HandlerFunc {
	_, router := loadSpec(t)
	return func(c *gin.Context) {
		capture := &responseCapture{ResponseWriter: c.Writer, body: &bytes.Buffer{}}
		c.Writer = capture

		c.Next()

		route, pathParams, err := router.FindRoute(c.Request)
		if err != nil {
			return
		}

		requestInput := &openapi3filter.RequestValidationInput{
			Request:    c.Request,
			PathParams: pathParams,
			Route:      route,
			Options: &openapi3filter.Options{
				SkipSettingDefaults: true,
				AuthenticationFunc: func(_ context.Context, _ *openapi3filter.AuthenticationInput) error {
					return nil
				},
			},
		}

		responseInput := &openapi3filter.ResponseValidationInput{
			RequestValidationInput: requestInput,
			Status:                 c.Writer.Status(),
			Header:                 c.Writer.Header(),
			Body:                   io.NopCloser(bytes.NewReader(capture.body.Bytes())),
			Options: &openapi3filter.Options{
				SkipSettingDefaults: true,
				AuthenticationFunc: func(_ context.Context, _ *openapi3filter.AuthenticationInput) error {
					return nil
				},
			},
		}

		if err := openapi3filter.ValidateResponse(c.Request.Context(), responseInput); err != nil {
			t.Errorf("OpenAPI contract violation: %s %s → %d\n%s",
				c.Request.Method, c.Request.URL.Path, c.Writer.Status(), err)
		}
	}
}
