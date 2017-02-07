package main

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"net/http"
	"time"

	"github.com/Sirupsen/logrus"
	"github.com/apognu/gocal"
	"github.com/leominov/kalendarus/backends"
	"github.com/leominov/kalendarus/messengers"
)

type Processor struct {
	config    Config
	stopChan  chan bool
	doneChan  chan bool
	errChan   chan error
	messenger messengers.Messenger
	backend   backends.Backend
	events    map[string]Event
	saved     bool
}

type Event struct {
	Summary           string    `json:"summary"`
	Location          string    `json:"location"`
	Description       string    `json:"description"`
	StartTime         time.Time `json:"start_time"`
	StartTimeLocal    time.Time `json:"start_time_local"`
	EndTime           time.Time `json:"end_time"`
	EndTimeLocal      time.Time `json:"end_time_local"`
	ModifiedTime      time.Time `json:"modified_time"`
	ModifiedTimeLocal time.Time `json:"modified_time_local"`
}

func NewProcessor(config Config, stopChan, doneChan chan bool, errChan chan error, messenger messengers.Messenger, backend backends.Backend) *Processor {
	p := &Processor{
		config:    config,
		stopChan:  stopChan,
		doneChan:  doneChan,
		errChan:   errChan,
		messenger: messenger,
		backend:   backend,
		events:    make(map[string]Event),
		saved:     true,
	}
	return p
}

func (p *Processor) Process() {
	defer close(p.doneChan)
	p.LoadState()
	for {
		if err := p.realProcess(); err != nil {
			p.errChan <- err
		}
		select {
		case <-p.stopChan:
			break
		case <-time.After(time.Duration(p.config.Interval) * time.Second):
			if err := p.SaveState(); err != nil {
				p.errChan <- err
			}
			continue
		}
	}
}

func (p *Processor) realProcess() error {
	events, err := p.getCalendarEvents()
	if err != nil {
		return err
	}
	if len(events) == 0 {
		return nil
	}
	for _, e := range events {
		if e.Summary == "" || e.Status != "CONFIRMED" || e.Start == nil {
			continue
		}
		sloc := e.Start.In(p.config.Location)
		if time.Now().In(p.config.Location).After(sloc) {
			continue
		}
		if _, ok := p.events[e.Uid]; !ok {
			p.saved = false
			event := Event{
				Summary:        e.Summary,
				Location:       e.Location,
				Description:    e.Description,
				StartTime:      *e.Start,
				StartTimeLocal: sloc,
			}
			if e.End != nil {
				event.EndTime = *e.End
				event.EndTimeLocal = e.End.In(p.config.Location)
			}
			if e.LastModified != nil {
				event.ModifiedTime = *e.LastModified
				event.ModifiedTimeLocal = e.LastModified.In(p.config.Location)
			}
			p.events[e.Uid] = event
			p.sendMessage(event, false)
		} else {
			event := p.events[e.Uid]
			if !event.ModifiedTime.Equal(*e.LastModified) && !event.StartTime.Equal(*e.Start) {
				p.saved = false
				event.ModifiedTime = *e.LastModified
				event.ModifiedTimeLocal = e.LastModified.In(p.config.Location)
				event.StartTime = *e.Start
				event.StartTimeLocal = sloc
				event.Summary = e.Summary
				event.Description = e.Description
				p.events[e.Uid] = event
				p.sendMessage(event, true)
			}
		}
	}
	if p.config.SkipFirstStart {
		p.config.SkipFirstStart = false
		if err := p.SaveState(); err != nil {
			p.errChan <- err
		}
	}
	return nil
}

func (p *Processor) sendMessage(event Event, updated bool) error {
	if p.config.SkipFirstStart {
		logrus.Infof("Skip sending message about %s", event.Summary)
		return nil
	}
	template := p.config.AddedEventFormat
	if updated {
		template = p.config.UpdatedEventFormat
		logrus.Debugf("UPD [%s] %s", event.StartTimeLocal.Format(p.config.TimeFormat), event.Summary)
	} else {
		logrus.Debugf("NEW [%s] %s", event.StartTimeLocal.Format(p.config.TimeFormat), event.Summary)
	}
	return p.messenger.Send(
		fmt.Sprintf(template, event.Summary, event.StartTimeLocal.Format(p.config.TimeFormat), event.Location, event.Description),
	)
}

func (p *Processor) getCalendarEvents() ([]gocal.Event, error) {
	var events []gocal.Event
	resp, err := http.Get(p.config.CalendarURL)
	if err != nil {
		return events, err
	}
	b, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return events, err
	}
	err = resp.Body.Close()
	if err != nil {
		return events, err
	}
	gc := gocal.NewParser(bytes.NewReader(b))
	gc.Parse()
	return gc.Events, nil
}

func (p *Processor) LoadState() error {
	logrus.Debug("Loading state...")
	if err := p.backend.Load("kalendarus/events", &p.events); err != nil {
		return err
	}
	logrus.Debug("Ok")
	return nil
}

func (p *Processor) SaveState() error {
	if p.saved == true {
		return nil
	}
	logrus.Debug("Saving state...")
	if err := p.backend.Save("kalendarus/events", p.events); err != nil {
		return err
	}
	p.saved = true
	logrus.Debug("Ok")
	return nil
}
