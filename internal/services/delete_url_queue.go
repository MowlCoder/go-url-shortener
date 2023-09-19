package services

import (
	"context"
	"fmt"
	"time"

	"github.com/MowlCoder/go-url-shortener/internal/domain"
)

type URLStorage interface {
	DoDeleteURLTasks(ctx context.Context, tasks []domain.DeleteURLsTask) error
}

type Logger interface {
	Info(msg string)
}

type DeleteURLQueue struct {
	ch         chan *domain.DeleteURLsTask
	urlStorage URLStorage
	logger     Logger
}

func NewDeleteURLQueue(urlStorage URLStorage, logger Logger, maxWorker int) *DeleteURLQueue {
	return &DeleteURLQueue{
		urlStorage: urlStorage,
		logger:     logger,
		ch:         make(chan *domain.DeleteURLsTask, maxWorker),
	}
}

func (q *DeleteURLQueue) Start() {
	ticker := time.NewTicker(time.Second * 5)
	deleteTasks := make([]domain.DeleteURLsTask, 0)

	go func() {
		for {
			select {
			case task := <-q.ch:
				deleteTasks = append(deleteTasks, *task)
			case <-ticker.C:
				if len(deleteTasks) == 0 {
					break
				}

				if err := q.urlStorage.DoDeleteURLTasks(context.Background(), deleteTasks); err != nil {
					q.logger.Info(err.Error())
					continue
				}

				q.logger.Info(fmt.Sprintf("Successfully did %d delete url tasks", len(deleteTasks)))
				deleteTasks = nil
			}
		}
	}()
}

func (q *DeleteURLQueue) Push(task *domain.DeleteURLsTask) {
	q.ch <- task
}
