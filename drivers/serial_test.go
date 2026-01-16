package drivers

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestValidateSerialConfig tests the configuration validation logic.
func TestValidateSerialConfig(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name    string
		cfg     SerialConfig
		wantErr bool
		errMsg  string
	}{
		{
			name:    "valid config",
			cfg:     SerialConfig{Port: "/dev/ttyUSB0", Baud: 9600},
			wantErr: false,
		},
		{
			name:    "empty port",
			cfg:     SerialConfig{Port: "", Baud: 9600},
			wantErr: true,
			errMsg:  "Port is required",
		},
		{
			name:    "zero baud",
			cfg:     SerialConfig{Port: "/dev/ttyUSB0", Baud: 0},
			wantErr: true,
			errMsg:  "Baud must be > 0",
		},
		{
			name:    "negative baud",
			cfg:     SerialConfig{Port: "/dev/ttyUSB0", Baud: -9600},
			wantErr: true,
			errMsg:  "Baud must be > 0",
		},
		{
			name:    "both port and baud invalid",
			cfg:     SerialConfig{Port: "", Baud: 0},
			wantErr: true,
			errMsg:  "Port is required",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			err := validateSerialConfig(tt.cfg)
			if tt.wantErr {
				require.Error(t, err)
				assert.Contains(t, err.Error(), tt.errMsg)
			} else {
				require.NoError(t, err)
			}
		})
	}
}

// TestSerialConfig tests SerialConfig struct initialization.
func TestSerialConfig(t *testing.T) {
	t.Parallel()

	cfg := SerialConfig{
		Port: "/dev/ttyUSB0",
		Baud: 115200,
	}

	assert.Equal(t, "/dev/ttyUSB0", cfg.Port)
	assert.Equal(t, 115200, cfg.Baud)
}

// TestSerialConfigZeroValues tests zero value behavior.
func TestSerialConfigZeroValues(t *testing.T) {
	t.Parallel()

	var cfg SerialConfig
	err := validateSerialConfig(cfg)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "Port is required")
}
