//go:generate go tool oapi-codegen -generate std-http,types -package gen -o gen/server.go spec.yaml
package demoapi

import (
	"encoding/json"
	"log/slog"
	"net/http"

	openapi_types "github.com/oapi-codegen/runtime/types"
	"github.com/simcap/sfuzz/demoapi/gen"
)

type option func(*server)

func New(options ...option) http.Handler {
	s := &server{
		logger: slog.New(slog.DiscardHandler),
	}
	for _, opt := range options {
		opt(s)
	}

	serverOptions := gen.StdHTTPServerOptions{
		Middlewares:      []gen.MiddlewareFunc{logRequestMiddleware(s.logger)},
		ErrorHandlerFunc: s.errorHandler,
	}

	return gen.HandlerWithOptions(s, serverOptions)
}

type server struct {
	logger       *slog.Logger
	errorHandler func(w http.ResponseWriter, r *http.Request, err error)
}

func (s server) PostBooks(w http.ResponseWriter, r *http.Request) {
	var request gen.BookRequest
	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
		s.errorHandler(w, r, &gen.InvalidParamFormatError{Err: err})
		return
	}
}

func (s server) DeleteBooksId(w http.ResponseWriter, r *http.Request, id openapi_types.UUID) {
	if err := json.NewEncoder(w).Encode(gen.Book{Id: &id}); err != nil {
		s.errorHandler(w, r, &gen.InvalidParamFormatError{Err: err})
		return
	}
}

func (s server) GetBooksId(w http.ResponseWriter, r *http.Request, id openapi_types.UUID) {
	if err := json.NewEncoder(w).Encode(gen.Book{Id: &id}); err != nil {
		s.errorHandler(w, r, &gen.InvalidParamFormatError{Err: err})
		return
	}
}

func (s server) PostCustomers(w http.ResponseWriter, r *http.Request) {
	var request gen.CustomerRequest
	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
		s.errorHandler(w, r, &gen.InvalidParamFormatError{Err: err})
		return
	}
}

func (s server) DeleteCustomersId(w http.ResponseWriter, r *http.Request, id openapi_types.UUID) {
	if err := json.NewEncoder(w).Encode(&gen.Customer{Id: &id}); err != nil {
		s.errorHandler(w, r, &gen.InvalidParamFormatError{Err: err})
		return
	}
}

func (s server) GetCustomersId(w http.ResponseWriter, r *http.Request, id openapi_types.UUID) {
	if err := json.NewEncoder(w).Encode(&gen.Customer{Id: &id}); err != nil {
		s.errorHandler(w, r, &gen.InvalidParamFormatError{Err: err})
		return
	}
}

func logRequestMiddleware(log *slog.Logger) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			log.Info("request", "path", r.URL.Path, "method", r.Method)
			next.ServeHTTP(w, r)
		})
	}
}

func WithLogger(log *slog.Logger) option {
	return func(s *server) {
		s.logger = log
	}
}
