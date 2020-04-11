package main

import (
	"net/http"

	"github.com/gin-gonic/gin"
	log "github.com/sirupsen/logrus"
)

type server struct {
	name string
}

func ping(c *gin.Context) {
	user, _, _ := c.Request.BasicAuth()
	c.JSON(200, gin.H{
		"user": user,
	})
}

func upload(c *gin.Context) {
	file, err := c.FormFile("file")
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "file not found",
		})
		log.WithError(err).Debug("file not found")
		return
	}

	// TODO: upload to minio
	err = c.SaveUploadedFile(file, file.Filename)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "upload file failed",
		})
		log.WithError(err).Debug("upload file failed")
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"created": file.Filename,
	})
}

func download(c *gin.Context) {

}
