package main

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"net/http"
	"sync"
	"time"

	"github.com/Sirupsen/logrus"
	"github.com/apognu/gocal"
	"github.com/leominov/kalendarus/backends"
	"github.com/leominov/kalendarus/messengers"
)

type Processor struct {
	config     Config
	stopChan   chan bool
	doneChan   chan bool
	errChan    chan error
	messenger  messengers.Messenger
	backend    backends.Backend
	wg         sync.WaitGroup
	cache      Cache
	saved      bool
	firstStart bool
}

func NewProcessor(config Config, stopChan, doneChan chan bool, errChan chan error, messenger messengers.Messenger, backend backends.Backend) *Processor {
	p := &Processor{
		config:     config,
		stopChan:   stopChan,
		doneChan:   doneChan,
		errChan:    errChan,
		messenger:  messenger,
		backend:    backend,
		saved:      true,
		firstStart: true,
	}
	p.cache = NewCache(config)
	return p
}

func (p *Processor) Process() {
	defer close(p.doneChan)

	p.LoadState()
	p.wg.Add(2)

	go p.PullDaemon()
	go p.NotifyDaemon()

	p.wg.Wait()
}

func (p *Processor) PullDaemon() {
	defer p.wg.Done()
	for {
		if err := p.realProcess(); err != nil {
			p.errChan <- err
		}
		select {
		case <-p.stopChan:
			break
		case <-time.After(time.Duration(p.config.PullInterval) * time.Second):
			continue
		}
	}
}

func (p *Processor) NotifyDaemon() {
	defer p.wg.Done()
	for {
		if err := p.notifyProcess(); err != nil {
			p.errChan <- err
		}
		select {
		case <-p.stopChan:
			break
		case <-time.After(time.Duration(p.config.NotifyInterval) * time.Second):
			continue
		}
	}
}

func (p *Processor) notifyProcess() error {
	if len(p.cache.Events) == 0 {
		return nil
	}
	cloc := time.Now().In(p.config.Location)
	for id, e := range p.cache.Events {
		notified := false
		diff := e.StartTimeLocal.Sub(cloc)
		for _, notify := range p.config.Notifications {
			if _, ok := e.Notificators[notify.BeforeStartRaw]; !ok {
				e.Notificators[notify.BeforeStartRaw] = false
			}
			if !e.Notificators[notify.BeforeStartRaw] && diff <= notify.BeforeStart {
				if notified {
					e.Notificators[notify.BeforeStartRaw] = true
					p.cache.Events[id] = e
					p.saved = false
					logrus.Debugf("SKIP_NOTIFY [%s] %s LESS %s (%s)", e.StartTimeLocal, e.Summary, notify.BeforeStart.String(), id)
					continue
				}
				notified = true
				if !p.firstStart {
					logrus.Infof("NOTIFY [%s] %s LESS %s (%s)", e.StartTimeLocal, e.Summary, notify.BeforeStart.String(), id)
					err := p.messenger.Send(
						fmt.Sprintf(p.config.NotifyTemplate, e.Summary, e.StartTimeLocal.Format(p.config.TimeFormat), e.Location, e.Description),
					)
					if err != nil {
						logrus.Error(err)
						continue
					}
				} else {
					logrus.Debugf("SKIP_ON_START [%s] %s LESS %s (%s)", e.StartTimeLocal, e.Summary, notify.BeforeStart.String(), id)
				}
				e.Notificators[notify.BeforeStartRaw] = true
				p.cache.Events[id] = e
				p.saved = false
			}
		}
	}
	p.firstStart = false
	return nil
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
		if !p.cache.IsActiveEvent(e) {
			continue
		}
		if !p.cache.HasEvent(e) {
			p.saved = false
			p.cache.SaveEvent(e)
			logrus.Infof("ADDED [%s] %s", e.Start.In(p.config.Location), e.Summary)
			continue
		}
		if changed, err := p.cache.IsEventChanged(e); err == nil && changed == true {
			p.saved = false
			p.cache.SaveEvent(e)
			logrus.Infof("UPDATED [%s] %s", e.Start.In(p.config.Location), e.Summary)
		}
	}
	if err := p.SaveState(); err != nil {
		p.errChan <- err
	}
	return nil
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
	if err := p.backend.Load("kalendarus/events/cache", &p.cache.Events); err != nil {
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
	if err := p.backend.Save("kalendarus/events/cache", p.cache.Events); err != nil {
		return err
	}
	p.saved = true
	logrus.Debug("Ok")
	return nil
}
