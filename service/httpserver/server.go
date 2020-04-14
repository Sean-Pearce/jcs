package main

import (
	"net/http"
	"path"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
)

const (
	CodeOK           = 20000
	CodeAuthFail     = 60204
	CodeUserNonexist = 50008
	CodeIllegalToken = 50008
	CodeLoggedIn     = 50012
	CodeTokenExpired = 50014
	CodeNoFile       = 50016
	CodeNonToken     = 50018
)

func TokenAuthMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		token := c.Request.FormValue("X-Token")
		if _, ok := tokenMap[token]; !ok {
			c.JSON(http.StatusUnauthorized, gin.H{
				"code": CodeNonToken,
			})
			return
		}

		c.Next()
	}
}

func login(c *gin.Context) {
	username := c.Request.FormValue("username")
	password := c.Request.FormValue("password")

	var user User
	db.First(&user, "name = ?", username)
	if user.Password != password {
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
		"data": token,
	})
}

func logout(c *gin.Context) {
	token := c.Request.FormValue("X-Token")
	delete(tokenMap, token)

	c.JSON(http.StatusOK, gin.H{
		"code": CodeOK,
		"data": "success",
	})
}

func info(c *gin.Context) {
	username := getUsernameByToken(c.Request.FormValue("X-Token"))

	var user User
	db.First(&user, "name = ?", username)

	c.JSON(http.StatusOK, gin.H{
		"code": CodeOK,
		"data": gin.H{
			"role":   user.Role,
			"avatar": "https://wpimg.wallstcn.com/f778738c-e4f8-4870-b634-56703b4acafe.gif",
		},
	})
}

func list(c *gin.Context) {
	username := getUsernameByToken(c.Request.FormValue("X-Token"))

	var user User
	db.First(&user, "name = ?", username)

	c.JSON(http.StatusOK, gin.H{
		"code": CodeOK,
		"data": gin.H{
			"total": len(user.Files),
			"items": user.Files,
		},
	})
}

func site(c *gin.Context) {
	username := getUsernameByToken(c.Request.FormValue("X-Token"))

	var user User
	db.First(&user, "name = ?", username)
	var sites []Site
	db.Find(&sites)

	c.JSON(http.StatusOK, gin.H{
		"code": CodeOK,
		"data": gin.H{
			"total":    len(sites),
			"items":    sites,
			"selected": user.Preference.Sites,
		},
	})
}

func preference(c *gin.Context) {
	username := getUsernameByToken(c.Request.FormValue("X-Token"))

	var prefMap map[string]interface{}
	c.BindJSON(&prefMap)
	rsites := prefMap["preference"].(map[string]interface{})["sites"].([]string)

	var sites []Site
	db.Find(&sites, "name IN (?)", rsites)
	db.Model(&User{}).Where("name = ?", username).Update("preference", Preference{Sites: sites})

	c.JSON(http.StatusOK, gin.H{
		"code": CodeOK,
		"data": "success",
	})
}

func upload(c *gin.Context) {
	username := getUsernameByToken(c.Request.FormValue("X-Token"))
	var user User
	db.First(&user, "name = ?", username)

	file, err := c.FormFile("file")
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"code": CodeNoFile,
		})
		return
	}
	// user1/testfile
	filename := path.Join(username, file.Filename)

	body, err := file.Open()
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "open file error",
		})
		return
	}
	defer body.Close()

	var sites []Site
	for _, site := range user.Preference.Sites {
		resp, err := clientMap[site.Name].upload(body, filename)
		if err != nil || resp.StatusCode != 200 {
			logrus.WithError(err).Errorf("upload %v to %v failed", filename, site.Name)
			continue
		}

		sites = append(sites, Site{Name: site.Name})
	}

	f := File{
		Name:         filename,
		Size:         file.Size,
		LastModified: time.Now().Unix(),
		Sites:        sites,
	}
	db.Save(&f)

	c.JSON(http.StatusOK, gin.H{
		"code": CodeOK,
		"data": "success",
	})
}

func download(c *gin.Context) {
	username := getUsernameByToken(c.Request.FormValue("X-Token"))
	filename := path.Join(username, c.Request.FormValue("filename"))

	var file File
	db.First(&file, "name = ?", filename)
	if file.Name != filename {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "file not exist",
		})
		return
	}

	resp, err := clientMap[file.Sites[0].Name].download(filename)
	if err != nil || resp.StatusCode != http.StatusOK {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "download from storage failed",
		})
		logrus.WithError(err).Errorf("download %v from %v failed", filename, file.Sites[0].Name)
		return
	}

	c.DataFromReader(http.StatusOK, file.Size, "multipart/form-data", resp.Body, map[string]string{
		"Content-Disposition": "attachment; filename=" + c.Request.FormValue("filename"),
	})
}
