package logger

import (
	"context"
	"io"
	"os"
	"pkg-common/utilities/timeutil"
	"time"

	"github.com/rs/zerolog"
)

type Logger struct {
	zl   zerolog.Logger
	sink func(string, zerolog.Level, timeutil.TimeUTC)
}

func New() *Logger {
	zerolog.TimeFieldFormat = time.RFC3339Nano

	logger := zerolog.New(os.Stdout).
		With().
		Timestamp().
		Caller().
		Logger()

	return &Logger{zl: logger}
}

func NewFromConfig(cfg LoggerConfig) *Logger {
	if cfg.LogLevel == zerolog.NoLevel {
		cfg.LogLevel = zerolog.InfoLevel
	}

	zerolog.TimeFieldFormat = zerolog.TimeFormatUnix
	zerolog.CallerSkipFrameCount = 3

	logger := zerolog.New(os.Stdout).
		With().
		Timestamp().
		Caller().
		Logger().
		Level(zerolog.Level(cfg.LogLevel))

	return &Logger{zl: logger}
}

func (l *Logger) WithOutput(w io.Writer) *Logger {
	l.zl = l.zl.Output(w)
	return l
}

func (l *Logger) WithLevel(level zerolog.Level) *Logger {
	l.zl = l.zl.Level(level)
	return l
}

func (l *Logger) WithContext(ctx context.Context) *Logger {
	return &Logger{zl: l.zl.With().Logger()}
}

func (l *Logger) With() zerolog.Context {
	return l.zl.With()
}

func (l *Logger) Debug(msg string) {
	l.zl.Debug().Msg(msg)
	l.activateSink(msg, zerolog.DebugLevel, timeutil.NowUTC())
}

func (l *Logger) Debugf(format string, v ...interface{}) {
	l.zl.Debug().Msgf(format, v...)
	l.activateSinkFormatted(format, zerolog.DebugLevel, timeutil.NowUTC(), v...)
}

func (l *Logger) Info(msg string) {
	l.zl.Info().Msg(msg)
	l.activateSink(msg, zerolog.InfoLevel, timeutil.NowUTC())
}

func (l *Logger) Infof(format string, v ...interface{}) {
	l.zl.Info().Msgf(format, v...)
	l.activateSinkFormatted(format, zerolog.InfoLevel, timeutil.NowUTC(), v...)
}

func (l *Logger) Warn(msg string) {
	l.zl.Warn().Msg(msg)
	l.activateSink(msg, zerolog.WarnLevel, timeutil.NowUTC())
}

func (l *Logger) Warnf(format string, v ...interface{}) {
	l.zl.Warn().Msgf(format, v...)
	l.activateSinkFormatted(format, zerolog.WarnLevel, timeutil.NowUTC(), v...)
}

func (l *Logger) Error(err error, msg string) {
	l.zl.Error().Err(err).Msg(msg)
	l.activateSink(msg, zerolog.ErrorLevel, timeutil.NowUTC())
}

func (l *Logger) Errorf(err error, format string, v ...interface{}) {
	l.zl.Error().Err(err).Msgf(format, v...)
	l.activateSinkFormatted(format, zerolog.ErrorLevel, timeutil.NowUTC(), v...)
}

func (l *Logger) Fatal(err error, msg string) {
	l.zl.Fatal().Err(err).Msg(msg)
	l.activateSink(msg, zerolog.FatalLevel, timeutil.NowUTC())
}

func (l *Logger) Fatalf(err error, format string, v ...interface{}) {
	l.zl.Fatal().Err(err).Msgf(format, v...)
	l.activateSinkFormatted(format, zerolog.FatalLevel, timeutil.NowUTC(), v...)
}

func (l *Logger) Panic(err error, msg string) {
	l.zl.Panic().Err(err).Msg(msg)
	l.activateSink(msg, zerolog.PanicLevel, timeutil.NowUTC())
}

func (l *Logger) Panicf(err error, format string, v ...interface{}) {
	l.zl.Panic().Err(err).Msgf(format, v...)
	l.activateSinkFormatted(format, zerolog.PanicLevel, timeutil.NowUTC(), v...)
}

func (l *Logger) Log(level zerolog.Level, msg string) {
	l.zl.WithLevel(level).Msg(msg)
	l.activateSink(msg, level, timeutil.NowUTC())
}

func (l *Logger) Logf(level zerolog.Level, format string, v ...interface{}) {
	l.zl.WithLevel(level).Msgf(format, v...)
	l.activateSinkFormatted(format, level, timeutil.NowUTC(), v...)
}
