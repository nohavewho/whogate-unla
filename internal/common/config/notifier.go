package config

type (
	// NotifierConfig represents the configuration for notifier
	NotifierConfig struct {
		Type   string       `yaml:"type"`
		Signal SignalConfig `yaml:"signal"`
		API    APIConfig    `yaml:"api"`
		Redis  RedisConfig  `yaml:"redis"`
	}

	// SignalConfig represents the configuration for signal-based notifier
	SignalConfig struct {
		Signal string `yaml:"signal"`
		PID    string `yaml:"pid"`
	}

	// APIConfig represents the configuration for API-based notifier
	APIConfig struct {
		Port int `yaml:"port"`
	}

	// RedisConfig represents the configuration for Redis-based notifier
	RedisConfig struct {
		Addr     string `yaml:"addr"`
		Password string `yaml:"password"`
		DB       int    `yaml:"db"`
		Topic    string `yaml:"topic"`
	}
)
