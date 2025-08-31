package handlers

import (
	"net/http"
	"time"

	"github.com/lamprosfasoulas/transfer/internal/auth"
	"github.com/lamprosfasoulas/transfer/internal/logger"
	"github.com/lamprosfasoulas/transfer/internal/sse"
	"github.com/lamprosfasoulas/transfer/internal/storage"
)

const cookieName = "auth-token"

type MainHandlers struct {
	Storage    storage.Storage
	Dispatcher sse.Dispatcher
	MAX_SPACE  int64
	Domain     string
	Logger     *logger.Logger
}

func NewMainHandler(s storage.Storage, di sse.Dispatcher, n int64, u string, l *logger.Logger) *MainHandlers {
	return &MainHandlers{
		Storage:    s,
		Dispatcher: di,
		MAX_SPACE:  n,
		Domain:     u,
		Logger:     l,
	}
}

type AuthHandler struct {
	AuthProvider auth.AuthProvider
	Logger       *logger.Logger
}

func NewAuthHandler(a auth.AuthProvider, l *logger.Logger) *AuthHandler {
	return &AuthHandler{
		AuthProvider: a,
		Logger:       l,
	}
}

// SetAuthCookie is used to initialize the session for the user.
// It stores a cookie named const cookieName with the value of
// the generated JWT token.
func (a *AuthHandler) SetAuthCookie(w http.ResponseWriter, value string, expiry time.Duration) {
	http.SetCookie(w, &http.Cookie{
		Name:     cookieName,
		Value:    value,
		Path:     "/",
		MaxAge:   int(expiry.Seconds()),
		HttpOnly: true,
		Secure:   true,
		SameSite: http.SameSiteLaxMode,
	})
}

// ClearAuthCookie terminates the user  session by deleting the cookie.
// It does not revoke the JWT token (i don't think it can be done),
// but logs out the browser client.
// Not used for cli sessions.
func (a *AuthHandler) ClearAuthCookie(w http.ResponseWriter) {
	http.SetCookie(w, &http.Cookie{
		Name:     cookieName,
		Value:    "",
		Path:     "",
		MaxAge:   -1,
		HttpOnly: true,
		Secure:   true,
		SameSite: http.SameSiteLaxMode,
	})
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
func GetUsedSpaceFromContext(r *http.Request) int64 {
	if v := r.Context().Value("space"); v != nil {
		if username, ok := v.(int64); ok {
			return username
		}
	}
	return 0
}
