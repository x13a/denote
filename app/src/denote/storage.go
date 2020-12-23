package denote

import (
	"database/sql"
	"sync"
	"time"

	"github.com/google/uuid"
	_ "github.com/mattn/go-sqlite3"
)

const CleanerInterval = 24 * time.Hour

var db = &database{}

type database struct {
	sync.Mutex
	db *sql.DB
}

func (d *database) open(dsn string) (err error) {
	d.db, err = sql.Open("sqlite3", dsn)
	return
}

func (d *database) close() error {
	return d.db.Close()
}

func (d *database) create() (err error) {
	tx, err := d.db.Begin()
	if err != nil {
		return
	}
	defer func() {
		if err != nil {
			tx.Rollback()
			return
		}
		err = tx.Commit()
	}()
	_, err = tx.Exec(`
		CREATE TABLE IF NOT EXISTS "denote" (
			"key" BLOB NOT NULL PRIMARY KEY,
			"data" BLOB NOT NULL,
			"view_count" INT NOT NULL DEFAULT 0,
			"view_limit" INT NOT NULL DEFAULT 1,
			"dt_limit" DATETIME NOT NULL
		)
	`)
	if err != nil {
		return
	}
	_, err = tx.Exec(`
		CREATE INDEX IF NOT EXISTS "key_dt_limit_index"
		ON "denote" ("key", "dt_limit")
	`)
	if err != nil {
		return
	}
	_, err = tx.Exec(`
		CREATE INDEX IF NOT EXISTS "view_count_limit_index"
		ON "denote" ("view_count", "view_limit")
	`)
	if err != nil {
		return
	}
	_, err = tx.Exec(`
		CREATE INDEX IF NOT EXISTS "dt_limit_index" 
		ON "denote" ("dt_limit")
	`)
	return
}

func (d *database) cleaner(stopChan chan struct{}) {
	ticker := time.NewTicker(CleanerInterval)
	defer ticker.Stop()
Loop:
	for {
		select {
		case <-stopChan:
			break Loop
		case <-ticker.C:
			d.clean()
		}
	}
	close(stopChan)
}

func (d *database) clean() (err error) {
	d.Lock()
	defer d.Unlock()
	_, err = d.db.Exec(`
		DELETE 
		FROM "denote" 
		WHERE 
			"view_count" >= "view_limit"
			OR
			datetime('now') >= "dt_limit"
	`)
	return
}

func (d *database) get(uid uuid.UUID) (data []byte, err error) {
	d.Lock()
	defer d.Unlock()
	tx, err := d.db.Begin()
	if err != nil {
		return
	}
	defer func() {
		if err != nil {
			tx.Rollback()
			return
		}
		err = tx.Commit()
	}()
	stmt, err := tx.Prepare(`
		SELECT "data", "view_count", "view_limit" 
		FROM "denote" 
		WHERE 
			"key" = ?
			AND
			datetime('now') < "dt_limit"
	`)
	if err != nil {
		return
	}
	defer stmt.Close()
	var viewCount int
	var viewLimit int
	if err = stmt.QueryRow(uid).Scan(
		&data,
		&viewCount,
		&viewLimit,
	); err != nil {
		return
	}
	viewCount++
	if viewCount < viewLimit {
		_, err = tx.Exec(
			`UPDATE "denote" SET "view_count" = ? WHERE "key" = ?`,
			viewCount,
			uid,
		)
	} else {
		_, err = tx.Exec(`DELETE FROM "denote" WHERE "key" = ?`, uid)
	}
	return
}

func (d *database) set(
	data []byte,
	viewLimit int,
	durationLimit time.Duration,
) (uuid.UUID, error) {
	uid, err := uuid.NewRandom()
	if err != nil {
		return uuid.Nil, err
	}
	d.Lock()
	defer d.Unlock()
	if _, err = d.db.Exec(
		`INSERT INTO "denote" (
			"key", 
			"data", 
			"view_limit", 
			"dt_limit"
		) VALUES (?, ?, ?, ?)`,
		uid,
		data,
		viewLimit,
		time.Now().Add(durationLimit).UTC(),
	); err != nil {
		return uuid.Nil, err
	}
	return uid, nil
}
