package dao

import (
	"context"
	"testing"

	"github.com/stretchr/testify/require"
)

const (
	database         = "test"
	userCollection   = "user"
	bucketCollection = "bucket"
	cloudCollection  = "cloud"
)

var d *Dao

func init() {
	var err error
	d, err = NewDao("mongodb://localhost:27017", database, userCollection, bucketCollection, cloudCollection)
	if err != nil {
		panic(err)
	}
}

func TestAll(t *testing.T) {
	// drop collection
	d.client.Database(database).Collection(userCollection).Drop(context.TODO())
	d.client.Database(database).Collection(bucketCollection).Drop(context.TODO())
	d.ensureIndex("username", true, d.userCollection)
	d.ensureIndex("bucket", true, d.bucketCollection)

	user := User{
		Username:  "admin",
		Password:  "secret",
		Role:      "admin",
		AccessKey: "ak",
		SecretKey: "sk",
	}

	bucket1 := Bucket{
		Name:  "testbucket",
		Owner: "admin",
	}
	bucket2 := Bucket{
		Name:  "testbucket2",
		Owner: "admin",
	}

	testCreateUser(t, user)
	testGetUser(t, user.Username, user)
	testGetUserByAccessKey(t, user.AccessKey, user)

	testCreateBucket(t, bucket1)
	testCreateBucket(t, bucket2)
	testGetBucket(t, bucket1.Name, bucket1)
	testGetUserBuckets(t, user.Username, []*Bucket{&bucket1, &bucket2})
}

func testCreateUser(t *testing.T, user User) {
	err := d.CreateUser(user)
	require.Nil(t, err)
}

func testGetUser(t *testing.T, username string, want User) {
	user, err := d.GetUser(username)
	require.Nil(t, err)
	require.Equal(t, want, *user)
}

func testGetUserByAccessKey(t *testing.T, ak string, want User) {
	user, err := d.GetUserByAccessKey(ak)
	require.Nil(t, err)
	require.Equal(t, want, *user)
}

func testCreateBucket(t *testing.T, bucket Bucket) {
	err := d.CreateBucket(bucket)
	require.Nil(t, err)
}

func testGetBucket(t *testing.T, bucket string, want Bucket) {
	b, err := d.GetBucket(bucket)
	require.Nil(t, err)
	require.Equal(t, want, *b)
}

func testGetUserBuckets(t *testing.T, username string, want []*Bucket) {
	b, err := d.GetUserBuckets(username)
	require.Nil(t, err)
	require.Equal(t, want, b)
}
