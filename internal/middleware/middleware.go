package middleware

import (
	"context"
	"fmt"
	"net/http"
	"strings"

	"github.com/lamprosfasoulas/transfer/internal/auth"
	"github.com/lamprosfasoulas/transfer/internal/logger"
	"github.com/lamprosfasoulas/transfer/internal/storage"
)

const authCookie = "auth-token"
const spaceCookie = "space-used"

type Middleware struct {
	AuthProvider auth.AuthProvider
	Storage      storage.Storage
	Logger       *logger.Logger
}

func NewMiddleware(a auth.AuthProvider, s storage.Storage, l *logger.Logger) *Middleware {
	return &Middleware{
		AuthProvider: a,
		Storage:      s,
		Logger:       l,
	}

}

// isTerminal is used to determine if client is
// connecting with browser or cli app
func isTerminal(userAgent string) bool {
	progs := []string{"curl", "wget", "HTTPie", "fetch", "Go-http-client"}
	for _, prog := range progs {
		if strings.Contains(userAgent, prog) {
			return true
		}
	}
	return false
}

// getHeaderUser tries to get the user's username from
// the Authorization Header. This is used for cli apps.
func (m *Middleware) getHeaderUser(r *http.Request) (string, error) {
	authh := r.Header.Get("Authorization")
	if authh == "" {
		return "", fmt.Errorf("No user found in request")
	}
	if strings.HasPrefix(authh, "Bearer") {
		authResp := m.AuthProvider.ValidateToken(strings.TrimPrefix(authh, "Bearer "))
		if authResp.Error == nil {
			return authResp.User.Username, nil
		}
	}
	return "", fmt.Errorf("Could not ParseJWT")
}

// getCookieUser is used to get the user's username from
// browser cookies. This is used for browser access.
func (m *Middleware) getCookieUser(r *http.Request) (string, error) {
	cookie, err := r.Cookie(authCookie)
	if err != nil {
		return "", err
	}
	authResp := m.AuthProvider.ValidateToken(cookie.Value)
	if authResp.Error != nil {
		return "", authResp.Error
	}
	return authResp.User.Username, nil
}

// GetUserFromRequest grabs the user from the getCookieUser
// and getHeaderUser methods. It is used by RequireAuth method.
func (m *Middleware) getUserFromRequest(r *http.Request, isTerm bool) (string, error) {

	var resUser string
	var err error

	if !isTerm {
		resUser, err = m.getCookieUser(r)
	} else {
		resUser, err = m.getHeaderUser(r)
	}

	if err != nil {
		return "", err
	}

	return resUser, nil
}

func (m *Middleware) getUserSpace(c context.Context) (int64, error) {
	if userVal := c.Value("username"); userVal != nil {
		if user, ok := userVal.(string); ok {
			return m.Storage.GetUserSpace(c, user)
		}
	}
	return 0, fmt.Errorf("Error getting user from context")
}

// RequireAuth is a method that grabs the user's username
// from GetUserFromRequest and generates Context Values that
// get passed to handler functions. Handlers use this info to
// access user values.
func (m *Middleware) RequireAuth(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		//Get the User-Agent
		isTerm := isTerminal(r.Header.Get("User-Agent"))

		user, err := m.getUserFromRequest(r, isTerm)
		if err != nil {
			m.Logger.Warn(logger.Middleware).Writef("Error getting user", err)
		}

		ctx := context.WithValue(r.Context(), "username", user)
		ctx = context.WithValue(ctx, "isTerminal", isTerm)
		next(w, r.WithContext(ctx))
	}
}

func (m *Middleware) PrepUpload(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()
		space, err := m.getUserSpace(ctx)
		if err != nil {
			m.Logger.Warn(logger.Middleware).Writef("Error getting user space", err)
			return
		}
		ctx = context.WithValue(ctx, "space", space)

		next(w, r.WithContext(ctx))
	}
}
