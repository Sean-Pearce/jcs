package main

import (
	"github.com/gin-gonic/gin"
	"github.com/jinzhu/gorm"
)

type User struct {
	gorm.Model
	Name     string
	Password string
	Files    []File
}

type File struct {
	gorm.Model
	Name   string
	Size   int64
	Sites  []Site
	UserID uint
}

type Site struct {
	gorm.Model
	Name     string
	Endpoint string
	FileID   uint
}

func login(c *gin.Context) {
	// todo
}

func logout(c *gin.Context) {
	// todo
}

func info(c *gin.Context) {
	// todo
}

func list(c *gin.Context) {
	// todo
}

func site(c *gin.Context) {
	// todo
}

func preference(c *gin.Context) {
	// todo
}

func upload(c *gin.Context) {
	// todo
}

func download(c *gin.Context) {
	// todo
}
