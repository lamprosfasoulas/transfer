package handlers

import (
	"encoding/json"
	"net/http"
	"time"
)


func (a *AuthHandler) Logout(w http.ResponseWriter, r *http.Request) {
	a.ClearAuthCookie(w)
	http.Redirect(w, r, "/login", http.StatusFound)
}

func (a *AuthHandler) LoginGet(w http.ResponseWriter, r *http.Request) {
	username := GetUsernameFromContext(r)
	if username != "" {
		http.Redirect(w, r, "/", http.StatusFound)
		return
	}
	if GetIsTerminalFromContext(r) {
		w.Write([]byte("Login with POST\n"))
	} else {
		LoginTmpl.Execute(w, nil)
	}
}


func (a *AuthHandler) Login(w http.ResponseWriter, r *http.Request) {
	//First look is user is signed in
	username := GetUsernameFromContext(r) //m.GetUserFromRequest(r)
	if username != "" {
		http.Redirect(w, r, "/", http.StatusFound)
		return
	}
	if a.AuthProvider == nil {
		a.Logger.Info().Write("The auth provider is not there")
	}
	resp := a.AuthProvider.Authenticate(r)
	if resp.Error != nil {
		a.Logger.Warn().Writef("Authentication error", resp.Error)
		http.Error(w, "Internal error during auth", http.StatusInternalServerError)
		return
	}

	if !resp.Success{
		//Redirect to login with error message
		http.Error(w, "Invalid username or password", http.StatusInternalServerError)
		return
	}
	resp = a.AuthProvider.GenerateToken(resp)
	a.Database.PutUser(r.Context(), resp.User.Username)

	// Response
	if GetIsTerminalFromContext(r) {//.Context().Value("isTerminal").(bool) {
		//fmt.Println("is terminal")
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]string{"token": resp.JwtToken})
		return
	} else {
		//w.Header().Set("Authorization", fmt.Sprintf("Bearer %s",resp.JwtToken))
		a.SetAuthCookie(w, resp.JwtToken, 1 * time.Hour)
		http.Redirect(w, r, "/", http.StatusFound)
	}
}

