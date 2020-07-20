package gogo

import (
	"database/sql"
)

type Sqlite3Interface struct {
	Schema string
}

func NewSqlite3Interface(schema string) *Sqlite3Interface {
	if "" == schema {
		schema = "gogo_migrate"
	}
	return &Sqlite3Interface{Schema: schema}
}

func (pg *Sqlite3Interface) schemaName(tbl string) string {
	if "" == pg.Schema {
		return tbl
	}
	return pg.Schema + "." + tbl
}

func (pg *Sqlite3Interface) TableExists(db *sql.DB, table string) (bool, error) {
	var tblname string
	schema := pg.Schema
	if "" == schema {
		schema = "current_schema"
	} else {
		schema = "'" + pg.Schema + "'"
	}
	res := db.QueryRow(`
	SELECT name
    FROM   sqlite_master 
    WHERE  name = ? AND type = ?
    `, table, `table`)
	err := res.Scan(&tblname)

	if nil == err {
		return true, nil
	}
	if sql.ErrNoRows == err {
		return false, nil
	}
	return false, err
}

func (pg *Sqlite3Interface) InsertVersionSql(gogoTable string) string {
	return `insert into ` + pg.schemaName(gogoTable) + `(version, migration_date) values($1, current_timestamp)`
}
func (pg *Sqlite3Interface) UpdateVersionSql(gogoTable string) string {
	return `update ` + pg.schemaName(gogoTable) + ` set version = $1, migration_date=current_timestamp`
}
func (pg *Sqlite3Interface) SelectVersionSql(gogoTable string) string {
	return `select version from ` + pg.schemaName(gogoTable)
}

func (pg *Sqlite3Interface) CreateGogoTableSql(gogoTable string) string {
	sql := `
			create table ` + pg.schemaName(gogoTable) + `
			(
				version int not null default 0,
				migration_date timestamp not null, 
				primary key(version)
			)`
	return sql
}

func (pg *Sqlite3Interface) PreMigrate(db *sql.DB) error {
	var err error
	if "" != pg.Schema {
		_, err = db.Exec(`create schema if not exists ` + pg.Schema)
	}
	return err
}

func (*Sqlite3Interface) PostMigrate(db *sql.DB) error {
	return nil
}
