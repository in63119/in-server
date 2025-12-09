package post

type Service struct{}

func New( /* deps */ ) *Service { return &Service{} }

type Post struct {
	ID    string `json:"id"`
	Title string `json:"title"`
	Body  string `json:"body"`
}

func (s *Service) List() ([]Post, error) { return nil, nil }
func (s *Service) Create(p Post) error   { return nil }
