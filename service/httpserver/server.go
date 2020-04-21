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
	// OK
	codeOK = 9200
	// BadRequest
	codeUploadError   = 9400
	codeAuthFail      = 9401
	codeInvalidToken  = 9402
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
			"code":    codeAuthFail,
			"message": "Account and password are incorrect.",
		})
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

	c.JSON(http.StatusOK, gin.H{
		"code":    codeOK,
		"message": "See you ~",
	})
}

func info(c *gin.Context) {
	username := getUsernameByToken(c.GetHeader("X-Token"))

	user, err := d.GetUserInfo(username)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{
			"code":    codeInvalidToken,
			"message": "User not exist.",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"code": codeOK,
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
			"code":    codeInternalError,
			"message": "Something is wrong.",
		})
		log.WithError(err).Errorf("get %v's files", username)
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"code": codeOK,
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
			"code":    codeInternalError,
			"message": "Something is wrong.",
		})
		log.WithError(err).Errorf("get %v's strategy", username)
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"code": codeOK,
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
			"code":    codeInternalError,
			"message": "Something is wrong.",
		})
		log.WithError(err).Errorf("set %v's strategy", username)
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"code":    codeOK,
		"message": "Set strategy successfully",
	})
}

func upload(c *gin.Context) {
	username := getUsernameByToken(c.GetHeader("X-Token"))

	// TODO: FormFile reads all c.body
	file, err := c.FormFile("file")
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"code":    codeUploadError,
			"message": "Form key must be 'file'",
		})
		return
	}

	_, err = d.GetFileInfo(username, file.Filename)
	if err == nil {
		c.JSON(http.StatusOK, gin.H{
			"code":    codeUploadError,
			"message": "File already exists",
		})
		return
	}

	// user1/testfile
	filename := path.Join(username, file.Filename)

	body, err := file.Open()
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"code":    codeUploadError,
			"message": "Cannot open file",
		})
		return
	}

	strategy, err := d.GetUserStrategy(username)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"code":    codeInternalError,
			"message": "Something is wrong.",
		})
		log.WithError(err).Errorf("get %v's strategy", username)
		return
	}

	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	resp, err := s.Schedule(
		ctx,
		&pb.ScheduleRequest{Sites: strategy.Sites},
	)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"code":    codeInternalError,
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

	if len(sites) == 0 {
		c.JSON(http.StatusOK, gin.H{
			"code":    codeUploadError,
			"message": "Upload to storage backends failed",
		})
		return
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
			"code":    codeInternalError,
			"message": "Something is wrong.",
		})
		log.WithError(err).Errorf("add file %v for %v", filename, username)
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"code": codeOK,
	})
}

func deleteFile(c *gin.Context) {
	username := getUsernameByToken(c.GetHeader("X-Token"))
	filename := c.Param("filename")

	file, err := d.GetFileInfo(username, filename)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"code":    codeFileNotExists,
			"message": "The given file not exists.",
		})
		return
	}

	for _, site := range file.Sites {
		resp, err := clientMap[site].Delete(path.Join(username, filename))
		if err != nil || resp.StatusCode != http.StatusOK {
			c.JSON(http.StatusInternalServerError, gin.H{
				"code":    codeInternalError,
				"message": "Something is wrong.",
			})
			logrus.WithError(err).Errorf("delete %v in %v failed", filename, file.Sites[0], resp.StatusCode)
			return
		}
		resp.Body.Close()
	}

	err = d.RemoveFile(username, filename)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"code":    codeInternalError,
			"message": "Something is wrong.",
		})
		logrus.WithError(err).Errorf("Remove file %v from database failed", filename)
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"code":    codeOK,
		"message": "Delete file successfully",
	})
}

func download(c *gin.Context) {
	username := getUsernameByToken(c.Query("t"))
	filename := c.Query("filename")

	file, err := d.GetFileInfo(username, filename)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"code":    codeFileNotExists,
			"message": "The given file not exists.",
		})
		return
	}

	// TODO: choose best site
	resp, err := clientMap[file.Sites[0]].Download(path.Join(username, filename))
	if err != nil || resp.StatusCode != http.StatusOK {
		c.JSON(http.StatusInternalServerError, gin.H{
			"code":    codeInternalError,
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
