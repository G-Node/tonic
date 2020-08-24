package db

import (
	_ "github.com/mattn/go-sqlite3"
	"xorm.io/xorm"
	"xorm.io/xorm/log"
	"xorm.io/xorm/names"
)

type Connection struct {
	engine *xorm.Engine
}

// Close the database.
func (conn *Connection) Close() error {
	return conn.engine.Close()
}

// New returns a database connection for the sqlite db file at the given path.
// If it does not exist it is created.
func New(path string) (*Connection, error) {
	db, err := xorm.NewEngine("sqlite3", path)
	if err != nil {
		return nil, err
	}
	db.Logger().SetLevel(log.LOG_DEBUG)
	db.SetMapper(names.GonicMapper{})

	if err := db.Sync2(new(Job), new(Session)); err != nil {
		return nil, err
	}
	return &Connection{db}, nil
}
