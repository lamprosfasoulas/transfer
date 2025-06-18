package middleware

import (
	"context"
	"fmt"
	"net/http"
	"strings"

	"github.com/lamprosfasoulas/transfer/pkg/auth"
	"github.com/lamprosfasoulas/transfer/pkg/database"
	"github.com/lamprosfasoulas/transfer/pkg/logger"
)

const cookieName = "auth-token"

type Middleware struct {
	AuthProvider auth.AuthProvider
	Database database.Database
	Logger *logger.Logger
}

func NewMiddleware(a auth.AuthProvider, d database.Database, l *logger.Logger) *Middleware{
	return &Middleware{
		AuthProvider: a,
		Database: d,
		Logger: l,
	}

}

// isTerminal is used to determine if client is 
// connecting with browser or cli app
func isTerminal(userAgent string) bool{
    progs := []string{"curl","wget","HTTPie","fetch","Go-http-client"}
    for _,prog := range progs {
        if strings.Contains(userAgent, prog){
            return true
        }
    }
    return false
}

// getHeaderUser tries to get the user's username from
// the Authorization Header. This is used for cli apps.
func (m *Middleware) getHeaderUser(r *http.Request) (string, error) {
	authz := r.Header.Get("Authorization")
	if authz == "" {
		return "", fmt.Errorf("No user found in request")
	}
	if strings.HasPrefix(authz, "Bearer") {
		authResp := m.AuthProvider.ValidateToken(strings.TrimPrefix(authz, "Bearer "))
		if authResp.Error == nil {
			return authResp.User.Username, nil
		}
	}
	return "", fmt.Errorf("Could not ParseJWT")
}

// getCookieUser is used to get the user's username from
// browser cookies. This is used for browser access.
func (m *Middleware) getCookieUser(r *http.Request) (string, error) {
	cookie, err := r.Cookie(cookieName)
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
func (m *Middleware) GetUserFromRequest(r *http.Request) (string, error) {
	resCookie, err := m.getCookieUser(r)
	if err == nil {
		return resCookie, nil
	}
	resHeader, err:= m.getHeaderUser(r)
	if err == nil {
		return resHeader, nil
	}
	return "", err
}

// RequireAuth is a method that grabs the user's username
// from GetUserFromRequest and generates Context Values that
// get passed to handler functions. Handlers use this info to 
// access user values.
func (m *Middleware) RequireAuth(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		//Get the User-Agent
		isItReally := isTerminal(r.Header.Get("User-Agent"))

		user, err := m.GetUserFromRequest(r)
		//Get user space to allow or not uploads
		space, err := m.Database.GetUserSpace(r.Context(),user)
		if err != nil {
			m.Logger.Warn(logger.Mid).Writef("Error getting user space", err)
		}
		ctx := context.WithValue(r.Context(), "username", user)
		ctx = context.WithValue(ctx, "isTerminal", isItReally)
		ctx = context.WithValue(ctx, "space", space)
		next(w, r.WithContext(ctx))
	}
}

