package log

import (
	"os"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

var logger *zap.SugaredLogger

type LevelLog = zapcore.Level

const (
	LevelDebug  = zapcore.DebugLevel
	LevelInfo   = zapcore.InfoLevel
	LevelWarn   = zapcore.WarnLevel
	LevelError  = zapcore.ErrorLevel
	LevelFatal  = zapcore.FatalLevel
	LevelDpanic = zapcore.DPanicLevel
	LevelPanic  = zapcore.PanicLevel
)

func InitLog(level any) {
	encoderConfig := zapcore.EncoderConfig{
		TimeKey:        "time",
		LevelKey:       "level",
		NameKey:        "logger",
		CallerKey:      "line",
		MessageKey:     "msg",
		StacktraceKey:  "stacktrace",
		LineEnding:     zapcore.DefaultLineEnding,
		EncodeLevel:    zapcore.LowercaseLevelEncoder,
		EncodeTime:     zapcore.ISO8601TimeEncoder,
		EncodeDuration: zapcore.SecondsDurationEncoder,
		EncodeCaller:   zapcore.ShortCallerEncoder,
		EncodeName:     zapcore.FullNameEncoder,
	}

	var writers []zapcore.WriteSyncer
	writers = append(writers, os.Stdout)

	atomicLevel := zap.NewAtomicLevel()
	var l zapcore.Level
	switch v := level.(type) {
	case string:
		l = getLevel(v)
	case LevelLog:
		l = v
	}
	atomicLevel.SetLevel(l)

	core := zapcore.NewCore(
		zapcore.NewConsoleEncoder(encoderConfig),
		zapcore.NewMultiWriteSyncer(writers...),
		atomicLevel,
	)
	logger = zap.New(core,
		zap.AddCaller(), zap.Development(),
		zap.AddCallerSkip(1), zap.AddStacktrace(zapcore.FatalLevel)).Sugar()
}

func getLevel(level string) (l zapcore.Level) {
	switch level {
	case "debug":
		l = zap.DebugLevel
	case "info":
		l = zap.InfoLevel
	case "warn":
		l = zap.WarnLevel
	case "error":
		l = zap.ErrorLevel
	case "panic":
		l = zap.PanicLevel
	case "fatal":
		l = zap.FatalLevel
	default:
		l = zap.InfoLevel
	}
	return
}

func Sync() error {
	return logger.Sync()
}

func Debug(args ...any) {
	logger.Debug(args...)
}

func Debugf(format string, args ...any) {
	logger.Debugf(format, args...)
}

func Debugw(msg string, args ...any) {
	logger.Debugw(msg, args...)
}

func Info(args ...any) {
	logger.Info(args...)
}

func Infof(format string, args ...any) {
	logger.Infof(format, args...)
}

func Infow(msg string, args ...any) {
	logger.Infow(msg, args...)
}

func Warn(args ...any) {
	logger.Warn(args...)
}

func Warnf(format string, args ...any) {
	logger.Warnf(format, args...)
}

func Warnw(msg string, args ...any) {
	logger.Warnw(msg, args...)
}

func Error(args ...any) {
	logger.Error(args...)
}

func Errorf(format string, args ...any) {
	logger.Errorf(format, args...)
}

func Errorw(msg string, args ...any) {
	logger.Errorw(msg, args...)
}

func DPanic(args ...any) {
	logger.DPanic(args...)
}

func DPanicf(format string, args ...any) {
	logger.DPanicf(format, args...)
}

func DPanicw(msg string, args ...any) {
	logger.DPanicw(msg, args...)
}

func Panic(args ...any) {
	logger.Panic(args...)
}

func Panicf(format string, args ...any) {
	logger.Panicf(format, args...)
}

func Panicw(msg string, args ...any) {
	logger.Panicw(msg, args...)
}

func Fatal(args ...any) {
	logger.Fatal(args...)
}

func Fatalf(format string, args ...any) {
	logger.Fatalf(format, args...)
}

func Fatalw(msg string, args ...any) {
	logger.Fatalw(msg, args...)
}

func GetLogger() *zap.SugaredLogger {
	return logger
}
