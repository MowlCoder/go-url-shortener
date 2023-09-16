package services

import (
	"context"
	"fmt"

	"github.com/MowlCoder/go-url-shortener/internal/domain"
)

type URLStorage interface {
	DeleteByShortURLs(ctx context.Context, shortURLs []string, userID string) error
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
	go func() {
		for task := range q.ch {
			if err := q.urlStorage.DeleteByShortURLs(context.Background(), task.ShortURLs, task.UserID); err != nil {
				q.logger.Info(err.Error())
				continue
			}

			q.logger.Info(fmt.Sprintf("Successfully deleted %d urls for user %s", len(task.ShortURLs), task.UserID))
		}
	}()
}

func (q *DeleteURLQueue) Push(task *domain.DeleteURLsTask) {
	q.ch <- task
}

func (q *DeleteURLQueue) Close() {
	close(q.ch)
}
