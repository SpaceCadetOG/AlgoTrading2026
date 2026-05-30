package system

type ServiceStatus string

const (
	ServiceCreated ServiceStatus = "CREATED"
	ServiceRunning ServiceStatus = "RUNNING"
	ServiceStopped ServiceStatus = "STOPPED"
	ServiceError   ServiceStatus = "ERROR"
)

type Service interface {
	Name() string
	Start() error
	Stop() error
	Status() ServiceStatus
}

type BaseService struct {
	name   string
	status ServiceStatus
	err    error
}

func NewBaseService(name string) *BaseService {
	return &BaseService{name: name, status: ServiceCreated}
}

func (s *BaseService) Name() string {
	return s.name
}

func (s *BaseService) Start() error {
	s.status = ServiceRunning
	s.err = nil
	return nil
}

func (s *BaseService) Stop() error {
	s.status = ServiceStopped
	s.err = nil
	return nil
}

func (s *BaseService) Status() ServiceStatus {
	return s.status
}

func (s *BaseService) setError(err error) error {
	s.err = err
	s.status = ServiceError
	return err
}
