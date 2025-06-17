package storage

import (
	"context"
	"io"
	"net/http"
	"os"
	"path/filepath"
)

type Filesystem struct {
	UploadDir string
	Error error
}

func NewFilesystem (upload string) *Filesystem {
	return &Filesystem{
		UploadDir: upload,
		Error: nil,
	}
}
func (f *Filesystem) GetError() error {
	return f.Error
}

func (f *Filesystem) PutObject(c context.Context, key string, r *ProgressReader) *FileInfo {
	userDir := filepath.Dir(key)
	saveDir := filepath.Join(f.UploadDir, userDir)
	err := os.MkdirAll(saveDir, os.ModePerm)
	if err != nil {
		return &FileInfo{
			Error: err,
			Message: "Failed to create user directory.",
		}
	}
	filePath := filepath.Join(f.UploadDir, key)
	saveFile, err := os.Create(filePath)
	if err != nil {
		return &FileInfo{
			Error: err,
			Message: "Failed to create file.",
		}
	}
	defer saveFile.Close()

	_, err = io.Copy(saveFile, r)
	if err != nil {
		return &FileInfo{
			Error: err,
			Message: "Failed to stream file.",
		}
	}

	info, err := saveFile.Stat()
	if err != nil {
		return &FileInfo{
			Error: err,
			Message: "Failed to stat file.",
		}
	}

	return &FileInfo{
		Key: key,
		Filename: r.Filename,
		Size: info.Size(),
		Error: err,
	}
}

func (f *Filesystem) GetObject(c context.Context, key string) *FileInfo {
	userFile := filepath.Join(f.UploadDir, key)
	file, err := os.Open(userFile)
	if err != nil {
		return &FileInfo{
			Error: err,
			Message: "Failed to open file.",
		}
	}

	info, err := file.Stat()
	if err != nil {
		file.Close()
		return &FileInfo{
			Error: err,
			Message: "Failed to stat file.",
		}
	}

	buf := make([]byte, 512)
	_, err = file.Read(buf)
	if err != nil {
		file.Close()
		return &FileInfo{
			Error: err,
			Message: "Failed to read file.",
		}
	}
	content := http.DetectContentType(buf)

	if _, err := file.Seek(0, io.SeekStart); err != nil {
		file.Close()
		return &FileInfo{
			Error: err,
			Message: "Failed to seek begining of file.",
		}
	}

	return &FileInfo{
		Object: file,
		Content: content,
		Size: info.Size(),
	}
}

func(f *Filesystem) DeleteObject(c context.Context, key string) *FileInfo{
	remFile := filepath.Join(f.UploadDir, key)
	err := os.Remove(remFile)
	if err != nil {
		return &FileInfo{
			Error: err,
			Message: "Failed to delete file.",
		}
	}
	return &FileInfo{}
}
