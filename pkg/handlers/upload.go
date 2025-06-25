package handlers

import (
	"archive/zip"
	"fmt"
	"io"
	"net/http"
	"path/filepath"
	"regexp"
	"time"

	"github.com/google/uuid"
	"github.com/lamprosfasoulas/transfer/pkg/database"
	"github.com/lamprosfasoulas/transfer/pkg/logger"
	"github.com/lamprosfasoulas/transfer/pkg/sse"
	"github.com/lamprosfasoulas/transfer/pkg/storage"
)

//Handler for Uploading files
func (m *MainHandlers) Upload(w http.ResponseWriter, r *http.Request) {
	username := GetUsernameFromContext(r)
	if username == "" {
		http.Redirect(w, r, "/login", http.StatusFound)
		return
	}
	ctx := r.Context()

	if r.ContentLength > (m.MAX_SPACE - GetUserUsedSpace(r)) {
		// Abort and say no more space
		m.Logger.Warn(logger.Upl).Write(fmt.Sprintf("User %s has no more space", username))
		http.Error(w, "No more space", http.StatusBadRequest)
		return
	}

	var fileName 	string
	var ext 		string
	var ch = make(chan string)

	mr, err := r.MultipartReader()
	pr, pw := io.Pipe()
	zw := zip.NewWriter(pw)

	go func(){
		defer pw.Close()
		defer zw.Close()
		for {
			part, err :=  mr.NextPart()
			if err == io.EOF {
				break
			}
			if err != nil {
				m.Logger.Error(logger.Upl).Writef("Error opening file", err)
				return
			}
			if part.FormName() == "filename" {
				tmpname := make([]byte, 32)
				n, err := part.Read(tmpname)

				if err != nil && err != io.EOF {
					tmpname = nil
					continue
				}
				if n == 0 {
					ch <- fmt.Sprintf("archive-%d.zip",time.Now().Unix())
				} else {
					ch <- fmt.Sprintf("%s.zip",sanitizeFilename(tmpname[:n]))
				}
				continue
			}

			w, err := zw.Create(part.FileName())
			if err != nil {
				m.Logger.Error(logger.Upl).Writef("Error creating zip entry", err)
				return
			}
			if _, err := io.Copy(w, part); err != nil {
				m.Logger.Error(logger.Upl).Writef("Error writing to zip", err)
				return
			}
		}
	}()

	fileName = <-ch //fmt.Sprintf("arcive-%d.zip", time.Now().Unix())
	ext = filepath.Ext(fileName)

	objID := uuid.New().String() + ext
	objectKey := username + "/" + objID
	expireAt := time.Now().Add(7 * 24 * time.Hour)
	uploadID := r.URL.Query().Get("id")

	prd := storage.NewProgressReader(pr, -1, fileName, uploadID, m.Dispatcher)

	uploadInfo, err := m.Storage.PutObject(ctx, objectKey, prd) 
	if err != nil {
		m.Logger.Error(logger.Upl).Writef("Failed to save to storge",err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	//Upload file info to db
	err = m.Database.PutFile(ctx, database.PutFileParams{
		Ownerid: username,
		Objkey: uploadInfo.Key,
		Filename: uploadInfo.Filename,
		ID: objID,
		Size: uploadInfo.Size,
		Expiresat: &expireAt,
	})
	if err != nil {
		m.Logger.Error(logger.Upl).Writef("Failed to upload metadata to database",err)
		//Those should go in a func together
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
	}

	//Update user's used space
	err = m.Database.RecalculateUserSpace(ctx, username)
	if err != nil {
		//Those should go in a func together
		m.Logger.Error(logger.Upl).Writef("Failed to Recalculate User Space",err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
	}

	m.Logger.Info(logger.Upl).Write(fmt.Sprintf("Uploaded: %s (%d bytes)\n", uploadInfo.Filename, uploadInfo.Size))
	m.Dispatcher.SendEvent(r.Context(),prd.UploadID, &sse.ProgressEvent{
		Filename: fileName,
		Bytes:      uploadInfo.Size,
		TotalBytes: uploadInfo.Size,
		Percentage: 100,
		Message:    "Upload complete",
	})

	downloadLink := fmt.Sprintf("%s/download/%s/%s", r.Host, username ,objID)
	//log.Printf("Download link is: %v", downloadLink)
	data := struct {
		DownloadLink 		string
		DirectPresigned 	string
		ExpiresInSeconds 	int64
	}{
		DownloadLink: downloadLink,
		DirectPresigned: "",
		ExpiresInSeconds: 4,
	}
	if GetIsTerminalFromContext(r) {
		ResultTmplTerm.Execute(w, data)
		return
	} else {
		fmt.Fprintf(w, "Successful upload: %v", data.DownloadLink)
		return
	}
}

func sanitizeFilename(b []byte) string {
	name := filepath.Base(string(b))
	re := regexp.MustCompile(`[^a-zA-Z0-9._-]`)
	name = re.ReplaceAllString(name, "_")
	//if len(name) > 15 {
	//	name = name[:15]
	//}
	return name
}
