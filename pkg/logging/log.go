package logging

import (
	"os"
	"time"

	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
)

// LogInit is a function that initializes the log package with a console writer, a trace level, a timestamp, and a caller.
// It takes a boolean as an argument, which determines the global level of the log package.
// If the argument is true, the global level is set to debug. Otherwise, it is set to info.
func LogInit(isDebug bool) {
	log.Logger = zerolog.New(
		zerolog.ConsoleWriter{Out: os.Stderr, TimeFormat: time.RFC3339},
	).Level(zerolog.TraceLevel).With().Timestamp().Logger()

	zerolog.SetGlobalLevel(zerolog.InfoLevel)
	if isDebug {
		zerolog.SetGlobalLevel(zerolog.DebugLevel)
	}
}
