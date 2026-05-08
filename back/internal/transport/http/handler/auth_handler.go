package handler

import (
	"net/http"

	"esports-backend/internal/transport/http/middleware"
)

type AuthHandler struct{ deps Deps }

func NewAuthHandler(deps Deps) *AuthHandler { return &AuthHandler{deps: deps} }

type registerRequest struct {
	FirstName string `json:"first_name" validate:"required,min=2,max=100"`
	LastName  string `json:"last_name" validate:"required,min=2,max=100"`
	Email     string `json:"email" validate:"required,email"`
	Phone     string `json:"phone" validate:"required,phone_ru"`
	Nickname  string `json:"nickname" validate:"required,min=2,max=50"`
	Password  string `json:"password" validate:"required,min=8,max=128"`
}

type loginRequest struct {
	Email    string `json:"email" validate:"required,email"`
	Password string `json:"password" validate:"required,min=8,max=128"`
}

type refreshRequest struct {
	RefreshToken string `json:"refresh_token" validate:"required"`
}

type forgotPasswordRequest struct {
	Email string `json:"email" validate:"required,email"`
}

type resetPasswordRequest struct {
	Token       string `json:"token" validate:"required"`
	NewPassword string `json:"new_password" validate:"required,min=8,max=128"`
}

func (h *AuthHandler) Register(w http.ResponseWriter, r *http.Request) {
	var req registerRequest
	if err := decodeJSON(r, &req); err != nil {
		writeError(w, err)
		return
	}
	if err := h.deps.Validate.Struct(req); err != nil {
		writeError(w, err)
		return
	}
	ua := r.UserAgent()
	ip := r.RemoteAddr
	user, tokens, err := h.deps.Auth.Register(r.Context(), serviceRegisterInput(req), &ua, &ip)
	if err != nil {
		writeError(w, err)
		return
	}
	writeJSON(w, http.StatusCreated, map[string]interface{}{"user": user, "tokens": tokens})
}

func (h *AuthHandler) Login(w http.ResponseWriter, r *http.Request) {
	var req loginRequest
	if err := decodeJSON(r, &req); err != nil {
		writeError(w, err)
		return
	}
	if err := h.deps.Validate.Struct(req); err != nil {
		writeError(w, err)
		return
	}
	ua := r.UserAgent()
	ip := r.RemoteAddr
	user, tokens, err := h.deps.Auth.Login(r.Context(), serviceLoginInput(req, &ua, &ip))
	if err != nil {
		writeError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, map[string]interface{}{"user": user, "tokens": tokens})
}

func (h *AuthHandler) Refresh(w http.ResponseWriter, r *http.Request) {
	var req refreshRequest
	if err := decodeJSON(r, &req); err != nil {
		writeError(w, err)
		return
	}
	if err := h.deps.Validate.Struct(req); err != nil {
		writeError(w, err)
		return
	}
	ua := r.UserAgent()
	ip := r.RemoteAddr
	tokens, err := h.deps.Auth.Refresh(r.Context(), req.RefreshToken, &ua, &ip)
	if err != nil {
		writeError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, map[string]interface{}{"tokens": tokens})
}

func (h *AuthHandler) Logout(w http.ResponseWriter, r *http.Request) {
	var req refreshRequest
	if err := decodeJSON(r, &req); err != nil {
		writeError(w, err)
		return
	}
	if err := h.deps.Auth.Logout(r.Context(), req.RefreshToken); err != nil {
		writeError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, map[string]string{"message": "logged out"})
}

func (h *AuthHandler) ForgotPassword(w http.ResponseWriter, r *http.Request) {
	var req forgotPasswordRequest
	if err := decodeJSON(r, &req); err != nil {
		writeError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, h.deps.Auth.ForgotPassword(r.Context(), req.Email))
}

func (h *AuthHandler) ResetPassword(w http.ResponseWriter, r *http.Request) {
	var req resetPasswordRequest
	if err := decodeJSON(r, &req); err != nil {
		writeError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, h.deps.Auth.ResetPassword(r.Context(), req.Token, req.NewPassword))
}

func mustUserID(r *http.Request) string {
	user := middleware.CurrentUser(r.Context())
	if user == nil {
		return ""
	}
	return user.UserID
}
