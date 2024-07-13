package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"time"

	"cloud.google.com/go/logging"
)

const LOCAL_LOGS = 0
const GCLOUD_LOGS = 1

type Logger struct {
	mode    int
	project string
	client  *logging.Client
	logs    *logging.Logger
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
		// client, err := logging.NewClient(context.Background(), logger.project)
		// if err == nil {
		// 	logger.client = client
		// 	logger.logs = client.Logger("scoresheet")
		// }
	} else {
		logger.mode = LOCAL_LOGS
	}
}

func (logger *Logger) debug(template string, args ...any) {
	logger.debug1(context.TODO(), template, args...)
}

func (logger *Logger) debug1(ctx context.Context, template string, args ...any) {
	switch logger.mode {
	case GCLOUD_LOGS:
		logger.gCloudLog(ctx, logging.Debug, template, args...)
	default:
		log.Printf("DEBUG "+template, args...)
	}
}

func (logger *Logger) info(template string, args ...any) {
	logger.info1(context.TODO(), template, args...)
}

func (logger *Logger) info1(ctx context.Context, template string, args ...any) {
	switch logger.mode {
	case GCLOUD_LOGS:
		logger.gCloudLog(ctx, logging.Info, template, args...)
	default:
		log.Printf("INFO  "+template, args...)
	}
}

func (logger *Logger) error(template string, args ...any) {
	logger.error1(context.TODO(), template, args...)
}

func (logger *Logger) error1(ctx context.Context, template string, args ...any) {
	switch logger.mode {
	case GCLOUD_LOGS:
		logger.gCloudLog(ctx, logging.Error, template, args...)
	default:
		log.Printf("ERROR "+template, args...)
	}
}

func (logger *Logger) gCloudLog(ctx context.Context, severity logging.Severity, template string, args ...any) {
	// labels := make(map[string]string)
	// values := ctx.Value(GameIdKey)
	// if values != nil {
	// 	labels["gameId"] = values.(GameRequestContext).GameId
	// 	labels["remoteAddr"] = values.(GameRequestContext).RemoteAddr
	// }
	// logger.logs.Log(logging.Entry{
	// 	Payload:  fmt.Sprintf(template, args...),
	// 	Severity: severity,
	// 	Labels:   labels,
	// })

	if logger.encoder == nil {
		logger.encoder = json.NewEncoder(os.Stderr)
	}

	entry := GcpLogEntry{
		Severity:  severity.String(),
		Timestamp: time.Now(),
		Message:   fmt.Sprintf(template, args...),
	}

	entry.Labels = map[string]string{
		"appname": logger.project,
	}

	logger.encoder.Encode(entry)
}
