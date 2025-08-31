// Handler are used to direct
// traffic to the right spots
package handlers

import (
	"fmt"
	"net/http"

	"github.com/google/uuid"
	"github.com/lamprosfasoulas/transfer/internal/logger"
	"github.com/lamprosfasoulas/transfer/internal/storage"
)

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

//i need work
//
// HandleHome gives the user the home screen
// It renders go templates with the struct
// below
//
// type PageData struct {
// 	Server string
// 	User string
// 	SSL string
// 	Space int64
// 	Files []database.File
// 	MAX int64
// }

type PageData struct {
	Server   string
	User     string
	UploadID string
	Space    int64
	Files    []storage.FileInfo
	MAX      int64
}

func (m *MainHandlers) Home(w http.ResponseWriter, r *http.Request) {
	username := GetUsernameFromContext(r)

	if username == "" {
		http.Redirect(w, r, "/login", http.StatusFound)
		return
	}

	ctx := r.Context()
	result, totalSize, err := m.Storage.ListFiles(ctx, username)
	if err != nil {
		m.Logger.Error(logger.Storage).Writef("Get files from backend error", err)
		//Those should go in a func together
		http.Error(w, "Internal server error", http.StatusInternalServerError)
	}

	data := PageData{
		Server:   m.Domain,
		Space:    totalSize,
		User:     username,
		UploadID: uuid.NewString(),
		Files:    result,
		MAX:      m.MAX_SPACE,
	}

	// Response
	if GetIsTerminalFromContext(r) {
		HomeTmplTerm.Execute(w, data)
	} else {
		HomeTmpl.Execute(w, data)
	}
}
