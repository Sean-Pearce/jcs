/*
 * MinIO Cloud Storage, (C) 2015, 2016, 2017, 2018 MinIO, Inc.
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 *     http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 */

package s3proxy

import "net/http"

// Error codes, non exhaustive list - http://docs.aws.amazon.com/AmazonS3/latest/API/ErrorResponses.html
const (
	ErrNone ErrorCode = iota
	ErrAccessDenied
	ErrBadDigest
	ErrEntityTooSmall
	ErrEntityTooLarge
	ErrPolicyTooLarge
	ErrIncompleteBody
	ErrInternalError
	ErrInvalidAccessKeyID
	ErrInvalidBucketName
	ErrInvalidDigest
	ErrInvalidRange
	ErrInvalidCopyPartRange
	ErrInvalidCopyPartRangeSource
	ErrInvalidMaxKeys
	ErrInvalidEncodingMethod
	ErrInvalidMaxUploads
	ErrInvalidMaxParts
	ErrInvalidPartNumber
	ErrInvalidPartNumberMarker
	ErrInvalidRequestBody
	ErrInvalidCopySource
	ErrInvalidMetadataDirective
	ErrInvalidCopyDest
	ErrInvalidPolicyDocument
	ErrInvalidObjectState
	ErrMalformedXML
	ErrMissingContentLength
	ErrMissingContentMD5
	ErrMissingRequestBodyError
	ErrNoSuchBucket
	ErrNoSuchBucketPolicy
	ErrNoSuchBucketLifecycle
	ErrNoSuchKey
	ErrNoSuchUpload
	ErrNoSuchVersion
	ErrNotImplemented
	ErrPreconditionFailed
	ErrRequestTimeTooSkewed
	ErrSignatureDoesNotMatch
	ErrMethodNotAllowed
	ErrInvalidPart
	ErrInvalidPartOrder
	ErrAuthorizationHeaderMalformed
	ErrMalformedPOSTRequest
	ErrPOSTFileRequired
	ErrSignatureVersionNotSupported
	ErrBucketNotEmpty
	ErrAllAccessDisabled
	ErrMalformedPolicy
	ErrMissingFields
	ErrMissingCredTag
	ErrCredMalformed
	ErrInvalidRegion
	ErrInvalidService
	ErrInvalidRequestVersion
	ErrMissingSignTag
	ErrMissingSignHeadersTag
	ErrMalformedDate
	ErrMalformedPresignedDate
	ErrMalformedCredentialDate
	ErrMalformedCredentialRegion
	ErrMalformedExpires
	ErrNegativeExpires
	ErrAuthHeaderEmpty
	ErrExpiredPresignRequest
	ErrRequestNotReadyYet
	ErrUnsignedHeaders
	ErrMissingDateHeader
	ErrInvalidQuerySignatureAlgo
	ErrInvalidQueryParams
	ErrBucketAlreadyOwnedByYou
	ErrInvalidDuration
	ErrBucketAlreadyExists
	ErrMetadataTooLarge
	ErrUnsupportedMetadata
	ErrMaximumExpires
	ErrSlowDown
	ErrInvalidPrefixMarker
	ErrBadRequest
	ErrKeyTooLongError
	ErrInvalidBucketObjectLockConfiguration
	ErrObjectLocked
	ErrInvalidRetentionDate
	ErrPastObjectLockRetainDate
	ErrUnknownWORMModeDirective
	ErrObjectLockInvalidHeaders
	// ErrNotFound is not a standard s3 error.
	// It must be converted to a standard error by caller manually.
	ErrNotFound
)

type ErrorCode int

func (c ErrorCode) String() string {
	return errMap[c].code
}

type errDetail struct {
	code           string
	description    string
	httpStatusCode int
}

