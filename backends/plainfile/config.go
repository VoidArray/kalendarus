package plainfile

import (
	"os"

	"github.com/pkg/errors"
)

const DefaultDataDir = "/etc/kalendarus/data"

type Config struct {
	Enabled bool   `toml:"enabled"`
	DataDir string `toml:"data_dir"`
}

func NewConfig() Config {
	return Config{
		DataDir: DefaultDataDir,
	}
}

func (c Config) Validate() error {
	if c.Enabled {
		if c.DataDir == "" {
			return errors.New("must specify data directory")
		}
		if _, err := os.Stat(c.DataDir); os.IsNotExist(err) {
			return err
		}
	}
	return nil
}
