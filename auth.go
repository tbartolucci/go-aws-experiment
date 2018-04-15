package main

import (
	"net/http"

	"github.com/gin-contrib/sessions"
	"github.com/gin-gonic/gin"
)

// AuthRequired an authentication middleware. If the user session is not found,
// the user is redirected to /signup.
func AuthRequired() gin.HandlerFunc {
	return func(c *gin.Context) {
		s := sessions.Default(c)
		uid := s.Get(userKey)

		if uid == nil {
			c.Redirect(http.StatusFound, "/signup")
			return
		}

		c.Next()
	}
}
