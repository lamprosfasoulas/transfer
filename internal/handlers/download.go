package handlers

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"

	"github.com/lamprosfasoulas/transfer/internal/logger"
)

// HTTP Handler for downloading files
func (m *MainHandlers) Download(w http.ResponseWriter, r *http.Request) {
	path := strings.TrimPrefix(r.URL.Path, "/download/")
	parts := strings.Split(path, "/")
	if len(parts) != 2 {
		http.Error(w, "Bad URL", http.StatusBadRequest)
		return
	}
	username := parts[0]
	fileID := parts[1]

	if fileID == "" || fileID == "download" || username == "" || username == "download" {
		http.Error(w, "Bad URL", http.StatusBadRequest)
		return
	}
	objectKey := username + "/" + fileID

	ctx, cancel := context.WithCancel(r.Context())
	defer cancel()

	info, err := m.Storage.GetObject(ctx, objectKey)
	if err != nil {
		m.Logger.Error(logger.Storage).Writef("Error getting object", err)
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
		m.Logger.Error(logger.Download).Writef(fmt.Sprintf("Error streaming info %q:", objectKey), err)
	}
}
