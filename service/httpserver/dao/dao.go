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
	client           *mongo.Client
	database         string
	userCollection   string
	bucketCollection string
	cloudCollection  string
}

type User struct {
	Username  string
	Password  string
	Role      string
	AccessKey string
	SecretKey string
}

type Bucket struct {
	Name  string
	Owner string

	// storage strategy
	Mode      string // "ec" | "replica"
	Locations []string
	Replica   int
	N         int
	K         int
}

type Cloud struct {
	Name      string
	Endpoint  string
	AccessKey string
	SecretKey string
	Status    string // "Online" | "Offline"
	Price     float64
	Latency   float64
}

// NewDao constructs a data access object (Dao).
func NewDao(mongoURI, database, userCollection, bucketCollection, cloudCollection string) (*Dao, error) {
	dao := &Dao{
		database:         database,
		userCollection:   userCollection,
		bucketCollection: bucketCollection,
		cloudCollection:  cloudCollection,
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
	err = dao.ensureIndex("username", true, dao.userCollection)
	if err != nil {
		return nil, err
	}
	err = dao.ensureIndex("name", true, dao.bucketCollection)
	if err != nil {
		return nil, err
	}
	err = dao.ensureIndex("name", true, dao.cloudCollection)
	if err != nil {
		return nil, err
	}

	return dao, nil
}

func (d *Dao) ensureIndex(index string, unique bool, collection string) error {
	col := d.client.Database(d.database).Collection(collection)
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

// CreateUser creates a new user.
func (d *Dao) CreateUser(user User) error {
	col := d.client.Database(d.database).Collection(d.userCollection)

	_, err := col.InsertOne(context.TODO(), user)
	if err != nil {
		return err
	}

	return nil
}

// GetUserBuckets returns user buckets.
func (d *Dao) GetUserBuckets(username string) ([]*Bucket, error) {
	col := d.client.Database(d.database).Collection(d.bucketCollection)

	var res []*Bucket
	cursor, err := col.Find(context.TODO(), bson.M{"owner": username})
	if err != nil {
		return nil, err
	}
	if err := cursor.All(context.TODO(), &res); err != nil {
		return nil, err
	}

	return res, nil
}

// GetUser returns the info of given user.
func (d *Dao) GetUser(username string) (*User, error) {
	col := d.client.Database(d.database).Collection(d.userCollection)

	var u User
	err := col.FindOne(context.TODO(), bson.M{"username": username}).Decode(&u)
	if err != nil {
		return nil, err
	}

	return &u, nil
}

// GetUserByAccessKey returns the info of given user.
func (d *Dao) GetUserByAccessKey(ak string) (*User, error) {
	col := d.client.Database(d.database).Collection(d.userCollection)

	var u User
	err := col.FindOne(context.TODO(), bson.M{"accesskey": ak}).Decode(&u)
	if err != nil {
		return nil, err
	}

	return &u, nil
}

// UpdateUser changes given user's password.
func (d *Dao) UpdateUser(user User) error {
	col := d.client.Database(d.database).Collection(d.userCollection)

	_, err := col.UpdateOne(
		context.TODO(),
		bson.M{
			"username": user.Username,
		},
		bson.M{
			"$set": bson.M{
				"password": user.Password,
			},
		},
	)
	if err != nil {
		return err
	}

	return nil
}

// CreateBucket creates a new bucket.
func (d *Dao) CreateBucket(bucket Bucket) error {
	col := d.client.Database(d.database).Collection(d.bucketCollection)

	_, err := col.InsertOne(context.TODO(), bucket)
	if err != nil {
		return err
	}

	return nil
}

// TODO: DeleteBucket

// GetBucket return the info of given bucket.
func (d *Dao) GetBucket(bucket string) (*Bucket, error) {
	col := d.client.Database(d.database).Collection(d.bucketCollection)

	var b Bucket
	err := col.FindOne(context.TODO(), bson.M{"name": bucket}).Decode(&b)
	if err != nil {
		return nil, err
	}

	return &b, nil
}

// InsertCloudInfo insert new cloud info to database.
func (d *Dao) InsertCloudInfo(cloud Cloud) error {
	col := d.client.Database(d.database).Collection(d.cloudCollection)
	upsert := true
	opt := options.UpdateOptions{Upsert: &upsert}
	_, err := col.UpdateOne(
		context.TODO(),
		bson.M{
			"name": cloud.Name,
		},
		cloud,
		&opt,
	)
	if err != nil {
		return err
	}

	return nil
}

// GetAllCloudInfo return the info of given bucket.
func (d *Dao) GetAllCloudInfo() ([]*Cloud, error) {
	col := d.client.Database(d.database).Collection(d.cloudCollection)

	var clouds []*Cloud
	cur, err := col.Find(context.TODO(), bson.D{})
	if err != nil {
		return nil, err
	}

	for cur.Next(context.TODO()) {

		// create a value into which the single document can be decoded
		var elem Cloud
		err := cur.Decode(&elem)
		if err != nil {
			return nil, err
		}
		clouds = append(clouds, &elem)
	}

	return clouds, nil
}
