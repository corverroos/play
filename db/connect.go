package db

import (
	"database/sql"
	"flag"
	"fmt"
	"runtime"
	"strings"
	"testing"

	"github.com/corverroos/play"
	"github.com/corverroos/unsure"
	"github.com/luno/jettison/log"
)

var (
	dbURI = flag.String("play_db", "mysql://root@unix("+unsure.SockFile()+")/%s?",
		"play DB URI (%s will be replaced with player_name flag)")
)

type PlayDB struct {
	DB        *sql.DB
	ReplicaDB *sql.DB
}

// ReplicaOrMaster returns the replica DB if available, otherwise the master.
func (db *PlayDB) ReplicaOrMaster() *sql.DB {
	if db.ReplicaDB != nil {
		return db.ReplicaDB
	}
	return db.DB
}

func Connect() (*PlayDB, error) {
	uri := *dbURI
	if strings.Contains(uri, "%s") {
		uri = fmt.Sprintf(uri, play.Name())
	}

	ok, err := unsure.MaybeRecreateSchema(uri, getSchemaPath())
	if err != nil {
		return nil, err
	} else if ok {
		log.Info(nil, "recreated schema")
	}

	dbc, err := unsure.Connect(uri)
	if err != nil {
		return nil, err
	}
	return &PlayDB{
		DB:        dbc,
		ReplicaDB: dbc,
	}, nil
}

func ConnectForTesting(t *testing.T) *sql.DB {
	return unsure.ConnectForTesting(t, getSchemaPath())
}

func getSchemaPath() string {
	_, filename, _, _ := runtime.Caller(0)
	return strings.Replace(filename, "connect.go", "schema.sql", 1)
}
