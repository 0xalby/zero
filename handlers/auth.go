package handlers

import (
	"net/http"
	"strings"
	"time"
	"zero/config"
	"zero/services"
	"zero/types"
	"zero/utils"

	"golang.org/x/crypto/bcrypt"
)

type AuthHandler struct {
	US *services.UsersService
	ES *services.EmailService
}

// Signs up the user
func (h *AuthHandler) SignUp(w http.ResponseWriter, r *http.Request) {
	logger := config.GetLogger()
	payload := new(types.UserSignUpPayload)
	if err := utils.Validate(w, r, payload); err != nil {
		return
	}
	// Hashing password
	hashed, err := Hash(payload.Password)
	if err != nil || hashed == "" {
		utils.Response(w, http.StatusInternalServerError, "internal server error")
		logger.Error("failed to hash", "payload.Password", payload.Password, "err", err)
		return
	}
	// Adding the user to the database
	if err := h.US.CreateUser(&types.User{Customer: CreateCustomer(payload.Username, payload.Email), Username: payload.Username, Email: payload.Email, Password: hashed}); err != nil {
		if strings.HasPrefix(err.Error(), "UNIQUE") {
			utils.Response(w, http.StatusConflict, "name or email already taken")
			return
		}
		utils.Response(w, http.StatusInternalServerError, "failed to create user")
		return
	}
	utils.Response(w, http.StatusCreated, "registered")
}

// Signs in the user
func (h *AuthHandler) SignIn(w http.ResponseWriter, r *http.Request) {
	logger := config.GetLogger()
	payload := new(types.UserSignInPayload)
	if err := utils.Validate(w, r, payload); err != nil {
		return
	}
	// Getting the user by email
	user, err := h.US.GetUserByEmail(payload.Email)
	if err != nil || user == nil {
		if strings.HasPrefix(err.Error(), "email") {
			utils.Response(w, http.StatusNotFound, "email not found")
			return
		}
		utils.Response(w, http.StatusInternalServerError, "internal server error")
		logger.Error("failed to get user email", "payload.Email", payload.Email, "err", err)
		return
	}
	// Comparing inputted password with the database one for that user
	if !h.US.ComparePasswords(user.Password, []byte(payload.Password)) {
		utils.Response(w, http.StatusUnauthorized, "wrong password sorry")
		return
	}
	// Generating a new JWT token providing access to protected routes for 12 hours
	expiration := time.Now().Add(time.Hour * 12)
	_, token, err := config.TokenAuth.Encode(map[string]interface{}{
		"user_id": user.Id,
		"exp":     expiration.Unix(),
	})
	if err != nil {
		utils.Response(w, http.StatusInternalServerError, "failes to generate jwt")
		logger.Error("failed to generate jwt", "user.Id", user.Id, "err", err)
		return
	}
	// Setting the JWT token as a Secure HttpOnly Cookie
	http.SetCookie(w, &http.Cookie{
		Name:     "jwt",
		Value:    token,
		HttpOnly: true,
		Secure:   true,
		Path:     "/",
		MaxAge:   int(time.Until(expiration).Seconds()),
		SameSite: http.SameSiteLaxMode,
		Expires:  expiration})
	response := map[string]string{"token": token, "redirect": "/dashboard"}
	utils.Response(w, http.StatusOK, response)
}

// Returns a hashed string using bcrypt
func Hash(password string) (string, error) {
	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return "", err
	}
	return string(hash), nil
}

// Sends an email containing a verification code
func (h *AuthHandler) SendVerificationEmail(w http.ResponseWriter, r *http.Request) {
	logger := config.GetLogger()
	// Generates a verfication code
	code, err := h.ES.GenerateVerificationCode()
	if err != nil {
		utils.Response(w, http.StatusInternalServerError, "internal server error")
		return
	}
	payload := new(types.SendVerificationEmailPayload)
	if err := utils.Validate(w, r, payload); err != nil {
		return
	}
	// Getting the user by email
	user, err := h.US.GetUserByEmail(payload.Email)
	if err != nil {
		utils.Response(w, http.StatusBadRequest, "email not found")
		return
	}
	// Adding the verification code to the database
	if err := h.ES.AddVerificationEmailCode(user.Id, code); err != nil {
		utils.Response(w, http.StatusInternalServerError, "internal server error")
		logger.Error("failed to add verification code to the database", "code", code, "err", err)
		return
	}
	// Sending verification email
	if err := h.ES.SendVerificationEmail(user.Email, user.Username, code); err != nil {
		utils.Response(w, http.StatusInternalServerError, "internal server error failed to send email")
		logger.Error("failed to send verification email", "email", user.Email, "code", code, "err", err)
		return
	}
	utils.Response(w, http.StatusOK, "verification email with code sent")
}

// Completes user verification asking for the code received by email
func (h *AuthHandler) CompleteVerification(w http.ResponseWriter, r *http.Request) {
	logger := config.GetLogger()
	payload := new(types.CompleteVerificationPayload)
	if err := utils.Validate(w, r, payload); err != nil {
		return
	}
	// Getting user by email
	user, err := h.US.GetUserByEmail(payload.Email)
	if err != nil {
		utils.Response(w, http.StatusBadRequest, "email not found")
		return
	}
	// Compares inputted code with the database ones for that user
	if err := h.ES.CompareVerificationCode(user.Id, payload.Code); err != nil {
		utils.Response(w, http.StatusUnauthorized, "verification code not valid")
		return
	}
	// Marks user as verified
	if err := h.ES.MarkUserAsVerified(user.Id); err != nil {
		utils.Response(w, http.StatusInternalServerError, "internal server error")
		logger.Error("failed to mark the user as verified", "user", user.Id, "err", err)
		return
	}
	utils.Response(w, http.StatusOK, "verified")
}
