package main

import (
	"context"
	"crypto/rand"
	"encoding/base64"
	"fmt"
	"log"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"

	_ "github.com/joho/godotenv/autoload"
	"github.com/lamprosfasoulas/transfer/pkg/auth"
	"github.com/lamprosfasoulas/transfer/pkg/database"
	"github.com/lamprosfasoulas/transfer/pkg/handlers"
	"github.com/lamprosfasoulas/transfer/pkg/logger"
	"github.com/lamprosfasoulas/transfer/pkg/middleware"
	"github.com/lamprosfasoulas/transfer/pkg/sse"
	"github.com/lamprosfasoulas/transfer/pkg/storage"
)

const (
	GREEN 	= "\033[32m"
	RED 	= "\033[31m"
	RESET 	= "\033[0m"
)
var ( 
	authProvider auth.AuthProvider
	databaseProvider database.Database
	storageProvider storage.Storage
	dispatchProvider sse.Dispatcher
	domain string
	MAX_SPACE int64
	logProvider *logger.Logger
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
		//log.SetPrefix(fmt.Sprintf("[\033[34mJANITOR INFO\033[0m] "))
		//log.Printf("The janitor has dProvider cleaning!!!\n")
		logProvider.Info(logger.Jan).Write("The janitor has started cleaning!!!")
		ctx := context.Background()

		//Deleting expired files
		files, err := databaseProvider.GetAllFiles(ctx)
		if err != nil {
			//log.SetPrefix(fmt.Sprintf("[\033[31mJANITOR ERR\033[0m] "))
			//log.Printf("Error getting files from database: %v\n", err)
			logProvider.Error(logger.Jan).Writef("Error getting files from database", err)
			continue
		}

		for _, file := range files {
			//fmt.Println(file.Expiresat.Sub(time.Now().Add(7 * 25 * time.Hour)))
			expired := 0 > time.Until(*file.Expiresat)
			if expired {
				//log.SetPrefix(fmt.Sprintf("[\033[34mJANITOR INFO\033[0m] "))
				logProvider.Warn(logger.Jan).Write(fmt.Sprintf("File %s is expired and is being deleted!\n", file.Objkey))
				_, err := storageProvider.DeleteObject(ctx, file.Objkey)
				//fmt.Println("Del info err:",delInfo.Error)
				if err != nil {
					//log.SetPrefix(fmt.Sprintf("[\033[31mJANITOR ERR\033[0m] "))
					logProvider.Error(logger.Jan).Write(fmt.Sprintf("File %s could not be deleted: %v!\n", file.Objkey, err))
					continue
				}
				err = databaseProvider.DeleteFile(ctx, file.Objkey)
				if err != nil {
					//log.SetPrefix(fmt.Sprintf("[\033[31mJANITOR ERR\033[0m] "))
					logProvider.Error(logger.Jan).Write(fmt.Sprintf("File %s could not be removed from database: %v!\n", file.Objkey, err))
					continue
				}
//				log.SetPrefix(fmt.Sprintf("[\033[34mJANITOR INFO\033[0m] "))
				logProvider.Warn(logger.Jan).Write(fmt.Sprintf("File %s was deleted successfully !\n", file.Objkey))
			}
		}

		//Recalculating User Used Space
		users, err := databaseProvider.GetAllUsers(ctx)
		if err != nil {
			//log.SetPrefix(fmt.Sprintf("[\033[31mJANITOR ERR\033[0m] "))
			logProvider.Error(logger.Dat).Writef("Error getting files from database: %v\n", err)
			continue
		}
		for _, user := range users {
			err = databaseProvider.RecalculateUserSpace(ctx, user.Username)
			if err != nil {
				//log.SetPrefix(fmt.Sprintf("[\033[31mJANITOR ERR\033[0m] "))
				logProvider.Error(logger.Dat).Writef(fmt.Sprintf("Could not recalculate user %s used space", user.Username), err)
				continue
			}
			//log.SetPrefix(fmt.Sprintf("[\033[34mJANITOR INFO\033[0m] "))
			logProvider.Info(logger.Jan).Write(fmt.Sprintf("User's %s space has been recalculated !\n", user.Username))
		}
	}
}

