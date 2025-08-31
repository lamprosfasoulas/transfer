package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"

	_ "github.com/joho/godotenv/autoload"
	"github.com/lamprosfasoulas/transfer/internal/auth"
	"github.com/lamprosfasoulas/transfer/internal/handlers"
	"github.com/lamprosfasoulas/transfer/internal/logger"
	"github.com/lamprosfasoulas/transfer/internal/middleware"
	"github.com/lamprosfasoulas/transfer/internal/sse"
	"github.com/lamprosfasoulas/transfer/internal/storage"
)

const (
	GREEN = "\033[32m"
	RED   = "\033[31m"
	RESET = "\033[0m"
)

var (
	authProvider     auth.AuthProvider
	storageProvider  storage.Storage
	dispatchProvider sse.Dispatcher
	logProvider      *logger.Logger

	domain    string
	MAX_SPACE int64
)

// Janitor is used to periodically check for expired files.
// if files have passed their expiration date then they AuthProvider
// deleted from the backend storage and then the backend database.
// Upon deletion the user's available space is recalculated.
func janitor() {
	ticker := time.NewTicker(1 * time.Hour)
	//ticker := time.NewTicker(5  * time.Second)
	defer ticker.Stop()

	for range ticker.C {
		logProvider.Info(logger.Janitor).Write("The janitor has started cleaning!!!")
		ctx := context.Background()

		//Deleting expired files
		files, _, err := storageProvider.ListFiles(ctx, "")
		if err != nil {
			logProvider.Error(logger.Janitor).Writef("Error getting files from storage", err)
			continue
		}

		for _, file := range files {
			//fmt.Println(file.Expiresat.Sub(time.Now().Add(7 * 25 * time.Hour)))
			expired := 0 > time.Until(file.ExpiresAt)
			if expired {
				logProvider.Warn(logger.Janitor).Write(fmt.Sprintf("File %s is expired and is being deleted!\n", file.ID))

				_, err := storageProvider.DeleteObject(ctx, file.ID)
				if err != nil {
					logProvider.Error(logger.Janitor).Write(fmt.Sprintf("File %s could not be deleted: %v!\n", file.ID, err))
					continue
				}

				logProvider.Warn(logger.Janitor).Write(fmt.Sprintf("File %s was deleted successfully !\n", file.ID))
			}
		}
	}
}

func initAuthProvider() {
	switch os.Getenv("AUTH_PROVIDER") {
	case "LDAP":
		authProvider = auth.NewLdapProvider(
			os.Getenv("LDAP_URL"),
			os.Getenv("LDAP_BindDN"),
			os.Getenv("LDAP_BindPW"),
			os.Getenv("LDAP_BaseDN"),
			os.Getenv("LDAP_Filter"),
			os.Getenv("JWT_Secret"),
			1*time.Hour,
		)
	case "dev":
		JWTExpiry := 12 * time.Hour
		JWTSecret := "thisisdeveloptestchangeme"
		authProvider = auth.NewDevProvider(
			"dev",
			"dev",
			JWTSecret,
			JWTExpiry,
		)
	default:
		//log.SetPrefix(fmt.Sprintf("[\033[31mSYSTEM ERR\033[0m] "))
		logProvider.Error().Fatal("You have not selected an Auth Provider")
	}
}

func initConfig() {
	domain = os.Getenv("DOMAIN")
	if domain == "" {
		//log.SetPrefix(fmt.Sprintf("[\033[31mSYSTEM ERR\033[0m] "))
		logProvider.Error().Fatal("Domain missing ...")
	}
	// I need work
	// get me from the env
	space := func(s string) int64 {
		var mult int64
		var size int64
		var atoi int
		var err error
		if strings.HasSuffix(s, "mb") {
			mult = 1024 * 1024
			atoi, err = strconv.Atoi(strings.TrimSuffix(s, "mb"))
			size = int64(atoi)
		} else if strings.HasSuffix(s, "gb") {
			mult = 1024 * 1024 * 1024
			atoi, err = strconv.Atoi(strings.TrimSuffix(s, "gb"))
			size = int64(atoi)
		}
		if err != nil {
			log.Fatal("Could not determine size limit")
		}

		return size * mult

	}
	//5 * 1024 * 1024 * 1024 //Max upload size per user
	MAX_SPACE = space(os.Getenv("MAX_SPACE"))
}

