package database

// Rows defines a interface for returned rows
type Rows interface {
	Next() bool
	Close()
	Err() error
	Scan(...interface{}) error
}

// Row defines a interface for returned row
type Row interface {
	Scan(...interface{}) error
}

// Database defines a interface for a database connection
type Database interface {
	Open() error
	Close() error
	NewTx() (Transaction, error)
}

// Transaction defined a interface for a database transaction
type Transaction interface {
	Query(string, ...interface{}) (Rows, error)
	QueryRow(string, ...interface{}) Row
	Exec(string, ...interface{}) error
	Commit() error
	Rollback()
}

var (
	readWriteDB Database
	readOnlyDB  Database
)

// RegisterDatabase defines a database to be used
func RegisterDatabase(db Database, ro bool) (err error) {
	if err = db.Open(); err != nil {
		return
	}

	if ro {
		readOnlyDB = db
	} else {
		readWriteDB = db
		if readOnlyDB == nil {
			readOnlyDB = db
		}
	}

	return
}

// Close closes all attached databases
func Close() (err error) {
	if readOnlyDB != nil {
		if err = readOnlyDB.Close(); err != nil {
			return
		}
	}

	if readWriteDB != nil {
		if err = readWriteDB.Close(); err != nil {
			return
		}
	}

	return
}

// NewTx creates a new database transaction
func NewTx(ro bool) (Transaction, error) {
	if ro {
		return readOnlyDB.NewTx()
	}

	return readWriteDB.NewTx()
}
