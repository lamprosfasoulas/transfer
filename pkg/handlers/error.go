package handlers

import "net/http"

func HandleError(w http.ResponseWriter, r *http.Request) {
	ErrorTmpl.Execute(w,nil)
}
