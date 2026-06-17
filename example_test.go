package grrgin_test

import (
	"context"
	"fmt"
	"net/http"
	"net/http/httptest"

	"github.com/gin-gonic/gin"

	"github.com/arpaad/grr"
	grrgin "github.com/arpaad/grr-gin"
)

func init() {
	gin.SetMode(gin.TestMode)
}

// Middleware injects a registry and a per-request scope into each request's
// context. Scoped dependencies resolve to the same instance within a request
// and are released when the request finishes.
func ExampleMiddleware() {
	r := grr.New()
	calls := 0
	r.RegisterScoped("greeting", func(_ context.Context) any {
		calls++
		return fmt.Sprintf("hello #%d", calls)
	})

	router := gin.New()
	router.Use(grrgin.Middleware(r))
	router.GET("/hi", func(c *gin.Context) {
		ctx := c.Request.Context()
		// Two resolves within the same request — same scoped instance.
		a := r.Resolve(ctx, "greeting")
		b := r.Resolve(ctx, "greeting")
		fmt.Fprintln(c.Writer, a)
		fmt.Fprintln(c.Writer, b)
		fmt.Fprintln(c.Writer, a == b)
	})

	req := httptest.NewRequest(http.MethodGet, "/hi", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	fmt.Print(w.Body.String())
	// Output:
	// hello #1
	// hello #1
	// true
}

// A second request gets a fresh scope — no shared state between requests.
func ExampleMiddleware_newScopePerRequest() {
	r := grr.New()
	calls := 0
	r.RegisterScoped("conn", func(_ context.Context) any {
		calls++
		return calls
	})

	router := gin.New()
	router.Use(grrgin.Middleware(r))
	router.GET("/", func(c *gin.Context) {
		v := r.Resolve(c.Request.Context(), "conn")
		fmt.Fprintln(c.Writer, v)
	})

	for i := 0; i < 2; i++ {
		req := httptest.NewRequest(http.MethodGet, "/", nil)
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)
		fmt.Print(w.Body.String())
	}
	// Output:
	// 1
	// 2
}
