package gogo

import (
	"database/sql"
	"fmt"
	"strings"
)

// Tx is a very lightweight wrapper around the standard SQL Tx.
// It provides a few extension methods.
type Tx struct {
	*sql.Tx
}

// Exec executes the sql with the given params, returning a result and error, 
// and logging to PrintError if there is an error.
func (db *Tx) Exec(sql string, params ...interface{}) (sql.Result, error) {
	res, err := db.Tx.Exec(sql, params...)
	if nil!=err {
		err = fmt.Errorf("%w: %s [%v]", err, sql, params)
		PrintError("ERROR: %s", err.Error())
	}
	return res, err
}

// MustExec functions like the standard Tx Exec, except that, if it encounters an
// error, it panics. This panic is caught by the gogo framework, and will
// automatically rollback. So Execp is a shortcut for query execution with
// error catching
func (db *Tx) MustExec(sql string, params ...interface{}) {
	_, err := db.Tx.Exec(sql, params...)
	PrintError("MustExec: %s [%v]", sql, params)
	if nil != err {	
		err = fmt.Errorf("%w: %s [%v]", err, sql, params)
		PrintError("ERROR: %s", err.Error())
		panic(err)
	}
}

// MustExecEach executes each statement in the arguments, and panics
// if an error occurs on any.
func (db *Tx) MustExecEach(sqlStatements ...string) {
	if err := db.ExecEach(sqlStatements...); nil!=err {
		panic(err)
	}
}

// ExecEach executes each statement passed as an argument. Unlike ExecAll, it doesn't 
// split the sql statement on some delimiter, but treats each statement as unique.
func (db *Tx) ExecEach(sqlStatements ...string) error {
	for _, s := range sqlStatements {
		s = strings.TrimSpace(s)
		if ``==s {
			continue
		}
		if _, err := db.Tx.Exec(s); nil!=err {
			err := fmt.Errorf("%w: %s", err, s)
			PrintError("ERROR: %s", err.Error())
			return err
		}
	}
	return nil
}

// MustExecAll takes a string of SQL commands delimited with semi-colons.
// It executes each command, failing if any command fails.
// NB: See the commentary on ExecAll about having semi-colons inside your
// SQL.
func (db *Tx) MustExecAll(sql string) {
	err := db.ExecAll(sql)
	if nil != err {
		panic(err)
	}
}


// ExecAll takes a string of SQL commands delimited with semi-colons.
// It executes each command, returning an error if any command fails..
// Note that ExecAll does a very simple string split on semi-colons,
// so that if you have a semi-colon in any SQL statement (for eg in a string),
// ExecAll will not be able to split the commands properly, and will fail.
// Easily corrected by using a separate Exec for that particular statement,
// but beware of this.
func (db *Tx) ExecAll(sql string) error {
	// Should actually try to parse this properly, but for now we'll leave
	// like this.
	return db.ExecEach(strings.Split(sql, ";")...)
}

// MustQuery executes the given query against the db, but panics if it encounters an
// error. Like sql.Query, it returns a *sql.Rows.
func (db *Tx) MustQuery(sql string, params ...interface{}) *sql.Rows {
	rows, err := db.Tx.Query(sql, params...)
	if nil != err {
		err = fmt.Errorf("%w: %s [%v]", err, sql, params)
		PrintError("ERROR: %s", err.Error())
		panic(err)
	}
	return rows
}

// MustPrepare prepares the given query against the db, but panics if it encounters an error.
// Like sql.Prepare, it returns a *sql.Stmt.
func (db *Tx) MustPrepare(sql string) *sql.Stmt {
	qry, err := db.Tx.Prepare(sql)
	if nil != err {
		err = fmt.Errorf("%w: %s", err, sql)
		PrintError("ERROR: %s", err.Error())
		panic(err)
	}
	return qry
}
