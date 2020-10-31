package s3proxy

import (
	"crypto/hmac"
	"crypto/sha256"
	"crypto/subtle"
	"encoding/hex"
	"net/http"
	"net/textproto"
	"regexp"
	"sort"
	"strconv"
	"strings"
	"time"
)

const (
	signV2Algorithm = "AWS"
	signV4Algorithm = "AWS4-HMAC-SHA256"

	amzAccessKeyID   = "AWSAccessKeyId"
	amzContentSha256 = "X-Amz-Content-Sha256"
	amzCredential    = "X-Amz-Credential"
	amzDate          = "X-Amz-Date"
	amzExpires       = "X-Amz-Expires"
	amzSignedHeaders = "X-Amz-SignedHeaders"
	amzSignature     = "X-Amz-Signature"
	iso8601Format    = "20060102T150405Z"
	yyyymmdd         = "20060102"
	unsignedPayload  = "UNSIGNED-PAYLOAD"
	emptySHA256      = "e3b0c44298fc1c149afbf4c8996fb92427ae41e4649b934ca495991b7852b855"
	// The maximum allowed time difference between the incoming request
	// date and server date during signature verification.
	globalMaxSkewTime = 15 * time.Minute // 15 minutes skew allowed.
)

var signV4CredFormat = regexp.MustCompile("^([^/]*)/([0-9]{8})/([^/]*)/s3/aws4_request$")

type signV4Values struct {
	accessKey     string
	secretKey     string
	credDate      time.Time
	date          time.Time
	region        string
	signedHeaders []string
	signature     string
	// expires is only used for presigned request.
	expires int64
}

func (sv *signV4Values) getScope() string {
	return strings.Join([]string{sv.credDate.Format(yyyymmdd), sv.region, "s3", "aws4_request"}, "/")
}

func parseSignDate(r *http.Request) (time.Time, error) {
	var dateStr string
	if dateStr = r.Header.Get(amzDate); dateStr == "" {
		if dateStr = r.URL.Query().Get(amzDate); dateStr == "" {
			return time.Parse(time.RFC1123, r.Header.Get("Date"))
		}
	}
	return time.Parse(iso8601Format, dateStr)
}

func parseSignV4(r *http.Request, getSecretKey func(accessKey string) (string, error)) (sv signV4Values, err error) {
	authV4 := strings.TrimPrefix(r.Header.Get("Authorization"), signV4Algorithm)
	authFields := strings.Split(authV4, ",")
	if len(authFields) != 3 {
		return sv, NewS3Error(ErrAuthorizationHeaderMalformed, nil)
	}
	for i := range authFields {
		authFields[i] = strings.TrimSpace(authFields[i])
	}
	// Parse credential scope.
	if !strings.HasPrefix(authFields[0], "Credential=") {
		return sv, NewS3Error(ErrAuthorizationHeaderMalformed, nil)
	}
	matches := signV4CredFormat.FindStringSubmatch(strings.TrimPrefix(authFields[0], "Credential="))
	if len(matches) != 4 {
		return sv, NewS3Error(ErrAuthorizationHeaderMalformed, nil)
	}
	var credDateStr string
	sv.accessKey, credDateStr, sv.region = matches[1], matches[2], matches[3]
	sv.credDate, err = time.Parse(yyyymmdd, credDateStr)
	if err != nil {
		return sv, NewS3Error(ErrAuthorizationHeaderMalformed, err)
	}
	// Parse sign date.
	sv.date, err = parseSignDate(r)
	if err != nil {
		return sv, NewS3Error(ErrAuthorizationHeaderMalformed, err)
	}
	// Parse signedHeaders.
	if !strings.HasPrefix(authFields[1], "SignedHeaders=") {
		return sv, NewS3Error(ErrAuthorizationHeaderMalformed, nil)
	}
	sv.signedHeaders = strings.Split(strings.TrimPrefix(authFields[1], "SignedHeaders="), ";")
	// Parse signature.
	if !strings.HasPrefix(authFields[2], "Signature=") {
		return sv, NewS3Error(ErrAuthorizationHeaderMalformed, nil)
	}
	sv.signature = strings.TrimPrefix(authFields[2], "Signature=")
	// Get secretKey by accessKey.
	sv.secretKey, err = getSecretKey(sv.accessKey)
	if err != nil {
		return sv, err
	}
	return sv, nil
}

