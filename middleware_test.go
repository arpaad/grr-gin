package grrgin_test

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"

	"github.com/arpaad/grr"
	grrgin "github.com/arpaad/grr-gin"
)

func TestMiddlewareInjectsRegistryAndScope(t *testing.T) {
	gin.SetMode(gin.TestMode)

	r := grr.New()
	calls := 0
	r.RegisterScoped("conn", func(ctx context.Context) any {
		calls++
		return calls
	})

	router := gin.New()
	router.Use(grrgin.Middleware(r))
	router.GET("/", func(c *gin.Context) {
		a := r.Resolve(c.Request.Context(), "conn")
		b := r.Resolve(c.Request.Context(), "conn")
		if a != b {
			t.Fatalf("expected same scoped instance within one request, got %v and %v", a, b)
		}
		c.Status(http.StatusOK)
	})

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("got status %d, want 200", w.Code)
	}
	if calls != 1 {
		t.Fatalf("factory called %d times, want 1", calls)
	}
}

func TestMiddlewareScopeEndsAfterRequest(t *testing.T) {
	gin.SetMode(gin.TestMode)

	r := grr.New()
	r.RegisterScoped("conn", func(ctx context.Context) any { return 1 })

	var savedCtx context.Context
	router := gin.New()
	router.Use(grrgin.Middleware(r))
	router.GET("/", func(c *gin.Context) {
		savedCtx = c.Request.Context()
		c.Status(http.StatusOK)
	})

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	defer func() {
		if recover() == nil {
			t.Fatal("expected panic resolving scoped key after request scope ended")
		}
	}()
	r.Resolve(savedCtx, "conn")
}
