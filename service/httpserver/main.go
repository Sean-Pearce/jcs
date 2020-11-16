package main

import (
	"flag"
	"strings"

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
	mongoURL      = flag.String("mongo", "mongodb://localhost:27017", "mongodb server address")
	schedulerAddr = flag.String("sched", "localhost:5001", "scheduler address")
	port          = flag.String("port", ":5000", "http server port")
	config        = flag.String("config", "httpserver.json", "httpserver config file")
	debug         = flag.Bool("debug", false, "debug mode")
	testMode      = flag.Bool("test", false, "enable test mode")
	tokenMap      map[string]string
	s3Map         map[string]*s3.S3
	d             *dao.Dao
)

func init() {
	flag.Parse()

	if *debug {
		log.SetLevel(log.DebugLevel)
	}

	var err error
	d, err = dao.NewDao(*mongoURL, "jcs", "user", "bucket", "cloud")
	if err != nil {
		panic(err)
	}

	if *testMode {
		err = d.CreateUser(dao.User{
			Username: "admin",
			Password: "admin",
			Role:     "admin",
		})
		if err != nil {
			log.WithError(err).Warnln("create test user failed")
		}
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
