package reqhandler

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"strconv"
	"time"

	"github.com/google/uuid"
	"github.com/q-sharafian/file-transfer/internal/auth"
	"github.com/q-sharafian/file-transfer/internal/common/file"
	"github.com/q-sharafian/file-transfer/internal/common/token"
	"github.com/q-sharafian/file-transfer/internal/storage"
	l "github.com/q-sharafian/file-transfer/pkg/logger"
)

type simpleReqHandler struct {
	// Maximum time after creating an upload link to destroy the link
	uploadExpireTime time.Duration
	// Maximum time after creating a download link to destroy the link
	downloadExpireTime time.Duration
	logger             l.Logger
	// Authentication service
	auth auth.Auth
	// Storage service
	storage storage.Storage
}

// Create a new instance of simpleReqHandler.
func NewSimpleReqHandler(auth auth.Auth, storage storage.Storage, logger l.Logger) ReqHandler {
	uploadExpireTime, _ := strconv.Atoi(os.Getenv("UPLOAD_EXPIRE_TIME"))
	downloadExpireTime, _ := strconv.Atoi(os.Getenv("DOWNLOAD_EXPIRE_TIME"))
	return &simpleReqHandler{
		time.Duration(uploadExpireTime) * time.Second,
		time.Duration(downloadExpireTime) * time.Second,
		logger,
		auth,
		storage,
	}
}

// Process An IO (i.e. download/upload) request and response to client
func (req *simpleReqHandler) HandleRequest(ioDetails *ReqDetails) {
	if ioDetails.Type == Upload {
		if ioDetails.Method != http.MethodPost {
			req.setResponse(ioDetails, downlaodResponse{
				StatusCode: http.StatusMethodNotAllowed,
				Message:    "Method not allowed. (To uploading a file, use POST method)",
			}, http.StatusMethodNotAllowed)
			return
		}
		req.uploadHander(ioDetails)
	} else {
		if ioDetails.Method != http.MethodGet {
			req.setResponse(ioDetails, downlaodResponse{
				StatusCode: http.StatusMethodNotAllowed,
				Message:    "Method not allowed. (To downloading a file, use GET method)",
			}, http.StatusMethodNotAllowed)
			return
		}
		req.downloadHander(ioDetails)
	}
}

func (rq *simpleReqHandler) downloadHander(req *ReqDetails) {
	downloadReq, err := rq.extractDownloadInfo(req)
	if err != nil {
		rq.logger.Debugf("Extracting download info error: %s", err.Error())
		rq.setResponse(req, downlaodResponse{
			StatusCode: http.StatusBadRequest,
			Message:    "Failed to extract download info",
		}, http.StatusBadRequest)
		return
	}
	allowInfo, err2 := rq.auth.IsAllowedDownload(*downloadReq)
	if err2 != nil {
		rq.logger.Debugf("Checking download permission error: %s", err2.Error())
		rq.setResponse(req, downlaodResponse{
			StatusCode: http.StatusInternalServerError,
			Message:    "Failed to check download permission",
		}, http.StatusInternalServerError)
		return
	}

	// Prepare http response to client
	var res downlaodResponse
	res.Tokens2URLs = make(map[string]string)
	for k, IsAllowDown := range allowInfo {
		if !IsAllowDown {
			res.Tokens2URLs[k.String()] = ""
		}
		downloadInfo := storage.DownloadFileInfo{
			FileName:     k.String(),
			DownloadedBy: downloadReq.AuthToken,
			DownloadedAt: time.Now().UTC(),
		}
		url, err := rq.storage.DownloadFile(downloadInfo, rq.downloadExpireTime)
		if err != nil {
			rq.logger.Debugf("Creating download link failed: %s", err.Error())
			rq.setResponse(req, downlaodResponse{
				StatusCode: http.StatusInternalServerError,
				Message:    "Failed to create download link",
			}, http.StatusInternalServerError)
			return
		}
		res.Tokens2URLs[k.String()] = url.String()
	}
	res.Message = "OK"
	res.StatusCode = http.StatusOK
	rq.setResponse(req, res, http.StatusOK)
}

