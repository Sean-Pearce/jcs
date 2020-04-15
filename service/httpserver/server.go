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
		token := c.GetHeader("X-Token")
		if _, ok := tokenMap[token]; !ok {
			c.JSON(http.StatusUnauthorized, gin.H{
				"code": CodeNonToken,
			})
			c.Abort()
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
		c.JSON(http.StatusOK, gin.H{
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
		"code": CodeOK,
		"data": "success",
	})
}

func info(c *gin.Context) {
	username := getUsernameByToken(c.GetHeader("X-Token"))

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
	username := getUsernameByToken(c.GetHeader("X-Token"))

	var user User
	db.Preload("Preference.Sites").Preload("Files.Sites").Find(&user, "name = ?", username)

	files := []gin.H{}
	for _, f := range user.Files {
		sites := []string{}
		for _, site := range f.Sites {
			sites = append(sites, site.Name)
		}

		files = append(files, gin.H{
			"filename":      f.Name,
			"size":          f.Size,
			"last_modified": f.LastModified,
			"sites":         sites,
		})
	}

	c.JSON(http.StatusOK, gin.H{
		"code": CodeOK,
		"data": gin.H{
			"total": len(files),
			"items": files,
		},
	})
}

func site(c *gin.Context) {
	username := getUsernameByToken(c.GetHeader("X-Token"))

	var user User
	db.Preload("Preference.Sites").Preload("Files.Sites").Find(&user, "name = ?", username)

	siteResp := []string{}
	selectedResp := []string{}
	for site := range clientMap {
		siteResp = append(siteResp, site)
	}
	for _, site := range user.Preference.Sites {
		selectedResp = append(selectedResp, site.Name)
	}

	c.JSON(http.StatusOK, gin.H{
		"code": CodeOK,
		"data": gin.H{
			"total":    len(siteResp),
			"items":    siteResp,
			"selected": selectedResp,
		},
	})
}

func preference(c *gin.Context) {
	username := getUsernameByToken(c.GetHeader("X-Token"))

	var user User
	db.Preload("Preference.Sites").Preload("Files.Sites").Find(&user, "name = ?", username)

	var req struct {
		Preference struct {
			Sites []string
		}
	}
	c.BindJSON(&req)

	var sites []Site
	for _, site := range req.Preference.Sites {
		sites = append(sites, Site{Name: site})
	}

	user.Preference = Preference{
		Sites: sites,
	}
	db.Save(&user)

	c.JSON(http.StatusOK, gin.H{
		"code": CodeOK,
		"data": "success",
	})
}

func upload(c *gin.Context) {
	username := getUsernameByToken(c.GetHeader("X-Token"))
	var user User
	db.Preload("Preference.Sites").Preload("Files.Sites").Find(&user, "name = ?", username)

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
		if err != nil || resp.StatusCode != http.StatusCreated {
			logrus.WithError(err).Errorf("upload %v to %v failed", filename, site.Name)
			continue
		}

		sites = append(sites, Site{Name: site.Name})
	}

	user.Files = append(user.Files, File{
		Name:         file.Filename,
		Size:         file.Size,
		LastModified: time.Now().Unix(),
		Sites:        sites,
	})
	db.Save(&user)

	c.JSON(http.StatusOK, gin.H{
		"code": CodeOK,
		"data": "success",
	})
}

func download(c *gin.Context) {
	username := getUsernameByToken(c.GetHeader("X-Token"))
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
