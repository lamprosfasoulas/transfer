package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/lamprosfasoulas/transfer/pkg/handlers"
	"github.com/lamprosfasoulas/transfer/pkg/middleware"
	"github.com/lamprosfasoulas/transfer/pkg/start"
	_ "github.com/joho/godotenv/autoload"
)

const (
	GREEN 	= "\033[32m"
	RED 	= "\033[31m"
	RESET 	= "\033[0m"
)

//Error logging
//log.SetPrefix(fmt.Sprintf("[\033[31mERR\033[0m] "))
//INFO logging
//log.SetPrefix(fmt.Sprintf("[\033[34mINFO\033[0m] "))

// Janitor is used to periodically check for expired files.
// if files have passed their expiration date then they AuthProvider
// deleted from the backend storage and then the backend database.
// Upon deletion the user's available space is recalculated.
//
//I need work
func janitor() {
	ticker := time.NewTicker(30  * time.Minute)
	//ticker := time.NewTicker(5  * time.Second)
	defer ticker.Stop()

	for range ticker.C {
		log.SetPrefix(fmt.Sprintf("[\033[34mJANITOR INFO\033[0m] "))
		log.Printf("The janitor has started cleaning!!!\n")
		ctx := context.Background()

		//Deleting expired files
		files, err := start.Database.GetAllFiles(ctx)
		if err != nil {
			log.SetPrefix(fmt.Sprintf("[\033[31mJANITOR ERR\033[0m] "))
			log.Printf("Error getting files from database: %v\n", err)
			continue
		}

		for _, file := range files {
			//fmt.Println(file.Expiresat.Sub(time.Now().Add(7 * 25 * time.Hour)))
			expired := 0 > file.Expiresat.Sub(time.Now())
			if expired {
				log.SetPrefix(fmt.Sprintf("[\033[34mJANITOR INFO\033[0m] "))
				log.Printf("File %s is expired and is being deleted!\n", file.Objkey)
				delInfo := start.Storage.DeleteObject(ctx, file.Objkey)
				//fmt.Println("Del info err:",delInfo.Error)
				if delInfo.Error != nil {
					log.SetPrefix(fmt.Sprintf("[\033[31mJANITOR ERR\033[0m] "))
					log.Printf("File %s could not be deleted: %v!\n", file.Objkey, delInfo.Error)
					continue
				}
				err := start.Database.DeleteFile(ctx, file.Objkey)
				if err != nil {
					log.SetPrefix(fmt.Sprintf("[\033[31mJANITOR ERR\033[0m] "))
					log.Printf("File %s could not be removed from database: %v!\n", file.Objkey, err)
					continue
				}
				log.SetPrefix(fmt.Sprintf("[\033[34mJANITOR INFO\033[0m] "))
				log.Printf("File %s was deleted successfully !\n", file.Objkey)
			}
		}

		//Recalculating User Used Space
		users, err := start.Database.GetAllUsers(ctx)
		if err != nil {
			log.SetPrefix(fmt.Sprintf("[\033[31mJANITOR ERR\033[0m] "))
			log.Printf("Error getting files from database: %v\n", err)
			continue
		}
		for _, user := range users {
			err = start.Database.RecalculateUserSpace(ctx, user.Username)
			if err != nil {
				log.SetPrefix(fmt.Sprintf("[\033[31mJANITOR ERR\033[0m] "))
				log.Printf("Could not recalculate user %s used space: %v \n", user.Username, err)
				continue
			}
			log.SetPrefix(fmt.Sprintf("[\033[34mJANITOR INFO\033[0m] "))
			log.Printf("User's %s space has been recalculated !\n", user.Username)
		}
	}
}


// ─────────────────────────────────────────────────────────────────
//   INIT
// ─────────────────────────────────────────────────────────────────
func init() {
	//Get the initial config
	//start.InitConfig()

	//Start the minio client
	//switch start.Cfg.AuthProvider {
	//case "LDAP":
	//case "OIDC":
	//	start.InitOIDC()
	//default:
	//log.Fatalln("You have not specified AUTH_PROVIDER")

	//}
	start.InitConfig()
	start.InitAuthProvider()
	start.InitStorage()
	start.InitDatabase()
	start.InitDispatcher()

	//Load the Templates
	handlers.LoadTemplates()
}


// ─────────────────────────────────────────────────────────────────
//   MAIN
// ─────────────────────────────────────────────────────────────────
func main() {

	ctx := context.Background()
	defer start.Database.Close(ctx)
	//This cleans older files
	go janitor()

	//Homepage handler
	http.HandleFunc("/", middleware.RequireAuth(handlers.HandleHome))

	//Login Handlers
	http.HandleFunc("GET /login", handlers.HandleLoginGet)
	http.HandleFunc("POST /login", handlers.HandleLoginPost)

	//Logout Handler
	http.HandleFunc("GET /logout", middleware.RequireAuth(handlers.HandleLogout))

	//Upload Handler
	http.HandleFunc("/upload", middleware.RequireAuth(handlers.HandleUpload))

	//Download Handler
	http.HandleFunc("GET /download/", handlers.HandleDownload)

	//Delete Handler
	http.HandleFunc("POST /delete/", middleware.RequireAuth(handlers.HandleDelete))

	//Status Handler
	http.HandleFunc("GET /status/", middleware.RequireAuth(handlers.SSEHandler))
	//http.HandleFunc("GET /status/", handlers.SSEHandler)

	//Error handler
	http.HandleFunc("/error", handlers.HandleError)
	//OIDC callback handler
	//http.HandleFunc("GET /status/", handlers.OIDCCallback))



	// Handle CSS files -- Change me
	fs := http.FileServer(http.Dir("static"))
	http.Handle("/static/", http.StripPrefix("/static/", fs))

	//Starting server
	log.SetPrefix(fmt.Sprintf("[\033[34mSYSTEM INFO\033[0m] "))
	log.Printf("Starting server on %s …\n", start.Domain)
	if err := http.ListenAndServe(start.Domain, nil); err != nil {
		log.SetPrefix(fmt.Sprintf("[\033[31mSYSTEM ERR\033[0m] "))
		log.Fatalf("Server failed: %v", err)
	}
}
