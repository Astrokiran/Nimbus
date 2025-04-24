package config

import (
	"strings"

	"github.com/spf13/viper"
)

// Config holds the application configuration.
type Config struct {
	Server struct {
		Port                int `mapstructure:"port"`
		ReadTimeoutSeconds  int `mapstructure:"read_timeout_seconds"`
		WriteTimeoutSeconds int `mapstructure:"write_timeout_seconds"`
		IdleTimeoutSeconds  int `mapstructure:"idle_timeout_seconds"`
	}
	Database struct {
		DSN            string `mapstructure:"dsn"`
		LogLevel       string `mapstructure:"log_level"`
		MigrateOnStart bool   `mapstructure:"migrate_on_start"`
		UseAutoMigrate bool   `mapstructure:"use_auto_migrate"`
	}
	Logger struct {
		Level  string `mapstructure:"level"`
		Format string `mapstructure:"format"`
	}
	Auth struct {
		Jwt struct {
			Secret              string `mapstructure:"secret"`
			AccessExpiryMinutes int    `mapstructure:"access_expiry_minutes"`
			RefreshExpiryDays   int    `mapstructure:"refresh_expiry_days"`
		} `mapstructure:"jwt"`
		Otp struct {
			TestModeEnabled bool   `mapstructure:"test_mode_enabled"`
			TestValue       string `mapstructure:"test_value"`
		} `mapstructure:"otp"`
	} `mapstructure:"auth"`
}

// LoadConfig reads configuration from file or environment variables.
func LoadConfig(path string) (*Config, error) {
	viper.AddConfigPath(path)
	viper.SetConfigName("config")
	viper.SetConfigType("yaml")

	// Read environment variables
	viper.AutomaticEnv()
	viper.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))
	viper.SetEnvPrefix("APP") // Example: APP_SERVER_PORT=8081

	// Set default values
	viper.SetDefault("server.port", 8080)
	viper.SetDefault("server.read_timeout_seconds", 15)
	viper.SetDefault("server.write_timeout_seconds", 15)
	viper.SetDefault("server.idle_timeout_seconds", 60)
	viper.SetDefault("database.log_level", "warn")
	viper.SetDefault("database.migrate_on_start", true)
	viper.SetDefault("database.use_auto_migrate", false) // Default to SQL migrations for backward compatibility
	viper.SetDefault("logger.level", "info")
	viper.SetDefault("logger.format", "json")

	// Set Auth defaults
	viper.SetDefault("auth.jwt.secret", "your-very-secret-key") // PLEASE CHANGE THIS IN PRODUCTION!
	viper.SetDefault("auth.jwt.access_expiry_minutes", 15)
	viper.SetDefault("auth.jwt.refresh_expiry_days", 7)
	viper.SetDefault("auth.otp.test_mode_enabled", false) // Secure default
	viper.SetDefault("auth.otp.test_value", "123456")     // Default test OTP

	if err := viper.ReadInConfig(); err != nil {
		if _, ok := err.(viper.ConfigFileNotFoundError); !ok {
			// Config file was found but another error was produced
			return nil, err
		}
		// Config file not found; ignore error if desired
		// We proceed with defaults and env vars
	}

	var cfg Config
	if err := viper.Unmarshal(&cfg); err != nil {
		return nil, err
	}

	return &cfg, nil
}
