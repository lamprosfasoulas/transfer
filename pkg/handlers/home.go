//Handler are used to direct
//traffic to the right spots
package handlers

import (
	"fmt"
	"log"
	"net/http"

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
	//prefix := username + "/"
	//We should be getting this from a cache
	//type File struct  {
	//  Owner 		string
	//	Filename 	string
	//	Size 		int64
	//	Id			string
	//}
	//objectCh := start.Storage.MinioClient.ListObjects(ctx, start.Storage.MinioBucket, minio.ListObjectsOptions{
	//	Prefix: prefix,
	//	Recursive: true,
	//})
	//type FileInfo struct {
	//	ID 			string
	//	URL			string
	//	Filename 	string
	//	Size		string
	//}

	//var files []FileInfo
	//for obj := range objectCh {
	//	if obj.Err != nil {
	//		log.Printf("Error listing objects: %v\n", obj.Err)
	//		continue
	//	}

	//	parts := strings.SplitN(obj.Key, "/", 2)
	//	if len(parts) != 2 {
	//		continue
	//	}
	//	fileID := parts[1]

	//	statInfo, err := start.Storage.MinioClient.StatObject(ctx, start.Storage.MinioBucket, obj.Key, minio.StatObjectOptions{})
	//	var origFilename string
	//	if err != nil {
	//		log.Printf("Error stat file: %v", obj.Err)
	//		continue
	//	}
	//	if meta := statInfo.Metadata["X-Amz-Meta-Filename"]; meta != nil {
	//		origFilename = meta[0]
	//	}

	//	files = append(files, FileInfo{
	//		ID: fileID,
	//		URL: fmt.Sprintf("http://localhost:42069/download/%s/%s", username, fileID),//presignedURL.String(),
	//		Filename: origFilename,
	//		Size: bites(obj.Size),
	//	})
	//}
	//fmt.Println(r.Context().Value("space").(int64))
	result, err := start.Database.GetUserFiles(ctx, username)
	if err != nil {
		log.SetPrefix(fmt.Sprintf("[\033[31mHOME ERR\033[0m] "))
		log.Printf("Upload to database failed: %v", err)
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
		SSL string
		Space int64
		Files []database.File
		MAX int64
	}{
		SSL: "http",
		Server: start.Domain,
		User: username,
		UploadID: "hi",
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
