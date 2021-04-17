package health

import (
	"fmt"
	"kraicklist/helper/errors"
	"kraicklist/helper/response"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"
)

const (
	healthHeaderToken   = "x-health-token"
	healthyMessage      = "service is alive"
	unHealthyMessage    = "service is not healthy"
	shuttingDownMessage = "service is shutting down"
)

var (
	isShuttingDown bool

	defaultHealthyResponse = func() healthResponse {
		return healthResponse{
			Message: healthyMessage,
		}
	}
)

type healthResponse struct {
	Message      string        `json:"message"`
	Persistences *Persistences `json:"persistences,omitempty"`
}

// HealthHandler contains list of persistences those need to be checked
// and also the delay duration before the service will be killed
type HealthHandler struct {
	persistences          *Persistences
	shutdownDelayDuration time.Duration

	token *string
}

func NewHealthHandler(persistences *Persistences, shutdownDelayDuration time.Duration) (*HealthHandler, error) {
	if persistences == nil {
		return nil, fmt.Errorf("persistences can't be nil")
	}

	handler := &HealthHandler{
		persistences:          persistences,
		shutdownDelayDuration: shutdownDelayDuration,
	}
	handler.gracefulShutdown()
	return handler, nil
}

func (h *HealthHandler) WithToken(token string) {
	h.token = &token
}

func (h HealthHandler) IsShuttingDown() (bool, *healthResponse) {
	if isShuttingDown {
		resp := healthResponse{}
		resp.Message = shuttingDownMessage
		return true, &resp
	}
	return false, nil
}

func (h HealthHandler) gracefulShutdown() {
	stopChan := make(chan os.Signal)
	signal.Notify(stopChan, os.Interrupt, syscall.SIGTERM)
	go h.listenToSigTerm(stopChan)
}

func (h HealthHandler) listenToSigTerm(stopChan chan os.Signal) {
	<-stopChan
	log.Println("Shutting down service... Will be killed on", h.shutdownDelayDuration)
	isShuttingDown = true

	time.Sleep(h.shutdownDelayDuration)
	log.Println("Bye..")
	os.Exit(0)
}

// GetHealth is an handler for health endpoint
func (h HealthHandler) GetHealth(w http.ResponseWriter, r *http.Request) {
	resp := defaultHealthyResponse()
	ctx := r.Context()

	// if use WithToken, the handler will validate token from incoming request
	// and if it doesn't match, will throw unauthorized error
	if h.token != nil && r.Header.Get(healthHeaderToken) != *h.token {
		response.Failed(ctx, w, http.StatusUnauthorized, errors.ErrorUnauthorized)
		return
	}

	if isShuttingDown, _ := h.IsShuttingDown(); isShuttingDown {
		response.Failed(ctx, w, http.StatusServiceUnavailable,
			errors.ErrorServiceUnavailable.AppendMessage(shuttingDownMessage))
		return
	}

	resp.Persistences = h.persistences
	if h.persistences.Ping() {
		response.Success(ctx, w, http.StatusOK, resp)
		return
	}
	resp.Message = unHealthyMessage
	response.Failed(ctx, w, http.StatusServiceUnavailable, errors.ErrorServiceUnavailable.SetData(resp))
}
