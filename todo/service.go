package todo

import (
	"context"
	"strconv"

	"github.com/golang/glog"
	"github.com/jmoiron/sqlx"
	"google.golang.org/grpc/metadata"
	"google.golang.org/protobuf/types/known/timestamppb"
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

func getAccountID(ctx context.Context) int {
	md, ok := metadata.FromIncomingContext(ctx)
	if !ok {
		glog.Fatal("can't read metadata")
	}
	strList := md.Get("Account-ID")
	if len(strList) == 0 {
		glog.Fatal("can't read metadata")
	}
	id, err := strconv.Atoi(strList[0])
	if err != nil {
		glog.Fatal("can't read metadata", err)
	}
	return id
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

func domainTodoToDTO(todo todoList) *TodoList {
	return &TodoList{
		Id:        int32(todo.id),
		AccountId: int32(todo.accountID),
		Name:      todo.name,
		CreatedAt: timestamppb.New(todo.createdAt),
		UpdatedAt: timestamppb.New(todo.updatedAt),
	}
}

// CreateTodoList create a todo list
func (s *Service) CreateTodoList(
	ctx context.Context,
	in *CreateTodoListRequest,
) (*CreateTodoListResponse, error) {
	id := getAccountID(ctx)

	todo, err := createTodoList(id, in.Name,
		s.repo.saveTodoList(ctx),
	)

	return &CreateTodoListResponse{
		Todo: domainTodoToDTO(todo),
	}, err
}

// UpdateTodoList update the name of todo list
func (s *Service) UpdateTodoList(
	ctx context.Context,
	in *UpdateTodoListRequest,
) (*UpdateTodoListResponse, error) {
	accountID := getAccountID(ctx)

	var todo todoList
	err := s.repo.transact(ctx, func(tx *sqlx.Tx) error {
		tmp, err := updateTodoList(int(
			in.Id), accountID, in.Name,
			s.repo.getTodoList(ctx, tx),
			s.repo.updateTodoList(ctx, tx),
		)
		todo = tmp
		return err
	})

	return &UpdateTodoListResponse{
		Todo: domainTodoToDTO(todo),
	}, err
}

// GetTodoList get all of todos from user
func (s *Service) GetTodoList(
	ctx context.Context,
	in *GetTodoListRequest,
) (*GetTotoListResponse, error) {
	accountID := getAccountID(ctx)
	todos, err := s.repo.getTodoListsByAccount(ctx)(accountID)
	if err != nil {
		return &GetTotoListResponse{}, err
	}

	result := make([]*TodoList, 0)
	for _, t := range todos {
		result = append(result, domainTodoToDTO(t))
	}

	return &GetTotoListResponse{Todos: result}, nil
}

// DeleteTodoList delete a todo list
func (s *Service) DeleteTodoList(
	ctx context.Context,
	in *DeleteTodoListResquest,
) (*DeleteTodoListResponse, error) {
	accountID := getAccountID(ctx)
	err := s.repo.transact(ctx, func(tx *sqlx.Tx) error {
		return deleteTodoList(
			int(in.Id), accountID,
			s.repo.getTodoList(ctx, tx),
			s.repo.deleteTodoList(ctx, tx),
		)
	})
	return &DeleteTodoListResponse{}, err
}
