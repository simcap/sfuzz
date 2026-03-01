package main

import (
	"log/slog"
	"net/http"
	"os"

	"github.com/simcap/sfuzz/demoapi"
)

var port = ":8080"

func main() {
	logger := slog.New(slog.NewTextHandler(os.Stdout, nil))
	handler := demoapi.New(demoapi.WithLogger(logger))

	logger.Info("Starting Demo API", "port", port)
	if err := http.ListenAndServe(port, handler); err != nil {
		logger.Error(err.Error())
	}
}
