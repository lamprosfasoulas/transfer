package handlers

import (
	"archive/zip"
	"context"
	"fmt"
	"io"
	"net/http"
	"path/filepath"
	"regexp"
	"time"

	"github.com/google/uuid"
	"github.com/lamprosfasoulas/transfer/internal/logger"
	"github.com/lamprosfasoulas/transfer/internal/sse"
	"github.com/lamprosfasoulas/transfer/internal/storage"
)

// Handler for Uploading files
func (m *MainHandlers) Upload(w http.ResponseWriter, r *http.Request) {
	username := GetUsernameFromContext(r)
	if username == "" {
		http.Redirect(w, r, "/login", http.StatusFound)
		return
	}
	ctx, cancel := context.WithCancel(r.Context())

	defer func() {
		io.Copy(io.Discard, r.Body)
		r.Body.Close()
		cancel()
	}()

	spaceUsed := GetUsedSpaceFromContext(r)
	if r.ContentLength > (m.MAX_SPACE - spaceUsed) {
		// Abort and say no more space
		m.Logger.Warn(logger.Upload).Write(fmt.Sprintf("User %s has no more space", username))
		// If we don't discard the request body we trigger NS_ERROR_NET_RESET
		// since the body was not read.
		io.Copy(io.Discard, r.Body)
		r.Body.Close()

		jsonErrResponder(w, "no more space", http.StatusBadRequest)
		//http.Error(w, "No more space", http.StatusBadRequest)
		return
	}

	var fileName string
	var ext string
	var ch = make(chan string)

	mr, err := r.MultipartReader()
	pr, pw := io.Pipe()
	zw := zip.NewWriter(pw)

	go func() {
		defer pw.Close()
		defer zw.Close()
		for {
			select {
			case <- ctx.Done():
				return
			default:
			}

			part, err := mr.NextPart()
			if err == io.EOF {
				break
			}
			if err != nil {
				m.Logger.Error(logger.Upload).Writef("Error opening file", err)
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
					ch <- fmt.Sprintf("archive-%d.zip", time.Now().Unix())
				} else {
					ch <- fmt.Sprintf("%s.zip", sanitizeFilename(tmpname[:n]))
				}
				continue
			}


			w, err := zw.Create(part.FileName())
			if err != nil {
				m.Logger.Error(logger.Upload).Writef("Error creating zip entry", err)
				return
			}
			
			if _, err := io.Copy(w, part); err != nil {
				m.Logger.Error(logger.Upload).Writef("Error writing to zip", err)
				return
			}
		}
	}()

	fileName = <-ch //fmt.Sprintf("arcive-%d.zip", time.Now().Unix())
	ext = filepath.Ext(fileName)

	objID := uuid.New().String() + ext
	objectKey := username + "/" + objID
	//expireAt := time.Now().Add(7 * 24 * time.Hour)
	uploadID := r.URL.Query().Get("id")

	// Here we set the total bytes equal to r.ContentLength to track the upload
	// progress (estimate). When the PutObject function is called the size is
	// set to -1. See pkg/storage/minio.go
	prd := storage.NewProgressReader(pr, r.ContentLength, fileName, uploadID, m.Dispatcher)

	uploadInfo, err := m.Storage.PutObject(ctx, objectKey, prd)
	if err != nil {
		errMin := storage.ToStorageError(err)
		switch errMin.StatusCode {
		case 0: break
		case 413:
			jsonErrResponder(w, "file too big", http.StatusRequestEntityTooLarge)
		default:
			jsonErrResponder(w, "Internal Server Error", http.StatusInternalServerError)
			m.Logger.Error(logger.Upload).Writef("Failed to save to storage", err)
		}
		return
	}

	m.Logger.Info(logger.Upload).Write(fmt.Sprintf("Uploaded: %s (%d bytes)\n", uploadInfo.Filename, uploadInfo.Size))
	m.Dispatcher.SendEvent(r.Context(), prd.UploadID, &sse.ProgressEvent{
		Filename:   fileName,
		Bytes:      uploadInfo.Size,
		TotalBytes: uploadInfo.Size,
		Percentage: 100,
		Message:    "Upload complete",
	})

	downloadLink := fmt.Sprintf("%s/download/%s/%s", r.Host, username, objID)
	//log.Printf("Download link is: %v", downloadLink)
	data := struct {
		DownloadLink     string
		DirectPresigned  string
		ExpiresInSeconds int64
	}{
		DownloadLink:     downloadLink,
		DirectPresigned:  "",
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
