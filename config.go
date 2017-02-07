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
	config Config

	configFile        = ""
	defaultConfigFile = "/etc/kalendarus/kalendarus.toml"
	defaultTimezone   = "Asia/Yekaterinburg"
	printVersion      bool
	logLevel          string
	interval          int
	calendarURL       string
	timezone          string

	messenger messengers.Messenger
	backend   backends.Backend
)

type Config struct {
	Interval           int              `toml:"interval"`
	LogLevel           string           `toml:"log-level"`
	CalendarURL        string           `toml:"calendar-url"`
	Telegram           telegram.Config  `toml:"telegram"`
	Plainfile          plainfile.Config `toml:"plainfile"`
	Timezone           string           `toml:"timezone"`
	TimeFormat         string           `toml:"time_format"`
	AddedEventFormat   string           `toml:"added_event_format"`
	UpdatedEventFormat string           `toml:"updated_event_format"`
	SkipFirstStart     bool             `toml:"skip_first_start"`
	Location           *time.Location   `toml:"-"`
}

func init() {
	flag.StringVar(&configFile, "config-file", "", "kalendarus config file")
	flag.StringVar(&calendarURL, "calendar-url", "", "url of the calendar")
	flag.IntVar(&interval, "interval", 600, "data polling interval")
	flag.StringVar(&logLevel, "log-level", "", "level which kalendarus should log messages")
	flag.BoolVar(&printVersion, "version", false, "print version and exit")
	flag.StringVar(&timezone, "timezone", "", "convert events to specific timezone")
}

func initConfig() error {
	if configFile == "" {
		if _, err := os.Stat(defaultConfigFile); !os.IsNotExist(err) {
			configFile = defaultConfigFile
		}
	}

	config = Config{
		Interval: 600,
		Timezone: defaultTimezone,
	}

	config.Telegram = telegram.NewConfig()
	config.Plainfile = plainfile.NewConfig()

	if configFile == "" {
		logrus.Debug("Skipping kalendarus config file.")
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

	return nil
}

func processFlags() {
	flag.Visit(setConfigFromFlag)
}

func setConfigFromFlag(f *flag.Flag) {
	switch f.Name {
	case "interval":
		config.Interval = interval
	case "log-level":
		config.LogLevel = logLevel
	case "calendar-url":
		config.CalendarURL = calendarURL
	}
}
