package s3proxy

import (
	"net/http"
	"net/url"
	"strings"

	"github.com/aws/aws-sdk-go/aws"
)

type requestType int

const (
	adminBucketReq requestType = iota
	listBucketsReq
	readBucketReq
	writeBucketReq
	listObjectsReq
	headBucketReq
	noriWriterReq

	notImplementReq requestType = 100
)

const NoriWriterHostFeature = "nori-write"

type copySource struct {
	Bucket string
	Key    string
	Range  string
}

type s3Query struct {
	Type   requestType
	Bucket string
	Key    string
	Source *copySource
}

// Copy from https://github.com/minio/minio/blob/master/cmd/handler-utils.go
func path2BucketAndObject(path string) (bucket, object string) {
	// Skip the first element if it is '/', split the rest.
	path = strings.TrimPrefix(path, "/")
	pathComponents := strings.SplitN(path, "/", 2)

	// Save the bucket and object extracted from path.
	switch len(pathComponents) {
	case 1:
		bucket = pathComponents[0]
	case 2:
		bucket = pathComponents[0]
		object = pathComponents[1]
	}
	return bucket, object
}

func getMetaData(h http.Header) map[string]*string {
	m := make(map[string]*string)
	for k, v := range h {
		k = strings.ToLower(k)
		if strings.HasPrefix(k, "x-amz-meta-") {
			key := strings.TrimPrefix(k, "x-amz-meta-")
			m[key] = aws.String(strings.Join(v, ","))
		}
	}
	return m
}

func parseS3Query(r *http.Request) (q *s3Query) {
	q = new(s3Query)
	q.Bucket, q.Key = path2BucketAndObject(r.URL.Path)
	query := r.URL.Query()

	if _, ok := query["acl"]; ok {
		q.Type = adminBucketReq
		return
	}
	if q.Key == "" {
		if q.Bucket == "" {
			q.Type = notImplementReq
			return
		}
		switch r.Method {
		case http.MethodGet:
			q.Type = listObjectsReq
		case http.MethodHead:
			q.Type = headBucketReq
		case http.MethodPost:
			// Delete objects request.
			if _, ok := query["delete"]; ok {
				q.Type = writeBucketReq
			} else {
				q.Type = adminBucketReq
			}
		case http.MethodPut, http.MethodDelete:
			q.Type = adminBucketReq
		default:
			q.Type = notImplementReq
		}
		return
	}

	switch r.Method {
	case http.MethodGet, http.MethodHead:
		q.Type = readBucketReq
	case http.MethodDelete:
		q.Type = writeBucketReq
	case http.MethodPost:
		if strings.Contains(r.Host, NoriWriterHostFeature) {
			q.Type = noriWriterReq
		} else {
			q.Type = writeBucketReq
		}
	case http.MethodPut:
		q.Type = writeBucketReq
		if v := r.Header.Get("x-amz-copy-source"); v != "" {
			if src, err := url.QueryUnescape(v); err == nil {
				v = src
			}
			q.Source = new(copySource)
			q.Source.Bucket, q.Source.Key = path2BucketAndObject(v)
			if q.Source.Key == "" {
				q.Type = notImplementReq
			}
			q.Source.Range = r.Header.Get("x-amz-copy-source-range")
		}
	default:
		q.Type = notImplementReq
	}
	return
}
