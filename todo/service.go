package todo

import (
	"context"
	"fmt"
	"strconv"

	"github.com/golang/glog"
	"github.com/jmoiron/sqlx"
	"google.golang.org/grpc/metadata"
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

func getAccountID(ctx context.Context) (int, bool) {
	md, ok := metadata.FromIncomingContext(ctx)
	if !ok {
		return 0, false
	}
	strList := md.Get("Account-ID")
	if len(strList) == 0 {
		return 0, false
	}
	id, err := strconv.Atoi(strList[0])
	if err != nil {
		return 0, false
	}
	return id, true
}

// CreateAccount : create a new account
func (s *Service) CreateAccount(
	ctx context.Context,
	in *CreateAccountRequest,
) (*CreateAccountResponse, error) {
	id, ok := getAccountID(ctx)
	if !ok {
		glog.Fatal("can't read metadata")
	}

	fmt.Println("Account id", id)

	err := createAccount(
		s.repo.saveAccount(ctx),
		in.Username, in.Password,
	)

	return &CreateAccountResponse{}, err
}
