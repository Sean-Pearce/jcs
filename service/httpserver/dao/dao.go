package dao

import (
	"context"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"go.mongodb.org/mongo-driver/mongo/readpref"
)

// Dao encapsulates database operations.
type Dao struct {
	client     *mongo.Client
	database   string
	collection string
}

type User struct {
	Username string
	Password string
	Role     string
	Strategy Strategy
	Files    []File
}

type File struct {
	Filename     string   `json:"filename"`
	Size         int64    `json:"size"`
	LastModified int64    `json:"last_modified"`
	Sites        []string `json:"sites"`
}

type Strategy struct {
	Sites []string `json:"sites"`
}

// NewDao constructs a data access object (Dao).
func NewDao(mongoURI, database, collection string) (*Dao, error) {
	dao := &Dao{
		database:   database,
		collection: collection,
	}

	client, err := mongo.Connect(context.TODO(), options.Client().ApplyURI(mongoURI))
	if err != nil {
		return nil, err
	}

	ctx, _ := context.WithTimeout(context.Background(), time.Second)
	err = client.Ping(ctx, readpref.Primary())
	if err != nil {
		return nil, err
	}

	dao.client = client
	err = dao.ensureIndex("username", true)
	if err != nil {
		return nil, err
	}

	return dao, nil
}

func (d *Dao) ensureIndex(index string, unique bool) error {
	col := d.client.Database(d.database).Collection(d.collection)
	idx := mongo.IndexModel{
		Keys: bson.M{
			index: 1,
		},
		Options: &options.IndexOptions{
			Unique: &unique,
		},
	}

	_, err := col.Indexes().CreateOne(context.TODO(), idx)
	if err != nil {
		return err
	}

	return nil
}

// CreateNewUser creates a new user.
func (d *Dao) CreateNewUser(user User) error {
	col := d.client.Database(d.database).Collection(d.collection)

	user.Strategy = Strategy{Sites: []string{}}
	user.Files = []File{}

	_, err := col.InsertOne(context.TODO(), user)
	if err != nil {
		return err
	}

	return nil
}

// GetUserInfo returns the info of given user.
func (d *Dao) GetUserInfo(username string) (*User, error) {
	col := d.client.Database(d.database).Collection(d.collection)

	var u User
	err := col.FindOne(context.TODO(), bson.M{"username": username}, &options.FindOneOptions{
		Projection: bson.M{
			"strategy": 0,
			"files":    0,
		},
	}).Decode(&u)
	if err != nil {
		return nil, err
	}

	return &u, nil
}

// GetUserFiles returns files of given user.
func (d *Dao) GetUserFiles(username string) (*[]File, error) {
	col := d.client.Database(d.database).Collection(d.collection)

	var u User
	err := col.FindOne(context.TODO(), bson.M{"username": username}, &options.FindOneOptions{
		Projection: bson.M{
			"files": 1,
		},
	}).Decode(&u)
	if err != nil {
		return nil, err
	}

	return &u.Files, nil
}

// GetUserStrategy returns the storage strategy of given user.
func (d *Dao) GetUserStrategy(username string) (*Strategy, error) {
	col := d.client.Database(d.database).Collection(d.collection)

	var u User
	err := col.FindOne(context.TODO(), bson.M{"username": username}, &options.FindOneOptions{
		Projection: bson.M{
			"strategy": 1,
		},
	}).Decode(&u)
	if err != nil {
		return nil, err
	}

	return &u.Strategy, nil
}

// SetUserStrategy sets the storage strategy of given user.
func (d *Dao) SetUserStrategy(username string, strategy Strategy) error {
	col := d.client.Database(d.database).Collection(d.collection)

	_, err := col.UpdateOne(
		context.TODO(),
		bson.M{
			"username": username,
		},
		bson.M{
			"$set": bson.M{
				"strategy": strategy,
			},
		},
	)
	if err != nil {
		return err
	}

	return nil
}

// AddFile adds given file for given user.
func (d *Dao) AddFile(username string, file File) error {
	col := d.client.Database(d.database).Collection(d.collection)

	_, err := col.UpdateOne(
		context.TODO(),
		bson.M{
			"username": username,
		},
		bson.M{
			"$push": bson.M{
				"files": file,
			},
		},
	)
	if err != nil {
		return err
	}

	return nil
}

// GetFileInfo returns the info of given file.
func (d *Dao) GetFileInfo(username, filename string) (*File, error) {
	col := d.client.Database(d.database).Collection(d.collection)

	var u User
	err := col.FindOne(
		context.TODO(),
		bson.D{
			{Key: "username", Value: username},
			{Key: "files.filename", Value: filename},
		},
		&options.FindOneOptions{
			Projection: bson.M{
				"files": 1,
			},
		}).Decode(&u)
	if err != nil {
		return nil, err
	}

	// TODO: too slow
	var f File
	for _, f = range u.Files {
		if f.Filename == filename {
			break
		}
	}

	return &f, nil
}

// RemoveFile removes the given file from database.
func (d *Dao) RemoveFile(username, filename string) error {
	col := d.client.Database(d.database).Collection(d.collection)

	_, err := col.UpdateOne(context.TODO(),
		bson.M{
			"username": username,
		},
		bson.M{
			"$pull": bson.M{
				"files": bson.M{
					"filename": filename,
				},
			},
		},
	)
	if err != nil {
		return err
	}

	return nil
}
