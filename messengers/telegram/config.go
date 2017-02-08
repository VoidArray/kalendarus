package telegram

import (
	"net/url"

	"github.com/pkg/errors"
)

const DefaultTelegramURL = "https://api.telegram.org/bot"
const DefaultTelegramLinksPreviewDisable = false
const DefaultTelegramNotificationDisable = false

type Config struct {
	Enabled               bool   `toml:"enabled"`
	URL                   string `toml:"url"`
	Token                 string `toml:"token"`
	ChatId                string `toml:"chat_id"`
	ParseMode             string `toml:"parse_mode"`
	DisableWebPagePreview bool   `toml:"disable_web_page_preview"`
	DisableNotification   bool   `toml:"disable_notification"`
}

func NewConfig() Config {
	return Config{
		URL: DefaultTelegramURL,
		DisableWebPagePreview: DefaultTelegramLinksPreviewDisable,
		DisableNotification:   DefaultTelegramNotificationDisable,
	}
}

func (c Config) Validate() error {
	if c.Enabled {
		if c.URL == "" {
			return errors.New("must specify url")
		}
		if c.Token == "" {
			return errors.New("must specify token")
		}
	}
	if _, err := url.Parse(c.URL); err != nil {
		return errors.Wrapf(err, "invalid url %q", c.URL)
	}
	return nil
}
