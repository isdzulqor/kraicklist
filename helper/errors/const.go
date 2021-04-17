package errors

import (
	"fmt"
	"net/http"
)

const (
	ParamInvalidError = "ParamInvalidError"
	ThirdPartyError   = "ThirdPartyError"
)

var (
	ErrorParamInvalid = WithMessage(ParamInvalidError, "param is invalid")
	ErrorThirdParty   = WithMessage(ThirdPartyError, "something's wrong with third party service")

	ErrorInternalServer = fmt.Errorf(http.StatusText(http.StatusInternalServerError))
)

var ErrorMappings = map[string]int{
	ParamInvalidError: http.StatusBadRequest,
	ThirdPartyError:   http.StatusBadGateway,
}

func GetStatusCode(err error) int {
	if f, ok := err.(Error); ok {
		return ErrorMappings[f.Code]
	}

	return http.StatusInternalServerError
}
