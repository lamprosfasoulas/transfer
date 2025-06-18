package handlers

import (
	"fmt"
	"net/http"
	"path"

	"github.com/lamprosfasoulas/transfer/pkg/logger"
)

//Handler for Deleting files
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
	//if err := start.Storage.MinioClient.RemoveObject(ctx, start.Cfg.MinioBucket, objectKey, minio.RemoveObjectOptions{}); err != nil {
	//	log.Printf("Error deleting %q: %v\n", objectKey, err)
	//}

	_, err := m.Storage.DeleteObject(ctx, objectKey)
	if err!= nil {
		//log.SetPrefix(fmt.Sprintf("[\033[31mDELETE ERR\033[0m] "))
		//log.Printf("Error deleting object: %v", info.Message)
		m.Logger.Error(logger.Sto).Writef("Erro deleting object: %v", err)
		//http.Error(w, info.Message, http.StatusBadRequest)
		// We still delete from db
		//return
	}

	err = m.Database.DeleteFile(ctx, objectKey)
	if err != nil {
		//log.SetPrefix(fmt.Sprintf("[\033[31mDELETE ERR\033[0m] "))
		//log.Printf("Error deleting file from database: %v", err)
		m.Logger.Error(logger.Del).Writef("Error deleting file from database", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}
	err = m.Database.RecalculateUserSpace(ctx, username)
	if err != nil {
		//log.SetPrefix(fmt.Sprintf("[\033[31mDELETE ERR\033[0m] "))
		//log.Printf("Error recalculating user used space: %v", err)
		m.Logger.Error(logger.Del).Writef("Error recalculating user used space", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}


	//log.SetPrefix(fmt.Sprintf("[\033[34mDELETE INFO\033[0m] "))
	//log.Printf("File %s deleted successfully\n",objectKey)
	m.Logger.Info(logger.Del).Write(fmt.Sprintf("File %s deleted successfully\n", objectKey))
	http.Redirect(w, r, "/", http.StatusFound)
}
