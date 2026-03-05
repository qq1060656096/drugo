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
				Level: "info",
				Outputs: []OutputConfig{
					{
						Type:   "file",
						Format: "json",
						File: &FileOutputConfig{
							Dir:        "/tmp/logs",
							MaxSize:    100,
							MaxBackups: 10,
							MaxAge:     30,
							Compress:   true,
						},
					},
				},
			},
			expectError: false,
		},
		{
			name: "valid config with minimal fields",
			config: Config{
				Outputs: []OutputConfig{
					{
						Type: "file",
						File: &FileOutputConfig{Dir: "/tmp/logs"},
					},
				},
			},
			expectError: false,
		},
		{
			name: "valid config with json format",
			config: Config{
				Level: "debug",
				Outputs: []OutputConfig{
					{
						Type:   "console",
						Format: "json",
					},
				},
			},
			expectError: false,
		},
		{
			name: "valid config with text format",
			config: Config{
				Level: "warn",
				Outputs: []OutputConfig{
					{
						Type:   "console",
						Format: "text",
					},
				},
			},
			expectError: false,
		},

		{
			name:        "empty outputs should error",
			config:      Config{},
			expectError: true,
			errorType:   IsEmptyLogOutputs,
		},
		{
			name: "negative max size should error",
			config: Config{
				Outputs: []OutputConfig{
					{
						Type: "file",
						File: &FileOutputConfig{
							Dir:     "/tmp/logs",
							MaxSize: -1,
						},
					},
				},
			},
			expectError: true,
			errorType:   IsInvalidConfigValue,
		},
		{
			name: "negative max backups should error",
			config: Config{
				Outputs: []OutputConfig{
					{
						Type: "file",
						File: &FileOutputConfig{
							Dir:        "/tmp/logs",
							MaxBackups: -1,
						},
					},
				},
			},
			expectError: true,
			errorType:   IsInvalidConfigValue,
		},
		{
			name: "negative max age should error",
			config: Config{
				Outputs: []OutputConfig{
					{
						Type: "file",
						File: &FileOutputConfig{
							Dir:    "/tmp/logs",
							MaxAge: -1,
						},
					},
				},
			},
			expectError: true,
			errorType:   IsInvalidConfigValue,
		},
		{
			name: "multiple negative values should error",
			config: Config{
				Outputs: []OutputConfig{
					{
						Type: "file",
						File: &FileOutputConfig{
							Dir:        "/tmp/logs",
							MaxSize:    -10,
							MaxBackups: -5,
							MaxAge:     -30,
						},
					},
				},
			},
			expectError: true,
			errorType:   IsInvalidConfigValue,
		},
		{
			name: "invalid log level should error",
			config: Config{
				Level: "invalid",
				Outputs: []OutputConfig{
					{
						Type: "console",
					},
				},
			},
			expectError: true,
			errorType:   IsInvalidLogLevel,
		},
		{
			name: "invalid log format should error",
			config: Config{
				Outputs: []OutputConfig{
					{
						Type:   "file",
						Format: "invalid",
						File:   &FileOutputConfig{Dir: "/tmp/logs"},
					},
				},
			},
			expectError: true,
			errorType:   IsInvalidLogFormat,
		},
		{
			name: "zero values for size/backups/age should be valid",
			config: Config{
				Outputs: []OutputConfig{
					{
						Type: "file",
						File: &FileOutputConfig{
							Dir:        "/tmp/logs",
							MaxSize:    0,
							MaxBackups: 0,
							MaxAge:     0,
						},
					},
				},
			},
			expectError: false,
		},
		{
			name: "positive values for size/backups/age should be valid",
			config: Config{
				Outputs: []OutputConfig{
					{
						Type: "file",
						File: &FileOutputConfig{
							Dir:        "/tmp/logs",
							MaxSize:    500,
							MaxBackups: 100,
							MaxAge:     365,
						},
					},
				},
			},
			expectError: false,
		},
		{
			name: "all valid log levels",
			config: Config{
				Level: "debug",
				Outputs: []OutputConfig{
					{
						Type: "console",
					},
				},
			},
			expectError: false,
		},
		{
			name: "info log level",
			config: Config{
				Level: "info",
				Outputs: []OutputConfig{
					{
						Type: "console",
					},
				},
			},
			expectError: false,
		},
		{
			name: "warn log level",
			config: Config{
				Level: "warn",
				Outputs: []OutputConfig{
					{
						Type: "console",
					},
				},
			},
			expectError: false,
		},
		{
			name: "error log level",
			config: Config{
				Level: "error",
				Outputs: []OutputConfig{
					{
						Type: "console",
					},
				},
			},
			expectError: false,
		},
		{
			name: "dpanic log level",
			config: Config{
				Level: "dpanic",
				Outputs: []OutputConfig{
					{
						Type: "console",
					},
				},
			},
			expectError: false,
		},
		{
			name: "panic log level",
			config: Config{
				Level: "panic",
				Outputs: []OutputConfig{
					{
						Type: "console",
					},
				},
			},
			expectError: false,
		},
		{
			name: "fatal log level",
			config: Config{
				Level: "fatal",
				Outputs: []OutputConfig{
					{
						Type: "console",
					},
				},
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
				Outputs: []OutputConfig{
					{
						Type: "console",
					},
				},
			},
			wantErr: false,
		},
		{
			name: "empty format should be valid",
			config: Config{
				Outputs: []OutputConfig{
					{
						Type: "file",
						File: &FileOutputConfig{Dir: "/tmp/logs"},
					},
				},
			},
			wantErr: false,
		},
		{
			name: "empty dir should error",
			config: Config{
				Outputs: []OutputConfig{
					{
						Type: "file",
						File: &FileOutputConfig{Dir: ""},
					},
				},
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

	validFormats := []string{"json", "text"}
	invalidFormats := []string{"", "xml", "yaml", "txt", "log", "custom"}

	for _, format := range validFormats {
		t.Run("valid format_"+format, func(t *testing.T) {
			config := Config{
				Outputs: []OutputConfig{
					{
						Type:   "file",
						Format: format,
						File:   &FileOutputConfig{Dir: "/tmp/logs"},
					},
				},
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
				Outputs: []OutputConfig{
					{
						Type:   "file",
						Format: format,
						File:   &FileOutputConfig{Dir: "/tmp/logs"},
					},
				},
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
				Level: level,
				Outputs: []OutputConfig{
					{
						Type: "console",
					},
				},
			}
			err := config.Validate()
			assert.NoError(t, err)
		})
	}

	for _, level := range invalidLevels {
		t.Run("invalid level_"+level, func(t *testing.T) {
			config := Config{
				Level: level,
				Outputs: []OutputConfig{
					{
						Type: "console",
					},
				},
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
				Outputs: []OutputConfig{
					{
						Type: "file",
						File: &FileOutputConfig{
							Dir:     "/tmp/logs",
							MaxSize: 10000, // 10GB
						},
					},
				},
			},
			expectError: false,
			description: "Large max size should be valid",
		},
		{
			name: "very large max backups",
			config: Config{
				Outputs: []OutputConfig{
					{
						Type: "file",
						File: &FileOutputConfig{
							Dir:        "/tmp/logs",
							MaxBackups: 1000,
						},
					},
				},
			},
			expectError: false,
			description: "Large max backups should be valid",
		},
		{
			name: "very large max age",
			config: Config{
				Outputs: []OutputConfig{
					{
						Type: "file",
						File: &FileOutputConfig{
							Dir:    "/tmp/logs",
							MaxAge: 3650, // 10 years
						},
					},
				},
			},
			expectError: false,
			description: "Large max age should be valid",
		},
		{
			name: "all boolean combinations",
			config: Config{
				Outputs: []OutputConfig{
					{
						Type: "file",
						File: &FileOutputConfig{
							Dir:      "/tmp/logs",
							Compress: true,
						},
					},
					{
						Type: "console",
					},
				},
			},
			expectError: false,
			description: "Both compress and console enabled should be valid",
		},
		{
			name: "minimal valid config",
			config: Config{
				Outputs: []OutputConfig{
					{
						Type: "file",
						File: &FileOutputConfig{Dir: "/tmp/logs"},
					},
				},
			},
			expectError: false,
			description: "Only outputs should be required",
		},
		{
			name: "config with all fields set",
			config: Config{
				Level: "info",
				Outputs: []OutputConfig{
					{
						Type:   "file",
						Format: "json",
						File: &FileOutputConfig{
							Dir:        "/tmp/logs",
							MaxSize:    100,
							MaxBackups: 10,
							MaxAge:     30,
							Compress:   true,
						},
					},
					{
						Type:   "console",
						Format: "text",
					},
				},
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
			name:        "empty outputs error message",
			config:      Config{},
			expectedMsg: "log outputs cannot be empty",
		},
		{
			name: "invalid config value error message",
			config: Config{
				Outputs: []OutputConfig{
					{
						Type: "file",
						File: &FileOutputConfig{
							Dir:     "/tmp/logs",
							MaxSize: -1,
						},
					},
				},
			},
			expectedMsg: "invalid config value: max_size, max_backups, max_age must be non-negative",
		},
		{
			name: "invalid log level error message",
			config: Config{
				Level: "invalid",
				Outputs: []OutputConfig{
					{
						Type: "console",
					},
				},
			},
			expectedMsg: "invalid log level",
		},
		{
			name: "invalid log format error message",
			config: Config{
				Outputs: []OutputConfig{
					{
						Type:   "file",
						Format: "invalid",
						File:   &FileOutputConfig{Dir: "/tmp/logs"},
					},
				},
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
		Level: "info",
		Outputs: []OutputConfig{
			{
				Type:   "file",
				Format: "json",
				File: &FileOutputConfig{
					Dir:        "/tmp/logs",
					MaxSize:    100,
					MaxBackups: 10,
					MaxAge:     30,
					Compress:   true,
				},
			},
		},
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = config.Validate()
	}
}

func BenchmarkConfig_Validate_Minimal(b *testing.B) {
	config := Config{
		Outputs: []OutputConfig{
			{
				Type: "file",
				File: &FileOutputConfig{Dir: "/tmp/logs"},
			},
		},
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = config.Validate()
	}
}

func BenchmarkConfig_Validate_WithLevel(b *testing.B) {
	config := Config{
		Level: "info",
		Outputs: []OutputConfig{
			{
				Type: "console",
			},
		},
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = config.Validate()
	}
}

func BenchmarkConfig_Validate_WithFormat(b *testing.B) {
	config := Config{
		Outputs: []OutputConfig{
			{
				Type:   "file",
				Format: "json",
				File:   &FileOutputConfig{Dir: "/tmp/logs"},
			},
		},
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = config.Validate()
	}
}
