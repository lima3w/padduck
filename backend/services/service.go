package services

type Service struct {
	// repositories will be injected here
}

func NewService() *Service {
	return &Service{}
}
