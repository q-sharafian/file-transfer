/*
Responsible for authentication and authorization
*/
package auth

import (
	"github.com/q-sharafian/file-transfer/internal/common/file"
	"github.com/q-sharafian/file-transfer/internal/common/token"
	e "github.com/q-sharafian/file-transfer/pkg/error"
)

type UploadAccessReq struct {
	// authentication token. It maybe jwt or something that is agreed upon between two parties.
	AuthToken token.Token
	// List of tokens, each representing a file type. The value of each key is the number
	// of files we want to upload with that extension
	ObjectTypes map[file.FileExtension]uint
}

type DownloadAccessReq struct {
	// authentication token. It maybe jwt or something that is agreed upon between two parties.
	AuthToken token.Token
	// list of tokens that each represents a file
	ObjectTokens []token.Token
}

type allowType struct {
	FileType file.FileExtension
	IsAllow  bool
	// Maximum size of the file with with FileType in Kbytes
	MaxSize uint64
}

// Specified which files are allowed to be downloaded
type allowDownload map[token.Token]bool

type errTypes int

const (
	// An internal error could be database error, network error, etc
	ErrInternal errTypes = iota
	// There's not any matched user with this auth token
	ErrUnauthorized
	// User with this token exists but can't upload/download any object.
	// (e.g., it has not download/upload permission or it's disabled)
	ErrForbidden
)

type Auth interface {
	// Check if each file specified in the input is allowed to be downloaded by specified
	// client that has 'AuthToken'.
	//
	// Possible error codes:
	// ErrInternal- ErrForbidden- ErrUnauthorized
	IsAllowedDownload(accessInfo DownloadAccessReq) (allowDownload, *e.Error)

	// Check if the file type specified in the input is allowed to be uploaded and what
	// is the maximum size of each type that could be uploaded then, return the result. these details
	// are only usesable for the client with 'AuthToken' not anyone else.
	//
	// Possible error codes:
	// ErrInternal- ErrForbidden- ErrUnauthorized
	IsAllowedUpload(accessInfo UploadAccessReq) ([]allowType, *e.Error)
}
