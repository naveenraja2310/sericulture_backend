package utils

import (
	"context"
	"encoding/base64"
	"fmt"
	"strings"
	"time"

	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/exporters/otlp/otlplog/otlploghttp"
	logs "go.opentelemetry.io/otel/log"
	"go.opentelemetry.io/otel/log/global"
	otellog "go.opentelemetry.io/otel/sdk/log"
	"go.opentelemetry.io/otel/sdk/resource"
	semconv "go.opentelemetry.io/otel/semconv/v1.26.0"
)

const ServiceName = "sericulture-backend"

func BuildGrafanaBasicAuth(instanceID, apiKey string) string {
	instanceID = strings.TrimSpace(instanceID)
	apiKey = strings.TrimSpace(apiKey)
	if instanceID == "" || apiKey == "" {
		return ""
	}

	return base64.StdEncoding.EncodeToString([]byte(instanceID + ":" + apiKey))
}

func InitGrafanaOTLPLogger(ctx context.Context, instanceID, apiKey string) (*otellog.LoggerProvider, error) {
	instanceID = strings.TrimSpace(instanceID)
	apiKey = strings.TrimSpace(apiKey)
	if instanceID == "" || apiKey == "" {
		return nil, nil
	}

	auth := BuildGrafanaBasicAuth(instanceID, apiKey)
	exporter, err := otlploghttp.New(
		ctx,
		otlploghttp.WithEndpoint("otlp-gateway-prod-ap-south-1.grafana.net"),
		otlploghttp.WithURLPath("/otlp/v1/logs"),
		otlploghttp.WithHeaders(map[string]string{
			"Authorization": "Basic " + auth,
		}),
		otlploghttp.WithTimeout(10*time.Second),
	)
	if err != nil {
		return nil, fmt.Errorf("failed to create grafana otlp log exporter: %w", err)
	}

	provider := otellog.NewLoggerProvider(
		otellog.WithProcessor(otellog.NewBatchProcessor(exporter)),
		otellog.WithResource(resource.NewWithAttributes(
			semconv.SchemaURL,
			semconv.ServiceName(ServiceName),
			semconv.ServiceVersion("1.0.0"),
		)),
	)
	global.SetLoggerProvider(provider)

	logger := global.Logger(ServiceName)
	record := logs.Record{}
	record.SetSeverity(logs.SeverityInfo)
	record.SetBody(attribute.StringValue("Backend started and Grafana OTLP logger initialized"))
	logger.Emit(ctx, record)

	return provider, nil
}
