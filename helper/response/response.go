package response

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/isdzulqor/kraicklist/helper/errors"
)

// Success will write a default template response when returning a success response
func Success(ctx context.Context, w http.ResponseWriter, status int, data interface{}) {
	resp := map[string]interface{}{
		"data":  data,
		"error": nil,
	}
	js, err := json.Marshal(resp)
	if err != nil {
		resp := map[string]interface{}{
			"data":  nil,
			"error": fmt.Sprintf("%s", err),
		}
		js, _ = json.Marshal(resp)
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	w.Write(js)
}

// Failed will write a default template response when returning a failed response
func Failed(ctx context.Context, w http.ResponseWriter, status int, err error) {
	var errResp map[string]interface{}
	if err != nil {
		errCode := "InternalServerError"
		errMsg := err.Error()
		var errData interface{}
		if f, ok := err.(errors.Error); ok {
			errCode = f.Code
			errMsg = f.Message
			errData = f.Data
		}
		errResp = map[string]interface{}{
			"code":    errCode,
			"message": errMsg,
		}
		if errData != nil {
			errResp["data"] = errData
		}
	}
	resp := map[string]interface{}{
		"data":  nil,
		"error": errResp,
	}
	js, _ := json.Marshal(resp)
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	w.Write(js)
}
