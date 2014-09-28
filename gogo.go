package gogo

import (
	"fmt"
	"database/sql"
	"strconv"
	"reflect"
)

var db *sql.DB
var gogoTable = "_gogo_migrations"
	
// A Migration is a one-way advancement of the database.
type Migration struct  {
	Apply func(*Tx)
	Rollback func(*Tx)
}

// MetaTable lets you set the name of the table where gogomigrate should
// store its information about the current db version. The default is _gogo_migrations
func MetaTable(table string) {
	gogoTable = table
}

// Version returns the current version of the database.
// Your migrations are numbered from 1, so a version of 0 means that no 
// migrations have been applied.
func Version(db *sql.DB) (int, error) {
	var version int

	var tblname string
	res := db.QueryRow(`show tables like '` + gogoTable + `'`)
	err := res.Scan(&tblname)
	if nil!=err && sql.ErrNoRows!=err{
		return -1, fmt.Errorf("Executing `show tables like '%v'`: %v", gogoTable, err.Error())
	}
	if sql.ErrNoRows == err {
		_, err = db.Exec(fmt.Sprintf(`
			create table %s
			(
				version int not null default 0,
				migration_date datetime not null, 
				primary key(version)
			)`, gogoTable))
		if nil!=err {
			return -1, fmt.Errorf("Creating gogomigration versioning table %v: %v", gogoTable, err.Error())
		}
		_, err = db.Exec(fmt.Sprintf(`insert into %s(version, migration_date) values(0, now())`, gogoTable))
		if nil!=err {
			return -1, fmt.Errorf("Inserting 0 version into %v: %v", gogoTable, err.Error())
		}
		version = 0
	} else {		
		res := db.QueryRow(`select version from ` + gogoTable)
		err = res.Scan(&version)
		if nil!=err {
			return -1, fmt.Errorf("Scanning version from %v: %v", gogoTable, err.Error())
		}
	}
	return version, nil
}

// Apply migrates the database to the latest migration version
func Migrate(db *sql.DB, migrations []Migration) (err error) {
	_, err = db.Exec(`set autocommit=0`)
	if nil!=err {
		return err
	}
	defer func() {
		_, e := db.Exec(`set autocommit=1`)
		if nil==err {
			err = e
		}
	}()
	
	version, err := Version(db)
	if nil!=err {
		return err
	}
	
	// safeExec applies the migration, and returns an error if the migration returns
	// an error, or if the migration panics
	safeExec := func(tx *sql.Tx, migration Migration) (err error) {
		defer func() {
			if e := recover(); e!=nil {
				ok := true
				err, ok = e.(error)
				if !ok {
					err = fmt.Errorf("Unexpected panic value of type %v: %v", reflect.TypeOf(e).Name(), e)
				}
			}
		}()		
		migration.Apply(&Tx{tx})
		return nil
	}
	
	for version < len(migrations) {
		Tx, err := db.Begin()
		if nil!=err {
			return fmt.Errorf("Starting transaction: %v", err.Error())
		}
		if err := safeExec(Tx, migrations[version]); nil!=err {
			if err := Tx.Rollback(); nil!=err {
				fmt.Printf("ERROR DURING ROLLBACK: %v\n", err.Error())
			}			
			return fmt.Errorf("Migrating to version %v: %v", version+1, err.Error()) 
		}
		version++
		_, err = Tx.Exec(`update ` + gogoTable + ` set version = ?, migration_date=now()`, version)
		if nil!=err {
			Tx.Rollback()
			return fmt.Errorf("Updating version to %v: %v", version+1, err.Error())
		}
		err = Tx.Commit()
		if nil!=err {
			Tx.Rollback()
			return fmt.Errorf("Committing transaction for migration %v: %v", version+1, err.Error())
		}
	}
	return nil
}

// Rollback rolls the database back to the destination version. Versions are 
// numbered from 1, so a Rollback to 0 would rollback all migrations.
func Rollback(db *sql.DB, destinationVersionString string, migrations []Migration) (err error) {
	if ""==destinationVersionString {
		return nil
	}
	destinationVersion, err := strconv.Atoi(destinationVersionString)
	if nil!=err {
		return err
	}

	_, err = db.Exec(`set autocommit=0`)
	if nil!=err {
		return err
	}
	defer func() {
		_, e := db.Exec(`set autocommit=1`)
		if nil==err {
			err = e
		}
	}()

	version, err := Version(db)
	if nil!=err {
		return err
	}
	
	// Negative destinationVersions mean we rollback by a delta
	if 0>destinationVersion {
		destinationVersion = version + destinationVersion
		if 0>destinationVersion {
			return fmt.Errorf("Cannot rollback to negative version %d from current version %d", destinationVersion, version)
		}
	}
	
	fmt.Println("Rollback to version ", destinationVersion)
	// safeRollback rollsback the migration, and returns an error if the migration rollback returns
	// an error, or if the migration rollback panics
	safeRollback := func(tx *sql.Tx, migration Migration) (err error) {
		defer func() {
			if e := recover(); e!=nil {
				ok := true
				err, ok = e.(error)
				if !ok {
					err = fmt.Errorf("Unexpected panic value of type %v: %v", reflect.TypeOf(e).Name(), e)
				}
			}
		}()		
		migration.Rollback(&Tx{tx})
		return nil
	}

	if 1>version {
		return fmt.Errorf("Cannot rollback: currently at version %v", version)
	}
	
	for version > destinationVersion {
		fmt.Printf("Rolling back from version %v\n", version)
		version--
		Tx, err := db.Begin()
		if nil!=err {
			return fmt.Errorf("Starting transaction: %v", err.Error())
		}
		if err := safeRollback(Tx, migrations[version]); nil!=err {
			Tx.Rollback()
			return fmt.Errorf("Rollingback version %v: %v", version+1, err.Error()) 
		}
		_, err = Tx.Exec(`update ` + gogoTable + ` set version = ?, migration_date=now()`, version)
		if nil!=err {
			Tx.Rollback()
			return fmt.Errorf("Updating version to %v: %v", version+1, err.Error())
		}
		err = Tx.Commit()
		if nil!=err {
			Tx.Rollback()
			return fmt.Errorf("Committing transaction for rollback %v: %v", version+1, err.Error())
		}
	}
	return nil
}
