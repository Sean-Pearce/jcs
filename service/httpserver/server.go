package main

import (
	"context"
	"net/http"
	"path"
	"time"

	"github.com/Sean-Pearce/jcs/service/httpserver/dao"
	pb "github.com/Sean-Pearce/jcs/service/scheduler/proto"
	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
	log "github.com/sirupsen/logrus"
)

const (
	CodeOK            = 9200
	CodeUploadError   = 9400
	CodeAuthFail      = 9401
	CodeInvalidToken  = 9402
	CodeFileNotExists = 9404
	CodeInternalError = 9500
)

func TokenAuthMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		token := c.GetHeader("X-Token")
		if token == "" {
			token = c.Query("t")
		}
		if _, ok := tokenMap[token]; !ok {
			c.JSON(http.StatusUnauthorized, gin.H{
				"code":    CodeInvalidToken,
				"message": "Invalid token",
			})
			c.Abort()
		}

		c.Next()
	}
}

func login(c *gin.Context) {
	username := c.Request.FormValue("username")
	password := c.Request.FormValue("password")

	user, err := d.GetUserInfo(username)
	if err != nil || user.Password != password {
		c.JSON(http.StatusUnauthorized, gin.H{
			"code":    CodeAuthFail,
			"message": "Account and password are incorrect.",
		})
		return
	}

	token := genToken()
	tokenMap[token] = username

	c.JSON(http.StatusOK, gin.H{
		"code": CodeOK,
		"data": gin.H{
			"token": token,
		},
	})
}

func logout(c *gin.Context) {
	token := c.GetHeader("X-Token")
	delete(tokenMap, token)

	c.JSON(http.StatusOK, gin.H{
		"code":    CodeOK,
		"message": "See you ~",
	})
}

func info(c *gin.Context) {
	username := getUsernameByToken(c.GetHeader("X-Token"))

	user, err := d.GetUserInfo(username)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{
			"code":    CodeInvalidToken,
			"message": "User not exist.",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"code": CodeOK,
		"data": gin.H{
			"role":   user.Role,
			"avatar": "https://wpimg.wallstcn.com/f778738c-e4f8-4870-b634-56703b4acafe.gif",
		},
	})
}

func list(c *gin.Context) {
	username := getUsernameByToken(c.GetHeader("X-Token"))

	files, err := d.GetUserFiles(username)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"code":    CodeInternalError,
			"message": "Something is wrong.",
		})
		log.WithError(err).Errorf("get %v's files", username)
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"code": CodeOK,
		"data": gin.H{
			"total": len(*files),
			"items": files,
		},
	})
}

func getStrategy(c *gin.Context) {
	username := getUsernameByToken(c.GetHeader("X-Token"))

	strategy, err := d.GetUserStrategy(username)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"code":    CodeInternalError,
			"message": "Something is wrong.",
		})
		log.WithError(err).Errorf("get %v's strategy", username)
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"code": CodeOK,
		"data": gin.H{
			"sites":    clientList,
			"strategy": strategy,
		},
	})
}

func setStrategy(c *gin.Context) {
	username := getUsernameByToken(c.GetHeader("X-Token"))

	// TODO: validate form
	var strategy dao.Strategy
	c.BindJSON(&strategy)

	err := d.SetUserStrategy(username, strategy)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"code":    CodeInternalError,
			"message": "Something is wrong.",
		})
		log.WithError(err).Errorf("set %v's strategy", username)
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"code":    CodeOK,
		"message": "Set strategy successfully",
	})
}

func upload(c *gin.Context) {
	username := getUsernameByToken(c.GetHeader("X-Token"))

	// TODO: FormFile reads all c.body
	file, err := c.FormFile("file")
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"code":    CodeUploadError,
			"message": "Form key must be 'file'",
		})
		return
	}
	// user1/testfile
	filename := path.Join(username, file.Filename)

	body, err := file.Open()
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"code":    CodeUploadError,
			"message": "Cannot open file",
		})
		return
	}

	strategy, err := d.GetUserStrategy(username)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"code":    CodeInternalError,
			"message": "Something is wrong.",
		})
		log.WithError(err).Errorf("get %v's strategy", username)
		return
	}

	ctx, _ := context.WithTimeout(context.Background(), time.Second)
	resp, err := s.Schedule(
		ctx,
		&pb.ScheduleRequest{Sites: strategy.Sites},
	)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"code":    CodeInternalError,
			"message": "Something is wrong.",
		})
		log.WithError(err).Errorf("schedule for %v, sites are %v", username, strategy.Sites)
		return
	}

	var sites []string
	for _, site := range resp.Sites {
		resp, err := clientMap[site].Upload(body, filename)
		if err != nil || resp.StatusCode != http.StatusCreated {
			log.WithError(err).Errorf("upload %v to %v failed", filename, site)
			continue
		}
		sites = append(sites, site)
	}

	item := dao.File{
		Filename:     file.Filename,
		Size:         file.Size,
		LastModified: time.Now().Unix(),
		Sites:        sites,
	}
	err = d.AddFile(username, item)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"code":    CodeInternalError,
			"message": "Something is wrong.",
		})
		log.WithError(err).Errorf("add file %v for %v", filename, username)
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"code": CodeOK,
	})
}

func download(c *gin.Context) {
	username := getUsernameByToken(c.Query("t"))
	filename := c.Query("filename")

	file, err := d.GetFileInfo(username, filename)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"code":    CodeFileNotExists,
			"message": "The given file not exists.",
		})
		return
	}

	// TODO: choose best site
	resp, err := clientMap[file.Sites[0]].Download(path.Join(username, filename))
	if err != nil || resp.StatusCode != http.StatusOK {
		c.JSON(http.StatusInternalServerError, gin.H{
			"code":    CodeInternalError,
			"message": "Something is wrong.",
		})
		logrus.WithError(err).Errorf("download %v from %v failed", filename, file.Sites[0], resp.StatusCode)
		return
	}
	defer resp.Body.Close()

	c.DataFromReader(http.StatusOK, resp.ContentLength, "multipart/form-data", resp.Body, map[string]string{
		"Content-Disposition": "attachment; filename=" + filename,
	})
}
