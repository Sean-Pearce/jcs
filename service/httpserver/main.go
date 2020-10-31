package main

import (
	"flag"

	"github.com/Sean-Pearce/jcs/service/httpserver/dao"
	"github.com/gin-gonic/gin"
	log "github.com/sirupsen/logrus"
)

const (
	version = "v0.1"
)

var (
	mongoURL      = flag.String("mongo", "mongodb://localhost:27017", "mongodb server address")
	schedulerAddr = flag.String("sched", "localhost:5001", "scheduler address")
	port          = flag.String("port", ":5000", "http server port")
	config        = flag.String("config", "httpserver.json", "httpserver config file")
	debug         = flag.Bool("debug", false, "debug mode")
	testMode      = flag.Bool("test", false, "enable test mode")
	tokenMap      map[string]string
	clientList    []string
	d             *dao.Dao
	promAddr      = ":10090"
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
