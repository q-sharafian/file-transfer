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
	storage  storage.Storage
	isDevEnv bool
}

// Create a new instance of simpleReqHandler.
func NewSimpleReqHandler(auth auth.Auth, storage storage.Storage, logger l.Logger) ReqHandler {
	uploadExpireTime, _ := strconv.Atoi(os.Getenv("UPLOAD_EXPIRE_TIME"))
	downloadExpireTime, _ := strconv.Atoi(os.Getenv("DOWNLOAD_EXPIRE_TIME"))
	isDevEnv := os.Getenv("APP_MODE") == "development"
	return &simpleReqHandler{
		time.Duration(uploadExpireTime) * time.Second,
		time.Duration(downloadExpireTime) * time.Second,
		logger,
		auth,
		storage,
		isDevEnv,
	}
}

// Process An IO (i.e. download/upload) request and response to client
func (req *simpleReqHandler) HandleRequest(ioDetails *ReqDetails) {
	if ioDetails.Type == Upload {
		if ioDetails.Method != http.MethodPost {
			msg := "HTTP method not allowed. (To uploading a file, use POST method)"
			req.prepareErrResponse(ioDetails, http.StatusMethodNotAllowed, msg, msg)
			return
		}
		req.uploadHander(ioDetails)
	} else {
		if ioDetails.Method != http.MethodGet {
			msg := "HTTP method not allowed. (To downloading a file, use GET method)"
			req.prepareErrResponse(ioDetails, http.StatusMethodNotAllowed, msg, msg)
			return
		}
		req.downloadHander(ioDetails)
	}
}

func (rq *simpleReqHandler) downloadHander(req *ReqDetails) {
	downloadReq, err := rq.extractDownloadInfo(req)
	if err != nil {
		msg := fmt.Sprintf("Extracting download info error: %s", err.Error())
		rq.logger.Debugf(msg)
		rq.prepareErrResponse(req, http.StatusBadRequest, msg, "Failed to extract download info")
		return
	}
	allowInfo, err2 := rq.auth.IsAllowedDownload(*downloadReq)
	if err2 != nil {
		msg := fmt.Sprintf("Checking download permission error: %s", err2.Error())
		rq.logger.Debugf(msg)
		rq.prepareErrResponse(req, http.StatusInternalServerError, msg, "Failed to check download permission")
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
			msg := fmt.Sprintf("Creating download link failed: %s", err.Error())
			rq.logger.Debugf(msg)
			rq.prepareErrResponse(req, http.StatusInternalServerError, msg, "Failed to create download link")
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
		msg := fmt.Sprintf("Extracting upload info error: %s", err.Error())
		rq.logger.Debugf(msg)
		rq.prepareErrResponse(req, http.StatusBadRequest, msg, "Failed to extract upload info")
		return
	}
	allowInfo, err2 := rq.auth.IsAllowedUpload(*uploadReq)
	if err2 != nil {
		msg := fmt.Sprintf("Checking upload permission error: %s", err2.Error())
		rq.logger.Debugf(msg)
		rq.prepareErrResponse(req, http.StatusInternalServerError, msg, "Failed to check upload permission")
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
				msg := fmt.Sprintf("Failed to create upload link: can't create uuid: %s", err.Error())
				rq.logger.Debugf(msg)
				rq.prepareErrResponse(req, http.StatusInternalServerError, msg, "Failed to create upload link")
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
				msg := fmt.Sprintf("Creating upload link failed: %s", err.Error())
				rq.logger.Debugf(msg)
				rq.prepareErrResponse(req, http.StatusInternalServerError, msg, "Failed to create upload link")
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

func (rq *simpleReqHandler) prepareErrResponse(req *ReqDetails, statusCode int, devMsg, prodMsg string) {
	msg := prodMsg
	if rq.isDevEnv {
		msg = devMsg
	}
	rq.setResponse(req, downlaodResponse{
		StatusCode: statusCode,
		Message:    msg,
	}, statusCode)
}
