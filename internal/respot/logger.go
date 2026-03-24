package respot

import (
	"context"
	"fmt"

	respot "github.com/devgianlu/go-librespot"
	"github.com/rs/zerolog"

	"github.com/pyrorhythm/spqt/pkg/log"
)

type respotLogger struct {
	logger *zerolog.Logger
}

func NewRespotLogger(logger *zerolog.Logger) respot.Logger {
	return &respotLogger{logger: logger}
}

func Logger() respot.Logger {
	return &respotLogger{logger: log.Logger()}
}

func FromContext(ctx context.Context) respot.Logger {
	return NewRespotLogger(log.Ctx(ctx))
}

func (l *respotLogger) Tracef(format string, args ...interface{}) {
	l.logger.Trace().Msgf(format, args...)
}

func (l *respotLogger) Debugf(format string, args ...interface{}) {
	l.logger.Debug().Msgf(format, args...)
}

func (l *respotLogger) Infof(format string, args ...interface{}) {
	l.logger.Info().Msgf(format, args...)
}

func (l *respotLogger) Warnf(format string, args ...interface{}) {
	l.logger.Warn().Msgf(format, args...)
}

func (l *respotLogger) Errorf(format string, args ...interface{}) {
	l.logger.Error().Msgf(format, args...)
}

func (l *respotLogger) Trace(args ...interface{}) {
	l.logger.Trace().Msg(formatArgs(args...))
}

func (l *respotLogger) Debug(args ...interface{}) {
	l.logger.Debug().Msg(formatArgs(args...))
}

func (l *respotLogger) Info(args ...interface{}) {
	l.logger.Info().Msg(formatArgs(args...))
}

func (l *respotLogger) Warn(args ...interface{}) {
	l.logger.Warn().Msg(formatArgs(args...))
}

func (l *respotLogger) Error(args ...interface{}) {
	l.logger.Error().Msg(formatArgs(args...))
}

func (l *respotLogger) WithField(key string, value interface{}) respot.Logger {
	return &respotLogger{logger: new(l.logger.With().Interface(key, value).Logger())}
}

func (l *respotLogger) WithError(err error) respot.Logger {
	return &respotLogger{logger: new(l.logger.With().Err(err).Logger())}
}

func formatArgs(args ...interface{}) string {
	if len(args) == 0 {
		return ""
	}
	if len(args) == 1 {
		if s, ok := args[0].(string); ok {
			return s
		}
	}
	return fmt.Sprintf("%v", args...)
}
