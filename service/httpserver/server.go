package main

import (
	"net/http"

	"github.com/Sean-Pearce/jcs/service/httpserver/dao"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/gin-gonic/gin"
	log "github.com/sirupsen/logrus"
)

const (
	// OK
	codeOK = 9200
	// BadRequest
	codeUploadError   = 9400
	codeAuthFail      = 9401
	codeInvalidToken  = 9402
	codeSignUpError   = 9403
	codeFileNotExists = 9404
	// InternalError
	codeInternalError = 9500
)

func tokenAuthMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		token := c.GetHeader("X-Token")
		if token == "" {
			token = c.Query("t")
		}
		if _, ok := tokenMap[token]; !ok {
			c.JSON(http.StatusUnauthorized, gin.H{
				"code":    codeInvalidToken,
				"message": "Invalid token",
			})
			log.Infof("Invalid token: %s", token)
			c.Abort()
		}

		c.Next()
	}
}

func signup(c *gin.Context) {
	username := c.Request.FormValue("username")
	password := c.Request.FormValue("password")
	user := dao.User{
		Username:  username,
		Password:  password,
		Role:      "user",
		AccessKey: genAccessKey(),
		SecretKey: genAccessKey(),
	}
	log.Infof("username: %s, password: %s", username, password)

	err := d.CreateUser(user)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{
			"code":    codeSignUpError,
			"message": "Sign up failed. Username has been used",
		})
		log.WithError(err).Infof("d.CreateUser(%v)", user)
		return
	}

	token := genToken()
	tokenMap[token] = username

	c.JSON(http.StatusOK, gin.H{
		"code": codeOK,
		"data": gin.H{
			"token": token,
		},
	})
}

func login(c *gin.Context) {
	username := c.Request.FormValue("username")
	password := c.Request.FormValue("password")
	log.Infof("username: %s, password: %s", username, password)

	user, err := d.GetUser(username)
	if err != nil || user.Password != password {
		c.JSON(http.StatusUnauthorized, gin.H{
			"code":    codeAuthFail,
			"message": "Account and password are incorrect.",
		})
		log.WithError(err).Infof("d.GetUser(%s) failed or password incorrect. user: %v", username, user)
		return
	}

	token := genToken()
	tokenMap[token] = username

	c.JSON(http.StatusOK, gin.H{
		"code": codeOK,
		"data": gin.H{
			"token": token,
		},
	})
}

func logout(c *gin.Context) {
	token := c.GetHeader("X-Token")
	delete(tokenMap, token)
	log.Infof("token: %s", token)

	c.JSON(http.StatusOK, gin.H{
		"code":    codeOK,
		"message": "See you ~",
	})
}

func userInfo(c *gin.Context) {
	token := c.GetHeader("X-Token")
	if token == "" {
		token = c.Query("t")
	}
	username := getUsernameByToken(token)
	log.Infof("username: %s, token: %s", username, token)

	user, err := d.GetUser(username)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{
			"code":    codeInvalidToken,
			"message": "User not exist.",
		})
		log.WithError(err).Infof("d.GetUser(%v) failed", username)
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"code": codeOK,
		"data": gin.H{
			"role":       user.Role,
			"access_key": user.AccessKey,
			"secret_key": user.SecretKey,
			"avatar":     "https://wpimg.wallstcn.com/f778738c-e4f8-4870-b634-56703b4acafe.gif",
		},
	})
}

func allCloudInfo(c *gin.Context) {
	clouds, err := d.GetAllCloudInfo()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"code":    codeInternalError,
			"message": "Get cloud info failed.",
		})
		log.WithError(err).Errorf("d.GetAllCloudInfo() failed", getFunctionName())
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"code": codeOK,
		"data": clouds,
	})
}

func passwd(c *gin.Context) {
	username := getUsernameByToken(c.GetHeader("X-Token"))
	password := c.Request.FormValue("password")
	newPassword := c.Request.FormValue("new_password")

	log.Infof("username: %s, password: %s, newPassword: %s", username, password, newPassword)

	user, err := d.GetUser(username)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{
			"code":    codeInvalidToken,
			"message": "User not exist.",
		})
		return
	}

	if user.Password != password {
		c.JSON(http.StatusOK, gin.H{
			"code":    codeAuthFail,
			"message": "Incorrect password.",
		})
		return
	}
	user.Password = newPassword
	err = d.UpdateUser(*user)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"code":    codeInternalError,
			"message": "Something is wrong.",
		})
		log.WithError(err).Errorf("set %v's password", username)
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"code":    codeOK,
		"message": "Successfully set password",
	})
}

func createBucket(c *gin.Context) {
	username := getUsernameByToken(c.Query("t"))
	var bucket dao.Bucket
	err := c.BindJSON(&bucket)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"code":    codeAuthFail,
			"message": "Bad request.",
		})
		return
	}

	bucket.Owner = username
	log.Infof("bucket: %v", bucket)
	err = d.CreateBucket(bucket)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"code":    codeInternalError,
			"message": "Something is wrong.",
		})
		log.WithError(err).Errorf("d.CreateBucket(%v) failed", bucket)
		return
	}

	clouds := append(bucket.Locations, minioName)
	for _, cloud := range clouds {
		client := s3Map[cloud]
		if client == nil {
			log.Errorf("cloud %s not found", cloud)
			continue
		}
		_, err = client.CreateBucket(&s3.CreateBucketInput{
			Bucket: aws.String(getBucketName(cloud, bucket.Name)),
		})
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{
				"code":    codeInternalError,
				"message": "create bucket failed.",
			})
			log.WithError(err).Errorf("%s.CreateBucket() failed", cloud)
			continue
		}
	}

	c.JSON(http.StatusOK, gin.H{
		"code": codeOK,
		"msg":  "success",
	})
}