func initStorage() {
	switch os.Getenv("STORAGE") {
	case "minio":
		MinioUseSSL := func(s string) bool {
			if s == "true" {
				return true
			}
			return false
		}
		storageProvider = storage.NewMinio(
			os.Getenv("MINIO_ENDPOINT"),
			os.Getenv("MINIO_ACCESSKEY"),
			os.Getenv("MINIO_SECRETKEY"),
			os.Getenv("MINIO_BUCKET"),
			MinioUseSSL(os.Getenv("MINIO_USESSL")),
		)

		if err := storageProvider.GetError(); err != nil {
			logProvider.Error().Fatalf("Failed to init MinIO client", err)
			return
		}
	case "filesystem":
		upDir := func(s string) string {
			if s == "" {
				return "uploads"
			}
			return s
		}
		storageProvider = storage.NewFilesystem(
			upDir(os.Getenv("UPLOAD_DIR")),
		)
	default:
		logProvider.Error().Fatal("You have not selected a Data Store option")
	}
}

func initDispatcher() {
	switch os.Getenv("DISPATCHER") {
	case "redis":
		dispatchProvider = sse.NewRedisDispatcher(
			os.Getenv("REDIS_ADDR"),
		)
	default:
		dispatchProvider = sse.NewMemDispatcher()
	}
}

func initLogger() {
	var err error
	logProvider, err = logger.NewLogger("logfile.log")
	if err != nil {
		log.Fatalf("Failed to init logger: %v", err)
	}
}

// ─────────────────────────────────────────────────────────────────
//
//	INIT
//
// ─────────────────────────────────────────────────────────────────
func init() {

	//Get the initial config
	initLogger()
	initConfig()
	initAuthProvider()
	initStorage()
	initDispatcher()

	//Load the Templates
	handlers.LoadTemplates()
}

// ─────────────────────────────────────────────────────────────────
//
//	MAIN
//
// ─────────────────────────────────────────────────────────────────
func main() {
	authHandler := handlers.NewAuthHandler(
		authProvider,
		logProvider,
	)

	middleware := middleware.NewMiddleware(
		authProvider,
		storageProvider,
		logProvider,
	)

	mainHandler := handlers.NewMainHandler(
		storageProvider,
		dispatchProvider,
		MAX_SPACE,
		domain,
		logProvider,
	)

	//This cleans older files
	go janitor()

	homeHandler := middleware.RequireAuth(
		mainHandler.Home,
	)

	loginRenderer := middleware.RequireAuth(
		authHandler.LoginGet,
	)

	loginHandler := middleware.RequireAuth(
		authHandler.Login,
	)

	logoutHandler := middleware.RequireAuth(
		authHandler.Logout,
	)

	uploadHandler := middleware.RequireAuth(
		middleware.PrepUpload(
			mainHandler.Upload,
		),
	)

	downloadHandler := mainHandler.Download

	deleteHandler := middleware.RequireAuth(
		mainHandler.Delete,
	)

	statusHandler := middleware.RequireAuth(
		mainHandler.SSEHandler,
	)

	//Homepage handler
	http.HandleFunc("/", homeHandler)

	//Login Handlers
	http.HandleFunc("GET /login", loginRenderer)
	http.HandleFunc("POST /login", loginHandler)

	//Logout Handler
	http.HandleFunc("GET /logout", logoutHandler)

	//Upload Handler
	http.HandleFunc("POST /upload", uploadHandler)

	//Download Handler
	http.HandleFunc("GET /download/", downloadHandler)

	//Delete Handler
	http.HandleFunc("POST /delete/", deleteHandler)

	//Status Handler
	http.HandleFunc("GET /status/", statusHandler)

	// Handle CSS files -- Change me
	fs := http.FileServer(http.Dir("static"))
	http.Handle("/static/", http.StripPrefix("/static/", fs))

	//Starting server
	logProvider.Info().Write("Starting server on :42069")
	if err := http.ListenAndServe(":42069", nil); err != nil {
		logProvider.Error().Fatalf("Server failed: %v", err)
	}
}
