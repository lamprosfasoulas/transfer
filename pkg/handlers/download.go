package handlers

import (
	"context"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/url"
	"path"

	"github.com/lamprosfasoulas/transfer/pkg/logger"
)

//HTTP Handler for downloading files
func (m *MainHandlers) Download(w http.ResponseWriter, r *http.Request) {
	username := path.Base(path.Dir(r.URL.Path))
	if username == "" {
		http.Redirect(w, r, "/login", http.StatusFound)
		return
	}
	if username == "download" {
		user := GetUsernameFromContext(r)
		if user == "" {
			http.Error(w, "Please login", http.StatusUnauthorized)
			return
		}
		username = user
	}

	fileID := path.Base(r.URL.Path)
	if fileID == "" || fileID == "download" {
		http.NotFound(w, r)
		return
	}
	objectKey := username + "/" + fileID

	ctx, cancel:= context.WithCancel(r.Context())
	defer cancel()

	info, err := m.Storage.GetObject(ctx, objectKey)
	if err != nil {
		m.Logger.Error(logger.Sto).Writef("Error getting object", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}
	defer info.Object.Close()

	reqParam := make(url.Values)
	reqParam.Set("response-content-disposition", fmt.Sprintf(`attachment; filename="%s"`, info.Filename))
	reqParam.Set("response-content-type", info.Content)
	w.Header().Set("Content-Disposition", fmt.Sprintf("attachment; filename=\"%s\"", info.Filename))
	w.Header().Set("Content-Type", info.Content)
	w.Header().Set("Content-Length", fmt.Sprintf("%d", info.Size))

	if _, err := io.Copy(w, info.Object); err != nil {
		log.SetPrefix(fmt.Sprintf("[\033[31mERR\033[0m] "))
		log.Printf("Error streaming info %q: %v", objectKey, err)
	}
}
