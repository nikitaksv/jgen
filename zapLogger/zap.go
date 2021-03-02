package zapLogger

import (
	"context"
	"fmt"
	"os"

	"github.com/micro/micro/v3/service/logger"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

type zaplog struct {
	zap  *zap.Logger
	opts logger.Options
}

func (l *zaplog) Init(opts ...logger.Option) error {
	var err error

	for _, o := range opts {
		o(&l.opts)
	}

	zapConfig := zap.NewProductionConfig()
	if zconfig, ok := l.opts.Context.Value(configKey{}).(zap.Config); ok {
		zapConfig = zconfig
	}

	if zcconfig, ok := l.opts.Context.Value(encoderConfigKey{}).(zapcore.EncoderConfig); ok {
		zapConfig.EncoderConfig = zcconfig
	}

	skip, ok := l.opts.Context.Value(callerSkipKey{}).(int)
	if !ok || skip < 1 {
		skip = 1
	}

	// Set log Level if not default
	zapConfig.Level = zap.NewAtomicLevel()
	if l.opts.Level != logger.InfoLevel {
		zapConfig.Level.SetLevel(loggerToZapLevel(l.opts.Level))
	}

	log, err := zapConfig.Build(zap.AddCallerSkip(skip))
	if err != nil {
		return err
	}

	// Adding seed fields if exist
	if l.opts.Fields != nil {
		var data []zap.Field
		for k, v := range l.opts.Fields {
			data = append(data, zap.Any(k, v))
		}
		log = log.With(data...)
	}

	// Adding namespace
	if namespace, ok := l.opts.Context.Value(namespaceKey{}).(string); ok {
		log = log.With(zap.Namespace(namespace))
	}

	l.zap = log

	return nil
}

func (l *zaplog) Fields(fields map[string]interface{}) logger.Logger {
	data := make([]zap.Field, 0, len(fields))
	for k, v := range fields {
		data = append(data, zap.Any(k, v))
	}

	zl := &zaplog{
		zap:  l.zap.With(data...),
		opts: l.opts,
	}

	return zl
}

func (l *zaplog) Log(level logger.Level, args ...interface{}) {
	msg := ""
	fields := make([]zap.Field, 0, len(args))
	for _, v := range args {
		if zapField, ok := v.(zap.Field); ok {
			fields = append(fields, zapField)
		} else {
			msg += " " + fmt.Sprint(v)
		}
	}

	l.log(level, msg, fields...)
}

func (l *zaplog) log(level logger.Level, msg string, fields ...zap.Field) {
	switch level {
	case logger.TraceLevel, logger.DebugLevel:
		l.zap.Debug(msg, fields...)
	case logger.WarnLevel:
		l.zap.Warn(msg, fields...)
	case logger.ErrorLevel:
		l.zap.Error(msg, fields...)
	case logger.FatalLevel:
		l.zap.Fatal(msg, fields...)
	default:
		l.zap.Info(msg, fields...)
	}
}

func (l *zaplog) Logf(level logger.Level, format string, args ...interface{}) {
	fmtArgs := make([]interface{}, 0, len(args))
	fields := make([]zap.Field, 0, len(args))
	for _, v := range args {
		if zapField, ok := v.(zap.Field); ok {
			fields = append(fields, zapField)
		} else {
			fmtArgs = append(fmtArgs, v)
		}
	}

	l.log(level, fmt.Sprintf(format, fmtArgs...), fields...)
}

func (l *zaplog) String() string {
	return "zap"
}

func (l *zaplog) Options() logger.Options {
	return l.opts
}

// New builds a new logger based on options
func NewLogger(opts ...logger.Option) (logger.Logger, error) {
	// Default options
	options := logger.Options{
		Level:   logger.InfoLevel,
		Fields:  make(map[string]interface{}),
		Out:     os.Stderr,
		Context: context.Background(),
	}

	l := &zaplog{opts: options}
	if err := l.Init(opts...); err != nil {
		return nil, err
	}

	return l, nil
}

func loggerToZapLevel(level logger.Level) zapcore.Level {
	switch level {
	case logger.TraceLevel, logger.DebugLevel:
		return zap.DebugLevel
	case logger.InfoLevel:
		return zap.InfoLevel
	case logger.WarnLevel:
		return zap.WarnLevel
	case logger.ErrorLevel:
		return zap.ErrorLevel
	case logger.FatalLevel:
		return zap.FatalLevel
	default:
		return zap.InfoLevel
	}
}
