package handlers

import (
	"net/http"
	"strings"
	"zero/config"
	"zero/services"
	"zero/types"
	"zero/utils"
)

type UsersHandler struct {
	US *services.UsersService
}

// Updates user name
func (h *UsersHandler) UsersUpdateUsername(w http.ResponseWriter, r *http.Request) {
	logger := config.GetLogger()
	payload := new(types.UpdateUserNamePayload)
	if err := utils.Validate(w, r, payload); err != nil {
		return
	}
	// Getting user id from request context
	id, err := services.GetUserIdFromContext(r)
	if err != nil {
		utils.Response(w, http.StatusUnauthorized, "unauthorized")
		return
	}
	// Getting user by id
	user, err := h.US.GetUserById(id)
	if err != nil {
		utils.Response(w, http.StatusInternalServerError, "internal server error")
		logger.Info("failed to get user id", "user.Id", user.Id, "user.Name", user.Username, "err", err)
		return
	}
	if user.Username != payload.Old {
		utils.Response(w, http.StatusBadRequest, "name not found")
		return
	}
	// Updating user's name in the database
	err = h.US.UpdateUserName(id, payload.New, payload.Old)
	if err != nil {
		if strings.Contains(err.Error(), "UNIQUE") {
			utils.Response(w, http.StatusConflict, "name already taken")
			return
		}
		utils.Response(w, http.StatusBadRequest, "failed to update user name")
		return
	}
	utils.Response(w, http.StatusOK, "updated")
}

// Updates user email
func (h *UsersHandler) UsersUpdateEmail(w http.ResponseWriter, r *http.Request) {
	logger := config.GetLogger()
	payload := new(types.UpdateUserEmailPayload)
	if err := utils.Validate(w, r, payload); err != nil {
		return
	}
	// Getting user id from request context
	id, err := services.GetUserIdFromContext(r)
	if err != nil {
		utils.Response(w, http.StatusUnauthorized, "unauthorized")
		return
	}
	// Getting user by id
	user, err := h.US.GetUserById(id)
	if err != nil {
		utils.Response(w, http.StatusInternalServerError, "internal server error")
		logger.Info("failed to get user by id", "id", id, "err", err)
		return
	}
	if user.Email != payload.Old {
		utils.Response(w, http.StatusBadRequest, "wrong email")
		logger.Warn("someone crafted a request with a wrong email", "user.Id", user.Id, "user.Email", user.Email)
		return
	}
	// Updating user's email in the database
	err = h.US.UpdateUserEmail(id, payload.New, payload.Old)
	if err != nil {
		if strings.Contains(err.Error(), "UNIQUE") {
			utils.Response(w, http.StatusConflict, "email already used")
			return
		}
		utils.Response(w, http.StatusBadRequest, "failed to update user email")
		return
	}
	utils.Response(w, http.StatusOK, "updated")
}

// Updates user's password
func (h *UsersHandler) UsersUpdatePassword(w http.ResponseWriter, r *http.Request) {
	logger := config.GetLogger()
	payload := new(types.UpdateUserPasswordPayload)
	if err := utils.Validate(w, r, payload); err != nil {
		return
	}
	// Getting user id from request context
	id, err := services.GetUserIdFromContext(r)
	if err != nil {
		utils.Response(w, http.StatusUnauthorized, "unauthorized")
		return
	}
	// Hashing user's new password
	hashed, err := Hash(payload.Password)
	if err != nil || hashed == "" {
		utils.Response(w, http.StatusInternalServerError, "internal server error")
		logger.Info("failed to hash", "err", err)
		return
	}
	// Updating user's password in the database
	err = h.US.UpdateUserPassword(id, hashed)
	if err != nil {
		utils.Response(w, http.StatusInternalServerError, "internal server error")
		return
	}
	utils.Response(w, http.StatusOK, "updated")
}

// Deletes a user
func (h *UsersHandler) UsersDelete(w http.ResponseWriter, r *http.Request) {
	logger := config.GetLogger()
	payload := new(types.DeleteAccountPayload)
	if err := utils.Validate(w, r, payload); err != nil {
		return
	}
	// Getting user id by request context
	id, err := services.GetUserIdFromContext(r)
	if err != nil {
		utils.Response(w, http.StatusUnauthorized, "unauthorized")
		return
	}
	// Getting user by id
	user, err := h.US.GetUserById(id)
	if err != nil {
		utils.Response(w, http.StatusBadRequest, "user not found")
		return
	}
	// Comparing user inputted password and the database one for that user as confirmation
	if !h.US.ComparePasswords(user.Password, []byte(payload.Password)) {
		utils.Response(w, http.StatusUnauthorized, "wrong password")
		return
	}
	// Deleting the user from the database
	if err := h.US.DeleteUser(id); err != nil {
		utils.Response(w, http.StatusInternalServerError, "internal server error")
		logger.Info("failed to delete user", "id", id, "err", err)
		return
	}
	utils.Response(w, http.StatusOK, "deleted")
}