func parsePreSignV4(r *http.Request, getSecretKey func(accessKey string) (string, error)) (sv signV4Values, err error) {
	query := r.URL.Query()
	// Parse credential scope.
	matches := signV4CredFormat.FindStringSubmatch(query.Get(amzCredential))
	if len(matches) != 4 {
		return sv, NewS3Error(ErrCredMalformed, nil)
	}
	var dateStr string
	sv.accessKey, dateStr, sv.region = matches[1], matches[2], matches[3]
	sv.credDate, err = time.Parse(yyyymmdd, dateStr)
	if err != nil {
		return sv, NewS3Error(ErrMalformedCredentialDate, err)
	}
	// Parse date.
	sv.date, err = parseSignDate(r)
	if err != nil {
		return sv, NewS3Error(ErrMalformedPresignedDate, err)
	}
	// Parse expires.
	sv.expires, err = strconv.ParseInt(query.Get(amzExpires), 10, 64)
	if err != nil {
		return sv, NewS3Error(ErrMalformedExpires, err)
	}
	// Parse signedHeaders and signature.
	sv.signedHeaders = strings.Split(query.Get(amzSignedHeaders), ";")
	sv.signature = query.Get(amzSignature)
	// Get secretKey by accessKey.
	sv.secretKey, err = getSecretKey(sv.accessKey)
	if err != nil {
		return sv, err
	}
	return sv, nil
}

func isRequestSignatureV2(r *http.Request) bool {
	return !strings.HasPrefix(r.Header.Get("Authorization"), signV4Algorithm) &&
		strings.HasPrefix(r.Header.Get("Authorization"), signV2Algorithm)
}

func isRequestPresignedSignatureV2(r *http.Request) bool {
	_, ok := r.URL.Query()[amzAccessKeyID]
	return ok
}

func isRequestSignatureV4(r *http.Request) bool {
	return strings.HasPrefix(r.Header.Get("Authorization"), signV4Algorithm)
}

func isRequestPresignedSignatureV4(r *http.Request) bool {
	_, ok := r.URL.Query()[amzCredential]
	return ok
}

// extractSignedHeaders extract signed headers from Authorization header
func extractSignedHeaders(signedHeaders []string, r *http.Request) (http.Header, error) {
	reqHeaders := r.Header
	reqQueries := r.URL.Query()
	if !contains(signedHeaders, "host") {
		return nil, NewS3Error(ErrUnsignedHeaders, nil)
	}
	extractedSignedHeaders := make(http.Header)
	for _, header := range signedHeaders {
		val, ok := reqHeaders[http.CanonicalHeaderKey(header)]
		if !ok {
			val, ok = reqQueries[header]
		}
		if ok {
			for _, enc := range val {
				extractedSignedHeaders.Add(header, enc)
			}
			continue
		}
		switch header {
		case "expect":
			extractedSignedHeaders.Set(header, "100-continue")
		case "host":
			host := r.Host
			if host == "" {
				host = r.URL.Host
			}
			extractedSignedHeaders.Set(header, host)
		case "transfer-encoding":
			for _, enc := range r.TransferEncoding {
				extractedSignedHeaders.Add(header, enc)
			}
		case "content-length":
			extractedSignedHeaders.Set(header, strconv.FormatInt(r.ContentLength, 10))
		default:
			return nil, NewS3Error(ErrUnsignedHeaders, nil)
		}
	}
	return extractedSignedHeaders, nil
}

func getContentSha256Cksum(r *http.Request) string {
	if isRequestPresignedSignatureV4(r) {
		if v, ok := r.URL.Query()[amzContentSha256]; ok {
			return v[0]
		}
		if v, ok := r.Header[amzContentSha256]; ok {
			return v[0]
		}
		return unsignedPayload
	} else {
		if v, ok := r.Header[amzContentSha256]; ok {
			return v[0]
		}
		return emptySHA256
	}
}

func getCanonicalString(method string, encodedPath string, queryStr string, signedHeaders http.Header, hashedPayload string) string {
	var headers []string
	for k := range signedHeaders {
		headers = append(headers, strings.ToLower(k))
	}
	sort.Strings(headers)
	headerValues := make([]string, len(headers))
	for i, k := range headers {
		v := strings.Join(signedHeaders[textproto.CanonicalMIMEHeaderKey(k)], ",")
		headerValues[i] = k + ":" + strings.Join(strings.Fields(v), " ") + "\n"
	}
	return strings.Join([]string{
		method,
		encodedPath,
		queryStr,
		strings.Join(headerValues, ""),
		strings.Join(headers, ";"),
		hashedPayload,
	}, "\n")
}

