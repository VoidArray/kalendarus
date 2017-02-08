package main

import (
	"errors"
	"flag"
	"io/ioutil"
	"os"
	"time"

	"github.com/BurntSushi/toml"
	"github.com/Sirupsen/logrus"
	"github.com/leominov/kalendarus/backends"
	"github.com/leominov/kalendarus/backends/plainfile"
	"github.com/leominov/kalendarus/messengers"
	"github.com/leominov/kalendarus/messengers/telegram"
)

var (
	config            Config
	configFile        = ""
	defaultConfigFile = "/etc/kalendarus/kalendarus.toml"
	defaultTimezone   = "Asia/Yekaterinburg"
	printVersion      bool
	logLevel          string
	pullInterval      int
	calendarURL       string
	messenger         messengers.Messenger
	backend           backends.Backend
)

type Config struct {
	PullInterval   int                     `toml:"pull_interval"`
	NotifyInterval int                     `toml:"notify_interval"`
	NotifyTemplate string                  `toml:"notify_template"`
	LogLevel       string                  `toml:"log_level"`
	CalendarURL    string                  `toml:"calendar_url"`
	Timezone       string                  `toml:"timezone"`
	TimeFormat     string                  `toml:"time_format"`
	Notifications  map[string]Notification `toml:"notification"`
	Telegram       telegram.Config         `toml:"telegram"`
	Plainfile      plainfile.Config        `toml:"plainfile"`
	Location       *time.Location          `toml:"-"`
}

type Notification struct {
	BeforeStartRaw string        `toml:"before_start"`
	BeforeStart    time.Duration `toml:"-"`
}

func init() {
	flag.StringVar(&configFile, "config-file", "", "kalendarus config file")
	flag.StringVar(&calendarURL, "calendar-url", "", "url of the calendar")
	flag.IntVar(&pullInterval, "pull-interval", 1800, "calendar polling interval")
	flag.StringVar(&logLevel, "log-level", "", "level which kalendarus should log messages")
	flag.BoolVar(&printVersion, "version", false, "print version and exit")
}

func initConfig() error {
	if configFile == "" {
		if _, err := os.Stat(defaultConfigFile); !os.IsNotExist(err) {
			configFile = defaultConfigFile
		}
	}

	config = Config{
		PullInterval:   1800,
		NotifyInterval: 600,
		Timezone:       defaultTimezone,
	}

	config.Telegram = telegram.NewConfig()
	config.Plainfile = plainfile.NewConfig()

	if configFile == "" {
		logrus.Debug("Skipping kalendarus config file")
	} else {
		logrus.Debug("Loading " + configFile)
		configBytes, err := ioutil.ReadFile(configFile)
		if err != nil {
			return err
		}
		_, err = toml.Decode(string(configBytes), &config)
		if err != nil {
			return err
		}
	}

	processFlags()

	if config.LogLevel != "" {
		level, err := logrus.ParseLevel(config.LogLevel)
		if err != nil {
			return err
		}
		logrus.SetLevel(level)
	}

	if config.Timezone != "" {
		loc, err := time.LoadLocation(config.Timezone)
		if err != nil {
			return err
		}
		config.Location = loc
	}

	if config.CalendarURL == "" {
		return errors.New("CalendarURL must be specified")
	}

	if err := config.Telegram.Validate(); err != nil {
		return err
	}

	if err := config.Plainfile.Validate(); err != nil {
		return err
	}

	messenger = telegram.NewService(config.Telegram)
	backend = plainfile.NewService(config.Plainfile)

	for id, notify := range config.Notifications {
		d, err := time.ParseDuration(notify.BeforeStartRaw)
		if err != nil {
			return err
		}
		notify.BeforeStart = d
		config.Notifications[id] = notify
	}

	return nil
}

func processFlags() {
	flag.Visit(setConfigFromFlag)
}

func setConfigFromFlag(f *flag.Flag) {
	switch f.Name {
	case "pull-interval":
		config.PullInterval = pullInterval
	case "log-level":
		config.LogLevel = logLevel
	case "calendar-url":
		config.CalendarURL = calendarURL
	}
}
