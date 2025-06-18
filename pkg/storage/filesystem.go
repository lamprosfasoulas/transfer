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

func (f *Filesystem) PutObject(c context.Context, key string, r *ProgressReader) (*FileInfo, error) {
	userDir := filepath.Dir(key)
	saveDir := filepath.Join(f.UploadDir, userDir)
	err := os.MkdirAll(saveDir, os.ModePerm)
	if err != nil {
		return nil, err
	}
	filePath := filepath.Join(f.UploadDir, key)
	saveFile, err := os.Create(filePath)
	if err != nil {
		return nil, err
	}
	defer saveFile.Close()

	_, err = io.Copy(saveFile, r)
	if err != nil {
		return nil, err
	}

	info, err := saveFile.Stat()
	if err != nil {
		return nil, err
	}

	return &FileInfo{
		Key: key,
		Filename: r.Filename,
		Size: info.Size(),
	}, nil
}

func (f *Filesystem) GetObject(c context.Context, key string) (*FileInfo, error) {
	userFile := filepath.Join(f.UploadDir, key)
	file, err := os.Open(userFile)
	if err != nil {
		return nil, err
	}

	info, err := file.Stat()
	if err != nil {
		file.Close()
		return nil, err
	}

	buf := make([]byte, 512)
	_, err = file.Read(buf)
	if err != nil {
		file.Close()
		return nil, err
	}
	content := http.DetectContentType(buf)

	if _, err := file.Seek(0, io.SeekStart); err != nil {
		file.Close()
		return nil, err
	}

	return &FileInfo{
		Object: file,
		Content: content,
		Size: info.Size(),
	}, nil
}

func(f *Filesystem) DeleteObject(c context.Context, key string) (*FileInfo, error){
	remFile := filepath.Join(f.UploadDir, key)
	err := os.Remove(remFile)
	if err != nil {
		return nil, err
	}
	return nil, nil
}
