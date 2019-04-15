package data

import (
	"os"
	"path"

	"github.com/dgraph-io/badger"
	log "github.com/sirupsen/logrus"
)

// DefaultOptions returns default options for the database, customized based on environment variables.
func DefaultOptions() badger.Options {
	opts := badger.DefaultOptions
	dbPath, ok := os.LookupEnv("DATABASE_DIR")
	if !ok {
		dbPath = path.Join(os.TempDir(), "vogon")
	}
	opts.Dir = dbPath
	opts.ValueDir = dbPath
	// Optimize options for low memory and disk usage
	opts.MaxTableSize = 1 << 24
	return opts
}

// DBService provides services for reading and writing structs in the database.
type DBService struct {
	db *badger.DB
}

// Open opens the database with options and returns a DBService instance.
func Open(options badger.Options) (*DBService, error) {
	log.WithField("dir", options.Dir).Info("Opening database")
	db, err := badger.Open(options)
	if err != nil {
		return nil, err
	}
	return &DBService{db: db}, nil
}

// Close closes the underlying database.
func (service *DBService) Close() {
	log.Info("Closing database")
	if service != nil && service.db != nil {
		err := service.db.Close()
		if err != nil {
			log.Fatal(err)
		}
		service.db = nil
	}
}
