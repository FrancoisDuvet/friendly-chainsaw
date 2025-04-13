package middleware

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/gorilla/sessions"
)

var Store *sessions.CookieStore

func InitSessionStore(secret []byte) {
	Store = sessions.NewCookieStore(secret)
}

// Middleware to extract email from session and inject into context
func RequireSession() gin.HandlerFunc {
	return func(c *gin.Context) {
		session, err := Store.Get(c.Request, "session")
		if err != nil {
			c.String(http.StatusUnauthorized, "Session error")
			c.Abort()
			return
		}

		email, ok := session.Values["user_email"].(string)
		if !ok || email == "" {
			c.String(http.StatusUnauthorized, "Unauthorized: no email in session")
			c.Abort()
			return
		}

		c.Set("user_email", email)
		c.Next()
	}
}
func RequireRole(expectedRole string) gin.HandlerFunc {
	return func(c *gin.Context) {
		session, err := Store.Get(c.Request, "session")
		if err != nil {
			c.String(http.StatusUnauthorized, "Session error")
			c.Abort()
			return
		}

		role, ok := session.Values["user_role"].(string)
		if !ok || role != expectedRole {
			c.String(http.StatusForbidden, "Forbidden: You do not have access to this resource")
			c.Abort()
			return
		}

		c.Set("user_role", role)
		c.Next()
	}
}
