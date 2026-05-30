package system

type SystemSupervisor struct {
	services []Service
}

type ServiceSummary struct {
	Name   string        `json:"name"`
	Status ServiceStatus `json:"status"`
}

func NewSystemSupervisor() *SystemSupervisor {
	return &SystemSupervisor{}
}

func NewChapter7Supervisor() *SystemSupervisor {
	supervisor := NewSystemSupervisor()
	supervisor.Register(NewLoggingService())
	supervisor.Register(NewPositionService())
	supervisor.Register(NewMarketDataService())
	supervisor.Register(NewOrderService())
	supervisor.Register(NewRiskService())
	return supervisor
}

func (s *SystemSupervisor) Register(service Service) {
	s.services = append(s.services, service)
}

func (s *SystemSupervisor) StartAll() error {
	for _, service := range s.services {
		if err := service.Start(); err != nil {
			return err
		}
	}
	return nil
}

func (s *SystemSupervisor) StopAll() error {
	for i := len(s.services) - 1; i >= 0; i-- {
		if err := s.services[i].Stop(); err != nil {
			return err
		}
	}
	return nil
}

func (s *SystemSupervisor) Summary() []ServiceSummary {
	out := make([]ServiceSummary, 0, len(s.services))
	for _, service := range s.services {
		out = append(out, ServiceSummary{Name: service.Name(), Status: service.Status()})
	}
	return out
}

func (s *SystemSupervisor) CountByStatus(status ServiceStatus) int {
	count := 0
	for _, service := range s.services {
		if service.Status() == status {
			count++
		}
	}
	return count
}

func (s *SystemSupervisor) ServiceCount() int {
	return len(s.services)
}