func (rq *simpleReqHandler) uploadHander(req *ReqDetails) {
	uploadReq, err := rq.extractUploadInfo(req)
	if err != nil {
		rq.logger.Debugf("Extracting upload info error: %s", err.Error())
		rq.setResponse(req, downlaodResponse{
			StatusCode: http.StatusBadRequest,
			Message:    "Failed to extract upload info",
		}, http.StatusBadRequest)
		return
	}
	allowInfo, err2 := rq.auth.IsAllowedUpload(*uploadReq)
	if err2 != nil {
		rq.logger.Debugf("Checking upload permission error: %s", err2.Error())
		rq.setResponse(req, downlaodResponse{
			StatusCode: http.StatusInternalServerError,
			Message:    "Failed to check upload permission",
		}, http.StatusInternalServerError)
		return
	}

	// Prepare http response to client
	var res uploadResponse
	res.Tokens2URLs = make(map[string][]string)
	for _, upInfo := range allowInfo {
		if !upInfo.IsAllow {
			continue
		}
		fileType := upInfo.FileType.String()
		if res.Tokens2URLs[fileType] == nil {
			res.Tokens2URLs[fileType] = make([]string, 0)
		}
		for i := uint(0); i < uploadReq.ObjectTypes[file.FileExtension(fileType)]; i++ {
			id, err := uuid.NewRandom()
			if err != nil {
				rq.logger.Debugf("Failed to create upload link: can't create uuid: %s", err.Error())
				rq.setResponse(req, downlaodResponse{
					StatusCode: http.StatusInternalServerError,
					Message:    "Failed to create upload link",
				}, http.StatusInternalServerError)
				return
			}

			uploadInfo := storage.UploadFileInfo{
				FileName:      id.String(),
				UploadedBy:    uploadReq.AuthToken,
				UploadedAt:    time.Now().UTC(),
				FileExtension: upInfo.FileType,
			}
			url, err := rq.storage.UploadFile(uploadInfo, rq.uploadExpireTime)
			if err != nil {
				rq.logger.Debugf("Creating upload link failed: %s", err.Error())
				rq.setResponse(req, downlaodResponse{
					StatusCode: http.StatusInternalServerError,
					Message:    "Failed to create upload link",
				}, http.StatusInternalServerError)
				return
			}
			res.Tokens2URLs[fileType] = append(res.Tokens2URLs[fileType], url.String())
		}
	}
	res.Message = "OK"
	res.StatusCode = http.StatusOK
	rq.setResponse(req, res, http.StatusOK)
}

// Send response to the client
func (io *simpleReqHandler) setResponse(req *ReqDetails, response any, statusCode int) {
	req.Header().Set("Content-Type", "application/json")
	req.WriteHeader(statusCode)
	err := json.NewEncoder(req.ResponseWriter).Encode(response)
	if err != nil {
		io.logger.Errorf("Encoding response error: %s", err.Error())
		req.WriteHeader(http.StatusInternalServerError)
		return
	}
}

// Extract needded info from http request and return
func (ioh *simpleReqHandler) extractDownloadInfo(ioDetails *ReqDetails) (*auth.DownloadAccessReq, error) {
	body, err := io.ReadAll(ioDetails.Body)
	if err != nil {
		return nil, fmt.Errorf("getting http body error: %s", err.Error())
	}
	defer ioDetails.Body.Close()

	var authData struct {
		AuthToken    token.Token   `json:"auth-token" validate:"required"`
		ObjectTokens []token.Token `json:"object-tokens" validate:"required"`
	}
	err = json.Unmarshal(body, &authData)
	if err != nil {
		return nil, fmt.Errorf("unmarshaling http body error: %s", err.Error())
	}
	return &auth.DownloadAccessReq{
		AuthToken:    authData.AuthToken,
		ObjectTokens: authData.ObjectTokens,
	}, nil
}

// Extract needded info from http request and return
func (ioh *simpleReqHandler) extractUploadInfo(ioDetails *ReqDetails) (*auth.UploadAccessReq, error) {
	body, err := io.ReadAll(ioDetails.Request.Body)
	if err != nil {
		return nil, fmt.Errorf("getting http body error: %s", err.Error())
	}
	defer ioDetails.Request.Body.Close()

	var authData struct {
		AuthToken   token.Token                 `json:"auth-token" validate:"required"`
		ObjectTypes map[file.FileExtension]uint `json:"object-types" validate:"required"`
	}
	err = json.Unmarshal(body, &authData)
	if err != nil {
		return nil, fmt.Errorf("unmarshaling http body error: %s", err.Error())
	}
	ioh.logger.Debugf("Extracted upload info: %+v", authData)
	return &auth.UploadAccessReq{
		AuthToken:   authData.AuthToken,
		ObjectTypes: authData.ObjectTypes,
	}, nil
}
