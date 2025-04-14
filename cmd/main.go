package main

import (
	"fmt"
	"os"
	"strconv"
	"time"

	"github.com/joho/godotenv"
	"github.com/q-sharafian/file-transfer/internal/auth"
	"github.com/q-sharafian/file-transfer/internal/endpoints"
	"github.com/q-sharafian/file-transfer/internal/reqhandler"
	"github.com/q-sharafian/file-transfer/internal/server"
	"github.com/q-sharafian/file-transfer/internal/storage"
	l "github.com/q-sharafian/file-transfer/pkg/logger"
)

func main() {
	logger := l.NewSLogger(l.Info, nil, os.Stdout)
	var appMode = os.Getenv("APP_MODE")
	if appMode == "development" || appMode == "" {
		if err := godotenv.Load(".env"); err != nil {
			panic(fmt.Sprintf("Error loading .env file: %s", err.Error()))
		}
		logger.ChangeLogLevel(l.Debug)
	}

	maxQueryTime, err := strconv.Atoi(os.Getenv("AUTH_QUERY_MAX_TIME"))
	if err != nil {
		logger.Panicf("Failed to parse AUTH_QUERY_MAX_TIME: %s", err.Error())
	}
	authService := auth.NewSimpleAuth(os.Getenv("AUTH_SERVER_ADDR"), time.Duration(maxQueryTime)*time.Second, logger)
	// authService := auth.NewDummyAuth()
	storageService := storage.NewS3Storage(logger)
	server := server.NewSimpleServer(logger)
	requestHandler := reqhandler.NewSimpleReqHandler(authService, storageService, logger)
	endpoints.InitEndpoints(server, logger, requestHandler)

	// Keep the main function running
	select {}
}
