package auth

import "github.com/q-sharafian/file-transfer/pkg/error"

// Its purpose is just for testing
type dummyAuth struct {
}

func NewDummyAuth() Auth {
	return &dummyAuth{}
}

func (d *dummyAuth) IsAllowedDownload(accessInfo DownloadAccessReq) (allowDownload, *error.Error) {
	allowDownload := make(allowDownload)
	for _, t := range accessInfo.ObjectTokens {
		allowDownload[t] = true
	}
	return allowDownload, nil
}

// IsAllowedUpload implements Auth.
func (d *dummyAuth) IsAllowedUpload(accessInfo UploadAccessReq) ([]allowType, *error.Error) {
	allowTypes := make([]allowType, 0, len(accessInfo.ObjectTypes))
	for fe := range accessInfo.ObjectTypes {
		allowTypes = append(allowTypes, allowType{
			FileType: fe,
			IsAllow:  true,
			MaxSize:  10240,
		})
	}
	return allowTypes, nil
}
