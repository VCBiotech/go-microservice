package file

import (
	"context"
	"database/sql"

	"vcbiotech/microservice/telemetry"
)

type SQLRepo struct {
	Client *sql.DB
}

func (r *SQLRepo) FindById(ctx context.Context, id uint64) (File, error) {
	logger := telemetry.SLogger(ctx)
	logger.Info("Not implemented.")
	return File{}, nil
}
