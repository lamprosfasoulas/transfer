package middleware

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"strings"

	//"strings"
	"time"

	"github.com/lamprosfasoulas/transfer/pkg/start"
)


const cookieName = "auth-token"

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

// SetAuthCookie is used to initialize the session for the user.
// It stores a cookie named const cookieName with the value of 
// the generated JWT token.
func SetAuthCookie(w http.ResponseWriter, value string, expiry time.Duration) {
	http.SetCookie(w, &http.Cookie{
		Name: cookieName,
		Value: value,
		Path: "/",
		MaxAge: int(expiry.Seconds()),
		HttpOnly: true,
		Secure: false,
		SameSite: http.SameSiteLaxMode,
	})
}

// ClearAuthCookie terminates the user  session by deleting the cookie.
// It does not revoke the JWT token (i don't think it can be done), 
// but logs out the browser client.
// Not used for cli sessions.
func ClearAuthCookie(w http.ResponseWriter) {
	http.SetCookie(w, &http.Cookie{
		Name: cookieName,
		Value: "",
		Path: "",
		MaxAge: -1,
		HttpOnly: true,
		Secure: false,
		SameSite: http.SameSiteLaxMode,
	})
}

// getHeaderUser tries to get the user's username from
// the Authorization Header. This is used for cli apps.
func getHeaderUser(r *http.Request) (string, error) {
	authz := r.Header.Get("Authorization")
	if authz == "" {
		return "", fmt.Errorf("No user found in request")
	}
	if strings.HasPrefix(authz, "Bearer") {
		authResp := start.AuthProvider.ValidateToken(strings.TrimPrefix(authz, "Bearer "))
		if authResp.Error == nil {
			return authResp.User.Username, nil
		}
	}
	return "", fmt.Errorf("Could not ParseJWT")
}

// getCookieUser is used to get the user's username from
// browser cookies. This is used for browser access.
func getCookieUser(r *http.Request) (string, error) {
	cookie, err := r.Cookie(cookieName)
	if err != nil {
		return "", err 
	}
	authResp := start.AuthProvider.ValidateToken(cookie.Value)
	if authResp.Error != nil {
		return "", authResp.Error
	}
	return authResp.User.Username, nil
}

// GetUserFromRequest grabs the user from the getCookieUser
// and getHeaderUser methods. It is used by RequireAuth method.
func GetUserFromRequest(r *http.Request) (string, error) {
	resCookie, err := getCookieUser(r)
	if err == nil {
		return resCookie, nil
	}
	resHeader, err:= getHeaderUser(r)
	if err == nil {
		return resHeader, nil
	}
	return "", err
}

// RequireAuth is a method that grabs the user's username
// from GetUserFromRequest and generates Context Values that
// get passed to handler functions. Handlers use this info to 
// access user values.
func RequireAuth(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		//Get the User-Agent
		isItReally := isTerminal(r.Header.Get("User-Agent"))

		user, err := GetUserFromRequest(r)
		if err != nil || user == "" {
			http.Redirect(w, r, "/login", http.StatusFound)
			return
		}
		//Get user space to allow or not uploads
		space, err := start.Database.GetUserSpace(r.Context(),user)
		if err != nil {
			log.Printf("Error getting user space: %v", err)
		}
		ctx := context.WithValue(r.Context(), "username", user)
		ctx = context.WithValue(ctx, "isTerminal", isItReally)
		ctx = context.WithValue(ctx, "space", space)
		next(w, r.WithContext(ctx))
	}
}

// GetIsTerminalFromContext returns the context values for
// isTerminal
func GetIsTerminalFromContext(r *http.Request) bool {
	if v := r.Context().Value("isTerminal"); v != nil {
		if isIt, ok := v.(bool); ok {
			return isIt
		}
	}
	return false
}

// GetUsernameFromContext returns the context values for
// user's username
func GetUsernameFromContext(r *http.Request) string {
	if v := r.Context().Value("username"); v != nil {
		if username, ok := v.(string); ok {
			return username
		}
	}
	return ""
}

// GetUserUsedSpace returns the context values for
// user's used space
func GetUserUsedSpace(r *http.Request) int64 {
	if v := r.Context().Value("space"); v != nil {
		if username, ok := v.(int64); ok {
			return username
		}
	}
	return 0
}

