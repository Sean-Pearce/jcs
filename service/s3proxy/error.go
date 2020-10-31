package s3proxy

import (
	"bytes"
	"encoding/xml"
	"net/http"
	"strconv"

	"github.com/aws/aws-sdk-go/aws/awserr"
	"github.com/aws/aws-sdk-go/service/s3"
	log "github.com/sirupsen/logrus"
)

// S3Error structure
type S3Error struct {
	code ErrorCode
	orig error
}

func NewS3Error(code ErrorCode, orig error) *S3Error {
	return &S3Error{
		code: code,
		orig: orig,
	}
}

func (se *S3Error) Description() string {
	return errMap[se.code].description
}

func (se *S3Error) Error() string {
	if se.orig == nil {
		return se.Description()
	}
	return se.orig.Error()
}

func (se *S3Error) Code() ErrorCode {
	return se.code
}

func (se *S3Error) SetCode(code ErrorCode) {
	se.code = code
}

func (se *S3Error) Unwrap() error {
	return se.orig
}

func (se *S3Error) HTTPStatusCode() int {
	return errMap[se.code].httpStatusCode
}

func convertError(err error) *S3Error {
	if serr, ok := err.(*S3Error); ok {
		return serr
	}
	if code := parseError2HTTPCode(err); code == http.StatusNotFound {
		return NewS3Error(ErrNotFound, err)
	}
	code := ErrInternalError
	if aerr, ok := err.(awserr.Error); ok {
		switch aerr.Code() {
		case s3.ErrCodeNoSuchKey, "NotFound":
			code = ErrNoSuchKey
		case s3.ErrCodeNoSuchBucket:
			code = ErrNoSuchBucket
		case "EntityTooSmall":
			code = ErrEntityTooSmall
		case "EntityTooLarge":
			code = ErrEntityTooLarge
		}
	}
	return NewS3Error(code, err)
}

func writeError(r *http.Request, w http.ResponseWriter, err error) {
	var (
		s3Code   ErrorCode
		httpCode int
	)
	logger := log.WithFields(log.Fields{
		"url":    r.URL,
		"method": r.Method,
		"header": r.Header,
	})
	if serr, ok := err.(*S3Error); ok {
		s3Code = serr.Code()
		httpCode = serr.HTTPStatusCode()
		if httpCode == 0 {
			httpCode = http.StatusInternalServerError
		}
	} else {
		s3Code = ErrInternalError
		httpCode = http.StatusInternalServerError
	}
	if httpCode == http.StatusMethodNotAllowed {
		w.Header().Set("Allow", "GET, HEAD")
	}
	if s3Code == ErrInternalError {
		logger.WithError(err).Warn("request error")
	}
	w.Header().Set("Content-Type", "application/xml")
	body, mErr := xml.MarshalIndent(&xmlErrorResponse{
		Code:    s3Code.String(),
		Message: errMap[s3Code].description,
	}, "", "  ")
	if mErr != nil {
		logger.WithError(mErr).Warn("marshal xml failed")
	}
	body = bytes.Join([][]byte{[]byte(xml.Header), body, {'\n'}}, []byte{})
	w.Header().Set("Content-Length", strconv.Itoa(len(body)))
	w.WriteHeader(httpCode)
	if r.Method != http.MethodHead {
		if _, wErr := w.Write(body); wErr != nil {
			logger.WithError(wErr).Warn("write xml body failed")
		}
	}
}
