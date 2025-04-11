package order

import (
	"context"
	"database/sql"

	"vcbiotech/microservice/telemetry"
)

type SQLRepo struct {
	Client *sql.DB
}

func (r *SQLRepo) Insert(ctx context.Context, order Order) error {
	logger := telemetry.SLogger(ctx)
	logger.Info("Not implemented.")
	return nil
}

func (r *SQLRepo) FindById(ctx context.Context, id uint64) (Order, error) {
	logger := telemetry.SLogger(ctx)
	logger.Info("Not implemented.")
	return Order{}, nil
}

func (r *SQLRepo) DeleteById(ctx context.Context, id uint64) error {
	logger := telemetry.SLogger(ctx)
	logger.Info("Not implemented.")
	return nil
}

func (r *SQLRepo) Update(ctx context.Context, order Order) error {
	logger := telemetry.SLogger(ctx)
	logger.Info("Not implemented.")
	return nil
}

type FindAllPage struct {
	Size   uint64
	Offset uint64
}

type FindResult struct {
	Orders []Order
	Cursor uint64
}

func (r *SQLRepo) FindAll(ctx context.Context, page FindAllPage) (FindResult, error) {

	cursor := uint64(1)
	orders := make([]Order, 1)

	result := FindResult{
		Orders: orders,
		Cursor: cursor,
	}

	return result, nil
}
