package storage

import (
	"net/url"
	"time"

	"github.com/q-sharafian/file-transfer/internal/common/file"
	"github.com/q-sharafian/file-transfer/internal/common/metadata"
	"github.com/q-sharafian/file-transfer/internal/common/token"
)

type DownloadFileInfo struct {
	// The name of the file in the storage. If it has an extension, add it to the filename.
	FileName string
	// token represents the user who downloaded the file
	DownloadedBy token.Token
	// Time the file is downloaded
	DownloadedAt time.Time
}
type UploadFileInfo struct {
	// Filename without extension. The name of the file in the storage will be renamed to this name
	FileName string
	file.FileExtension
	metadata.Metadata
	// token represents the user who uploaded the file. (e.g. auth token)
	UploadedBy token.Token
	// Time the file is uploaded
	UploadedAt time.Time
}

// Each implementation must create a one-time link to download/upload file with
// a maximum time to use the link. The link should be expired after the expiration time.
// Also manage file metadata. (e.g. removing sensitive metadata during downloading)
type Storage interface {
	// Create a link to upload one file and expire the link after the expiration time
	UploadFile(fileInfo UploadFileInfo, expireTime time.Duration) (url.URL, error)
	// Create a link to download one file and expire the link after the expiration time.
	DownloadFile(fileInfo DownloadFileInfo, expireTime time.Duration) (url.URL, error)
}
