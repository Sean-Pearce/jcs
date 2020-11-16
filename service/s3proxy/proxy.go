package s3proxy

import (
	"net/http"
	"net/http/httputil"
	"net/url"
	"strings"

	"github.com/Sean-Pearce/jcs/service/httpserver/dao"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3"
	log "github.com/sirupsen/logrus"
	"go.mongodb.org/mongo-driver/mongo"
)

const (
	constBufferLength = 32 * 1024
	constRGWBackend   = "radosgw"
	constDatabase     = "jcs"
	constUserTable    = "user"
	constBucketTable  = "bucket"
	constCloudTable   = "cloud"

	minioName     = "minio"
	minioEndpoint = "http://localhost:9002"
	minioAK       = "minioadmin"
	minioSK       = "minioadmin"
)

type Proxy struct {
	// storage backends
	dao     *dao.Dao
	backend *httputil.ReverseProxy
	s3Map   map[string]*s3.S3
	tmpPath string
}

func NewProxy(endpoint, ak, sk string, mongoURL string, tmpPath string) (*Proxy, error) {
	p := &Proxy{}

	target, err := url.Parse(endpoint)
	if err != nil {
		return nil, err
	}

	rp := &httputil.ReverseProxy{
		Director: getDirector(target, getSigner(ak, sk), constRGWBackend),
	}
	rp.Transport = newProxyTransport()
	rp.BufferPool = newBufferPool()

	d, err := dao.NewDao(mongoURL, constDatabase, constUserTable, constBucketTable, constCloudTable)
	if err != nil {
		return nil, err
	}

	// init s3 clients
	s3map := make(map[string]*s3.S3)

	clouds, err := d.GetAllCloudInfo()
	if err != nil {
		return nil, err
	}
	clouds = append(clouds, &dao.Cloud{
		Name:      minioName,
		AccessKey: minioAK,
		SecretKey: minioSK,
		Endpoint:  minioEndpoint,
	})

	for _, cloud := range clouds {
		pathStyle := true
		if strings.HasPrefix(cloud.Name, "aliyun") {
			pathStyle = false
		}
		sess := session.Must(session.NewSession(
			&aws.Config{
				Endpoint: aws.String(cloud.Endpoint),
				Region:   aws.String("us-east-1"),
				Credentials: credentials.NewStaticCredentials(
					cloud.AccessKey,
					cloud.SecretKey,
					"",
				),
				DisableSSL:       aws.Bool(true),
				S3ForcePathStyle: aws.Bool(pathStyle),
			}),
		)
		s3map[cloud.Name] = s3.New(sess)
	}

	p.dao = d
	p.backend = rp
	p.s3Map = s3map
	p.tmpPath = tmpPath
	return p, nil
}

func (p *Proxy) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	// forward
	if err := p.forward(w, r); err != nil {
		// log.Error(err)
	}
}

func (p *Proxy) forward(w http.ResponseWriter, r *http.Request) error {
	// parse request
	query := parseS3Query(r)
	if query.Type == notImplementReq {
		writeError(r, w, NewS3Error(ErrNotImplemented, nil))
		return nil
	}

	// authentication
	user, err := p.checkSignature(r)
	if err != nil {
		writeError(r, w, NewS3Error(ErrInvalidAccessKeyID, nil))
		return nil
	}

	// check authorization
	bucket, err := p.getBucket(query.Bucket)
	if err != nil {
		return err
	}
	if bucket.Owner != user.Username {
		writeError(r, w, NewS3Error(ErrAccessDenied, nil))
		return nil
	}

	if query.Type == writeBucketReq && query.Bucket != "" && query.Key != "" {
		// PutObject
		p.backend.ServeHTTP(w, r)

		err = p.upload(bucket, query.Key)
		if err != nil {
			writeError(r, w, NewS3Error(ErrInternalError, nil))
			return err
		}
	} else if query.Type == readBucketReq && query.Bucket != "" && query.Key != "" {
		// GetBucket
		err = p.download(bucket, query.Key)
		if err != nil {
			log.WithError(err).Error("download failed.")
			writeError(r, w, NewS3Error(ErrInternalError, nil))
			return err
		}
		p.backend.ServeHTTP(w, r)
	} else {
		p.backend.ServeHTTP(w, r)
	}

	return nil
}

func (p *Proxy) checkSignature(r *http.Request) (*dao.User, error) {
	var user *dao.User
	err := verifyRequestSignature(r, func(accessKey string) (secretKey string, err error) {
		user, err = p.dao.GetUserByAccessKey(accessKey)
		if err != nil {
			return "", NewS3Error(ErrInvalidAccessKeyID, err)
		}
		return user.SecretKey, nil
	})
	return user, err
}

func (p *Proxy) getBucket(name string) (*dao.Bucket, error) {
	bucket, err := p.dao.GetBucket(name)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			return nil, NewS3Error(ErrNotFound, err)
		} else {
			return nil, NewS3Error(ErrInternalError, err)
		}
	}
	return bucket, nil
}
