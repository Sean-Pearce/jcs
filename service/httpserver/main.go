package main

import (
	"flag"
	"fmt"
	"path"
	"runtime"
	"strings"
	"time"

	"github.com/Sean-Pearce/jcs/service/httpserver/dao"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/gin-gonic/gin"
	log "github.com/sirupsen/logrus"
)

const (
	version = "v0.1"

	minioName     = "minio"
	minioEndpoint = "http://localhost:9002"
	minioAK       = "minioadmin"
	minioSK       = "minioadmin"
)

var (
	mongoURL = flag.String("mongo", "mongodb://localhost:27017", "mongodb server address")
	port     = flag.String("port", ":5000", "http server port")
	tokenMap map[string]string
	s3Map    map[string]*s3.S3
	d        *dao.Dao
)

func init() {
	flag.Parse()

	log.SetReportCaller(true)
	log.SetFormatter(&log.TextFormatter{
		FullTimestamp:   true,
		TimestampFormat: time.RFC3339,
		CallerPrettyfier: func(f *runtime.Frame) (string, string) {
			filename := path.Base(f.File)
			return fmt.Sprintf("%s()", f.Function), fmt.Sprintf("%s:%d", filename, f.Line)
		},
	})

	var err error
	d, err = dao.NewDao(*mongoURL, "jcs", "user", "bucket", "cloud")
	if err != nil {
		panic(err)
	}

	tokenMap = make(map[string]string)
	s3Map = make(map[string]*s3.S3)

	clouds, err := d.GetAllCloudInfo()
	if err != nil {
		panic(err)
	}
	clouds = append(clouds, &dao.Cloud{
		Name:      minioName,
		AccessKey: minioAK,
		SecretKey: minioSK,
		Endpoint:  minioEndpoint,
	})

	for _, cloud := range clouds {
		pathStyle := true
		if strings.HasPrefix(cloud.Name, "aliyun") {
			pathStyle = false
		}
		sess := session.Must(session.NewSession(
			&aws.Config{
				Endpoint: aws.String(cloud.Endpoint),
				Region:   aws.String("us-east-1"),
				Credentials: credentials.NewStaticCredentials(
					cloud.AccessKey,
					cloud.SecretKey,
					"",
				),
				DisableSSL:       aws.Bool(true),
				S3ForcePathStyle: aws.Bool(pathStyle),
			}),
		)
		s3Map[cloud.Name] = s3.New(sess)
	}
}

func main() {
	log.Infoln("Starting httpserver", version)

	r := gin.Default()

	r.POST("/api/user/login", login)
	r.POST("/api/user/signup", signup)

	r.Use(tokenAuthMiddleware())

	r.GET("/api/cloudinfo", allCloudInfo)

	r.POST("/api/createbucket", createBucket)

	r.GET("/api/user/userinfo", userInfo)
	r.POST("/api/user/logout", logout)
	r.POST("/api/user/passwd", passwd)

	r.Run(*port)
}
