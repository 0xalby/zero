package main

import (
	"database/sql"
	"zero/config"
	"zero/handlers"
	"zero/middleware"
	"zero/services"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/jwtauth/v5"
)

func Register(r chi.Router, db *sql.DB) {
	// Services
	usersService := &services.UsersService{DB: db}
	emailService := &services.EmailService{DB: db}

	// Handlers
	authHandler := handlers.AuthHandler{US: usersService, ES: emailService}
	usersHandler := handlers.UsersHandler{US: usersService}
	stripeHandler := handlers.StripeHandler{US: usersService}

	// Routes
	r.Route("/auth", func(r chi.Router) {
		r.Post("/signin", authHandler.SignIn)
		r.Post("/signup", authHandler.SignUp)
		r.Route("/email", func(r chi.Router) {
			r.Get("/send", authHandler.SendVerificationEmail)
			r.Post("/verify", authHandler.CompleteVerification)
		})
	})
	r.Route("/account", func(r chi.Router) {
		r.Use(jwtauth.Verifier(config.TokenAuth))
		r.Use(jwtauth.Authenticator(config.TokenAuth))
		r.Use(middleware.Verified(&usersHandler))
		r.Put("/update/name", usersHandler.UsersUpdateUsername)
		r.Put("/update/email", usersHandler.UsersUpdateEmail)
		r.Put("/update/password", usersHandler.UsersUpdatePassword)
		r.Delete("/delete", usersHandler.UsersDelete)
	})
	r.Route("/shop", func(r chi.Router) {
		r.With(jwtauth.Verifier(config.TokenAuth)).
			With(jwtauth.Authenticator(config.TokenAuth)).
			With(middleware.Verified(&usersHandler)).
			Post("/checkout", stripeHandler.StripeCheckout)
		r.Post("/webhook", stripeHandler.StripeWebhook)
	})
}
