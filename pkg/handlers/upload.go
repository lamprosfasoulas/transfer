package handlers

import (
	"archive/zip"
	"context"
	"fmt"
	"io"
	"log"
	"net/http"
	"path/filepath"
	"time"

	"github.com/google/uuid"
	"github.com/lamprosfasoulas/transfer/pkg/database"
	m "github.com/lamprosfasoulas/transfer/pkg/middleware"
	"github.com/lamprosfasoulas/transfer/pkg/sse"
	"github.com/lamprosfasoulas/transfer/pkg/start"
	"github.com/lamprosfasoulas/transfer/pkg/storage"
)

//Handler for Uploading files
func HandleUpload(w http.ResponseWriter, r *http.Request) {
	username := m.GetUsernameFromContext(r)
	if username == "" {
		http.Redirect(w, r, "/login", http.StatusFound)
		return
	}
	switch r.Method {
	case http.MethodGet:
		data := struct {UploadID string}{UploadID: "hi"}//uuid.NewString()}
		if !m.GetIsTerminalFromContext(r) {
			UploadTmpl.Execute(w, data)
		} else {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		}
	case http.MethodPost:
		ctx, cancel:= context.WithCancel(r.Context())
		defer cancel()

		//Use this to somehow validate sse access
		//fmt.Println("From post upload",r.URL.Query().Get("uploadID"))

		//sse.SendEventToSubscriber("hi", &sse.ProgressEvent{
		//	Filename: "hi",
		//	Bytes:      0,
		//	TotalBytes: 0,
		//	Percentage: 0,
		//	Message:    "Start",
		//})


		err := r.ParseMultipartForm(50 << 20)
		if err != nil {
			log.Println("Error parsing form: ", err)
			return
		}

		if r.ContentLength > (start.MAX_SPACE - m.GetUserUsedSpace(r)) {
			//Delete the object and return no more space
			log.SetPrefix(fmt.Sprintf("[\033[34mUPLOAD INFO\033[0m] "))
			log.Printf("User %s has no more space\n", username)
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
			log.SetPrefix(fmt.Sprintf("[\033[31mUPLOAD ERR\033[0m] "))
			http.Error(w, "You sent no files", http.StatusBadRequest)
			return
		}
		if len(files) > 1 {
			pr, pw := io.Pipe()
			zw := zip.NewWriter(pw)
			fileName = fmt.Sprintf("archive-%d.zip", time.Now().Unix())
			ext = filepath.Ext(fileName)

			go func(){
				log.SetPrefix(fmt.Sprintf("[\033[31mUPLOAD ERR\033[0m] "))
				defer pw.Close()
				defer zw.Close()
				for _, fh := range files {
					file, err := fh.Open()
					if err != nil {
						log.Println("Error opening file: ", err)
						return
					}

					w,err := zw.Create(fh.Filename)
					if err != nil {
						log.Println("Error creating zip entry: ", err)
						return
					}
					if _, err := io.Copy(w, file); err != nil {
						log.Println("Error writing to zip: ", err)
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
				log.SetPrefix(fmt.Sprintf("[\033[31mUPLOAD ERR\033[0m] "))
				log.Println("Failed to read file from form")
				return
			}
			defer file.Close()
			reader = file
		}

		objID := uuid.New().String() + ext
		objectKey := username + "/" + objID
		expireAt := time.Now().Add(7 * 24 * time.Hour)

		prd := storage.NewProgressReader(reader, totalSize, fileName, "hi")//r.URL.Query().Get("uploadID"))
		//fgmt.Println("uploadid",prd.UploadID)

		//uploadInfo, err := start.Storage.MinioClient.PutObject(ctx, start.Cfg.MinioBucket, objectKey, prd, prd.Total, minio.PutObjectOptions{
		//	//ContentType: prd.filename.Header.Get("Content-Type"),
		//	UserMetadata: map[string]string {
		//		"filename": prd.Filename,
		//	},
		//})
		uploadInfo := start.Storage.PutObject(ctx, objectKey, prd) 
		if uploadInfo.Error != nil {
			log.SetPrefix(fmt.Sprintf("[\033[31mUPLOAD ERR\033[0m] "))
			log.Printf("Failed to upload to MinIO: %v", uploadInfo.Error)
			http.Error(w, fmt.Sprintf("Failed to upload to MinIO: %v",uploadInfo.Error), http.StatusInternalServerError)
			return
		}

		//Upload file info to db
		err = start.Database.PutFile(ctx, database.PutFileParams{
			Ownerid: username,
			Objkey: uploadInfo.Key,
			Filename: uploadInfo.Filename,
			ID: objID,
			Size: uploadInfo.Size,
			Expiresat: &expireAt,
		})
		if err != nil {
			log.SetPrefix(fmt.Sprintf("[\033[31mUPLOAD ERR\033[0m] "))
			log.Printf("Upload to database failed: %v\n", err)
			//Those should go in a func together
			http.Error(w, fmt.Sprintf("Failed to upload to Database: %v",err), http.StatusInternalServerError)
		}

		//Update user's used space
		err = start.Database.RecalculateUserSpace(ctx, username)
		if err != nil {
			log.SetPrefix(fmt.Sprintf("[\033[31mUPLOAD ERR\033[0m] "))
			log.Printf("Failed to recalculate user space: %v\n", err)
			//Those should go in a func together
			//http.Error(w, fmt.Sprintf("Failed to upload to Database: %v",err), http.StatusInternalServerError)
		}

		log.SetPrefix(fmt.Sprintf("[\033[34mUPLOAD INFO\033[0m] "))
		log.Printf("Uploaded: %s (%d bytes)\n", uploadInfo.Filename, uploadInfo.Size)
		sse.SendEventToSubscriber(prd.UploadID, &sse.ProgressEvent{
			Filename: fileName,
			Bytes:      uploadInfo.Size,
			TotalBytes: uploadInfo.Size,
			Percentage: 100,
			Message:    "Upload complete",
		})

		downloadLink := fmt.Sprintf("%s/download/%s/%s", r.Host, username ,objID)
		log.Printf("Download link is: %v", downloadLink)
		data := struct {
		DownloadLink 		string
		DirectPresigned 	string
		ExpiresInSeconds 	int64
		}{
			DownloadLink: downloadLink,
			DirectPresigned: "",
			ExpiresInSeconds: 4,
		}
		if m.GetIsTerminalFromContext(r) {
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
