package db

import (
	"database/sql"
	"fmt"
	"sync"

	_ "github.com/go-sql-driver/mysql"
	"github.com/golang/glog"
	_ "github.com/lib/pq"
	_ "github.com/mattn/go-sqlite3"
	"gopkg.in/gorp.v1"
)

type Table interface {
	// return table_name, primary_key, model
	Meta() (string, string, interface{})
}

type TableCreator interface {
	// sql to create table
	CreationSql() string
}

var _tables map[string]Table
var _tableFreezed = false
var _tableLk = &sync.Mutex{}

func RegisterTable(table Table) {
	name, _, _ := table.Meta()
	_tableLk.Lock()
	defer _tableLk.Unlock()

	if _tableFreezed {
		glog.Fatalf("table freezed")
	}
	if _, exists := _tables[name]; exists {
		glog.Fatalf("reregister table: `%s`", name)
	}
	_tables[name] = table
}

type Config struct {
	Host   string
	Port   int
	User   string
	Passwd string
	DB     string `default:"autoservice.db"`
	DBType string `default:"sqlite3"`

	MaxIdleConns int
	MaxOpenConns int

	SqlTrace bool
}

func (cfg *Config) DSN() string {
	switch cfg.DBType {
	case "mysql":
		return fmt.Sprintf("%s:%s@tcp(%s:%d)/%s",
			cfg.User, cfg.Passwd, cfg.Host, cfg.Port, cfg.DB)
	case "postgres":
		return fmt.Sprintf("postgres://%s:%s@tcp(%s:%d)/%s",
			cfg.User, cfg.Passwd, cfg.Host, cfg.Port, cfg.DB)
	case "sqlite3":
		return cfg.DB
	}

	return ""
}

type logger struct{}

func (l logger) Printf(format string, v ...interface{}) {
	glog.Infof(format, v...)
}

var db *gorp.DbMap

func InitDB(cfg *Config) (err error) {
	if cfg.DBType == "sqlite" {
		cfg.DBType = "sqlite3"
	}
	dsn := cfg.DSN()
	if dsn == "" {
		return fmt.Errorf("unsupported database: `%s`", cfg.DBType)
	}

	var raw_db *sql.DB
	if raw_db, err = sql.Open(cfg.DBType, dsn); err != nil {
		return
	}
	if cfg.MaxIdleConns > 0 {
		raw_db.SetMaxIdleConns(cfg.MaxIdleConns)
	}
	if cfg.MaxOpenConns > 0 {
		raw_db.SetMaxOpenConns(cfg.MaxOpenConns)
	}

	var dialect gorp.Dialect
	switch cfg.DBType {
	case "mysql":
		dialect = gorp.MySQLDialect{Engine: "InnoDB", Encoding: "UTF-8"}
	case "postgres":
		dialect = gorp.PostgresDialect{}
	case "sqlite3":
		dialect = gorp.SqliteDialect{}
	}

	db = &gorp.DbMap{Db: raw_db, Dialect: dialect}
	if cfg.SqlTrace {
		db.TraceOn("[db]", logger{})
	}
	_tableLk.Lock()
	_tableFreezed = true
	_tableLk.Unlock()

	for _, table := range _tables {
		name, primary, model := table.Meta()
		t := db.AddTableWithName(model, name)
		if primary != "" {
			t.SetKeys(true, primary)
		}
	}
	return nil
}

func CloseDB() error {
	if db == nil {
		return nil
	}
	return db.Db.Close()
}
