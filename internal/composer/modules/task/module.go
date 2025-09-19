package task

import (
	"errors"

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
