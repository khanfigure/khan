package khan

import (
	"errors"
)

type Service struct {
	Name string

	Running bool
	Enabled bool

	id int
}

func (s *Service) String() string {
	return s.Name
}

func (s *Service) SetID(id int) {
	s.id = id
}
func (s *Service) ID() int {
	return s.id
}
func (s *Service) Clone() Item {
	r := *s
	r.id = 0
	return &r
}

func (s *Service) Validate() error {
	if s.Name == "" {
		return errors.New("Service name is required")
	}
	return nil
}

func (s *Service) StaticFiles() []string {
	return nil
}

func (s *Service) After() []string {
	return nil
}
func (s *Service) Before() []string {
	return nil
}
func (s *Service) Provides() []string {
	return []string{"service:" + s.Name}
}

func (s *Service) Apply(host *Host) (Status, error) {
	return Created, nil
}
