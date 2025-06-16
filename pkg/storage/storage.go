package storage

import (
	"context"
	"fmt"
	"io"
	"time"

	"github.com/lamprosfasoulas/transfer/pkg/sse"
)

//Convertion chart
func bites(bytes int64) string {
    const (
        KB = 1 << (10 * 1) // 1024
        MB = 1 << (10 * 2) // 1,048,576
        GB = 1 << (10 * 3) // 1,073,741,824
        TB = 1 << (10 * 4) // 1,099,511,627,776
    )

    b := float64(bytes)

    switch {
    case bytes >= TB:
        return fmt.Sprintf("%.2f TB", b/float64(TB))
    case bytes >= GB:
        return fmt.Sprintf("%.2f GB", b/float64(GB))
    case bytes >= MB:
        return fmt.Sprintf("%.2f MB", b/float64(MB))
    case bytes >= KB:
        return fmt.Sprintf("%.2f KB", b/float64(KB))
    default:
        return fmt.Sprintf("%d B", bytes)
    }
}

// Storage is the interface that stores the actuall files.
// It is used to store, retrieve and delete files from 
// storage.
type Storage interface {
	PutObject		(context.Context, string, *ProgressReader) *FileInfo
	GetObject		(context.Context, string) *FileInfo
	//ListObjects		(context.Context, string) *[]FileInfo
	DeleteObject 	(context.Context, string) *FileInfo
	GetError 		() error
}

// ProgressReader is used to wrap io.Reader so that it can
// send SSE updates while the file is being streamed to a
// backend storage
type ProgressReader struct {
	Src 		io.Reader
	Total 		int64
	Red 		int64
	Filename	string
	UploadID	string
	Dispatch sse.Dispatcher
}

// NewProgressReader creates a new progress reader for each upload
func NewProgressReader (src io.Reader, total int64, file, upId string, d sse.Dispatcher) *ProgressReader{
	return &ProgressReader{
			Src: src,
			Total: total,
			Filename: file,
			UploadID: upId,
			Dispatch: d,
	}
}

// Read is a wrapper of io.Reader Read() method.
// It is called to read the files and gives us upload info.
func (pr *ProgressReader) Read(p []byte) (n int, err error) {
    n, err = pr.Src.Read(p)
    if n > 0 {
        pr.Red += int64(n)
		var pct float64
		if pr.Total > 0 {
			pct = float64(pr.Red) / float64(pr.Total) * 100
		}
		ev := sse.NewProgressEvent(
			pr.Filename,
			"Uploading",
			pr.Red,
			pr.Total,
			pct,
			)
		pr.Dispatch.SendEvent(context.Background(), pr.UploadID, ev)
    }
    return n, err
}

// FileInfo is the struct that is used for communication 
// with the storage backend.
// Not to be confused with the File struct that is used
// for communication with the database backend.
type FileInfo struct {
	Object 		io.ReadCloser
	Filename 	string // Filename
	ID			string // UUID
	Size 		int64 // Size in Bytes
	DispSize 	string // Not Used
	URL			string // Not USed

	Error 		error // Captures the Error
	Content 	string //Content-Type
	Message		string // Captures the message

	LastMod		time.Time // Not Used
	ExpireAt	int // Not Used
	Key			string // Object Key
}
