package handlers

import (
	"fmt"
	"log"
	"net/http"
	"path"

	m "github.com/lamprosfasoulas/transfer/pkg/middleware"
	"github.com/lamprosfasoulas/transfer/pkg/start"
)

//Handler for Deleting files
func HandleDelete(w http.ResponseWriter, r *http.Request) {
	username := m.GetUsernameFromContext(r)
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
	//if err := start.Storage.MinioClient.RemoveObject(ctx, start.Cfg.MinioBucket, objectKey, minio.RemoveObjectOptions{}); err != nil {
	//	log.Printf("Error deleting %q: %v\n", objectKey, err)
	//}

	info := start.Storage.DeleteObject(ctx, objectKey)
	if info.Error != nil {
		log.SetPrefix(fmt.Sprintf("[\033[31mDELETE ERR\033[0m] "))
		log.Printf("Error deleting object: %v", info.Message)
		//http.Error(w, info.Message, http.StatusBadRequest)
		// We still delete from db
		//return
	}

	err := start.Database.DeleteFile(ctx, objectKey)
	if err != nil {
		log.SetPrefix(fmt.Sprintf("[\033[31mDELETE ERR\033[0m] "))
		log.Printf("Error deleting file from database: %v", err)
		http.Error(w, info.Message, http.StatusBadRequest)
		return
	}
	err = start.Database.RecalculateUserSpace(ctx, username)
	if err != nil {
		log.SetPrefix(fmt.Sprintf("[\033[31mDELETE ERR\033[0m] "))
		log.Printf("Error recalculating user used space: %v", err)
		http.Error(w, info.Message, http.StatusBadRequest)
		return
	}


	log.SetPrefix(fmt.Sprintf("[\033[34mDELETE INFO\033[0m] "))
	log.Printf("File %s deleted successfully\n",objectKey)
	http.Redirect(w, r, "/", http.StatusFound)
}
