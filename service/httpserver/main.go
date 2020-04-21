package main

import (
	"encoding/json"
	"flag"
	"io/ioutil"

	"github.com/Sean-Pearce/jcs/service/httpserver/dao"
	pb "github.com/Sean-Pearce/jcs/service/scheduler/proto"
	"github.com/Sean-Pearce/jcs/service/storage/client"
	"github.com/gin-gonic/gin"
	log "github.com/sirupsen/logrus"
	"google.golang.org/grpc"
)

const (
	version = "v0.1"
)

var (
	mongoURL      = flag.String("mongo", "mongodb://localhost:27017", "mongodb server address")
	schedulerAddr = flag.String("sched", "localhost:5001", "scheduler address")
	port          = flag.String("port", ":5000", "http server port")
	config        = flag.String("accounts", "accounts.json", "accounts for storage backends")
	debug         = flag.Bool("debug", false, "debug mode")
	testMode      = flag.Bool("test", false, "enable test mode")
	tokenMap      map[string]string
	clientMap     map[string]*client.StorageClient
	clientList    []string
	d             *dao.Dao
	s             pb.SchedulerClient
)

func init() {
	flag.Parse()

	if *debug {
		log.SetLevel(log.DebugLevel)
	}

	var err error
	d, err = dao.NewDao(*mongoURL, "jcs", "user")
	if err != nil {
		panic(err)
	}

	if *testMode {
		err = d.CreateNewUser(dao.User{
			Username: "admin",
			Password: "admin",
			Role:     "admin",
		})
		if err != nil {
			log.WithError(err).Warnln("create test user failed")
		}
	}

	tokenMap = make(map[string]string)
	clientMap = make(map[string]*client.StorageClient)

	var clients []client.StorageClient
	data, err := ioutil.ReadFile(*config)
	if err != nil {
		panic(err)
	}
	json.Unmarshal(data, &clients)
	for i := range clients {
		clientMap[clients[i].Name] = &clients[i]
		clientList = append(clientList, clients[i].Name)
	}

	conn, err := grpc.Dial(*schedulerAddr, grpc.WithInsecure())
	if err != nil {
		panic(err)
	}

	s = pb.NewSchedulerClient(conn)
}

func main() {
	log.Infoln("Starting httpserver", version)

	r := gin.Default()
	r.POST("/api/user/login", login)

	r.Use(tokenAuthMiddleware())

	r.GET("/api/user/info", info)
	r.GET("/api/user/strategy", getStrategy)
	r.POST("/api/user/strategy", setStrategy)
	r.POST("/api/user/logout", logout)

	r.GET("/api/storage/list", list)
	r.GET("/api/storage/download", download)
	r.POST("/api/storage/upload", upload)
	r.DELETE("api/storage/delete/:filename", deleteFile)

	r.Run(*port)
}
