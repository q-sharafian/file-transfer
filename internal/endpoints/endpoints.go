package endpoints

import (
	"os"

	reqh "github.com/q-sharafian/file-transfer/internal/reqhandler"
	s "github.com/q-sharafian/file-transfer/internal/server"
	l "github.com/q-sharafian/file-transfer/pkg/logger"
)

// Initialize permanent endpoints.
func InitEndpoints(server s.Server, logger l.Logger, reqHandler reqh.ReqHandler) {
	logger.Info("Initializing endpoints")
	uploadPath := os.Getenv("UPLOAD_PATH")
	downloadPath := os.Getenv("DOWNLOAD_PATH")

	server.AddHandler(uploadPath, func(w s.ResponseWriter, r *s.Request) {
		// Method 1
		reqHandler.HandleRequest(&reqh.ReqDetails{
			Type: reqh.Upload, ResponseWriter: w, Request: r,
		})

		// Method 2
		// type data struct {
		// 	s.ResponseWriter
		// 	*s.Request
		// }
		// dataChann := make(chan data, 1)
		// defer close(dataChann)
		// var wg sync.WaitGroup
		// wg.Add(1)
		// go func() {
		// 	defer wg.Done()
		// 	rr := <-dataChann
		// 	if rr.Request.Method != http.MethodPost {

		// 	}
		// 	reqHandler.HandleRequest(&reqh.ReqDetails{
		// 		// Type: reqh.Upload, ResponseWriter: rr.ResponseWriter, Request: &rr.Request,
		// 		Type: reqh.Upload, ResponseWriter: rr.ResponseWriter, Request: rr.Request,
		// 	})
		// }()
		// dataChann <- data{ResponseWriter: w, Request: r}
		// wg.Wait()
	})

	server.AddHandler(downloadPath, func(w s.ResponseWriter, r *s.Request) {
		reqHandler.HandleRequest(&reqh.ReqDetails{
			Type: reqh.Download, ResponseWriter: w, Request: r,
		})
	})
}
