package log

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestConfig_Validate(t *testing.T) {
	tests := []struct {
		name        string
		config      Config
		expectError bool
		errorType   func(error) bool
	}{
		{
			name: "valid config with all fields",
			config: Config{
				Dir:        "/tmp/logs",
				Level:      "info",
				Format:     "json",
				MaxSize:    100,
				MaxBackups: 10,
				MaxAge:     30,
				Compress:   true,
				Console:    false,
			},
			expectError: false,
		},
		{
			name: "valid config with minimal fields",
			config: Config{
				Dir: "/tmp/logs",
			},
			expectError: false,
		},
		{
			name: "valid config with console format",
			config: Config{
				Dir:    "/tmp/logs",
				Level:  "debug",
				Format: "console",
			},
			expectError: false,
		},
		{
			name: "valid config with text format",
			config: Config{
				Dir:    "/tmp/logs",
				Level:  "warn",
				Format: "text",
			},
			expectError: false,
		},
		{
			name: "valid config with standard format",
			config: Config{
				Dir:    "/tmp/logs",
				Level:  "error",
				Format: "standard",
			},
			expectError: false,
		},
		{
			name:        "empty dir should error",
			config:      Config{},
			expectError: true,
			errorType:   IsEmptyLogDir,
		},
		{
			name: "negative max size should error",
			config: Config{
				Dir:     "/tmp/logs",
				MaxSize: -1,
			},
			expectError: true,
			errorType:   IsInvalidConfigValue,
		},
		{
			name: "negative max backups should error",
			config: Config{
				Dir:        "/tmp/logs",
				MaxBackups: -1,
			},
			expectError: true,
			errorType:   IsInvalidConfigValue,
		},
		{
			name: "negative max age should error",
			config: Config{
				Dir:    "/tmp/logs",
				MaxAge: -1,
			},
			expectError: true,
			errorType:   IsInvalidConfigValue,
		},
		{
			name: "multiple negative values should error",
			config: Config{
				Dir:        "/tmp/logs",
				MaxSize:    -10,
				MaxBackups: -5,
				MaxAge:     -30,
			},
			expectError: true,
			errorType:   IsInvalidConfigValue,
		},
		{
			name: "invalid log level should error",
			config: Config{
				Dir:   "/tmp/logs",
				Level: "invalid",
			},
			expectError: true,
			errorType:   IsInvalidLogLevel,
		},
		{
			name: "invalid log format should error",
			config: Config{
				Dir:    "/tmp/logs",
				Format: "invalid",
			},
			expectError: true,
			errorType:   IsInvalidLogFormat,
		},
		{
			name: "zero values for size/backups/age should be valid",
			config: Config{
				Dir:        "/tmp/logs",
				MaxSize:    0,
				MaxBackups: 0,
				MaxAge:     0,
			},
			expectError: false,
		},
		{
			name: "positive values for size/backups/age should be valid",
			config: Config{
				Dir:        "/tmp/logs",
				MaxSize:    500,
				MaxBackups: 100,
				MaxAge:     365,
			},
			expectError: false,
		},
		{
			name: "all valid log levels",
			config: Config{
				Dir:   "/tmp/logs",
				Level: "debug",
			},
			expectError: false,
		},
		{
			name: "info log level",
			config: Config{
				Dir:   "/tmp/logs",
				Level: "info",
			},
			expectError: false,
		},
		{
			name: "warn log level",
			config: Config{
				Dir:   "/tmp/logs",
				Level: "warn",
			},
			expectError: false,
		},
		{
			name: "error log level",
			config: Config{
				Dir:   "/tmp/logs",
				Level: "error",
			},
			expectError: false,
		},
		{
			name: "dpanic log level",
			config: Config{
				Dir:   "/tmp/logs",
				Level: "dpanic",
			},
			expectError: false,
		},
		{
			name: "panic log level",
			config: Config{
				Dir:   "/tmp/logs",
				Level: "panic",
			},
			expectError: false,
		},
		{
			name: "fatal log level",
			config: Config{
				Dir:   "/tmp/logs",
				Level: "fatal",
			},
			expectError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.config.Validate()

			if tt.expectError {
				require.Error(t, err)
				if tt.errorType != nil {
					assert.True(t, tt.errorType(err), "Expected error type mismatch")
				}
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestConfig_Validate_EmptyStringFields(t *testing.T) {
	tests := []struct {
		name     string
		config   Config
		wantErr  bool
		errorMsg string
	}{
		{
			name: "empty level should be valid",
			config: Config{
				Dir: "/tmp/logs",
			},
			wantErr: false,
		},
		{
			name: "empty format should be valid",
			config: Config{
				Dir: "/tmp/logs",
			},
			wantErr: false,
		},
		{
			name: "empty dir should error",
			config: Config{
				Dir: "",
			},
			wantErr:  true,
			errorMsg: "log directory cannot be empty",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.config.Validate()
			if tt.wantErr {
				require.Error(t, err)
				if tt.errorMsg != "" {
					assert.Contains(t, err.Error(), tt.errorMsg)
				}
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestConfig_Validate_LogFormats(t *testing.T) {
	validFormats := []string{"json", "console", "text", "standard"}
	invalidFormats := []string{"", "xml", "yaml", "txt", "log", "custom"}

	for _, format := range validFormats {
		t.Run("valid format_"+format, func(t *testing.T) {
			config := Config{
				Dir:    "/tmp/logs",
				Format: format,
			}
			err := config.Validate()
			assert.NoError(t, err)
		})
	}

	for _, format := range invalidFormats {
		if format == "" {
			continue // empty format is valid
		}
		t.Run("invalid format_"+format, func(t *testing.T) {
			config := Config{
				Dir:    "/tmp/logs",
				Format: format,
			}
			err := config.Validate()
			require.Error(t, err)
			assert.True(t, IsInvalidLogFormat(err))
		})
	}
}

func TestConfig_Validate_LogLevels(t *testing.T) {
	validLevels := []string{"debug", "info", "warn", "error", "dpanic", "panic", "fatal", "DEBUG", "INFO", "WARN", "ERROR"}
	invalidLevels := []string{"trace", "verbose", "critical", "invalid"}

	for _, level := range validLevels {
		t.Run("valid level_"+level, func(t *testing.T) {
			config := Config{
				Dir:   "/tmp/logs",
				Level: level,
			}
			err := config.Validate()
			assert.NoError(t, err)
		})
	}

	for _, level := range invalidLevels {
		t.Run("invalid level_"+level, func(t *testing.T) {
			config := Config{
				Dir:   "/tmp/logs",
				Level: level,
			}
			err := config.Validate()
			require.Error(t, err)
			assert.True(t, IsInvalidLogLevel(err))
		})
	}
}

func TestConfig_Validate_EdgeCases(t *testing.T) {
	tests := []struct {
		name        string
		config      Config
		expectError bool
		description string
	}{
		{
			name: "very large max size",
			config: Config{
				Dir:     "/tmp/logs",
				MaxSize: 10000, // 10GB
			},
			expectError: false,
			description: "Large max size should be valid",
		},
		{
			name: "very large max backups",
			config: Config{
				Dir:        "/tmp/logs",
				MaxBackups: 1000,
			},
			expectError: false,
			description: "Large max backups should be valid",
		},
		{
			name: "very large max age",
			config: Config{
				Dir:    "/tmp/logs",
				MaxAge: 3650, // 10 years
			},
			expectError: false,
			description: "Large max age should be valid",
		},
		{
			name: "all boolean combinations",
			config: Config{
				Dir:      "/tmp/logs",
				Compress: true,
				Console:  true,
			},
			expectError: false,
			description: "Both compress and console enabled should be valid",
		},
		{
			name: "minimal valid config",
			config: Config{
				Dir: "/tmp/logs",
			},
			expectError: false,
			description: "Only dir should be required",
		},
		{
			name: "config with all fields set",
			config: Config{
				Dir:        "/tmp/logs",
				Level:      "info",
				Format:     "json",
				MaxSize:    100,
				MaxBackups: 10,
				MaxAge:     30,
				Compress:   true,
				Console:    true,
			},
			expectError: false,
			description: "All fields set with valid values should be valid",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.config.Validate()
			if tt.expectError {
				assert.Error(t, err, tt.description)
			} else {
				assert.NoError(t, err, tt.description)
			}
		})
	}
}

func TestConfig_Validate_ErrorMessages(t *testing.T) {
	tests := []struct {
		name        string
		config      Config
		expectedMsg string
	}{
		{
			name: "empty dir error message",
			config: Config{
				Dir: "",
			},
			expectedMsg: "log directory cannot be empty",
		},
		{
			name: "invalid config value error message",
			config: Config{
				Dir:     "/tmp/logs",
				MaxSize: -1,
			},
			expectedMsg: "invalid config value: max_size, max_backups, max_age must be non-negative",
		},
		{
			name: "invalid log level error message",
			config: Config{
				Dir:   "/tmp/logs",
				Level: "invalid",
			},
			expectedMsg: "invalid log level",
		},
		{
			name: "invalid log format error message",
			config: Config{
				Dir:    "/tmp/logs",
				Format: "invalid",
			},
			expectedMsg: "invalid log format",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.config.Validate()
			require.Error(t, err)
			assert.Contains(t, err.Error(), tt.expectedMsg)
		})
	}
}

// Benchmark tests
func BenchmarkConfig_Validate_Valid(b *testing.B) {
	config := Config{
		Dir:        "/tmp/logs",
		Level:      "info",
		Format:     "json",
		MaxSize:    100,
		MaxBackups: 10,
		MaxAge:     30,
		Compress:   true,
		Console:    false,
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = config.Validate()
	}
}

func BenchmarkConfig_Validate_Minimal(b *testing.B) {
	config := Config{
		Dir: "/tmp/logs",
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = config.Validate()
	}
}

func BenchmarkConfig_Validate_WithLevel(b *testing.B) {
	config := Config{
		Dir:   "/tmp/logs",
		Level: "info",
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = config.Validate()
	}
}

func BenchmarkConfig_Validate_WithFormat(b *testing.B) {
	config := Config{
		Dir:    "/tmp/logs",
		Format: "json",
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = config.Validate()
	}
}
