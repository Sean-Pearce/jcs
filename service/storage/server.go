package main

import (
	"flag"
	"net/http"
	"path"

	"github.com/gin-gonic/gin"
	"github.com/minio/minio-go/v6"
	log "github.com/sirupsen/logrus"
)

var minioClient *minio.Client

const bucketName = "jcs"

func init() {
	flag.Parse()

	if *debug {
		log.SetLevel(log.DebugLevel)
	}

	// Initialize minio client
	var err error
	minioClient, err = minio.New(*endpoint, *accessKey, *secretKey, *useSSL)
	if err != nil {
		log.Fatal(err)
	}

	// Ensure that bucket 'jcs' exists
	location := "us-east-1"
	err = minioClient.MakeBucket(bucketName, location)
	if err != nil {
		// Check to see if we already own this bucket (which happens if you run this twice)
		exists, errBucketExists := minioClient.BucketExists(bucketName)
		if errBucketExists == nil && exists {
			log.Infof("We already own %s\n", bucketName)
		} else {
			log.Fatal(err)
		}
	} else {
		log.Infof("Successfully created %s\n", bucketName)
	}
}

func ping(c *gin.Context) {
	user, _, _ := c.Request.BasicAuth()
	c.JSON(http.StatusOK, gin.H{
		"user": user,
	})
}

func upload(c *gin.Context) {
	file, err := c.FormFile("file")
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "file not found",
		})
		return
	}

	body, err := file.Open()
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "open file error",
		})
		return
	}
	defer body.Close()

	user, _, _ := c.Request.BasicAuth()
	filename := c.PostForm("filename")
	if !validateFilename(filename) {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "invalid filename",
		})
		return
	}
	objName := path.Join(user, filename)

	_, err = minioClient.PutObject(bucketName, objName, body, file.Size, minio.PutObjectOptions{ContentType: c.ContentType()})
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "upload file error",
		})
		log.WithError(err).Errorf("upload object %v error", objName)
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"created": file.Filename,
	})
}

func download(c *gin.Context) {
	user, _, _ := c.Request.BasicAuth()
	filename := c.PostForm("filename")
	if !validateFilename(filename) {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "invalid filename",
		})
		return
	}
	objName := path.Join(user, filename)

	objInfo, err := minioClient.StatObject(bucketName, objName, minio.StatObjectOptions{})
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "stat object error",
		})
		log.WithError(err).Errorf("stat object %v error", objName)
		return
	}

	obj, err := minioClient.GetObject(bucketName, objName, minio.GetObjectOptions{})
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "get object error",
		})
		log.WithError(err).Errorf("get object %v error", objName)
		return
	}

	c.DataFromReader(http.StatusOK, objInfo.Size, objInfo.ContentType, obj, map[string]string{
		"Content-Disposition": "attachment; filename=" + objName,
	})
}

func validateFilename(filename string) bool {
	// TODO: use regexp to validate
	return filename != ""
}
