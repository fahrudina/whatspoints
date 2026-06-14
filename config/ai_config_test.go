package config

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestLoadAIConfig_Parsing(t *testing.T) {
	cases := []struct {
		name        string
		enableVal   string
		setEnable   bool
		wantEnabled bool
	}{
		{"missing defaults disabled", "", false, false},
		{"empty disabled", "", true, false},
		{"true enabled", "true", true, true},
		{"one enabled", "1", true, true},
		{"yes enabled", "yes", true, true},
		{"on enabled", "on", true, true},
		{"TRUE case-insensitive", "TRUE", true, true},
		{"garbage disabled", "maybe", true, false},
		{"false disabled", "false", true, false},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			t.Setenv("AI_SERVICE_URL", "") // isolate
			if tc.setEnable {
				t.Setenv("ENABLE_AI_RESPONSE", tc.enableVal)
			}
			cfg := LoadAIConfig()
			assert.Equal(t, tc.wantEnabled, cfg.Enabled)
		})
	}
}

func TestLoadAIConfig_ServiceURLDefaultOnlyWhenEnabled(t *testing.T) {
	t.Run("disabled leaves URL empty", func(t *testing.T) {
		t.Setenv("ENABLE_AI_RESPONSE", "false")
		t.Setenv("AI_SERVICE_URL", "")
		cfg := LoadAIConfig()
		assert.False(t, cfg.Enabled)
		assert.Empty(t, cfg.ServiceURL)
	})

	t.Run("enabled defaults URL", func(t *testing.T) {
		t.Setenv("ENABLE_AI_RESPONSE", "true")
		t.Setenv("AI_SERVICE_URL", "")
		cfg := LoadAIConfig()
		assert.True(t, cfg.Enabled)
		assert.Equal(t, "http://localhost:8090", cfg.ServiceURL)
	})

	t.Run("enabled honors explicit URL", func(t *testing.T) {
		t.Setenv("ENABLE_AI_RESPONSE", "1")
		t.Setenv("AI_SERVICE_URL", "http://ai-agent:8090")
		cfg := LoadAIConfig()
		assert.Equal(t, "http://ai-agent:8090", cfg.ServiceURL)
	})
}

func TestLoadAIConfig_AutoSendDefaultsFalse(t *testing.T) {
	t.Setenv("ENABLE_AI_AUTO_SEND", "")
	cfg := LoadAIConfig()
	assert.False(t, cfg.AutoSend)

	t.Setenv("ENABLE_AI_AUTO_SEND", "true")
	cfg = LoadAIConfig()
	assert.True(t, cfg.AutoSend)
}
