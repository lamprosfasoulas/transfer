package handlers

import (
	"archive/zip"
	"fmt"
	"io"
	"net/http"
	"path/filepath"
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
	switch r.Method {
	//case http.MethodGet:
	//	data := struct {UploadID string}{UploadID: "hi"}//uuid.NewString()}
	//	if !m.GetIsTerminalFromContext(r) {
	//		UploadTmpl.Execute(w, data)
	//	} else {
	//		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
	//	}
	case http.MethodPost:
		ctx := r.Context()

		err := r.ParseMultipartForm(50 << 20)
		if err != nil {
			//log.Println("Error parsing form: ", err)
			m.Logger.Error(logger.Upl).Writef("Error parsing form:", err)
			return
		}

		if r.ContentLength > (m.MAX_SPACE - GetUserUsedSpace(r)) {
			//Delete the object and return no more space
			//log.SetPrefix(fmt.Sprintf("[\033[34mUPLOAD INFO\033[0m] "))
			//log.Printf("User %s has no more space\n", username)
			m.Logger.Warn(logger.Upl).Write(fmt.Sprintf("User %s has no more space", username))
			http.Error(w, "No more space", http.StatusBadRequest)
			return
		}

		files := r.MultipartForm.File["file"]

		//http.Redirect(w, r,"/status?id=hi",http.StatusFound)
		var reader 		io.Reader
		var totalSize 	int64
		var fileName 	string
		var ext 		string

		if len(files) == 0 {
			//log.SetPrefix(fmt.Sprintf("[\033[31mUPLOAD ERR\033[0m] "))
			m.Logger.Warn(logger.Upl).Write("No files sent")
			http.Error(w, "You sent no files", http.StatusBadRequest)
			return
		}
		if len(files) > 1 {
			pr, pw := io.Pipe()
			zw := zip.NewWriter(pw)
			fileName = fmt.Sprintf("archive-%d.zip", time.Now().Unix())
			ext = filepath.Ext(fileName)

			go func(){
				//log.SetPrefix(fmt.Sprintf("[\033[31mUPLOAD ERR\033[0m] "))
				defer pw.Close()
				defer zw.Close()
				for _, fh := range files {
					file, err := fh.Open()
					if err != nil {
						//log.Println("Error opening file: ", err)
						m.Logger.Error(logger.Upl).Writef("Error opening file", err)
						return
					}

					w,err := zw.Create(fh.Filename)
					if err != nil {
						//log.Println("Error creating zip entry: ", err)
						m.Logger.Error(logger.Upl).Writef("Error creating zip entry", err)
						return
					}
					if _, err := io.Copy(w, file); err != nil {
						//log.Println("Error writing to zip: ", err)
						m.Logger.Error(logger.Upl).Writef("Error writing to zip", err)
						return
					}
					file.Close()
				}
			}()
			totalSize = -1
			reader = pr

		} else {
			fh := files[0]
			file, err := fh.Open()
			totalSize = fh.Size
			fileName = fh.Filename
			ext = filepath.Ext(fh.Filename)
			if err != nil {
				//http.Error(w, "Failed to read file from form", http.StatusBadRequest)
				//log.SetPrefix(fmt.Sprintf("[\033[31mUPLOAD ERR\033[0m] "))
				//log.Println("Failed to read file from form")
				m.Logger.Error(logger.Upl).Writef("Failed to read file from form", err)
				return
			}
			defer file.Close()
			reader = file
		}

		objID := uuid.New().String() + ext
		objectKey := username + "/" + objID
		expireAt := time.Now().Add(7 * 24 * time.Hour)
		uploadID := r.URL.Query().Get("id")

		prd := storage.NewProgressReader(reader, totalSize, fileName, uploadID, m.Dispatcher)

		uploadInfo, err := m.Storage.PutObject(ctx, objectKey, prd) 
		if err != nil {
			//log.SetPrefix(fmt.Sprintf("[\033[31mUPLOAD ERR\033[0m] "))
			//log.Printf("Failed to upload to MinIO: %v", uploadInfo.Error)
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
			//log.SetPrefix(fmt.Sprintf("[\033[31mUPLOAD ERR\033[0m] "))
			//log.Printf("Upload to database failed: %v\n", err)
			m.Logger.Error(logger.Upl).Writef("Failed to upload metadata to database",err)
			//Those should go in a func together
			http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		}

		//Update user's used space
		err = m.Database.RecalculateUserSpace(ctx, username)
		if err != nil {
			//log.SetPrefix(fmt.Sprintf("[\033[31mUPLOAD ERR\033[0m] "))
			//log.Printf("Failed to recalculate user space: %v\n", err)
			//Those should go in a func together
			//http.Error(w, fmt.Sprintf("Failed to upload to Database: %v",err), http.StatusInternalServerError)
			m.Logger.Error(logger.Upl).Writef("Failed to Recalculate User Space",err)
		}

		//log.SetPrefix(fmt.Sprintf("[\033[34mUPLOAD INFO\033[0m] "))
		//log.Printf("Uploaded: %s (%d bytes)\n", uploadInfo.Filename, uploadInfo.Size)
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
	
	default:
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
	}
}
