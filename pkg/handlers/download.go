package handlers

import (
	"context"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/url"
	"path"

	m "github.com/lamprosfasoulas/transfer/pkg/middleware"
	"github.com/lamprosfasoulas/transfer/pkg/start"
)

//HTTP Handler for downloading files
func HandleDownload(w http.ResponseWriter, r *http.Request) {
	username := path.Base(path.Dir(r.URL.Path))
	if username == "" {
		http.Redirect(w, r, "/login", http.StatusFound)
		return
	}
	if username == "download" {
		user, err := m.GetUserFromRequest(r)
		if err != nil {
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

	info := start.Storage.GetObject(ctx, objectKey)
	defer info.Object.Close()

	if info.Error != nil {
		log.SetPrefix(fmt.Sprintf("[\033[31mERR\033[0m] "))
		log.Println(info.Message)
		http.Error(w, info.Message, http.StatusInternalServerError)
		return
	}
	//object, err := start.Storage.MinioClient.GetObject(ctx, start.Cfg.MinioBucket, objectKey, minio.GetObjectOptions{})
	//if err != nil {
	//	w.WriteHeader(http.StatusInternalServerError)
	//	json.NewEncoder(w).Encode(map[string]string{"error": "could not fetch object"})
	//	return
	//}
	//defer object.Close()

	//stat, err := object.Stat()
	//if err != nil {
	//	w.WriteHeader(http.StatusNotFound)
	//	json.NewEncoder(w).Encode(map[string]string{"error": "file not found"})
	//	return
	//}


	//filename := fileID
	//if meta := stat.Metadata["X-Amz-Meta-Filename"]; meta != nil {
	//	filename = meta[0]
	//}
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
	//ctx := context.Background()
	//presignedURL, err := start.MinioClient.PresignedGetObject(ctx, start.Cfg.MinioBucket, objectKey, 1 * time.Minute, reqParam)
	//if err != nil {
	//	http.Error(w, "Failed to generate download link", http.StatusInternalServerError)
	//	return
	//}
	//http.Redirect(w, r, presignedURL.String(), http.StatusFound)
}
