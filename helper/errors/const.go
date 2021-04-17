package errors

import (
	"net/http"
)

const (
	ParamInvalidError = "ParamInvalidError"
	ThirdPartyError   = "ThirdPartyError"

	ServiceUnavailableError = "ServiceUnavailableError"
	InternalServerError     = "InternalServerError"
	UnauthorizedError       = "UnauthorizedError"
)

var (
	ErrorParamInvalid = WithMessage(ParamInvalidError, "param is invalid")
	ErrorThirdParty   = WithMessage(ThirdPartyError, "something's wrong with third party service")

	ErrorInternalServer     = WithMessage(InternalServerError, "internal server error")
	ErrorUnauthorized       = WithMessage(UnauthorizedError, "unauthorized")
	ErrorServiceUnavailable = WithMessage(ServiceUnavailableError, "service is unavailable")
)

var ErrorMappings = map[string]int{
	ParamInvalidError: http.StatusBadRequest,
	ThirdPartyError:   http.StatusBadGateway,

	UnauthorizedError:       http.StatusUnauthorized,
	InternalServerError:     http.StatusInternalServerError,
	ServiceUnavailableError: http.StatusServiceUnavailable,
}

func GetStatusCode(err error) int {
	if f, ok := err.(Error); ok {
		return ErrorMappings[f.Code]
	}

	return http.StatusInternalServerError
}
