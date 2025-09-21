package task

import (
	"errors"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
)

type Module struct {
	conn *pgxpool.Pool
}

var (
	ErrNoTasks = errors.New("no tasks")
)

func New(conn *pgxpool.Pool) *Module {
	return &Module{
		conn: conn,
	}
}

func (m *Module) GetForSplitting(hostname string) (Task, error) {
	return Task{}, nil
}

func (m *Module) TaskCanceled(id uuid.UUID) (bool, error) {
	return false, nil
}
