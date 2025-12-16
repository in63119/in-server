package visitor

type Service struct{}

func New() *Service { return &Service{} }

func (s *Service) List() ([]any, error) {
	return nil, nil
}

func (s *Service) Create(v any) error {
	return nil
}

// HasVisited is a placeholder that always returns false.
func (s *Service) HasVisited(ip string) (bool, error) {
	return false, nil
}
