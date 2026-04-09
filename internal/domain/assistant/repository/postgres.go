// package repository
package repository

import (
	"assistant/internal/config"
	"context"

	"github.com/jackc/pgx/v5/pgxpool"
	"go.uber.org/zap"
)

type Repository struct {
	ctx context.Context
	db  *pgxpool.Pool
	log *zap.Logger
	cfg *config.Config
}

func NewRepo(ctx context.Context, logger *zap.Logger, cfg *config.Config) *Repository {
	return &Repository{
		ctx: ctx,
		cfg: cfg,
		log: logger,
	}
}

func (r *Repository) OnStart(_ context.Context) (err error) {
	r.db, err = pgxpool.New(r.ctx, r.cfg.DSN)
	if err != nil {
		r.log.Error("failed to connect to database", zap.Error(err))
		return err
	}

	return nil
}
