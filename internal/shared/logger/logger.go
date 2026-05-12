package logger

import (
	"fmt"
	"os"

	"github.com/bagusyanuar/erp-digital-printing-be/internal/shared/config"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"gopkg.in/natefinch/lumberjack.v2"
)

func NewLogger(cfg *config.Config) (*zap.Logger, error) {
	if cfg.Log.File == "" {
		return nil, fmt.Errorf("LOG_FILE is required")
	}

	// Lumberjack for log rotation
	lumberjackLogger := &lumberjack.Logger{
		Filename:   cfg.Log.File,
		MaxSize:    cfg.Log.MaxSize,
		MaxBackups: cfg.Log.MaxBackups,
		MaxAge:     cfg.Log.MaxAge,
		Compress:   cfg.Log.Compress,
	}

	// Encoder config
	encoderConfig := zap.NewProductionEncoderConfig()
	encoderConfig.EncodeTime = zapcore.ISO8601TimeEncoder
	encoderConfig.EncodeLevel = zapcore.CapitalLevelEncoder

	var encoder zapcore.Encoder
	if cfg.App.Env == "development" {
		encoder = zapcore.NewConsoleEncoder(encoderConfig)
	} else {
		encoder = zapcore.NewJSONEncoder(encoderConfig)
	}

	// Log Level
	var level zapcore.Level
	if err := level.UnmarshalText([]byte(cfg.Log.Level)); err != nil {
		level = zapcore.InfoLevel
	}

	core := zapcore.NewCore(
		encoder,
		zapcore.NewMultiWriteSyncer(zapcore.AddSync(os.Stdout), zapcore.AddSync(lumberjackLogger)),
		level,
	)

	return zap.New(core, zap.AddCaller(), zap.AddStacktrace(zapcore.ErrorLevel)), nil
}
