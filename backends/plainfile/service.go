package plainfile

import (
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"
	"sync/atomic"

	"github.com/leominov/kalendarus/backends"
)

var (
	replacer = strings.NewReplacer("/", "_")
)

type Service struct {
	configValue atomic.Value
}

func NewService(c Config) backends.Backend {
	s := &Service{}
	s.configValue.Store(c)
	return s
}

func (s *Service) config() Config {
	return s.configValue.Load().(Config)
}

func transform(key string) string {
	k := strings.TrimPrefix(key, "/")
	return strings.ToLower(replacer.Replace(k))
}

func (s *Service) Load(key string, vars interface{}) error {
	c := s.config()

	if !c.Enabled {
		return errors.New("service is not enabled")
	}

	filename := filepath.Join(c.DataDir, fmt.Sprintf("%s.json", transform(key)))
	if _, err := os.Stat(filename); err != nil {
		return err
	}
	file, err := ioutil.ReadFile(filename)
	if err != nil {
		return err
	}
	err = json.Unmarshal(file, &vars)
	if err != nil {
		return err
	}
	return nil
}

func (s *Service) Save(key string, vars interface{}) error {
	c := s.config()

	if !c.Enabled {
		return errors.New("service is not enabled")
	}

	filename := filepath.Join(c.DataDir, fmt.Sprintf("%s.json", transform(key)))
	keyFile, err := os.OpenFile(filename, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0666)
	if err != nil {
		return fmt.Errorf("Could not open file: %v", err)
	}
	defer keyFile.Close()
	json, err := json.Marshal(vars)
	if err != nil {
		return err
	}
	_, err = keyFile.WriteString(string(json))
	if err != nil {
		return fmt.Errorf("Could not write to file: %s", err)
	}
	return nil
}
