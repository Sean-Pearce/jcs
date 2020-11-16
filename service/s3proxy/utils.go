package s3proxy

import (
	"encoding/xml"
	"fmt"
	"net"
	"net/http"
	"net/url"
	"sync"
	"time"

	"github.com/aws/aws-sdk-go/aws/credentials"
	v4 "github.com/aws/aws-sdk-go/aws/signer/v4"
	log "github.com/sirupsen/logrus"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type xmlErrorResponse struct {
	XMLName   xml.Name `xml:"Error"`
	Code      string   `xml:"Code"`
	Message   string   `xml:"Message"`
	Resource  string   `xml:"Resource"`
	RequestId string   `xml:"RequestId"`
}

func getDirector(target *url.URL, signer *v4.Signer, backend string) func(req *http.Request) {
	return func(req *http.Request) {
		req.URL.Scheme = target.Scheme
		req.URL.Host = target.Host

		// Sign the request if it is radosgw backend
		if backend == constRGWBackend {
			// Req.Host should be "" or req.URL.Host. Because v4.Signer will prefer req.host to req.URL.host.
			req.Host = target.Host

			// RGW will ignore Content-Length if value is 0, which will lead to SignatureNotMatch.
			if req.Header.Get("Content-Length") == "0" {
				req.Header.Del("Content-Length")
			}

			_, err := signer.Sign(req, nil, "s3", "", time.Now())
			if err != nil {
				log.WithError(err).Errorln("sign failed")
			}
		}
	}
}

func getSigner(accessKey, secretKey string) *v4.Signer {
	creds := credentials.NewStaticCredentials(accessKey, secretKey, "")
	s := v4.NewSigner(creds)
	// Prevents setting the HTTPRequest's Body. Since the Body could be
	// wrapped in a custom io.Closer that we do not want to be stompped
	// on top of by the signer.
	s.DisableRequestBodyOverwrite = true
	// S3 service should not have any escaping applied
	s.DisableURIPathEscaping = true
	// Prevent signing and checking body because we pass nil
	// into Sign func within reverse proxy
	s.UnsignedPayload = true
	return s
}

func newProxyTransport() *http.Transport {
	return &http.Transport{
		DialContext: (&net.Dialer{
			Timeout:   10 * time.Second,
			KeepAlive: 100 * time.Second,
		}).DialContext,
		MaxIdleConns:          4096,
		MaxIdleConnsPerHost:   4096,
		IdleConnTimeout:       90 * time.Second,
		TLSHandshakeTimeout:   10 * time.Second,
		ExpectContinueTimeout: 1 * time.Second,
	}
}

func newBufferPool() *bufferPool {
	return &bufferPool{
		pool: &sync.Pool{
			New: func() interface{} {
				b := make([]byte, constBufferLength)
				return b
			},
		},
	}
}

type bufferPool struct {
	pool *sync.Pool
}

func (b *bufferPool) Get() []byte {
	return b.pool.Get().([]byte)
}

func (b *bufferPool) Put(buf []byte) {
	b.pool.Put(buf)
}

func contains(s []string, e string) bool {
	for _, a := range s {
		if a == e {
			return true
		}
	}
	return false
}

func parseError2HTTPCode(err error) (code int) {
	switch c := status.Code(err); c {
	case codes.DeadlineExceeded: // GRPC timeout
		return http.StatusGatewayTimeout
	case codes.NotFound:
		return http.StatusNotFound
	case codes.Unknown: // not GRPC error
		return http.StatusUnauthorized
	default:
		return http.StatusInternalServerError
	}
}

func getBucketName(cloud string, bucket string) string {
	if cloud == "minio" {
		return bucket
	}
	return fmt.Sprintf("jcs-%s-%s", cloud, bucket)
}
