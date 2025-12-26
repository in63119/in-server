package apperr

import (
	"errors"
	"fmt"
	"net/http"
)

type Error struct {
	Code    string
	Message string
	Status  int
	Err     error
}

func (e *Error) Error() string {
	if e == nil {
		return ""
	}
	if e.Err != nil {
		return fmt.Sprintf("%s: %s: %v", e.Code, e.Message, e.Err)
	}
	return fmt.Sprintf("%s: %s", e.Code, e.Message)
}

func (e *Error) Unwrap() error { return e.Err }

func New(code, message string, status int) *Error {
	return &Error{Code: code, Message: message, Status: status}
}

func Wrap(err error, code, message string, status int) *Error {
	return &Error{Code: code, Message: message, Status: status, Err: err}
}

func (e *Error) Is(target error) bool {
	var t *Error
	if !errors.As(target, &t) {
		return false
	}
	return e.Code != "" && e.Code == t.Code
}

var System = struct {
	ErrInvalidMetadataURL   *Error
	ErrLoadMetadata         *Error
	ErrMissingAuthHash      *Error
	ErrMissingAuthAdminCode *Error
	ErrNotImplemented       *Error
}{
	ErrInvalidMetadataURL:   New("INVALID_METADATA_URL", "invalid post metadata url", http.StatusBadRequest),
	ErrLoadMetadata:         New("FAILED_LOAD_METADATA", "failed to load metadata", http.StatusInternalServerError),
	ErrMissingAuthHash:      New("ADMIN_AUTH_CODE_HASH_NOT_FOUND", "auth hash (salt) is empty", http.StatusInternalServerError),
	ErrMissingAuthAdminCode: New("ADMIN_AUTH_CODE_NOT_FOUND", "auth admin code is empty", http.StatusInternalServerError),
	ErrNotImplemented:       New("NOT_IMPLEMENTED", "not implemented", http.StatusNotImplemented),
}

var Blockchain = struct {
	ErrRPCURLMissing      *Error
	ErrContractNotFound   *Error
	ErrNoAvailableRelayer *Error
	ErrInvalidWallet      *Error
}{
	ErrRPCURLMissing:      New("BLOCKCHAIN_RPC_URL_MISSING", "blockchain rpc url is empty", http.StatusInternalServerError),
	ErrContractNotFound:   New("CONTRACT_NOT_FOUND", "contract not found", http.StatusNotFound),
	ErrNoAvailableRelayer: New("NO_AVAILABLE_RELAYER", "no available relayer", http.StatusInternalServerError),
	ErrInvalidWallet:      New("INVALID_WALLET", "invalid wallet", http.StatusBadRequest),
}

var Visitors = struct {
	ErrInvalidBody *Error
	ErrCheckVisit  *Error
	ErrAddVisit    *Error
	ErrVisitCount  *Error
}{
	ErrInvalidBody: New("INVALID_BODY", "url is required", http.StatusForbidden),
	ErrCheckVisit:  New("FAILED_TO_CHECK_VISIT", "failed to check visit", http.StatusInternalServerError),
	ErrAddVisit:    New("FAILED_TO_ADD_VISIT", "failed to add visit", http.StatusInternalServerError),
	ErrVisitCount:  New("FAILED_TO_GET_VISIT_COUNT", "failed to get visit count", http.StatusInternalServerError),
}

var Subscriber = struct {
	ErrGetSubscribers *Error
	ErrInvalidBody    *Error
	ErrCreate         *Error
	ErrAlreadyExists  *Error
}{
	ErrGetSubscribers: New("FAILED_TO_GET_SUBSCRIBERS", "failed to get subscribers", http.StatusInternalServerError),
	ErrInvalidBody:    New("INVALID_BODY", "invalid request body", http.StatusBadRequest),
	ErrCreate:         New("FAILED_TO_SUBSCRIBE", "failed to subscribe", http.StatusInternalServerError),
	ErrAlreadyExists:  New("ALREADY_EXISTS_SUBSCRIBER", "subscriber already exists", http.StatusConflict),
}

var Post = struct {
	ErrInvalidBody      *Error
	ErrAdminCodeMissing *Error
	ErrInvalidAdminCode *Error
	ErrDuplicatePost    *Error
	ErrInvalidRequest   *Error
	ErrS3BucketMissing  *Error
	ErrUploadMetadata   *Error
	ErrPublishFailed    *Error
}{
	ErrInvalidBody:      New("INVALID_BODY", "invalid request body", http.StatusBadRequest),
	ErrAdminCodeMissing: New("ADMIN_CODE_MISSING", "admin code is required", http.StatusBadRequest),
	ErrInvalidAdminCode: New("INVALID_ADMIN_CODE", "invalid admin code", http.StatusBadRequest),
	ErrDuplicatePost:    New("DUPLICATE_POST", "a post with the same slug already exists", http.StatusConflict),
	ErrInvalidRequest:   New("INVALID_REQUEST", "invalid request", http.StatusBadRequest),
	ErrS3BucketMissing:  New("AWS_S3_BUCKET_MISSING", "aws s3 bucket is empty", http.StatusInternalServerError),
	ErrUploadMetadata:   New("FAILED_UPLOAD_METADATA", "failed to upload metadata", http.StatusInternalServerError),
	ErrPublishFailed:    New("FAILED_PUBLISH_POST", "failed to publish post", http.StatusInternalServerError),
}

var Email = struct {
	ErrInvalidBody        *Error
	ErrFailedSendingEmail *Error
	ErrClaimPinCode       *Error
	ErrVerifyPinCode      *Error
	ErrInvalidEmail       *Error
}{
	ErrInvalidBody:        New("INVALID_BODY", "invalid request body", http.StatusBadRequest),
	ErrFailedSendingEmail: New("FAILED_SENDING_EMAIL", "failed sending email", http.StatusInternalServerError),
	ErrClaimPinCode:       New("FAILED_TO_CLAIM_PIN_CODE", "failed to claim pin code", http.StatusInternalServerError),
	ErrVerifyPinCode:      New("FAILED_TO_VERIFY_PIN_CODE", "failed to verify pin code", http.StatusInternalServerError),
	ErrInvalidEmail:       New("INVALID_EMAIL", "invalid email address", http.StatusBadRequest),
}
