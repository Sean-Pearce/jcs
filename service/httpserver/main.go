package main

import (
	"flag"
	"fmt"

	"github.com/jinzhu/gorm"
	_ "github.com/jinzhu/gorm/dialects/sqlite"
)

var (
	dbPath   = flag.String("db", "db.db", "database path")
	port     = flag.String("port", ":5002", "http server port")
	tokenMap map[string]string
	db       *gorm.DB
)

func init() {
	var err error
	db, err = gorm.Open("sqlite3", *dbPath)
	if err != nil {
		panic(err)
	}
	db.AutoMigrate(&User{}, &File{}, &Site{})
	tokenMap = make(map[string]string)
}

func main() {
	defer db.Close()

	// 创建
	db.Create(&User{Name: "admin", Password: "admin"})

	// 读取
	var user User
	db.First(&user)
	fmt.Printf("%#v\n", user)

	// 更新 - 更新product的price为2000
	db.Model(&user).Update("password", "123")

	user = User{}
	db.First(&user, "name = ?", "admin")
	fmt.Printf("%#v\n", user)

	// 删除 - 删除product
	db.Delete(&user)

	user = User{}
	db.First(&user, "name = ?", "admin")
	fmt.Printf("%#v\n", user)
}
