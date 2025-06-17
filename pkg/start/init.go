package start

import (
	"context"
	"crypto/rand"
	"encoding/base64"
	"fmt"
	"log"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/lamprosfasoulas/transfer/pkg/auth"
	"github.com/lamprosfasoulas/transfer/pkg/database"
	"github.com/lamprosfasoulas/transfer/pkg/sse"
	"github.com/lamprosfasoulas/transfer/pkg/storage"
)

var (
	Storage 		storage.Storage
	AuthProvider 	auth.AuthProvider
	Database 		database.Database
	Dispatcher		sse.Dispatcher
	Domain 			string
	MAX_SPACE		int64
)


func InitConfig() {
	Domain = os.Getenv("DOMAIN")
	if Domain == "" {
		log.SetPrefix(fmt.Sprintf("[\033[31mSYSTEM ERR\033[0m] "))
		log.Fatalln("Domain missing ...")
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


func InitAuthProvider() {
	switch os.Getenv("AUTH_PROVIDER") {
	case "LDAP":
		JWTExpiry := 1 * time.Hour
		JWTSecret := func (s string) string {
			if s == "" {
				bytes := make([]byte, 32)
				_, err := rand.Read(bytes)
				if err != nil {
					log.SetPrefix(fmt.Sprintf("[\033[31mSYSTEM ERR\033[0m] "))
					log.Fatalln("Failed to create JWTSecret:", err)
				}
				return base64.URLEncoding.EncodeToString(bytes)
			} 
			return base64.URLEncoding.EncodeToString([]byte(s))

		}
		AuthProvider = auth.NewLdapProvider(
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
		AuthProvider = auth.NewDevProvider(
			"dev",
			"dev",
			JWTSecret,
			JWTExpiry,
			)
	default:
		log.SetPrefix(fmt.Sprintf("[\033[31mSYSTEM ERR\033[0m] "))
		log.Fatalln("You have not selected an Auth Provider")
	}
}

func InitStorage() {
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
		Storage = storage.NewMinio(
			os.Getenv("MINIO_ENDPOINT"),
			os.Getenv("MINIO_ACCESSKEY"),
			os.Getenv("MINIO_SECRETKEY"),
			os.Getenv("MINIO_BUCKET"),
			MinioUseSSL(os.Getenv("MINIO_USESSL")),
			)

		if err := Storage.GetError(); err != nil {
			log.SetPrefix(fmt.Sprintf("[\033[31mSYSTEM ERR\033[0m] "))
			log.Fatalf("Failed to init MinIO client: %v\n", err)
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
		Storage = storage.NewFilesystem(
			upDir(os.Getenv("UPLOAD_DIR")),
			)
	default:
		log.SetPrefix(fmt.Sprintf("[\033[31mSYSTEM ERR\033[0m] "))
		log.Fatalln("You have not selected a Data Store option")
	}
}

func InitDatabase() {
	switch os.Getenv("DATABASE") {
	case "postgres":
		postgresUrl := func (s string) string {
			if s == "" {
				log.SetPrefix(fmt.Sprintf("[\033[31mSYSTEM ERR\033[0m] "))
				log.Fatalln("Database url is empty")
			}
			return s
		}

		//conn, err := pgx.Connect(context.Background(),postgresUrl(os.Getenv("POSTGRES_URL")))
		//if err != nil {
		//	log.SetPrefix(fmt.Sprintf("[\033[31mSYSTEM ERR\033[0m] "))
		//	log.Fatalln("Failed to Connect to Postgres")
		//}
		//defer conn.Close(context.Background())
		Database = database.NewPostgres(
			context.Background(),
			postgresUrl(os.Getenv("DB_URL")),
			)
		if err := Database.GetError(); err != nil {
			log.SetPrefix(fmt.Sprintf("[\033[31mSYSTEM ERR\033[0m] "))
			log.Fatalf("Failed to connect to postgres: %v\n", err)
		}
	default:
		log.SetPrefix(fmt.Sprintf("[\033[31mSYSTEM ERR\033[0m] "))
		log.Println("You have not selected a Database option")
	}
}

func InitDispatcher() {
	switch os.Getenv("DISPATCHER") {
	case "redis":
		Dispatcher = sse.NewRedisDispatcher(
			os.Getenv("REDIS_ADDR"),
			)
	default:
		Dispatcher = sse.NewMemDispatcher()
	}
}
