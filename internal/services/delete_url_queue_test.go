package services

import (
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/mock/gomock"

	"github.com/MowlCoder/go-url-shortener/internal/domain"
	servicesmocks "github.com/MowlCoder/go-url-shortener/internal/services/mocks"
)

func TestNewDeleteURLQueue(t *testing.T) {
	ctrl := gomock.NewController(t)
	urlStorageInstance := servicesmocks.NewMockurlStorage(ctrl)
	loggerInstance := servicesmocks.NewMocklogger(ctrl)
	maxWorker := 10

	t.Run("new", func(t *testing.T) {
		queue := NewDeleteURLQueue(urlStorageInstance, loggerInstance, maxWorker)
		require.NotNil(t, queue)
		assert.Equal(t, cap(queue.ch), maxWorker)
	})
}

func TestDeleteURLQueue_Push(t *testing.T) {
	ctrl := gomock.NewController(t)
	urlStorageInstance := servicesmocks.NewMockurlStorage(ctrl)
	loggerInstance := servicesmocks.NewMocklogger(ctrl)
	maxWorker := 10
	queue := NewDeleteURLQueue(urlStorageInstance, loggerInstance, maxWorker)

	t.Run("valid", func(t *testing.T) {
		queue.Push(&domain.DeleteURLsTask{})
		assert.Equal(t, len(queue.ch), 1)

		queue.Push(&domain.DeleteURLsTask{})
		assert.Equal(t, len(queue.ch), 2)
	})
}

func TestDeleteURLQueue_doDeleteTasks(t *testing.T) {
	ctrl := gomock.NewController(t)
	urlStorageInstance := servicesmocks.NewMockurlStorage(ctrl)
	loggerInstance := servicesmocks.NewMocklogger(ctrl)
	maxWorker := 10
	queue := NewDeleteURLQueue(urlStorageInstance, loggerInstance, maxWorker)

	type TestCase struct {
		PrepareServiceFunc func()
		Name               string
		Tasks              []domain.DeleteURLsTask
		IsError            bool
	}

	testCases := []TestCase{
		{
			Name:    "valid",
			IsError: false,
			PrepareServiceFunc: func() {
				urlStorageInstance.
					EXPECT().
					DoDeleteURLTasks(gomock.Any(), gomock.Any()).
					Return(nil)

				loggerInstance.
					EXPECT().
					Info(gomock.Any())
			},
			Tasks: []domain.DeleteURLsTask{
				{
					UserID: "1",
				},
				{
					UserID: "2",
				},
			},
		},
		{
			Name:    "valid (no tasks)",
			IsError: false,
		},
		{
			Name:    "invalid",
			IsError: true,
			PrepareServiceFunc: func() {
				urlStorageInstance.
					EXPECT().
					DoDeleteURLTasks(gomock.Any(), gomock.Any()).
					Return(errors.New("undefined behavior"))
			},
			Tasks: []domain.DeleteURLsTask{
				{
					UserID: "1",
				},
				{
					UserID: "2",
				},
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.Name, func(t *testing.T) {
			queue.tasks = tc.Tasks

			if tc.PrepareServiceFunc != nil {
				tc.PrepareServiceFunc()
			}

			err := queue.doDeleteTasks()

			if tc.IsError {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
			}
		})
	}
}
