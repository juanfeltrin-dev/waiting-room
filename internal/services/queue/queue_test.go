package queue_test

import (
	"context"
	"errors"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"testing"
	"waitingroom/internal/model"
	queuesvc "waitingroom/internal/services/queue"
)

// Mock Repositories
type MockLoadRepository struct {
	mock.Mock
}

func (m *MockLoadRepository) GetStatus(ctx context.Context) model.Load {
	args := m.Called(ctx)
	return args.Get(0).(model.Load)
}

func (m *MockLoadRepository) IsMember(ctx context.Context, sessionID string) bool {
	args := m.Called(ctx, sessionID)
	return args.Bool(0)
}

func (m *MockLoadRepository) Increment(ctx context.Context, sessionID string) {
	m.Called(ctx, sessionID)
}

func (m *MockLoadRepository) Decrement(ctx context.Context, sessionID string) {
	m.Called(ctx, sessionID)
}

type MockQueueRepository struct {
	mock.Mock
}

func (m *MockQueueRepository) Enter(ctx context.Context, sessionID string) {
	m.Called(ctx, sessionID)
}

func (m *MockQueueRepository) Exit(ctx context.Context, sessionID string) {
	m.Called(ctx, sessionID)
}

func (m *MockQueueRepository) GetPosition(ctx context.Context, sessionID string) model.Queue {
	args := m.Called(ctx, sessionID)
	return args.Get(0).(model.Queue)
}

func (m *MockQueueRepository) IsMember(ctx context.Context, sessionID string) bool {
	args := m.Called(ctx, sessionID)
	return args.Bool(0)
}

func (m *MockQueueRepository) First(ctx context.Context) (string, error) {
	args := m.Called(ctx)
	return args.String(0), args.Error(1)
}

type MockSessionRepository struct {
	mock.Mock
}

func (m *MockSessionRepository) Init(ctx context.Context, sessionID string) error {
	args := m.Called(ctx, sessionID)
	return args.Error(0)
}

func (m *MockSessionRepository) Exit(ctx context.Context, sessionID string) {
	m.Called(ctx, sessionID)
}

func (m *MockSessionRepository) Exist(ctx context.Context, sessionID string) bool {
	args := m.Called(ctx, sessionID)

	return args.Bool(0)
}

func (m *MockSessionRepository) GetAverageQueueTime(ctx context.Context) int64 {
	args := m.Called(ctx)

	return args.Get(0).(int64)
}

func setupTestService() (*queuesvc.QueueService, *MockLoadRepository, *MockQueueRepository, *MockSessionRepository) {
	loadRepo := &MockLoadRepository{}
	queueRepo := &MockQueueRepository{}
	sessionRepo := &MockSessionRepository{}

	return &queuesvc.QueueService{
		LoadRepository:    loadRepo,
		QueueRepository:   queueRepo,
		SessionRepository: sessionRepo,
	}, loadRepo, queueRepo, sessionRepo
}

func TestQueueService_Enter(t *testing.T) {
	tests := []struct {
		name           string
		setupMocks     func(*MockLoadRepository, *MockQueueRepository, *MockSessionRepository)
		expectedResult bool
	}{
		{
			name: "User already in queue",
			setupMocks: func(ml *MockLoadRepository, mq *MockQueueRepository, ms *MockSessionRepository) {
				mq.On("IsMember", mock.Anything, "session123").Return(true)
			},
			expectedResult: true,
		},
		{
			name: "User already in load",
			setupMocks: func(ml *MockLoadRepository, mq *MockQueueRepository, ms *MockSessionRepository) {
				mq.On("IsMember", mock.Anything, "session123").Return(false)
				ml.On("IsMember", mock.Anything, "session123").Return(true)
			},
			expectedResult: false,
		},
		{
			name: "Capacity not reached, add to load",
			setupMocks: func(ml *MockLoadRepository, mq *MockQueueRepository, ms *MockSessionRepository) {
				mq.On("IsMember", mock.Anything, "session123").Return(false)
				ml.On("IsMember", mock.Anything, "session123").Return(false)
				ml.On("GetStatus", mock.Anything).Return(model.Load{Count: 0})
				ml.On("Increment", mock.Anything, "session123").Once()
			},
			expectedResult: false,
		},
		{
			name: "Capacity reached, add to queue",
			setupMocks: func(ml *MockLoadRepository, mq *MockQueueRepository, ms *MockSessionRepository) {
				mq.On("IsMember", mock.Anything, "session123").Return(false)
				ml.On("IsMember", mock.Anything, "session123").Return(false)
				ml.On("GetStatus", mock.Anything).Return(model.Load{Count: 10})
				ms.On("Init", mock.Anything, "session123").Return(nil)
				mq.On("Enter", mock.Anything, "session123").Once()
			},
			expectedResult: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			service, loadRepo, queueRepo, sessionRepo := setupTestService()
			tt.setupMocks(loadRepo, queueRepo, sessionRepo)

			result := service.Enter(context.Background(), "session123")

			assert.Equal(t, tt.expectedResult, result)
			loadRepo.AssertExpectations(t)
			queueRepo.AssertExpectations(t)
			sessionRepo.AssertExpectations(t)
		})
	}
}

