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
}{
	ErrInvalidMetadataURL:   New("INVALID_METADATA_URL", "invalid post metadata url", http.StatusBadRequest),
	ErrLoadMetadata:         New("FAILED_LOAD_METADATA", "failed to load metadata", http.StatusInternalServerError),
	ErrMissingAuthHash:      New("ADMIN_AUTH_CODE_HASH_NOT_FOUND", "auth hash (salt) is empty", http.StatusInternalServerError),
	ErrMissingAuthAdminCode: New("ADMIN_AUTH_CODE_NOT_FOUND", "auth admin code is empty", http.StatusInternalServerError),
}

var Blockchain = struct {
	ErrRPCURLMissing      *Error
	ErrContractNotFound   *Error
	ErrNoAvailableRelayer *Error
}{
	ErrRPCURLMissing:      New("BLOCKCHAIN_RPC_URL_MISSING", "blockchain rpc url is empty", http.StatusInternalServerError),
	ErrContractNotFound:   New("CONTRACT_NOT_FOUND", "contract not found", http.StatusNotFound),
	ErrNoAvailableRelayer: New("NO_AVAILABLE_RELAYER", "no available relayer", http.StatusInternalServerError),
}
