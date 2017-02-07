package telegram

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"net/url"
	"path"
	"sync/atomic"

	"github.com/leominov/kalendarus/messengers"
	"github.com/pkg/errors"
)

type Service struct {
	configValue atomic.Value
}

func NewService(c Config) messengers.Messenger {
	s := &Service{}
	s.configValue.Store(c)
	return s
}

func (s *Service) config() Config {
	return s.configValue.Load().(Config)
}

func (s *Service) Send(message string) error {
	url, post, err := s.preparePost(message)
	if err != nil {
		return err
	}

	resp, err := http.Post(url, "application/json", post)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		body, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			return err
		}
		type response struct {
			Description string `json:"description"`
			ErrorCode   int    `json:"error_code"`
			Ok          bool   `json:"ok"`
		}
		res := &response{}

		err = json.Unmarshal(body, res)

		if err != nil {
			return fmt.Errorf("failed to understand Telegram response (err: %s). code: %d content: %s", err.Error(), resp.StatusCode, string(body))
		}
		return fmt.Errorf("sendMessage error (%d) description: %s", res.ErrorCode, res.Description)
	}
	return nil
}

func (s *Service) preparePost(message string) (string, io.Reader, error) {
	c := s.config()

	if !c.Enabled {
		return "", nil, errors.New("service is not enabled")
	}

	if c.ParseMode != "" && c.ParseMode != "Markdown" && c.ParseMode != "HTML" {
		return "", nil, fmt.Errorf("parseMode %s is not valid, please use 'Markdown' or 'HTML'", c.ParseMode)
	}

	postData := make(map[string]interface{})
	postData["chat_id"] = c.ChatId
	postData["text"] = message

	if c.ParseMode != "" {
		postData["parse_mode"] = c.ParseMode
	}

	if c.DisableWebPagePreview {
		postData["disable_web_page_preview"] = true
	}

	if c.DisableNotification {
		postData["disable_notification"] = true
	}

	var post bytes.Buffer
	enc := json.NewEncoder(&post)
	err := enc.Encode(postData)
	if err != nil {
		return "", nil, err
	}

	u, err := url.Parse(c.URL)
	if err != nil {
		return "", nil, errors.Wrap(err, "invalid URL")
	}

	u.Path = path.Join(u.Path+c.Token, "sendMessage")
	return u.String(), &post, nil
}
