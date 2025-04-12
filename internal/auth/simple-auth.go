package auth

import (
	"context"
	"time"

	"github.com/q-sharafian/file-transfer/internal/common/file"
	"github.com/q-sharafian/file-transfer/internal/common/token"
	e "github.com/q-sharafian/file-transfer/pkg/error"
	l "github.com/q-sharafian/file-transfer/pkg/logger"
	pbAuth "github.com/q-sharafian/file-transfer/pkg/pb/auth"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

// This authentication method create a gRPC connection to the auth server and
// make the authentication request.
type simpleAuth struct {
	authClient pbAuth.AuthClient
	// Maximum time needed to send a request to auth server and receive its response
	maxQueryTime time.Duration
	logger       l.Logger
}

// Connect to the authentication server to query whether an auth query is allowed or not.
// Any query time must be less than maxQueryTime or it will be rejected.
func NewSimpleAuth(serverAddr string, maxQueryTime time.Duration, logger l.Logger) Auth {
	logger.Infof("Connecting to gRPC server with address %s", serverAddr)
	conn, err := grpc.NewClient(serverAddr, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		logger.Panicf("Failed to create gRPC connection to auth server: %s", err.Error())
	}
	authClient := pbAuth.NewAuthClient(conn)
	return &simpleAuth{
		authClient,
		maxQueryTime,
		logger,
	}
}

func (s *simpleAuth) IsAllowedDownload(accessInfo DownloadAccessReq) (allowDownload, *e.Error) {
	ctx, cancel := context.WithTimeout(context.Background(), s.maxQueryTime)
	defer cancel()
	objectTokens := tokens2Strings(accessInfo.ObjectTokens)

	result, err := s.authClient.IsAllowedDownload(ctx, &pbAuth.DownloadAccessReq{
		AuthToken:    accessInfo.AuthToken.String(),
		ObjectTokens: objectTokens,
	})
	if err != nil {
		return nil, e.NewErrorP("Failed to check download access privileges: %s", ErrInternal, err.Error())
	}
	switch result.GetStatusCode() {
	case pbAuth.StatusCode_ErrForbidden:
		return nil, e.NewErrorP("Download access is forbidden for this specified user: %s", ErrForbidden, result.GetErrmsg())
	case pbAuth.StatusCode_ErrUnauthorized:
		return nil, e.NewErrorP("There's not any matched user with this auth token: %s", ErrUnauthorized, result.GetErrmsg())
	case pbAuth.StatusCode_ErrInternal:
		return nil, e.NewErrorP("Failed to check download access privileges: %s", ErrInternal, result.GetErrmsg())
	case pbAuth.StatusCode_OK:
		allowDownload := make(allowDownload)
		for k, v := range result.GetFiles() {
			allowDownload[token.Token(k)] = v
		}
		return allowDownload, nil
	default:
		s.logger.Panicf("Unknown status code %d: %s", result.GetStatusCode(), result.GetErrmsg())
		return nil, nil
	}
}

func (s *simpleAuth) IsAllowedUpload(accessInfo UploadAccessReq) ([]allowType, *e.Error) {
	ctx, cancel := context.WithTimeout(context.Background(), s.maxQueryTime)
	defer cancel()
	fileTypes := make(map[string]int64, 0)
	for k, v := range accessInfo.ObjectTypes {
		fileTypes[k.String()] = int64(v)
	}

	result, err := s.authClient.IsAllowedUpload(ctx, &pbAuth.UploadAccessReq{
		AuthToken:   accessInfo.AuthToken.String(),
		ObjectTypes: fileTypes,
	})
	if err != nil {
		return nil, e.NewErrorP("failed to check upload access privileges: %s", ErrInternal, err.Error())
	}
	switch result.GetStatusCode() {
	case pbAuth.StatusCode_ErrForbidden:
		return nil, e.NewErrorP("Upload access is forbidden for this specified user: %s", ErrForbidden, result.GetErrmsg())
	case pbAuth.StatusCode_ErrUnauthorized:
		return nil, e.NewErrorP("There's not any matched user with this auth token: %s", ErrUnauthorized, result.GetErrmsg())
	case pbAuth.StatusCode_ErrInternal:
		return nil, e.NewErrorP("failed to check upload access privileges: %s", ErrInternal, result.GetErrmsg())
	case pbAuth.StatusCode_OK:
		var allowTypes []allowType
		for _, v := range result.GetFileTypes() {
			allowTypes = append(allowTypes, allowType{
				FileType: file.FileExtension(v.FileType),
				MaxSize:  v.MaxSize,
				IsAllow:  v.IsAllow,
			})
		}
		return allowTypes, nil
	default:
		s.logger.Panicf("Unknown status code %d: %s", result.GetStatusCode(), result.GetErrmsg())
		return nil, nil
	}
}

func tokens2Strings(tokens []token.Token) []string {
	var strs []string
	for _, token := range tokens {
		strs = append(strs, token.String())
	}
	return strs
}
