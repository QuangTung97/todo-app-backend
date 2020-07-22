package todo

import (
	"context"
	"github.com/jmoiron/sqlx"
)

// Service : gRPC endpoint
type Service struct {
	repo *repository
}

// NewService : create a new service
func NewService(db *sqlx.DB) *Service {
	repo := newRepository(db)
	return &Service{
		repo: repo,
	}
}

// CreateAccount : create a new account
func (s *Service) CreateAccount(
	ctx context.Context,
	in *CreateAccountRequest,
) (*CreateAccountResponse, error) {
	err := createAccount(
		s.repo.saveAccount(ctx),
		in.Username, in.Password,
	)

	return &CreateAccountResponse{}, err
}
