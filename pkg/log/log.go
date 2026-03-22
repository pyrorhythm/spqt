package log

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"sync"

	"github.com/rs/zerolog"
)

var (
	logger     zerolog.Logger
	loggerOnce = sync.Once{}
)

func Logger() *zerolog.Logger {
	loggerOnce.Do(func() {
		zerolog.CallerMarshalFunc = func(pc uintptr, file string, line int) string {
			parts := strings.Split(runtime.FuncForPC(pc).Name(), "/")
			pkgparts := strings.Split(parts[len(parts)-1], ".")
			if len(parts) == 1 {
				pkg := pkgparts[0]
				return fmt.Sprintf("%s:%sL%d", pkg, filepath.Base(file), line)
			}

			parent, pkg := parts[len(parts)-2], pkgparts[0]

			return fmt.Sprintf("%s/%s:%sL%d", parent, pkg, filepath.Base(file), line)
		}

		logger = zerolog.New(zerolog.MultiLevelWriter(
			zerolog.ConsoleWriter{Out: os.Stdout},
			// os.Stderr,
		)).
			Level(zerolog.DebugLevel).
			With().
			Timestamp().
			Caller().
			Logger()
	})

	return &logger
}

func Ctx(ctx context.Context) *zerolog.Logger {
	l := zerolog.Ctx(ctx)
	if l.GetLevel() == zerolog.Disabled {
		return Logger()
	}
	return l
}

func WithContext(ctx context.Context, l zerolog.Logger) context.Context {
	return l.WithContext(ctx)
}
