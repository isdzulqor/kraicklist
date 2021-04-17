package logging

import (
	"context"
	"fmt"
	"log"
	"os"
	"runtime"
	"runtime/debug"
)

const (
	requestIDKey = 0

	Debug = "DEBUG"
	Info  = "INFO"
	Warn  = "WARN"
	Error = "ERROR"
	Fatal = "FATAL"
)

var (
	activeLogLevel = ""

	logLevel = map[string]int{
		Debug: 4,
		Info:  3,
		Warn:  2,
		Error: 1,
		Fatal: 0,
	}
)

func Init(logLevel string) {
	activeLogLevel = logLevel
}

func getActiveLogLevel() int {
	return logLevel[activeLogLevel]
}

func extractReqID(ctx context.Context) string {
	if value, ok := ctx.Value(requestIDKey).(string); ok {
		return value
	}
	return ""
}

func DebugContext(ctx context.Context, format string, v ...interface{}) {
	if getActiveLogLevel() >= logLevel[Debug] {
		message := fmt.Sprintf("DEBUG: "+getFunctionCaller()+format, v...)
		if reqID := extractReqID(ctx); reqID != "" {
			message = fmt.Sprintf("DEBUG: ReqID "+extractReqID(ctx)+" - "+getFunctionCaller()+format, v...)
		}
		log.Print(message)
	}
}

// InfoContext prints Info message to logs
func InfoContext(ctx context.Context, format string, v ...interface{}) {
	if getActiveLogLevel() >= logLevel[Info] {
		message := fmt.Sprintf("\033[32mINFO: \033[0m"+format, v...)
		if reqID := extractReqID(ctx); reqID != "" {
			message = fmt.Sprintf("\033[32mINFO: \033[0mReqID "+extractReqID(ctx)+" - "+format, v...)
		}
		log.Print(message)
	}
}

// WarnContext prints warning message to logs
func WarnContext(ctx context.Context, format string, v ...interface{}) {
	if getActiveLogLevel() >= logLevel[Warn] {
		message := fmt.Sprintf("\033[33mWARN: \033[0m"+format, v...)
		if reqID := extractReqID(ctx); reqID != "" {
			message = fmt.Sprintf("\033[33mWARN: \033[0mReqID "+extractReqID(ctx)+" - "+format, v...)
		}
		log.Print(message)
	}
}

// ErrContext prints warning message to logs
func ErrContext(ctx context.Context, format string, v ...interface{}) {
	if getActiveLogLevel() >= logLevel[Error] {
		message := []interface{}{fmt.Sprintf("\033[31mERROR: \033[0m"+getFunctionCaller()+format, v...)}
		if reqID := extractReqID(ctx); reqID != "" {
			message = []interface{}{fmt.Sprintf("\033[31mERROR: \033[0mReqID "+extractReqID(ctx)+" - "+getFunctionCaller()+format, v...)}
		}
		message = append(message, "\n", string(debug.Stack()))
		log.Print(message...)
	}
}

func ErrContextNoStackTrace(ctx context.Context, format string, v ...interface{}) {
	if getActiveLogLevel() >= logLevel[Error] {
		message := []interface{}{fmt.Sprintf("\033[31mERROR: \033[0m"+getFunctionCaller()+format, v...)}
		if reqID := extractReqID(ctx); reqID != "" {
			message = []interface{}{fmt.Sprintf("\033[31mERROR: \033[0mReqID "+extractReqID(ctx)+" - "+getFunctionCaller()+format, v...)}
		}
		log.Print(message...)
	}
}

// FatalContext calls Err and then os.Exit(1)
func FatalContext(ctx context.Context, format string, v ...interface{}) {
	ErrContext(ctx, format, v...)
	os.Exit(1)
}

func WithRequestIDContext(ctx context.Context, reqID string) context.Context {
	return context.WithValue(ctx, requestIDKey, reqID)
}

func getFunctionCaller() string {
	pc, _, _, ok := runtime.Caller(2)
	details := runtime.FuncForPC(pc)
	if ok && details != nil {
		return fmt.Sprintf("%s ", details.Name())
	}
	return ""
}
