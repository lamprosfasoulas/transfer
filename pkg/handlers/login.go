package handlers

import (
	"fmt"
	"log"
	"net/http"
	"time"

	m "github.com/lamprosfasoulas/transfer/pkg/middleware"
	"github.com/lamprosfasoulas/transfer/pkg/start"
)


func HandleLoginGet(w http.ResponseWriter, r *http.Request) {
	username, _ := m.GetUserFromRequest(r)
	if username != "" {
		http.Redirect(w, r, "/", http.StatusFound)
	}

	//Response
	if m.GetIsTerminalFromContext(r) {//.Context().Value("isTerminal").(bool) {
		fmt.Fprintln(w, "Method Not Allowed")
	} else {
		LoginTmpl.Execute(w, nil)
	}
}

func HandleLoginPost(w http.ResponseWriter, r *http.Request) {
	log.SetPrefix(fmt.Sprintf("[\033[31mERR\033[0m] "))
	var username string
	//First look is user is signed in
	username, err:= m.GetUserFromRequest(r)
	if username != "" && err == nil {
		http.Redirect(w, r, "/", http.StatusFound)
		return
	}
	if start.AuthProvider == nil {
		log.Fatalln("fuck you")
	}
	resp := start.AuthProvider.Authenticate(r)
	if resp.Error != nil {
		log.Printf("Authentication error: %v\n", resp.Error)
		http.Error(w, "Internal error during auth", http.StatusInternalServerError)
		return
	}

	if !resp.Success{
		//Redirect to login with error message
		http.Error(w, "Invalid username or password", http.StatusInternalServerError)
		return
	}
	resp = start.AuthProvider.GenerateToken(resp)
	start.Database.PutUser(r.Context(), resp.User.Username)

	////If user is not signed in log him in
	//if err := r.ParseForm(); err != nil {
	//	http.Error(w, "Invalid form data", http.StatusBadRequest)
	//	return
	//}

	//username = r.FormValue("username")
	//password = r.FormValue("password")
	//if username == "" || password == "" {
	//	http.Error(w, "Please provide credentials", http.StatusBadRequest)
	//	return
	//}

	//// Here we get our session credentials
	//ok, err, tokenString := auth.Authenticate(username, password);
	//log.SetPrefix(fmt.Sprintf("[\033[34mINFO\033[0m] "))
	//log.Printf("New login: %v\n", *r)

	// Response
	if m.GetIsTerminalFromContext(r) {//.Context().Value("isTerminal").(bool) {
		w.Header().Set("Content-Type", "application/json")
		//json.NewEncoder(w).Encode(map[string]string{"token": tokenString})
		return
	} else {
		//w.Header().Set("Authorization", fmt.Sprintf("Bearer %s",resp.JwtToken))
		m.SetAuthCookie(w, resp.JwtToken, 1 * time.Hour)
		http.Redirect(w, r, "/", http.StatusFound)
	}
}

//func HandleLoginOIDC(w http.ResponseWriter, r *http.Request) {
//	if r.URL.Query().Get("state") != start.Cfg.OIDCState {
//		http.Error(w, "State mismatch", http.StatusBadRequest)
//	}
//
//}

