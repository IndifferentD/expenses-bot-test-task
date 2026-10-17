package api

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"

	"expenses-bot/backend/internal"
)

type UserHandler struct {
	service *internal.UserService
}

func NewUserHandler(service *internal.UserService) *UserHandler {
	return &UserHandler{service: service}
}

func (h *UserHandler) CreateUser(w http.ResponseWriter, r *http.Request) {
	user, err := readUser(r)
	if err != nil {
		writeError(w, r, err)
		return
	}
	created, err := h.service.Create(r.Context(), user)
	if err != nil {
		writeError(w, r, err)
		return
	}
	writeJSON(w, http.StatusCreated, created)
}

func (h *UserHandler) ListUsers(w http.ResponseWriter, r *http.Request) {
	users, err := h.service.List(r.Context())
	if err != nil {
		writeError(w, r, err)
		return
	}
	writeJSON(w, http.StatusOK, struct {
		Items []internal.User `json:"items"`
	}{Items: users})
}

func (h *UserHandler) GetUser(w http.ResponseWriter, r *http.Request) {
	id, err := resourceID(r, "user")
	if err != nil {
		writeError(w, r, err)
		return
	}
	user, err := h.service.Get(r.Context(), id)
	if err != nil {
		writeError(w, r, err)
		return
	}
	writeJSON(w, http.StatusOK, user)
}

func (h *UserHandler) UpdateUser(w http.ResponseWriter, r *http.Request) {
	id, err := resourceID(r, "user")
	if err != nil {
		writeError(w, r, err)
		return
	}
	user, err := readUser(r)
	if err != nil {
		writeError(w, r, err)
		return
	}
	user.ID = id
	updated, err := h.service.Update(r.Context(), user)
	if err != nil {
		writeError(w, r, err)
		return
	}
	writeJSON(w, http.StatusOK, updated)
}

func (h *UserHandler) DeleteUser(w http.ResponseWriter, r *http.Request) {
	id, err := resourceID(r, "user")
	if err != nil {
		writeError(w, r, err)
		return
	}
	if err := h.service.Delete(r.Context(), id); err != nil {
		writeError(w, r, err)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

func readUser(r *http.Request) (internal.User, error) {
	body, err := io.ReadAll(r.Body)
	if err != nil {
		return internal.User{}, fmt.Errorf("failed to read user request: %w", &requestError{
			Code: codeInvalidBody, Reason: "failed to read request body", Cause: err,
		})
	}
	var user internal.User
	if err := json.Unmarshal(body, &user); err != nil {
		return internal.User{}, fmt.Errorf("failed to decode user request: %w", &requestError{
			Code: codeInvalidJSON, Reason: "invalid JSON request", Cause: err,
		})
	}
	return user, nil
}