func initAuthProvider() {
	switch os.Getenv("AUTH_PROVIDER") {
	case "LDAP":
		JWTExpiry := 1 * time.Hour
		JWTSecret := func (s string) string {
			if s == "" {
				bytes := make([]byte, 32)
				_, err := rand.Read(bytes)
				if err != nil {
					//log.SetPrefix(fmt.Sprintf("[\033[31mSYSTEM ERR\033[0m] "))
					logProvider.Error().Fatalf("Failed to create JWTSecret:", err)
				}
				return base64.URLEncoding.EncodeToString(bytes)
			} 
			return base64.URLEncoding.EncodeToString([]byte(s))

		}
		authProvider = auth.NewLdapProvider(
			os.Getenv("LDAP_URL"),
			os.Getenv("LDAP_BindDN"),
			os.Getenv("LDAP_BindPW"),
			os.Getenv("LDAP_BaseDN"),
			os.Getenv("LDAP_Filter"),
			JWTSecret(os.Getenv("JWT_Secret")),
			JWTExpiry,
			)
	case "dev":
		JWTExpiry := 12 * time.Hour
		JWTSecret := base64.URLEncoding.EncodeToString([]byte("thisisdeveloptestchangeme"))
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

func initDatabase() {
	switch os.Getenv("DATABASE") {
	case "postgres":
		postgresUrl := func (s string) string {
			if s == "" {
				//log.SetPrefix(fmt.Sprintf("[\033[31mSYSTEM ERR\033[0m] "))
				logProvider.Error().Fatal("Database url is empty")
			}
			return s
		}

		//conn, err := pgx.Connect(context.Background(),postgresUrl(os.Getenv("POSTGRES_URL")))
		//if err != nil {
		//	log.SetPrefix(fmt.Sprintf("[\033[31mSYSTEM ERR\033[0m] "))
		//	log.Fatalln("Failed to Connect to Postgres")
		//}
		//defer conn.Close(context.Background())
		databaseProvider = database.NewPostgres(
			context.Background(),
			postgresUrl(os.Getenv("DB_URL")),
			)
		if err := databaseProvider.GetError(); err != nil {
			//log.SetPrefix(fmt.Sprintf("[\033[31mSYSTEM ERR\033[0m] "))
			logProvider.Error().Fatalf("Failed to connect to postgres", err)
		}
	default:
		//log.SetPrefix(fmt.Sprintf("[\033[31mSYSTEM ERR\033[0m] "))
		logProvider.Error().Fatal("You have not selected a Database option")
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
	space := func (s string) int64 {
		var mult int64
		var size int64
		var atoi int
		var err error
		if strings.HasSuffix(s,"mb") {
			mult = 1024 * 1024
			atoi, err = strconv.Atoi(strings.TrimSuffix(s, "mb"))
			size = int64(atoi)
		} else if strings.HasSuffix(s, "gb"){
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
	switch os.Getenv("STORAGE"){
	case "minio":
		//MinioClient, err = minio.New(Cfg.MinioEndpoint, &minio.Options{
		//	Creds: credentials.NewStaticV4(Cfg.MinioAccessKey, Cfg.MinioSecretKey, ""),
		//	Secure: Cfg.MinioUseSSL,
		//})
		MinioUseSSL := func (s string) bool {
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
			//log.SetPrefix(fmt.Sprintf("[\033[31mSYSTEM ERR\033[0m] "))
			logProvider.Error().Fatalf("Failed to init MinIO client", err)
			return
		}
		//ctx := context.Background()
		//_, errBucket := MinioClient.BucketExists(ctx, Cfg.MinioBucket)
		//if errBucket != nil {
		//	log.Fatalf("Failed to get MinIO Bucket: %v\n", err)
		//	return
		//}
	case "filesystem":
		upDir := func (s string) string {
			if s == "" {
				return "uploads"
			}
			return s
		}
		storageProvider = storage.NewFilesystem(
			upDir(os.Getenv("UPLOAD_DIR")),
			)
	default:
		//log.SetPrefix(fmt.Sprintf("[\033[31mSYSTEM ERR\033[0m] "))
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
	initLogger()
	initConfig()
	initAuthProvider()
	initStorage()
	initDatabase()
	initDispatcher()

	//Load the Templates
	handlers.LoadTemplates()
}


// ─────────────────────────────────────────────────────────────────
//   MAIN
// ─────────────────────────────────────────────────────────────────
func main() {
	authHandler := handlers.NewAuthHandler(
		authProvider,
		databaseProvider,
		logProvider,
		)
	middleware := middleware.NewMiddleware(
		authProvider,
		databaseProvider,
		logProvider,
		)
	mainHandler := handlers.NewMainHandler(
		storageProvider,
		databaseProvider,
		dispatchProvider,
		MAX_SPACE,
		domain,
		logProvider,
		)


	ctx := context.Background()
	defer databaseProvider.Close(ctx)
	//This cleans older files
	go janitor()


	//Homepage handler
	http.HandleFunc("/", middleware.RequireAuth(mainHandler.Home))

	//Login Handlers
	http.HandleFunc("GET /login", middleware.RequireAuth(authHandler.LoginGet))
	http.HandleFunc("POST /login", middleware.RequireAuth(authHandler.Login))

	//Logout Handler
	http.HandleFunc("GET /logout", middleware.RequireAuth(authHandler.Logout))

	//Upload Handler
	http.HandleFunc("POST /upload", middleware.RequireAuth(mainHandler.Upload))

	//Download Handler
	http.HandleFunc("GET /download/", mainHandler.Download)

	//Delete Handler
	http.HandleFunc("POST /delete/", middleware.RequireAuth(mainHandler.Delete))

	//Status Handler
	http.HandleFunc("GET /status/", middleware.RequireAuth(mainHandler.SSEHandler))
	//http.HandleFunc("GET /status/", handlers.SSEHandler)

	//Error handler
	//http.HandleFunc("/error", handlers.HandleError)
	//OIDC callback handler
	//http.HandleFunc("GET /status/", handlers.OIDCCallback))

	// Handle CSS files -- Change me
	fs := http.FileServer(http.Dir("static"))
	http.Handle("/static/", http.StripPrefix("/static/", fs))

	//Starting server
	//log.SetPrefix(fmt.Sprintf("[\033[34mSYSTEM INFO\033[0m] "))
	logProvider.Info().Write("Starting server on :42069")
	//logProvider.Error().Write("tesing error messages")
	//logProvider.Info().Write("Testing once more")
	if err := http.ListenAndServe(":42069", nil); err != nil {
		//log.SetPrefix(fmt.Sprintf("[\033[31mSYSTEM ERR\033[0m] "))
		logProvider.Error().Fatalf("Server failed: %v", err)
	}
}