func TestQueueService_GetPosition(t *testing.T) {
	tests := []struct {
		name          string
		setupMocks    func(*MockQueueRepository)
		expectedError string
		expectedQueue model.Queue
	}{
		{
			name: "User in queue",
			setupMocks: func(mq *MockQueueRepository) {
				mq.On("GetPosition", mock.Anything, "session123").
					Return(model.NewQueue(0))
			},
			expectedQueue: model.NewQueue(0),
		},
		{
			name: "User not in queue",
			setupMocks: func(mq *MockQueueRepository) {
				mq.On("GetPosition", mock.Anything, "session123").
					Return(model.NewQueue(-1))
			},
			expectedError: "user is not in queue",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			service, _, queueRepo, _ := setupTestService()
			tt.setupMocks(queueRepo)

			result, err := service.GetPosition(context.Background(), "session123")

			if tt.expectedError != "" {
				assert.EqualError(t, err, tt.expectedError)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.expectedQueue.GetPosition(), result.GetPosition())
			}
			queueRepo.AssertExpectations(t)
		})
	}
}

func TestQueueService_Exit(t *testing.T) {
	tests := []struct {
		name           string
		setupMocks     func(*MockLoadRepository)
		expectedResult bool
	}{
		{
			name: "User in load, decrement",
			setupMocks: func(ml *MockLoadRepository) {
				ml.On("IsMember", mock.Anything, "session123").Return(true)
				ml.On("Decrement", mock.Anything, "session123").Once()
			},
			expectedResult: true,
		},
		{
			name: "User not in load",
			setupMocks: func(ml *MockLoadRepository) {
				ml.On("IsMember", mock.Anything, "session123").Return(false)
			},
			expectedResult: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			service, loadRepo, _, _ := setupTestService()
			tt.setupMocks(loadRepo)

			result := service.Exit(context.Background(), "session123")

			assert.Equal(t, tt.expectedResult, result)
			loadRepo.AssertExpectations(t)
		})
	}
}

func TestQueueService_ReleaseEntry(t *testing.T) {
	tests := []struct {
		name       string
		setupMocks func(*MockLoadRepository, *MockQueueRepository, *MockSessionRepository)
		expectedID string
	}{
		{
			name: "Reached capacity, do nothing",
			setupMocks: func(ml *MockLoadRepository, mq *MockQueueRepository, ms *MockSessionRepository) {
				ml.On("GetStatus", mock.Anything).Return(model.Load{Count: 10})
			},
			expectedID: "",
		},
		{
			name: "No one in queue",
			setupMocks: func(ml *MockLoadRepository, mq *MockQueueRepository, ms *MockSessionRepository) {
				ml.On("GetStatus", mock.Anything).Return(model.Load{Count: 0})
				mq.On("First", mock.Anything).Return("", errors.New("queue empty"))
			},
			expectedID: "",
		},
		{
			name: "Release user from queue",
			setupMocks: func(ml *MockLoadRepository, mq *MockQueueRepository, ms *MockSessionRepository) {
				ml.On("GetStatus", mock.Anything).Return(model.Load{Count: 0})
				mq.On("First", mock.Anything).Return("session123", nil)
				ml.On("Increment", mock.Anything, "session123").Once()
				mq.On("Exit", mock.Anything, "session123").Once()
				ms.On("Exit", mock.Anything, "session123").Once()
			},
			expectedID: "session123",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			service, loadRepo, queueRepo, sessionRepo := setupTestService()
			tt.setupMocks(loadRepo, queueRepo, sessionRepo)

			result := service.ReleaseEntry(context.Background())

			assert.Equal(t, tt.expectedID, result)
			loadRepo.AssertExpectations(t)
			queueRepo.AssertExpectations(t)
			sessionRepo.AssertExpectations(t)
		})
	}
}

func TestQueueService_GetAverageQueueTime(t *testing.T) {
	tests := []struct {
		name         string
		setupMocks   func(*MockSessionRepository)
		expectedTime int64
	}{
		{
			name: "Calculate average time",
			setupMocks: func(ms *MockSessionRepository) {
				ms.On("GetAverageQueueTime", mock.Anything).Return(int64(100)).Once()
			},
			expectedTime: 100, // Adjust based on your test case
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			service, _, _, sessionRepo := setupTestService()
			tt.setupMocks(sessionRepo)

			result := service.GetAverageQueueTime(context.Background())

			assert.Equal(t, tt.expectedTime, result)
			sessionRepo.AssertExpectations(t)
		})
	}
}
