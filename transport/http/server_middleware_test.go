package http

import (
	"bytes"
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/go-kratos/kratos/v2/middleware"
)

func TestEarlyJSONDecodingFailureMiddleware(t *testing.T) {
	var middlewareCalled bool
	var middlewareErr error

	m := func(handler middleware.Handler) middleware.Handler {
		return func(ctx context.Context, req interface{}) (interface{}, error) {
			middlewareCalled = true
			reply, err := handler(ctx, req)
			middlewareErr = err
			return reply, err
		}
	}

	srv := NewServer(Middleware(m))
	srv.Route("/").POST("/test", func(ctx Context) error {
		var in struct {
			Name string `json:"name"`
		}
		h := ctx.Middleware(func(ctx context.Context, req interface{}) (interface{}, error) {
			if err := ctx.Bind(req); err != nil {
				return nil, err
			}
			return req, nil
		})
		_, err := h(ctx, &in)
		if err != nil {
			return err
		}
		return ctx.Result(200, "ok")
	})

	ts := httptest.NewServer(srv)
	defer ts.Close()

	resp, err := http.Post(ts.URL+"/test", "application/json", bytes.NewReader([]byte(`{"name": `)))
	if err != nil {
		t.Fatal(err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusBadRequest {
		t.Errorf("expected status 400, got %d", resp.StatusCode)
	}
	if !middlewareCalled {
		t.Error("expected middleware to be called")
	}
	if middlewareErr == nil {
		t.Error("expected middleware to capture decoding error")
	}
}
