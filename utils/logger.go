package utils

import (
	"context"
	"fmt"
	"log"
	"os"
	"strings"

	"sericulture/model"

	"github.com/gofiber/fiber/v2"
	"go.opentelemetry.io/otel/attribute"
	logs "go.opentelemetry.io/otel/log"
	"go.opentelemetry.io/otel/log/global"
)

// ErrorLog writes error messages to stderr
var ErrorLog = log.New(os.Stderr, "ERROR: ", log.LstdFlags)

// WarnLog writes warnings to stdout
var WarnLog = log.New(os.Stdout, "WARN: ", log.LstdFlags)

// InfoLog writes informational messages to stdout
var InfoLog = log.New(os.Stdout, "", log.LstdFlags)

func LoggerAttrsFromUser(user model.User, deviceID string) []attribute.KeyValue {
	if deviceID == "" {
		deviceID = user.DeviceID
	}

	if user.ID.IsZero() {
		return []attribute.KeyValue{
			attribute.String("userId", ""),
			attribute.String("deviceId", deviceID),
		}
	}

	return []attribute.KeyValue{
		attribute.String("userId", user.ID.Hex()),
		attribute.String("deviceId", deviceID),
	}
}

func LoggerAttrsFromFiber(c *fiber.Ctx) []attribute.KeyValue {
	if c == nil {
		return nil
	}

	attrs := []attribute.KeyValue{
		attribute.String("method", c.Method()),
		attribute.String("path", c.Path()),
	}

	if user, ok := c.Locals("user").(model.User); ok {
		return append(attrs, LoggerAttrsFromUser(user, "")...)
	}

	if userID, ok := c.Locals("userId").(string); ok && strings.TrimSpace(userID) != "" {
		attrs = append(attrs, attribute.String("userId", userID))
	}
	if deviceID, ok := c.Locals("deviceID").(string); ok && strings.TrimSpace(deviceID) != "" {
		attrs = append(attrs, attribute.String("deviceId", deviceID))
	} else if deviceID, ok := c.Locals("deviceId").(string); ok && strings.TrimSpace(deviceID) != "" {
		attrs = append(attrs, attribute.String("deviceId", deviceID))
	}

	return attrs
}

func normalizeLogAttrs(args ...any) []attribute.KeyValue {
	if len(args) == 0 {
		return nil
	}
	if len(args)%2 != 0 {
		args = append(args, "detail")
	}

	attrs := make([]attribute.KeyValue, 0, len(args)/2)
	for i := 0; i < len(args); i += 2 {
		key, ok := args[i].(string)
		if !ok {
			key = fmt.Sprint(args[i])
		}

		switch v := args[i+1].(type) {
		case string:
			attrs = append(attrs, attribute.String(key, v))
		case bool:
			attrs = append(attrs, attribute.Bool(key, v))
		case int:
			attrs = append(attrs, attribute.Int(key, v))
		case int64:
			attrs = append(attrs, attribute.Int64(key, v))
		case float64:
			attrs = append(attrs, attribute.Float64(key, v))
		case float32:
			attrs = append(attrs, attribute.Float64(key, float64(v)))
		default:
			attrs = append(attrs, attribute.String(key, fmt.Sprintf("%v", v)))
		}
	}
	return attrs
}

func emitLog(level logs.Severity, c *fiber.Ctx, msg string, args ...any) {
	attrs := normalizeLogAttrs(args...)
	if c != nil {
		attrs = append(LoggerAttrsFromFiber(c), attrs...)
	}

	record := logs.Record{}
	record.SetSeverity(level)
	record.SetBody(attribute.StringValue(msg))
	record.AddAttributes(attrs...)
	global.Logger(ServiceName).Emit(context.Background(), record)
}

func Info(c *fiber.Ctx, msg string, args ...any) {
	emitLog(logs.SeverityInfo, c, msg, args...)
}

func Warn(c *fiber.Ctx, msg string, args ...any) {
	emitLog(logs.SeverityWarn, c, msg, args...)
}

func Error(c *fiber.Ctx, msg string, args ...any) {
	emitLog(logs.SeverityError, c, msg, args...)
}

func Infof(c *fiber.Ctx, format string, args ...any) {
	Info(c, fmt.Sprintf(format, args...))
}

func Warnf(c *fiber.Ctx, format string, args ...any) {
	Warn(c, fmt.Sprintf(format, args...))
}

func Errorf(c *fiber.Ctx, format string, args ...any) {
	Error(c, fmt.Sprintf(format, args...))
}

func Log(c *fiber.Ctx, level logs.Severity, msg string, args ...any) {
	emitLog(level, c, msg, args...)
}