var errMap = map[ErrorCode]errDetail{
	ErrInvalidCopyDest: {
		code:           "InvalidRequest",
		description:    "This copy request is illegal because it is trying to copy an object to itself without changing the object's metadata, storage class, website redirect location or encryption attributes.",
		httpStatusCode: http.StatusBadRequest,
	},
	ErrInvalidCopySource: {
		code:           "InvalidArgument",
		description:    "Copy Source must mention the source bucket and key: sourcebucket/sourcekey.",
		httpStatusCode: http.StatusBadRequest,
	},
	ErrInvalidMetadataDirective: {
		code:           "InvalidArgument",
		description:    "Unknown metadata directive.",
		httpStatusCode: http.StatusBadRequest,
	},
	ErrInvalidRequestBody: {
		code:           "InvalidArgument",
		description:    "Body shouldn't be set for this request.",
		httpStatusCode: http.StatusBadRequest,
	},
	ErrInvalidMaxUploads: {
		code:           "InvalidArgument",
		description:    "Argument max-uploads must be an integer between 0 and 2147483647.",
		httpStatusCode: http.StatusBadRequest,
	},
	ErrInvalidMaxKeys: {
		code:           "InvalidArgument",
		description:    "Argument maxKeys must be an integer between 0 and 2147483647.",
		httpStatusCode: http.StatusBadRequest,
	},
	ErrInvalidEncodingMethod: {
		code:           "InvalidArgument",
		description:    "Invalid Encoding Method specified in Request.",
		httpStatusCode: http.StatusBadRequest,
	},
	ErrInvalidMaxParts: {
		code:           "InvalidArgument",
		description:    "Argument max-parts must be an integer between 0 and 2147483647.",
		httpStatusCode: http.StatusBadRequest,
	},
	ErrInvalidPartNumber: {
		code:           "InvalidArgument",
		description:    "Argument partNumber must be an integer.",
		httpStatusCode: http.StatusBadRequest,
	},
	ErrInvalidPartNumberMarker: {
		code:           "InvalidArgument",
		description:    "Argument partNumberMarker must be an integer.",
		httpStatusCode: http.StatusBadRequest,
	},
	ErrInvalidPolicyDocument: {
		code:           "InvalidPolicyDocument",
		description:    "The content of the form does not meet the conditions specified in the policy document.",
		httpStatusCode: http.StatusBadRequest,
	},
	ErrAccessDenied: {
		code:           "AccessDenied",
		description:    "Access Denied.",
		httpStatusCode: http.StatusForbidden,
	},
	ErrBadDigest: {
		code:           "BadDigest",
		description:    "The Content-Md5 you specified did not match what we received.",
		httpStatusCode: http.StatusBadRequest,
	},
	ErrEntityTooSmall: {
		code:           "EntityTooSmall",
		description:    "Your proposed upload is smaller than the minimum allowed object size.",
		httpStatusCode: http.StatusBadRequest,
	},
	ErrEntityTooLarge: {
		code:           "EntityTooLarge",
		description:    "Your proposed upload exceeds the maximum allowed object size.",
		httpStatusCode: http.StatusBadRequest,
	},
	ErrPolicyTooLarge: {
		code:           "PolicyTooLarge",
		description:    "Policy exceeds the maximum allowed document size.",
		httpStatusCode: http.StatusBadRequest,
	},
	ErrIncompleteBody: {
		code:           "IncompleteBody",
		description:    "You did not provide the number of bytes specified by the Content-Length HTTP header.",
		httpStatusCode: http.StatusBadRequest,
	},
	ErrInternalError: {
		code:           "InternalError",
		description:    "We encountered an internal error, please try again.",
		httpStatusCode: http.StatusInternalServerError,
	},
	ErrInvalidAccessKeyID: {
		code:           "InvalidAccessKeyId",
		description:    "The access key ID you provided does not exist in our records.",
		httpStatusCode: http.StatusForbidden,
	},
	ErrInvalidBucketName: {
		code:           "InvalidBucketName",
		description:    "The specified bucket is not valid.",
		httpStatusCode: http.StatusBadRequest,
	},
	ErrInvalidDigest: {
		code:           "InvalidDigest",
		description:    "The Content-Md5 you specified is not valid.",
		httpStatusCode: http.StatusBadRequest,
	},
	ErrInvalidRange: {
		code:           "InvalidRange",
		description:    "The requested range is not satisfiable.",
		httpStatusCode: http.StatusRequestedRangeNotSatisfiable,
	},
	ErrMalformedXML: {
		code:           "MalformedXML",
		description:    "The XML you provided was not well-formed or did not validate against our published schema.",
		httpStatusCode: http.StatusBadRequest,
	},
	ErrMissingContentLength: {
		code:           "MissingContentLength",
		description:    "You must provide the Content-Length HTTP header.",
		httpStatusCode: http.StatusLengthRequired,
	},
	ErrMissingContentMD5: {
		code:           "MissingContentMD5",
		description:    "Missing required header for this request: Content-Md5.",
		httpStatusCode: http.StatusBadRequest,
	},
	ErrMissingRequestBodyError: {
		code:           "MissingRequestBodyError",
		description:    "Request body is empty.",
		httpStatusCode: http.StatusLengthRequired,
	},
	ErrNoSuchBucket: {
		code:           "NoSuchBucket",
		description:    "The specified bucket does not exist.",
		httpStatusCode: http.StatusNotFound,
	},
	ErrNoSuchBucketPolicy: {
		code:           "NoSuchBucketPolicy",
		description:    "The bucket policy does not exist.",
		httpStatusCode: http.StatusNotFound,
	},
	ErrNoSuchBucketLifecycle: {
		code:           "NoSuchBucketLifecycle",
		description:    "The bucket lifecycle configuration does not exist.",
		httpStatusCode: http.StatusNotFound,
	},
	ErrNoSuchKey: {
		code:           "NoSuchKey",
		description:    "The specified key does not exist.",
		httpStatusCode: http.StatusNotFound,
	},
	ErrNoSuchUpload: {
		code:           "NoSuchUpload",
		description:    "The specified multipart upload does not exist. The upload ID may be invalid, or the upload may have been aborted or completed.",
		httpStatusCode: http.StatusNotFound,
	},
	ErrNoSuchVersion: {
		code:           "NoSuchVersion",
		description:    "Indicates that the version ID specified in the request does not match an existing version.",
		httpStatusCode: http.StatusNotFound,
	},
	ErrNotImplemented: {
		code:           "NotImplemented",
		description:    "A header you provided implies functionality that is not implemented.",
		httpStatusCode: http.StatusNotImplemented,
	},
	ErrPreconditionFailed: {
		code:           "PreconditionFailed",
		description:    "At least one of the pre-conditions you specified did not hold.",
		httpStatusCode: http.StatusPreconditionFailed,
	},
	ErrRequestTimeTooSkewed: {
		code:           "RequestTimeTooSkewed",
		description:    "The difference between the request time and the server's time is too large.",
		httpStatusCode: http.StatusForbidden,
	},
	ErrSignatureDoesNotMatch: {
		code:           "SignatureDoesNotMatch",
		description:    "The request signature we calculated does not match the signature you provided. Check your key and signing method.",
		httpStatusCode: http.StatusForbidden,
	},
	ErrMethodNotAllowed: {
		code:           "MethodNotAllowed",
		description:    "The specified method is not allowed against this resource.",
		httpStatusCode: http.StatusMethodNotAllowed,
	},
	ErrInvalidPart: {
		code:           "InvalidPart",
		description:    "One or more of the specified parts could not be found.  The part may not have been uploaded, or the specified entity tag may not match the part's entity tag.",
		httpStatusCode: http.StatusBadRequest,
	},
	ErrInvalidPartOrder: {
		code:           "InvalidPartOrder",
		description:    "The list of parts was not in ascending order. The parts list must be specified in order by part number.",
		httpStatusCode: http.StatusBadRequest,
	},
	ErrInvalidObjectState: {
		code:           "InvalidObjectState",
		description:    "The operation is not valid for the current state of the object.",
		httpStatusCode: http.StatusForbidden,
	},
	ErrAuthorizationHeaderMalformed: {
		code:           "AuthorizationHeaderMalformed",
		description:    "The authorization header is malformed; the region is wrong; expecting 'us-east-1'.",
		httpStatusCode: http.StatusBadRequest,
	},
	ErrMalformedPOSTRequest: {
		code:           "MalformedPOSTRequest",
		description:    "The body of your POST request is not well-formed multipart/form-data.",
		httpStatusCode: http.StatusBadRequest,
	},
	ErrPOSTFileRequired: {
		code:           "InvalidArgument",
		description:    "POST requires exactly one file upload per request.",
		httpStatusCode: http.StatusBadRequest,
	},
	ErrSignatureVersionNotSupported: {
		code:           "InvalidRequest",
		description:    "The authorization mechanism you have provided is not supported. Please use AWS4-HMAC-SHA256.",
		httpStatusCode: http.StatusBadRequest,
	},
	ErrBucketNotEmpty: {
		code:           "BucketNotEmpty",
		description:    "The bucket you tried to delete is not empty.",
		httpStatusCode: http.StatusConflict,
	},
	ErrBucketAlreadyExists: {
		code:           "BucketAlreadyExists",
		description:    "The requested bucket name is not available. The bucket namespace is shared by all users of the system. Please select a different name and try again.",
		httpStatusCode: http.StatusConflict,
	},
	ErrAllAccessDisabled: {
		code:           "AllAccessDisabled",
		description:    "All access to this bucket has been disabled.",
		httpStatusCode: http.StatusForbidden,
	},
	ErrMalformedPolicy: {
		code:           "MalformedPolicy",
		description:    "Policy has invalid resource.",
		httpStatusCode: http.StatusBadRequest,
	},
	ErrMissingFields: {
		code:           "MissingFields",
		description:    "Missing fields in request.",
		httpStatusCode: http.StatusBadRequest,
	},
	ErrMissingCredTag: {
		code:           "InvalidRequest",
		description:    "Missing Credential field for this request.",
		httpStatusCode: http.StatusBadRequest,
	},
	ErrCredMalformed: {
		code:           "AuthorizationQueryParametersError",
		description:    "Error parsing the X-Amz-Credential parameter; the Credential is mal-formed; expecting \"<YOUR-AKID>/YYYYMMDD/REGION/SERVICE/aws4_request\".",
		httpStatusCode: http.StatusBadRequest,
	},
	ErrMalformedDate: {
		code:           "MalformedDate",
		description:    "Invalid date format header, expected to be in ISO8601, RFC1123 or RFC1123Z time format.",
		httpStatusCode: http.StatusBadRequest,
	},
	ErrMalformedPresignedDate: {
		code:           "AuthorizationQueryParametersError",
		description:    "X-Amz-Date must be in the ISO8601 Long Format \"yyyyMMdd'T'HHmmss'Z'\".",
		httpStatusCode: http.StatusBadRequest,
	},
	// FIXME: Should contain the invalid param set as seen in https://github.com/minio/minio/issues/2385.
	// right description:    "Error parsing the X-Amz-Credential parameter; incorrect date format \"%s\". This date in the credential must be in the format \"yyyyMMdd\".",
	// Need changes to make sure variable messages can be constructed.
	ErrMalformedCredentialDate: {
		code:           "AuthorizationQueryParametersError",
		description:    "Error parsing the X-Amz-Credential parameter; incorrect date format \"%s\". This date in the credential must be in the format \"yyyyMMdd\".",
		httpStatusCode: http.StatusBadRequest,
	},
	// FIXME: Should contain the invalid param set as seen in https://github.com/minio/minio/issues/2385.
	// right description:    "Error parsing the X-Amz-Credential parameter; the region 'us-east-' is wrong; expecting 'us-east-1'".
	// Need changes to make sure variable messages can be constructed.
	ErrMalformedCredentialRegion: {
		code:           "AuthorizationQueryParametersError",
		description:    "Error parsing the X-Amz-Credential parameter; the region is wrong.",
		httpStatusCode: http.StatusBadRequest,
	},
	ErrInvalidRegion: {
		code:           "InvalidRegion",
		description:    "Region does not match.",
		httpStatusCode: http.StatusBadRequest,
	},
	// FIXME: Should contain the invalid param set as seen in https://github.com/minio/minio/issues/2385.
	// right description:   "Error parsing the X-Amz-Credential parameter; incorrect service \"s4\". This endpoint belongs to \"s3\".".
	// Need changes to make sure variable messages can be constructed.
	ErrInvalidService: {
		code:           "AuthorizationQueryParametersError",
		description:    "Error parsing the X-Amz-Credential parameter; incorrect service. This endpoint belongs to \"s3\".",
		httpStatusCode: http.StatusBadRequest,
	},
	// FIXME: Should contain the invalid param set as seen in https://github.com/minio/minio/issues/2385.
	// description:   "Error parsing the X-Amz-Credential parameter; incorrect terminal "aws4_reque". This endpoint uses "aws4_request".
	// Need changes to make sure variable messages can be constructed.
	ErrInvalidRequestVersion: {
		code:           "AuthorizationQueryParametersError",
		description:    "Error parsing the X-Amz-Credential parameter; incorrect terminal. This endpoint uses \"aws4_request\".",
		httpStatusCode: http.StatusBadRequest,
	},
	ErrMissingSignTag: {
		code:           "AccessDenied",
		description:    "Signature header missing Signature field.",
		httpStatusCode: http.StatusBadRequest,
	},
	ErrMissingSignHeadersTag: {
		code:           "InvalidArgument",
		description:    "Signature header missing SignedHeaders field.",
		httpStatusCode: http.StatusBadRequest,
	},
	ErrMalformedExpires: {
		code:           "AuthorizationQueryParametersError",
		description:    "X-Amz-Expires should be a number.",
		httpStatusCode: http.StatusBadRequest,
	},
	ErrNegativeExpires: {
		code:           "AuthorizationQueryParametersError",
		description:    "X-Amz-Expires must be non-negative.",
		httpStatusCode: http.StatusBadRequest,
	},
	ErrAuthHeaderEmpty: {
		code:           "InvalidArgument",
		description:    "Authorization header is invalid -- one and only one ' ' (space) required.",
		httpStatusCode: http.StatusBadRequest,
	},
	ErrMissingDateHeader: {
		code:           "AccessDenied",
		description:    "AWS authentication requires a valid Date or x-amz-date header.",
		httpStatusCode: http.StatusBadRequest,
	},
	ErrInvalidQuerySignatureAlgo: {
		code:           "AuthorizationQueryParametersError",
		description:    "X-Amz-Algorithm only supports \"AWS4-HMAC-SHA256\".",
		httpStatusCode: http.StatusBadRequest,
	},
	ErrExpiredPresignRequest: {
		code:           "AccessDenied",
		description:    "Request has expired.",
		httpStatusCode: http.StatusForbidden,
	},
	ErrRequestNotReadyYet: {
		code:           "AccessDenied",
		description:    "Request is not valid yet.",
		httpStatusCode: http.StatusForbidden,
	},
	ErrSlowDown: {
		code:           "SlowDown",
		description:    "Please reduce your request.",
		httpStatusCode: http.StatusServiceUnavailable,
	},
	ErrInvalidPrefixMarker: {
		code:           "InvalidPrefixMarker",
		description:    "Invalid marker prefix combination.",
		httpStatusCode: http.StatusBadRequest,
	},
	ErrBadRequest: {
		code:           "BadRequest",
		description:    "400 BadRequest.",
		httpStatusCode: http.StatusBadRequest,
	},
	ErrKeyTooLongError: {
		code:           "KeyTooLongError",
		description:    "Your key is too long.",
		httpStatusCode: http.StatusBadRequest,
	},

	// FIXME: Actual XML error response also contains the header which missed in list of signed header parameters.
	ErrUnsignedHeaders: {
		code:           "AccessDenied",
		description:    "There were headers present in the request which were not signed.",
		httpStatusCode: http.StatusBadRequest,
	},
	ErrInvalidQueryParams: {
		code:           "AuthorizationQueryParametersError",
		description:    "Query-string authentication version 4 requires the X-Amz-Algorithm, X-Amz-Credential, X-Amz-Signature, X-Amz-Date, X-Amz-SignedHeaders, and X-Amz-Expires parameters.",
		httpStatusCode: http.StatusBadRequest,
	},
	ErrBucketAlreadyOwnedByYou: {
		code:           "BucketAlreadyOwnedByYou",
		description:    "Your previous request to create the named bucket succeeded and you already own it.",
		httpStatusCode: http.StatusConflict,
	},
	ErrInvalidDuration: {
		code:           "InvalidDuration",
		description:    "Duration provided in the request is invalid.",
		httpStatusCode: http.StatusBadRequest,
	},
	ErrInvalidBucketObjectLockConfiguration: {
		code:           "InvalidRequest",
		description:    "Bucket is missing ObjectLockConfiguration.",
		httpStatusCode: http.StatusBadRequest,
	},
	ErrObjectLocked: {
		code:           "InvalidRequest",
		description:    "Object is WORM protected and cannot be overwritten.",
		httpStatusCode: http.StatusBadRequest,
	},
	ErrInvalidRetentionDate: {
		code:           "InvalidRequest",
		description:    "Date must be provided in ISO 8601 format.",
		httpStatusCode: http.StatusBadRequest,
	},
	ErrPastObjectLockRetainDate: {
		code:           "InvalidRequest",
		description:    "the retain until date must be in the future.",
		httpStatusCode: http.StatusBadRequest,
	},
	ErrUnknownWORMModeDirective: {
		code:           "InvalidRequest",
		description:    "unknown wormMode directive.",
		httpStatusCode: http.StatusBadRequest,
	},
	ErrObjectLockInvalidHeaders: {
		code:           "InvalidRequest",
		description:    "x-amz-object-lock-retain-until-date and x-amz-object-lock-mode must both be supplied.",
		httpStatusCode: http.StatusBadRequest,
	},
	ErrInvalidCopyPartRange: {
		code:           "InvalidArgument",
		description:    "The x-amz-copy-source-range value must be of the form bytes=first-last where first and last are the zero-based offsets of the first and last bytes to copy.",
		httpStatusCode: http.StatusBadRequest,
	},
	ErrInvalidCopyPartRangeSource: {
		code:           "InvalidArgument",
		description:    "Range specified is not valid for source object.",
		httpStatusCode: http.StatusBadRequest,
	},
	ErrMetadataTooLarge: {
		code:           "InvalidArgument",
		description:    "Your metadata headers exceed the maximum allowed metadata size.",
		httpStatusCode: http.StatusBadRequest,
	},
	ErrUnsupportedMetadata: {
		code:           "InvalidArgument",
		description:    "Your metadata headers are not supported.",
		httpStatusCode: http.StatusBadRequest,
	},
	ErrMaximumExpires: {
		code:           "AuthorizationQueryParametersError",
		description:    "X-Amz-Expires must be less than a week (in seconds); that is, the given X-Amz-Expires must be less than 604800 seconds.",
		httpStatusCode: http.StatusBadRequest,
	},
	ErrNotFound: {
		code:           "NotFound",
		description:    "The resource you request is not found.",
		httpStatusCode: http.StatusNotFound,
	},
}
