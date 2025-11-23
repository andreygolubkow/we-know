package storage

import (
	"database/sql"
	"fmt"
	"strings"

	_ "github.com/mattn/go-sqlite3"

	"github.com/andreygolubkow/we-know/internal/config"
)

type Store struct {
	db *sql.DB
}

type Features struct {
	Id        int64
	FeatureID string
	FileId    int64
}

type Files struct {
	Id          int64
	ProjectName string
	Path        string
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
CREATE TABLE IF NOT EXISTS files (
    id   INTEGER PRIMARY KEY AUTOINCREMENT,
    project_name TEXT NOT NULL,
    path TEXT NOT NULL,
    UNIQUE(project_name, path)
);

CREATE TABLE IF NOT EXISTS features (
    id         INTEGER PRIMARY KEY AUTOINCREMENT,
    feature_id TEXT NOT NULL,
    file_id    INTEGER NOT NULL,
    FOREIGN KEY(file_id) REFERENCES files(id),
    UNIQUE(feature_id, file_id)
);

CREATE INDEX IF NOT EXISTS idx_features_feature_id
    ON features(feature_id);

CREATE INDEX IF NOT EXISTS idx_features_file_id
    ON features(file_id);
`
	_, err := db.Exec(schema)
	return err
}

// Close — закрыть БД.
func (s *Store) Close() error {
	if s == nil || s.db == nil {
		return nil
	}
	return s.db.Close()
}

func (s *Store) AppendAnalysis(featureID string, projectName string, changedFiles []string) (err error) {
	if featureID == "" {
		return fmt.Errorf("featureID is empty")
	}
	if projectName == "" {
		return fmt.Errorf("projectName is empty")
	}
	if len(changedFiles) == 0 {
		return nil
	}

	tx, err := s.db.Begin()
	if err != nil {
		return fmt.Errorf("begin tx: %w", err)
	}
	defer func() {
		if err != nil {
			_ = tx.Rollback()
		}
	}()

	for _, path := range changedFiles {
		path = strings.TrimSpace(path)
		if path == "" {
			continue
		}

		// 1) вставляем файл (или игнорируем, если уже есть такой project_name+path)
		_, err = tx.Exec(
			`INSERT OR IGNORE INTO files (project_name, path) VALUES (?, ?)`,
			projectName,
			path,
		)
		if err != nil {
			return fmt.Errorf("insert file %q (%s): %w", path, projectName, err)
		}

		// 2) получаем id файла по project_name + path
		var fileID int64
		err = tx.QueryRow(
			`SELECT id FROM files WHERE project_name = ? AND path = ?`,
			projectName,
			path,
		).Scan(&fileID)
		if err != nil {
			return fmt.Errorf("select file id for %q (%s): %w", path, projectName, err)
		}

		// 3) создаем связь feature ↔ file
		_, err = tx.Exec(
			`INSERT OR IGNORE INTO features (feature_id, file_id) VALUES (?, ?)`,
			featureID,
			fileID,
		)
		if err != nil {
			return fmt.Errorf("insert feature-file link %q/%d: %w", featureID, fileID, err)
		}
	}

	if err = tx.Commit(); err != nil {
		return fmt.Errorf("commit tx: %w", err)
	}

	return nil
}
