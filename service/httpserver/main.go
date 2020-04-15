package main

import (
	"encoding/json"
	"flag"
	"io/ioutil"

	"github.com/gin-gonic/gin"
	"github.com/jinzhu/gorm"
	_ "github.com/jinzhu/gorm/dialects/sqlite"
	log "github.com/sirupsen/logrus"
)

var (
	dbPath    = flag.String("db", "db.db", "database path")
	port      = flag.String("port", ":5000", "http server port")
	config    = flag.String("config", "config.json", "storage info")
	debug     = flag.Bool("debug", false, "debug mode")
	tokenMap  map[string]string
	clientMap map[string]*StorageClient
	db        *gorm.DB
)

func init() {
	flag.Parse()

	if *debug {
		log.SetLevel(log.DebugLevel)
	}

	var err error
	db, err = gorm.Open("sqlite3", *dbPath)
	if err != nil {
		panic(err)
	}
	db.AutoMigrate(&User{}, &File{}, &Site{})

	tokenMap = make(map[string]string)
	clientMap = make(map[string]*StorageClient)

	var clients []StorageClient
	data, err := ioutil.ReadFile(*config)
	if err != nil {
		panic(err)
	}
	json.Unmarshal(data, &clients)
	for i := range clients {
		clientMap[clients[i].name] = &clients[i]
	}
}

func main() {
	defer db.Close()

	r := gin.Default()
	r.POST("/api/user/login", login)

	r.Use(TokenAuthMiddleware())

	r.GET("/api/user/info", info)
	r.GET("/api/user/site", site)
	r.POST("/api/user/logout", logout)
	r.POST("/api/user/preference", preference)

	r.GET("/api/storage/list", list)
	r.GET("/api/storage/download", download)
	r.POST("/api/storage/upload", upload)

	r.Run(*port)
}
