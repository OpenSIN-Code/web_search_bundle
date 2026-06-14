// SPDX-License-Identifier: MIT
// Purpose: SQLite persistent storage for Alchemist experiment runs.
// Docs: history.doc.md

package alchemist

import (
	"database/sql"
	"fmt"
	"os"
	"path/filepath"
	"time"

	_ "modernc.org/sqlite"
)

// History persists experiment records to SQLite
type History struct {
	db   *sql.DB
	path string
}

// NewHistory creates a new history store
func NewHistory(repoPath string) (*History, error) {
	dbPath := filepath.Join(repoPath, ".sin-code", "alchemist.db")

	if err := os.MkdirAll(filepath.Dir(dbPath), 0750); err != nil {
		return nil, err
	}

	db, err := sql.Open("sqlite", dbPath)
	if err != nil {
		return nil, err
	}

	schema := `
	CREATE TABLE IF NOT EXISTS experiments (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		timestamp DATETIME NOT NULL,
		hypothesis TEXT NOT NULL,
		metric_before REAL,
		metric_after REAL,
		delta REAL,
		duration_ns INTEGER,
		decision TEXT NOT NULL,
		commit_sha TEXT,
		stdout_snippet TEXT
	);
	CREATE INDEX IF NOT EXISTS idx_experiments_timestamp ON experiments(timestamp);
	CREATE INDEX IF NOT EXISTS idx_experiments_decision ON experiments(decision);
	`
	if _, err := db.Exec(schema); err != nil {
		_ = db.Close()
		return nil, fmt.Errorf("schema: %w", err)
	}

	return &History{db: db, path: dbPath}, nil
}

func (h *History) Insert(r ExperimentRecord) error {
	_, err := h.db.Exec(`
		INSERT INTO experiments (timestamp, hypothesis, metric_before, metric_after, delta, duration_ns, decision, commit_sha, stdout_snippet)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?)`,
		r.Timestamp, r.Hypothesis, r.MetricBefore, r.MetricAfter, r.Delta,
		r.Duration.Nanoseconds(), r.Decision, r.CommitSHA, r.StdoutSnippet)
	return err
}

func (h *History) All() ([]ExperimentRecord, error) {
	rows, err := h.db.Query(`
		SELECT id, timestamp, hypothesis, metric_before, metric_after, delta, duration_ns, decision, commit_sha
		FROM experiments ORDER BY timestamp DESC`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var records []ExperimentRecord
	for rows.Next() {
		var r ExperimentRecord
		var dur int64
		var sha sql.NullString
		if err := rows.Scan(&r.ID, &r.Timestamp, &r.Hypothesis, &r.MetricBefore,
			&r.MetricAfter, &r.Delta, &dur, &r.Decision, &sha); err != nil {
			return nil, err
		}
		r.Duration = time.Duration(dur)
		if sha.Valid {
			r.CommitSHA = sha.String
		}
		records = append(records, r)
	}
	return records, nil
}

func (h *History) Summary() (map[string]any, error) {
	var total, committed, discarded, errors int
	err := h.db.QueryRow(`
		SELECT
			COUNT(*),
			COALESCE(SUM(CASE WHEN decision = 'committed' THEN 1 ELSE 0 END), 0),
			COALESCE(SUM(CASE WHEN decision = 'discarded' THEN 1 ELSE 0 END), 0),
			COALESCE(SUM(CASE WHEN decision IN ('error', 'timeout', 'commit-failed') THEN 1 ELSE 0 END), 0)
		FROM experiments`).Scan(&total, &committed, &discarded, &errors)
	if err != nil {
		return nil, err
	}

	var bestDelta sql.NullFloat64
	_ = h.db.QueryRow(`SELECT MAX(delta) FROM experiments WHERE decision = 'committed'`).Scan(&bestDelta)

	var totalRuntime int64
	_ = h.db.QueryRow(`SELECT COALESCE(SUM(duration_ns), 0) FROM experiments`).Scan(&totalRuntime)

	return map[string]any{
		"total_experiments": total,
		"committed":         committed,
		"discarded":         discarded,
		"errors":            errors,
		"success_rate":      safeRate(committed, total),
		"best_delta":        bestDelta.Float64,
		"total_runtime":     time.Duration(totalRuntime).String(),
	}, nil
}

func (h *History) Recent(n int) ([]ExperimentRecord, error) {
	rows, err := h.db.Query(`
		SELECT id, timestamp, hypothesis, metric_before, metric_after, delta, duration_ns, decision, commit_sha
		FROM experiments ORDER BY timestamp DESC LIMIT ?`, n)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var records []ExperimentRecord
	for rows.Next() {
		var r ExperimentRecord
		var dur int64
		var sha sql.NullString
		if err := rows.Scan(&r.ID, &r.Timestamp, &r.Hypothesis, &r.MetricBefore,
			&r.MetricAfter, &r.Delta, &dur, &r.Decision, &sha); err != nil {
			return nil, err
		}
		r.Duration = time.Duration(dur)
		if sha.Valid {
			r.CommitSHA = sha.String
		}
		records = append(records, r)
	}
	return records, nil
}

func (h *History) Close() error {
	return h.db.Close()
}

func safeRate(num, den int) float64 {
	if den == 0 {
		return 0
	}
	return float64(num) / float64(den)
}
