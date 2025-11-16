package storage

import (
	"database/sql"
	"fmt"
	"time"

	_ "github.com/mattn/go-sqlite3"

	"github.com/andreygolubkow/we-know/internal/config"
)

type Store struct {
	db *sql.DB
}

type AnalysisRecord struct {
	ID               int64
	FeatureID        string
	AnalyzedAt       time.Time
	FilesChanged     int
	LocAdded         int
	LocDeleted       int
	MonolithsTouched int
}

func New(cfg *config.Config) (*Store, error) {
	dsn := cfg.DB.Path
	db, err := sql.Open("sqlite3", dsn)
	if err != nil {
		return nil, fmt.Errorf("open sqlite db: %w", err)
	}

	// Немного тюнинга
	if _, err := db.Exec(`PRAGMA journal_mode = WAL;`); err != nil {
		_ = db.Close()
		return nil, fmt.Errorf("set WAL mode: %w", err)
	}
	if _, err := db.Exec(`PRAGMA foreign_keys = ON;`); err != nil {
		_ = db.Close()
		return nil, fmt.Errorf("enable foreign_keys: %w", err)
	}

	if err := migrate(db); err != nil {
		_ = db.Close()
		return nil, fmt.Errorf("migrate schema: %w", err)
	}

	return &Store{db: db}, nil
}

func migrate(db *sql.DB) error {
	schema := `
CREATE TABLE IF NOT EXISTS analyses (
    id                 INTEGER PRIMARY KEY AUTOINCREMENT,
    feature_id         TEXT NOT NULL,
    analyzed_at        DATETIME NOT NULL,
    files_changed      INTEGER NOT NULL,
    loc_added          INTEGER NOT NULL,
    loc_deleted        INTEGER NOT NULL,
    monoliths_touched  INTEGER NOT NULL
);

CREATE INDEX IF NOT EXISTS idx_analyses_feature_id
    ON analyses(feature_id);
`
	if _, err := db.Exec(schema); err != nil {
		return err
	}
	return nil
}

// Close — закрыть БД.
func (s *Store) Close() error {
	if s == nil || s.db == nil {
		return nil
	}
	return s.db.Close()
}

func (s *Store) SaveAnalysis(a *AnalysisRecord) error {
	if a.AnalyzedAt.IsZero() {
		a.AnalyzedAt = time.Now()
	}

	res, err := s.db.Exec(
		`INSERT INTO analyses (feature_id, analyzed_at, files_changed, loc_added, loc_deleted, monoliths_touched)
         VALUES (?, ?, ?, ?, ?, ?)`,
		a.FeatureID,
		a.AnalyzedAt,
		a.FilesChanged,
		a.LocAdded,
		a.LocDeleted,
		a.MonolithsTouched,
	)
	if err != nil {
		return fmt.Errorf("insert analysis: %w", err)
	}

	id, err := res.LastInsertId()
	if err == nil {
		a.ID = id
	}

	return nil
}
