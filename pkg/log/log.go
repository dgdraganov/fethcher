package log

import (
	"time"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

// type ZapLogger struct {
// 	*zap.SugaredLogger
// }

func NewZapLogger(serviceName string, logLevel zapcore.Level) *zap.SugaredLogger {
	atom := zap.NewAtomicLevel()
	atom.SetLevel(logLevel)

	encoderConfig := zapcore.EncoderConfig{
		TimeKey:       "timestamp",
		LevelKey:      "level",
		NameKey:       "logger",
		CallerKey:     "caller",
		FunctionKey:   zapcore.OmitKey,
		MessageKey:    "message",
		StacktraceKey: "stacktrace",
		LineEnding:    zapcore.DefaultLineEnding,
		EncodeLevel:   zapcore.LowercaseLevelEncoder,
		EncodeTime: func(t time.Time, enc zapcore.PrimitiveArrayEncoder) {
			enc.AppendString(t.UTC().Format(time.RFC3339Nano))
		},
		EncodeDuration: zapcore.SecondsDurationEncoder,
		EncodeCaller:   zapcore.ShortCallerEncoder,
	}

	config := zap.Config{
		Level:       atom,
		Development: false,
		Sampling: &zap.SamplingConfig{
			Initial:    100,
			Thereafter: 100,
		},
		Encoding:         "json",
		EncoderConfig:    encoderConfig,
		OutputPaths:      []string{"stderr"},
		ErrorOutputPaths: []string{"stderr"},
	}

	core, err := config.Build()
	if err != nil {
		panic("can not build logger config")
	}

	logger := zap.New(core.Core(), zap.AddCaller())

	sugar := logger.Sugar()

	// setting default fields
	sugar = sugar.With(
		"service", serviceName,
	)

	return sugar
}
