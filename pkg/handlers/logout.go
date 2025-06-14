package handlers

import (
	"net/http"

	m "github.com/lamprosfasoulas/transfer/pkg/middleware"
)

//Handler for loging out user
func HandleLogout(w http.ResponseWriter, r *http.Request) {
	m.ClearAuthCookie(w)
	http.Redirect(w, r, "/login", http.StatusFound)
}
