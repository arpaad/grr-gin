// Package grrgin provides Gin middleware for grr: injecting a registry
// into each request's context and managing a per-request scope. Lives in
// its own repo/module so neither grr nor gold need a Gin dependency.
package grrgin

import (
	"github.com/gin-gonic/gin"

	"github.com/arpaad/grr"
)

// Middleware attaches r to each request's context and begins a scope that
// lives for the duration of the request.
func Middleware(r *grr.Registry) gin.HandlerFunc {
	return func(c *gin.Context) {
		ctx := grr.WithRegistry(c.Request.Context(), r)
		ctx, endScope := r.BeginScope(ctx)
		defer endScope()
		c.Request = c.Request.WithContext(ctx)
		c.Next()
	}
}
