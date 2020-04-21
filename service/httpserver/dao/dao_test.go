package dao

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

const (
	database   = "test"
	collection = "user"
)

var d *Dao

func init() {
	var err error
	d, err = NewDao("mongodb://localhost:27017", database, collection)
	if err != nil {
		panic(err)
	}
}

func TestAll(t *testing.T) {
	// drop collection
	d.client.Database(database).Collection(collection).Drop(context.TODO())
	d.ensureIndex("username", true)

	user := User{
		Username: "admin",
		Password: "secret",
		Role:     "admin",
	}

	strategy := Strategy{
		Sites: []string{"bj", "sh"},
	}

	files := []File{
		{
			Filename:     "testfile1",
			Size:         1024,
			LastModified: time.Now().Unix(),
			Sites:        []string{"bj", "sh"},
		},
		{
			Filename:     "testfile2",
			Size:         2048,
			LastModified: time.Now().Unix(),
			Sites:        []string{"gz", "sh"},
		},
	}

	testCreateUser(t, user)
	testGetUserInfo(t, user.Username, user)

	testSetUserStrategy(t, user.Username, strategy)
	testGetUserStrategy(t, user.Username, strategy)

	testAddFile(t, user.Username, files[0])
	testAddFile(t, user.Username, files[1])
	testGetFileInfo(t, user.Username, files[0].Filename, files[0])
	testGetFileInfo(t, user.Username, files[1].Filename, files[1])
	testGetUserFiles(t, user.Username, files)

	testRemoveFile(t, user.Username, files[0].Filename)
	testFileNotExists(t, user.Username, files[0].Filename)
	testGetUserFiles(t, user.Username, files[1:])

	testRemoveFile(t, user.Username, files[1].Filename)
	testFileNotExists(t, user.Username, files[1].Filename)
	testGetUserFiles(t, user.Username, files[2:])
}

func testCreateUser(t *testing.T, user User) {
	err := d.CreateNewUser(user)
	require.Nil(t, err)
}

func testGetUserInfo(t *testing.T, username string, want User) {
	user, err := d.GetUserInfo(username)
	require.Nil(t, err)
	require.Equal(t, want, *user)
}

func testAddFile(t *testing.T, username string, file File) {
	err := d.AddFile(username, file)
	require.Nil(t, err)
}

func testRemoveFile(t *testing.T, username, filename string) {
	err := d.RemoveFile(username, filename)
	require.Nil(t, err)
}

func testGetFileInfo(t *testing.T, username, filename string, want File) {
	file, err := d.GetFileInfo(username, filename)
	require.Nil(t, err)
	require.Equal(t, want, *file)
}

func testFileNotExists(t *testing.T, username, filename string) {
	_, err := d.GetFileInfo(username, filename)
	require.NotNil(t, err)
}

func testGetUserFiles(t *testing.T, username string, want []File) {
	files, err := d.GetUserFiles(username)
	require.Nil(t, err)
	require.Equal(t, want, *files)
}

func testSetUserStrategy(t *testing.T, username string, startegy Strategy) {
	err := d.SetUserStrategy(username, startegy)
	require.Nil(t, err)
}

func testGetUserStrategy(t *testing.T, username string, want Strategy) {
	strategy, err := d.GetUserStrategy(username)
	require.Nil(t, err)
	require.Equal(t, want, *strategy)
}