// getStringToSign a string based on selected query values.
func getStringToSign(canonicalRequest string, t time.Time, scope string) string {
	stringToSign := signV4Algorithm + "\n" + t.Format(iso8601Format) + "\n"
	stringToSign = stringToSign + scope + "\n"
	canonicalRequestBytes := sha256.Sum256([]byte(canonicalRequest))
	stringToSign = stringToSign + hex.EncodeToString(canonicalRequestBytes[:])
	return stringToSign
}

// sumHMAC calculate hmac between two input byte array.
func sumHMAC(key []byte, data []byte) []byte {
	hash := hmac.New(sha256.New, key)
	hash.Write(data)
	return hash.Sum(nil)
}

// getSigningKey hmac seed to calculate final signature.
func getSigningKey(secretKey string, t time.Time, region string) []byte {
	date := sumHMAC([]byte("AWS4"+secretKey), []byte(t.Format(yyyymmdd)))
	regionBytes := sumHMAC(date, []byte(region))
	service := sumHMAC(regionBytes, []byte("s3"))
	signingKey := sumHMAC(service, []byte("aws4_request"))
	return signingKey
}

// getSignature final signature in hexadecimal form.
func getSignature(signingKey []byte, stringToSign string) string {
	return hex.EncodeToString(sumHMAC(signingKey, []byte(stringToSign)))
}

func doesSignV4Match(r *http.Request, isPresign bool, getSecretKey func(accessKey string) (string, error)) (err error) {
	var sv signV4Values
	if isPresign {
		sv, err = parsePreSignV4(r, getSecretKey)
	} else {
		sv, err = parseSignV4(r, getSecretKey)
	}
	if err != nil {
		return err
	}
	query := r.URL.Query()
	if sv.date.After(time.Now().Add(globalMaxSkewTime)) {
		return NewS3Error(ErrRequestTimeTooSkewed, nil)
	}
	if time.Now().Sub(sv.date) > time.Hour*24*7 {
		return NewS3Error(ErrExpiredPresignRequest, nil)
	}
	if isPresign {
		if sv.expires < 0 {
			return NewS3Error(ErrNegativeExpires, nil)
		}
		if sv.expires > int64(time.Hour*24*7/time.Second) {
			return NewS3Error(ErrMaximumExpires, nil)
		}
		if int64(time.Now().Sub(sv.date)/time.Second) > sv.expires {
			return NewS3Error(ErrExpiredPresignRequest, nil)
		}
		delete(query, amzSignature)
	}
	for key := range query {
		sort.Strings(query[key])
	}
	queryStr := strings.ReplaceAll(query.Encode(), "+", "%20")
	signedHeaders, err := extractSignedHeaders(sv.signedHeaders, r)
	if err != nil {
		return err
	}
	hashedPayload := getContentSha256Cksum(r)
	canonicalString := getCanonicalString(r.Method, r.URL.EscapedPath(), queryStr, signedHeaders, hashedPayload)
	stringToSign := getStringToSign(canonicalString, sv.date, sv.getScope())
	signingKey := getSigningKey(sv.secretKey, sv.date, sv.region)
	newSignature := getSignature(signingKey, stringToSign)
	if subtle.ConstantTimeCompare([]byte(newSignature), []byte(sv.signature)) != 1 {
		return NewS3Error(ErrSignatureDoesNotMatch, nil)
	}
	return nil
}

func verifyRequestSignature(r *http.Request, getSecretKey func(accessKey string) (string, error)) error {
	switch {
	case isRequestSignatureV4(r):
		return doesSignV4Match(r, false, getSecretKey)
	case isRequestPresignedSignatureV4(r):
		return doesSignV4Match(r, true, getSecretKey)
	case isRequestSignatureV2(r), isRequestPresignedSignatureV2(r):
		return NewS3Error(ErrSignatureVersionNotSupported, nil)
	default:
		return NewS3Error(ErrAuthorizationHeaderMalformed, nil)
	}
}
