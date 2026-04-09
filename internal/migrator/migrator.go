// Package migrator - пакет для миграции
package migrator

import (
	"assistant/internal/config"
	"context"
	"errors"
	"fmt"
	"os"
	"path/filepath"

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

func detectMigrationsDir() (string, error) {
	if wd, err := os.Getwd(); err == nil {
		p := filepath.Join(wd, "migrations")
		if st, err := os.Stat(p); err == nil && st.IsDir() {
			return p, nil
		}
	}

	if exe, err := os.Executable(); err == nil {
		exeDir := filepath.Dir(exe)
		candidates := []string{
			filepath.Join(exeDir, "migrations"),
			filepath.Clean(filepath.Join(exeDir, "..", "migrations")),
		}
		for _, p := range candidates {
			if st, err := os.Stat(p); err == nil && st.IsDir() {
				return p, nil
			}
		}
	}

	return "", fmt.Errorf("migrations directory not found (tried cwd and executable dir)")
}

func NewMigrator(cfg *config.Config, log *zap.Logger) (*Migrator, error) {
	migrationsDir, err := detectMigrationsDir()
	if err != nil {
		log.Error("failed to detect migrations directory", zap.Error(err))
		return nil, err
	}

	sourceURL := "file://" + migrationsDir
	log.Info("migrator configured", zap.String("source", sourceURL))

	m, err := migrate.New(
		sourceURL,
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
