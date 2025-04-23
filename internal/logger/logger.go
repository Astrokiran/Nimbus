package logger

import (
	"io"
	"os"
	"strings"
	"time"

	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"

	"nimbus-service/internal/config"
)

var Logger zerolog.Logger

// Init initializes the global logger based on configuration.
func Init(cfg *config.Config) {
	var output io.Writer = os.Stdout
	if strings.ToLower(cfg.Logger.Format) == "json" {
		// Default JSON format
	} else {
		// Pretty console output for development
		output = zerolog.ConsoleWriter{
			Out:        os.Stdout,
			TimeFormat: time.RFC3339,
		}
	}

	logLevel, err := zerolog.ParseLevel(strings.ToLower(cfg.Logger.Level))
	if err != nil {
		logLevel = zerolog.InfoLevel // Default to info level if parsing fails
		log.Warn().Msgf("Invalid logger level '%s', defaulting to 'info'", cfg.Logger.Level)
	}

	zerolog.SetGlobalLevel(logLevel)
	Logger = zerolog.New(output).With().Timestamp().Caller().Logger()

	log.Logger = Logger // Configure standard log package to use zerolog
	Logger.Info().Msgf("Logger initialized with level '%s' and format '%s'", logLevel, cfg.Logger.Format)
}
