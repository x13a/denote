package db

import (
	"context"
	"database/sql"
	"time"

	"github.com/google/uuid"
	_ "github.com/mattn/go-sqlite3"

	"github.com/x13a/denote/config"
)

var db *sql.DB

func Open() (err error) {
	defer func() {
		config.DSN = ""
	}()
	db, err = sql.Open("sqlite3", config.DSN)
	if err != nil {
		db.SetMaxOpenConns(1)
	}
	return
}

func Close() error {
	return db.Close()
}

func Create(ctx context.Context) (err error) {
	tx, err := db.BeginTx(ctx, nil)
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
	if _, err = tx.ExecContext(ctx, `
		CREATE TABLE IF NOT EXISTS "denote" (
			"key" BLOB NOT NULL PRIMARY KEY,
			"data" BLOB NOT NULL,
			"view_count" INT NOT NULL DEFAULT 0,
			"view_limit" INT NOT NULL DEFAULT 1,
			"dt_limit" DATETIME NOT NULL,
			"rm_key" BLOB NOT NULL
		)
	`); err != nil {
		return
	}
	if _, err = tx.ExecContext(ctx, `
		CREATE INDEX IF NOT EXISTS "key_dt_limit_index"
		ON "denote" ("key", "dt_limit")
	`); err != nil {
		return
	}
	if _, err = tx.ExecContext(ctx, `
		CREATE INDEX IF NOT EXISTS "dt_limit_index" 
		ON "denote" ("dt_limit")
	`); err != nil {
		return
	}
	_, err = tx.ExecContext(ctx, `
		CREATE UNIQUE INDEX IF NOT EXISTS "rm_key_index"
		ON "denote" ("rm_key")
	`)
	return
}

func Init(ctx context.Context) error {
	if err := Open(); err != nil {
		return err
	}
	return Create(ctx)
}

func Cleaner(
	ctx context.Context,
	interval time.Duration,
	stopChan chan struct{},
) {
	ticker := time.NewTicker(interval)
	defer ticker.Stop()
Loop:
	for {
		select {
		case <-stopChan:
			break Loop
		case <-ticker.C:
			Clean(ctx)
		}
	}
	close(stopChan)
}

func Clean(ctx context.Context) (err error) {
	_, err = db.ExecContext(ctx, `
		DELETE FROM "denote" WHERE datetime('now') >= "dt_limit"
	`)
	return
}

func Ping(ctx context.Context) error {
	return db.PingContext(ctx)
}

func Get(ctx context.Context, key uuid.UUID) (data []byte, err error) {
	tx, err := db.BeginTx(ctx, nil)
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
	var viewCount int
	var viewLimit int
	if err = tx.QueryRowContext(ctx, `
		SELECT "data", "view_count", "view_limit" 
		FROM "denote" 
		WHERE 
			"key" = ?
			AND
			datetime('now') < "dt_limit"
	`, key).Scan(
		&data,
		&viewCount,
		&viewLimit,
	); err != nil {
		return
	}
	viewCount++
	if viewCount < viewLimit {
		_, err = tx.ExecContext(ctx, `
			UPDATE "denote" SET "view_count" = ? WHERE "key" = ?
		`, viewCount, key)
	} else {
		_, err = tx.ExecContext(ctx, `
			DELETE FROM "denote" WHERE "key" = ?
		`, key)
	}
	return
}

func Set(
	ctx context.Context,
	data []byte,
	viewLimit int,
	durationLimit time.Duration,
) (uuid.UUID, uuid.UUID, error) {
	key, err := uuid.NewRandom()
	if err != nil {
		return uuid.Nil, uuid.Nil, err
	}
	rmKey, err := uuid.NewRandom()
	if err != nil {
		return uuid.Nil, uuid.Nil, err
	}
	if _, err = db.ExecContext(ctx,
		`INSERT INTO "denote" (
			"key", 
			"data", 
			"view_limit", 
			"dt_limit",
			"rm_key"
		) VALUES (?, ?, ?, ?, ?)`,
		key,
		data,
		viewLimit,
		time.Now().Add(durationLimit).UTC(),
		rmKey,
	); err != nil {
		return uuid.Nil, uuid.Nil, err
	}
	return key, rmKey, nil
}

func Delete(ctx context.Context, rmKey uuid.UUID) (err error) {
	_, err = db.ExecContext(ctx, `
		DELETE FROM "denote" WHERE "rm_key" = ?
	`, rmKey)
	return
}
