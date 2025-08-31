package handlers

import (
	"fmt"
	"net/http"
	"path"

	"github.com/lamprosfasoulas/transfer/internal/logger"
)

// Handler for Deleting files
func (m *MainHandlers) Delete(w http.ResponseWriter, r *http.Request) {
	username := GetUsernameFromContext(r)
	if username == "" {
		http.Redirect(w, r, "/login", http.StatusFound)
		return
	}

	fileIDPath := path.Base(r.URL.Path)
	if err := r.ParseMultipartForm(10 << 20); err != nil {
		http.Error(w, "Invalid form data", http.StatusBadRequest)
		return
	}
	fileIDForm := r.FormValue("file")
	var fileID string
	switch {
	case fileIDPath == "" && fileIDForm == "":
		http.Redirect(w, r, "/", http.StatusFound)
		return
	case fileIDForm == "" && fileIDPath != "delete":
		fileID = fileIDPath
	case fileIDPath == "" || fileIDPath == "delete":
		fileID = fileIDForm
	default:
		http.Redirect(w, r, "/", http.StatusFound)
		return
	}

	objectKey := username + "/" + fileID
	ctx := r.Context()

	_, err := m.Storage.DeleteObject(ctx, objectKey)
	if err != nil {
		m.Logger.Error(logger.Storage).Writef("Erro deleting object", err)
		//http.Error(w, info.Message, http.StatusBadRequest)
		// We still delete from db
		//return
	}

	m.Logger.Info(logger.Delete).Write(fmt.Sprintf("File %s deleted successfully\n", objectKey))
	http.Redirect(w, r, "/", http.StatusFound)
}
