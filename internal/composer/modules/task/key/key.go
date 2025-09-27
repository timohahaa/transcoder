package key

import "github.com/google/uuid"

func Skip(taskID uuid.UUID) string {
	return "transcoder:" + taskID.String() + ":skip"
}

func Counter(taskID uuid.UUID) string {
	return "transcoder:" + taskID.String() + ":counter"
}
