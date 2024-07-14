package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"strings"
	"time"
)

const LOCAL_LOGS = 0
const GCLOUD_LOGS = 1

type Logger struct {
	mode    int
	project string
	encoder *json.Encoder
}

type GcpLogEntry struct {
	Severity    string            `json:"severity"`
	Timestamp   time.Time         `json:"timestamp"`
	Message     interface{}       `json:"message,omitempty"`
	TextPayload interface{}       `json:"textPayload,omitempty"`
	Labels      map[string]string `json:"logging.googleapis.com/labels,omitempty"`
	TraceID     string            `json:"logging.googleapis.com/trace,omitempty"`
	SpanID      string            `json:"logging.googleapis.com/spanId,omitempty"`
	HttpRequest HttpRequestLog    `json:"httpRequest,omitempty"`
}

type HttpRequestLog struct {
	RequestMethod string `json:"requestMethod,omitempty"`
	RequestUrl    string `json:"requestUrl,omitempty"`
}

func (logger *Logger) init() {
	if runningOnGCloud() {
		logger.mode = GCLOUD_LOGS
		logger.project = "icehockeyscoresheet"
	} else {
		logger.mode = LOCAL_LOGS
	}
}

func (logger *Logger) debug(template string, args ...any) {
	logger.log(context.TODO(), "Debug", template, args...)
}

func (logger *Logger) debug1(ctx context.Context, template string, args ...any) {
	logger.log(ctx, "Debug", template, args...)
}

func (logger *Logger) info(template string, args ...any) {
	logger.log(context.TODO(), "Info", template, args...)
}

func (logger *Logger) info1(ctx context.Context, template string, args ...any) {
	logger.log(ctx, "Info", template, args...)
}

func (logger *Logger) error(template string, args ...any) {
	logger.log(context.TODO(), "Error", template, args...)
}

func (logger *Logger) error1(ctx context.Context, template string, args ...any) {
	logger.log(ctx, "Error", template, args...)
}

func (logger *Logger) log(ctx context.Context, severity string, template string, args ...any) {
	switch logger.mode {
	case GCLOUD_LOGS:
		logger.gCloudLog(ctx, severity, template, args...)
	default:
		severity = fmt.Sprintf("%-6s", strings.ToUpper(severity))
		log.Printf(severity+template, args...)
	}
}

func (logger *Logger) gCloudLog(ctx context.Context, severity string, template string, args ...any) {
	if logger.encoder == nil {
		logger.encoder = json.NewEncoder(os.Stderr)
	}

	entry := GcpLogEntry{
		Severity:  severity,
		Timestamp: time.Now(),
		Message:   fmt.Sprintf(template, args...),
	}

	entry.Labels = map[string]string{
		"appname": logger.project,
	}

	logger.encoder.Encode(entry)
}
