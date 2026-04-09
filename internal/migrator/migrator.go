// Package migrator - пакет для миграции
package migrator

import (
	"assistant/internal/config"
	"context"
	"errors"

	"github.com/golang-migrate/migrate/v4"
	_ "github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
	_ "github.com/jackc/pgx/v4/stdlib"
	"go.uber.org/zap"
)

type Migrator struct {
	*migrate.Migrate
	log *zap.Logger
}

// RunMigrations запускает миграции базы данных.
func NewMigrator(cfg *config.Config, log *zap.Logger) (*Migrator, error) {
	m, err := migrate.New(
		"file://migrations",
		cfg.DSN,
	)
	if err != nil {
		log.Error("failed to create migrator", zap.Error(err))
		return nil, err
	}

	return &Migrator{m, log}, nil
}

func (m *Migrator) OnStart(_ context.Context) error {
	defer m.Close()

	if err := m.Up(); err != nil && !errors.Is(err, migrate.ErrNoChange) {
		m.log.Error("failed to migrate", zap.Error(err))
		return err
	}

	m.log.Info("migrator up done")

	return nil
}
