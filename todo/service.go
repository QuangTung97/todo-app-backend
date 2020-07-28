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

func domainTodoItemToDTO(item todoItem) *TodoItem {
	return &TodoItem{
		Id:          int32(item.id),
		TodoListId:  int32(item.todoListID),
		Description: item.description,
		Completed:   item.completed,
		CreatedAt:   timestamppb.New(item.createdAt),
	}
}

// CreateTodoItem create a todo item
func (s *Service) CreateTodoItem(
	ctx context.Context,
	in *CreateTodoItemRequest,
) (*CreateTodoItemResponse, error) {
	accountID := getAccountID(ctx)

	var item todoItem
	var err error

	err = s.repo.transact(ctx, func(tx *sqlx.Tx) error {
		item, err = createTodoItem(
			int(in.TodoListId), accountID,
			in.Description,
			s.repo.getTodoList(ctx, tx),
			s.repo.createTodoItem(ctx, tx),
		)
		return err
	})
	return &CreateTodoItemResponse{
		Item: domainTodoItemToDTO(item),
	}, err
}

// GetTodoItems return list of todo items
func (s *Service) GetTodoItems(
	ctx context.Context,
	in *GetTodoItemsRequest,
) (*GetTodoItemsResponse, error) {
	accountID := getAccountID(ctx)

	items, err := selectTodoItems(
		int(in.TodoListId), accountID,
		s.repo.getTodoListNoTx(ctx),
		s.repo.selectTodoItemsNoTx(ctx),
	)
	if err != nil {
		return nil, err
	}

	result := make([]*TodoItem, 0)
	for _, item := range items {
		result = append(result, domainTodoItemToDTO(item))
	}

	return &GetTodoItemsResponse{
		TodoItems: result,
	}, nil
}

// UpdateTodoItemsCompleted update completed fields
func (s *Service) UpdateTodoItemsCompleted(
	ctx context.Context,
	in *UpdateTodoItemsCompletedRequest,
) (*UpdateTodoItemsCompletedResponse, error) {
	accountID := getAccountID(ctx)

	toBeCompleted := make([]int, 0)
	for _, e := range in.ToBeCompletedIds {
		toBeCompleted = append(toBeCompleted, int(e))
	}

	toBeActive := make([]int, 0)
	for _, e := range in.ToBeActiveIds {
		toBeActive = append(toBeActive, int(e))
	}

	err := s.repo.transact(ctx, func(tx *sqlx.Tx) error {
		return updateTodoItemsCompleted(
			int(in.TodoListId), accountID,
			toBeCompleted, toBeActive,
			s.repo.getTodoList(ctx, tx),
			s.repo.selectTodoItems(ctx, tx),
			s.repo.updateTodoItemsCompleted(ctx, tx),
		)
	})
	return &UpdateTodoItemsCompletedResponse{}, err
}

// DeleteTodoItemsCompleted delete completed todo items
func (s *Service) DeleteTodoItemsCompleted(
	ctx context.Context,
	in *DeleteTodoItemsCompletedRequest,
) (*DeleteTodoItemsCompletedResponse, error) {
	accountID := getAccountID(ctx)

	err := s.repo.transact(ctx, func(tx *sqlx.Tx) error {
		return deleteTodoItemsCompleted(
			int(in.TodoListId), accountID,
			s.repo.getTodoList(ctx, tx),
			s.repo.deleteTodoItemsCompleted(ctx, tx),
		)
	})

	return &DeleteTodoItemsCompletedResponse{}, err
}
