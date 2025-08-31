package handlers

import (
	"encoding/json"
	"net/http"
)

type JSONError struct {
	Error string `json:"error"`
}

func jsonResponder(w http.ResponseWriter, data any, code int) {
	h := w.Header()

	h.Del("Content-Length")

	h.Set("Content-Type", "application/json")
	h.Set("X-Content-Type-Options", "nosniff")
	w.WriteHeader(code)

	if err := json.NewEncoder(w).Encode(data); err != nil {
		http.Error(w, `{"error":"internal server error"}`, http.StatusInternalServerError)
	}
}

func jsonErrResponder(w http.ResponseWriter, msg string, code int) {
	jsonResponder(w, JSONError{Error: msg}, code)
}
