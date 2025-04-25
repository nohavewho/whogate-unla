package config

type (
	StorageConfig struct {
		Type     string            `yaml:"type"`     // disk or db
		Database DatabaseConfig    `yaml:"database"` // database configuration for db type
		Disk     DiskStorageConfig `yaml:"disk"`     // disk configuration for disk type
	}

	DiskStorageConfig struct {
		Path string `yaml:"path"` // path for disk storage
	}
)
