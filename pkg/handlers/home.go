// Handler are used to direct
// traffic to the right spots
package handlers

import (
	"fmt"
	"log"
	"net/http"

	"github.com/google/uuid"
	"github.com/lamprosfasoulas/transfer/pkg/database"
	m "github.com/lamprosfasoulas/transfer/pkg/middleware"
	"github.com/lamprosfasoulas/transfer/pkg/start"
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
func HandleHome(w http.ResponseWriter, r *http.Request) {
	username := m.GetUsernameFromContext(r)
	//username := w.Header().Get("Authorization")
	if username == "" {
		http.Redirect(w, r, "/login", http.StatusFound)
		return
	}

	ctx := r.Context()
	result, err := start.Database.GetUserFiles(ctx, username)
	if err != nil {
		log.SetPrefix(fmt.Sprintf("[\033[31mHOME ERR\033[0m] "))
		log.Printf("Get files from database failed: %v", err)
		//Those should go in a func together
		http.Error(w, fmt.Sprintf("Failed to get files: %v",err), http.StatusInternalServerError)
	}
	//res := start.Storage.ListObjects(ctx, username)
	//re := (*res)[0].ExpireAt
	//fmt.Println("Left: ", int(re)/24)
	//fmt.Println("Left: ", int(re)%24)
	data := struct {
		Server string
		User string
		UploadID string
		Space int64
		Files []database.File
		MAX int64
	}{
		Server: start.Domain,
		User: username,
		UploadID: uuid.NewString(), //"hi",
		Files: result,
		MAX: start.MAX_SPACE,
	}
	for _, r := range result {
		data.Space += r.Size
	}
	// Response
	if m.GetIsTerminalFromContext(r) {
		HomeTmplTerm.Execute(w, data)
	} else {
		HomeTmpl.Execute(w, data)
	}
}
