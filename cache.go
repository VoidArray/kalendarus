package main

import (
	"errors"
	"time"

	"github.com/apognu/gocal"
)

type Cache struct {
	config Config
	Events map[string]Event
}

type Event struct {
	Summary           string          `json:"summary"`
	Location          string          `json:"location"`
	Description       string          `json:"description"`
	StartTime         time.Time       `json:"start_time"`
	StartTimeLocal    time.Time       `json:"start_time_local"`
	EndTime           time.Time       `json:"end_time"`
	EndTimeLocal      time.Time       `json:"end_time_local"`
	ModifiedTime      time.Time       `json:"modified_time"`
	ModifiedTimeLocal time.Time       `json:"modified_time_local"`
	Notificators      map[string]bool `json:"notificators"`
}

func NewCache(config Config) Cache {
	return Cache{
		config: config,
		Events: make(map[string]Event),
	}
}

func (c *Cache) HasEvent(e gocal.Event) bool {
	if _, ok := c.Events[e.Uid]; ok {
		return true
	}
	return false
}

func (c *Cache) IsEventChanged(e gocal.Event) (bool, error) {
	if !c.HasEvent(e) {
		return false, errors.New("event not found")
	}
	event := c.Events[e.Uid]
	if !event.ModifiedTime.Equal(*e.LastModified) {
		return true, nil
	}
	return false, nil
}

func (c *Cache) IsActiveEvent(e gocal.Event) bool {
	if e.Summary == "" || e.Status != "CONFIRMED" || e.Start == nil {
		return false
	}
	sloc := e.Start.In(c.config.Location)
	if time.Now().In(c.config.Location).After(sloc) {
		return false
	}
	return true
}

func (c *Cache) SaveEvent(e gocal.Event) {
	event := Event{
		Summary:        e.Summary,
		Location:       e.Location,
		Description:    e.Description,
		StartTime:      *e.Start,
		StartTimeLocal: e.Start.In(c.config.Location),
		Notificators:   make(map[string]bool),
	}
	if e.End != nil {
		event.EndTime = *e.End
		event.EndTimeLocal = e.End.In(c.config.Location)
	}
	if e.LastModified != nil {
		event.ModifiedTime = *e.LastModified
		event.ModifiedTimeLocal = e.LastModified.In(c.config.Location)
	}
	c.Events[e.Uid] = event
}
