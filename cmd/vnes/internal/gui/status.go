package gui

import (
	"time"
)

var _ Component = &Status{}

type Status struct {
	*Message
	Tag    string
	Flash  string
	Status string
	TTL    time.Time
}

func (s *Status) tag() string {
	return s.Tag
}

func (s *Status) Expired() bool {
	return !s.TTL.IsZero() && time.Now().After(s.TTL)
}

func (s *Status) SetFlashMsg(m string, delta time.Duration) {
	s.Flash = m
	s.TTL = time.Now().Add(delta)
}

func (s *Status) SetStatusMsg(m string) {
	s.Status = m
	s.Flash = ""
	s.TTL = time.Time{}
}

func (s *Status) Update(v *View) {
	if s.Disabled {
		return
	}

	if s.Expired() {
		s.Flash = ""
	}

	if s.Flash != "" {
		s.Text = s.Flash
		s.TTL = s.TTL
	} else {
		s.Text = s.Status
		s.TTL = time.Time{}
	}

	s.viewRect = *v.Rect
}

func (s *Status) Draw(v *View) error {
	if s.Disabled || s.Expired() {
		return nil
	}

	return s.Message.Draw(v)
}
