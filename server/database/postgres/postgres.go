package postgres

import (
	"context"
	"runtime"
	"strconv"
	"strings"

	"server/database"

	"github.com/jackc/pgx/v4"
)

/*
	TODO: Add timeout to the context
*/

type pg struct {
	conn     *pgx.Conn
	settings *pgx.ConnConfig
}

type pgTx struct {
	tx pgx.Tx
}

// New Initializes a pg instance with the given settings
func New(connStr string) (database.Database, error) {
	var (
		p   = &pg{}
		err error
	)

	if p.settings, err = pgx.ParseConfig(connStr); err != nil {
		return nil, err
	}

	return p, nil
}

// Open tries to connect to the database
func (db *pg) Open() (err error) {
	if db.conn, err = pgx.ConnectConfig(
		context.Background(),
		db.settings,
	); err != nil {
		return
	}

	return
}

// Close attempts to close a database connection
func (db *pg) Close() (err error) {
	return db.conn.Close(context.Background())
}

// NewTx returns a transaction
func (db *pg) NewTx() (tx database.Transaction, err error) {
	var t = &pgTx{}

	if t.tx, err = db.conn.Begin(context.Background()); err != nil {
		return
	}

	return t, nil
}

// Query implements the Transaction interface
func (tx *pgTx) Query(query string, args ...interface{}) (database.Rows, error) {
	return tx.tx.Query(context.Background(), getCaller()+query, args...)
}

// QueryRow implements the Transaction interface
func (tx *pgTx) QueryRow(query string, args ...interface{}) database.Row {
	return tx.tx.QueryRow(context.Background(), getCaller()+query, args...)
}

// Exec implements the Transaction interface
func (tx *pgTx) Exec(query string, args ...interface{}) (err error) {
	_, err = tx.tx.Exec(context.Background(), getCaller()+query, args...)

	return
}

// Commit implements the Transaction interface
func (tx *pgTx) Commit() error {
	return tx.tx.Commit(context.Background())
}

// Rollback implements the Transaction interface
func (tx *pgTx) Rollback() {
	_ = tx.tx.Rollback(context.Background())
}

// getCaller get who in the server code called the database making easier to debug
func getCaller() (out string) {
	var (
		file string
		line int
	)

	out = "-- \r\n"

	for i := 2; i < 8; i++ {
		_, file, line, _ = runtime.Caller(i)
		if strings.Contains(file, "/server/") {
			out += "-- " + strconv.FormatInt(int64(i), 10) + " :: " + file + ":" + strconv.FormatInt(int64(line), 10) + "\r\n"
		}
	}

	return
}
