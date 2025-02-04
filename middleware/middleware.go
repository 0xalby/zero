package middleware

import (
	"context"
	"net/http"
	"zero/handlers"
	"zero/services"
	"zero/utils"
)

type key int

const userKey key = 0

// Middleware function that assures a user is verified and can access protected routes
func Verified(h *handlers.UsersHandler) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// Getting user id from request context
			id, err := services.GetUserIdFromContext(r)
			if err != nil {
				utils.Response(w, http.StatusUnauthorized, "unauthorized")
				return
			}
			// Getting user by id
			user, err := h.US.GetUserById(id)
			if err != nil {
				utils.Response(w, http.StatusUnauthorized, "unable to find user")
				return
			}
			// Checking if the user is verified
			if !user.Verified {
				utils.Response(w, http.StatusForbidden, "email not verified")
				return
			}
			ctx := context.WithValue(r.Context(), userKey, user)
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}
