package metadata

import (
	"fmt"
	"time"

	"github.com/q-sharafian/file-transfer/internal/common/token"
)

// list of all metadata
type Metadata map[string]string

const (
	// Real name of the file without any extension
	fileRealName = "RealName"
	uploadedAt   = "UploadedAt"
	// token represents the user who uploaded the file. (e.g. auth token)
	uploadedBy = "CreatedBy"
	// token represents the user who downloaded the file
	downloadedBy = "DownloadedBy"
	// Time the file is downloaded
	downloadedAt = "DownloadedAt"
)

type RequiredDownloadMetadata struct {
	DownloadedBy token.Token
}

// Remove all metadata, except the one that is needed to download the file
func (m *Metadata) PrepareDownloadMetadata(downloadBy token.Token) {
	newMetadata := Metadata{}

	newMetadata[fileRealName] = (*m)[fileRealName]
	newMetadata[uploadedAt] = (*m)[uploadedAt]
	newMetadata[uploadedBy] = (*m)[uploadedBy]
	newMetadata[downloadedBy] = downloadBy.String()
	newMetadata[downloadedAt] = fmt.Sprintf("%d", time.Now().UTC().Unix())

	*m = newMetadata
}

func (m *Metadata) PrepareUploadMetadata(uploadBy token.Token, realFileName string) {
	newMetadata := Metadata{}

	newMetadata[fileRealName] = realFileName
	newMetadata[uploadedAt] = fmt.Sprintf("%d", time.Now().UTC().Unix())
	newMetadata[uploadedBy] = uploadBy.String()

	*m = newMetadata
}
