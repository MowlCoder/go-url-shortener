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
	tasks      []domain.DeleteURLsTask
}

func NewDeleteURLQueue(urlStorage URLStorage, logger Logger, maxWorker int) *DeleteURLQueue {
	return &DeleteURLQueue{
		urlStorage: urlStorage,
		logger:     logger,
		ch:         make(chan *domain.DeleteURLsTask, maxWorker),
		tasks:      make([]domain.DeleteURLsTask, 0, 500),
	}
}

func (q *DeleteURLQueue) Start(ctx context.Context) {
	ticker := time.NewTicker(time.Second * 5)
	defer ticker.Stop()

	for {
		select {
		case task := <-q.ch:
			q.tasks = append(q.tasks, *task)
		case <-ctx.Done():
			if err := q.doDeleteTasks(); err != nil {
				q.logger.Info(err.Error())
			}
		case <-ticker.C:
			if err := q.doDeleteTasks(); err != nil {
				q.logger.Info(err.Error())
			}
		}
	}
}

func (q *DeleteURLQueue) Push(task *domain.DeleteURLsTask) {
	q.ch <- task
}

func (q *DeleteURLQueue) doDeleteTasks() error {
	if len(q.tasks) == 0 {
		return nil
	}

	if err := q.urlStorage.DoDeleteURLTasks(context.Background(), q.tasks); err != nil {
		return err
	}

	q.logger.Info(fmt.Sprintf("Successfully did %d delete url tasks", len(q.tasks)))
	q.tasks = q.tasks[0:]
	return nil
}
