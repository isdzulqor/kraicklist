package infra

import (
	"context"
	"net/http"
	"time"

	"github.com/isdzulqor/kraicklist/helper/errors"
	"github.com/isdzulqor/kraicklist/helper/health"
	"github.com/isdzulqor/kraicklist/helper/logging"
	"github.com/isdzulqor/kraicklist/helper/reqid"
	"github.com/isdzulqor/kraicklist/helper/response"
	"github.com/isdzulqor/kraicklist/helper/uuid"
)

func RecoverHandler(next http.Handler) http.Handler {
	fn := func(w http.ResponseWriter, r *http.Request) {
		defer func() {
			if err := recover(); err != nil {
				ctx := r.Context()

				logging.ErrContext(ctx, "%v", err)
				response.Failed(ctx, w, http.StatusInternalServerError, errors.ErrorInternalServer)
			}
		}()
		next.ServeHTTP(w, r)
	}
	return http.HandlerFunc(fn)
}

func LoggingHandler(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		reqID := r.Header.Get(reqid.RequestIDHeader)
		if reqID == "" {
			reqID = uuid.UUIDv4()
		}

		ctx := logging.WithRequestIDContext(context.Background(), reqID)

		start := time.Now()
		logging.InfoContext(ctx, "Requesting "+r.Method+" "+r.URL.Path)

		resp := responseLogger(w)
		resp.Header().Set(reqid.ResponseReqIDHeader, r.Header.Get(reqid.RequestIDHeader))
		next.ServeHTTP(resp, r.WithContext(ctx))

		if resp.statusCode >= 500 {
			logging.ErrContextNoStackTrace(ctx, "Response "+r.Method+" "+r.URL.Path+" took "+time.Since(start).String())
			return
		}

		if resp.statusCode >= 400 {
			logging.WarnContext(ctx, "Response "+r.Method+" "+r.URL.Path+" took "+time.Since(start).String())
			return

		}
		logging.InfoContext(ctx, "Response "+r.Method+" "+r.URL.Path+" took "+time.Since(start).String())
	})
}

// CheckShuttingDown works as middleware
// will check shuttingDown status (already received SIGTERM signal) on each incoming request
func CheckShuttingDown(healthHandler health.HealthHandler) func(next http.Handler) http.Handler {
	return func(h http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if isShuttingDown, resp := healthHandler.IsShuttingDown(); isShuttingDown {
				response.Failed(r.Context(), w, http.StatusServiceUnavailable,
					errors.ErrorServiceUnavailable.AppendMessage(resp.Message))
				return
			}
			h.ServeHTTP(w, r)
		})
	}
}

type responseLog struct {
	http.ResponseWriter
	statusCode int
	body       []byte
}

func responseLogger(w http.ResponseWriter) *responseLog {
	// default status is 200
	return &responseLog{ResponseWriter: w, statusCode: http.StatusOK, body: nil}
}

// WriteHeader override status and set response value
func (c *responseLog) WriteHeader(status int) {
	c.statusCode = status
	c.ResponseWriter.WriteHeader(status)
}

func (c *responseLog) Write(b []byte) (int, error) {
	c.body = b
	return c.ResponseWriter.Write(b)
}
