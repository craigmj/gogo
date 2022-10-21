gogo
====

Gogo handles database migrations in Go code like yoyo-migrate for python (well, a little like that).

It has been tested with MySQL, Postgresql and with SQLITE3 driver.

USAGE
=====
gogo considers a migration to be a `gogo.Migration` struct:
```go
type Migration struct {
	Apply    func(*Tx)
	Rollback func(*Tx)
}
```
In practice, you create an array like so:
```go
var myMigrations = []*gogo.Migration{
	{
		Apply: func(tx *gogo.Tx) {

		},
		Rollback: func(tx *gogo.Tx) {

		},
	},
}
```
To migrate the database to the latest version, gogo has:
```go
// Apply migrates the database to the latest migration version
func Migrate(db *sql.DB, migrations []Migration) (err error) {
```
In your code, you would use:
```go
	err := gogo.Migrate(db, myMigrations)
	if nil!=err {
		log.Fatal(err)
	}
```
gogo will always migrate to the latest version (which it tracks in the database using a table called, `_gogo_migrations`, although you can rename it with the function `gogo.MetaTable("gogo_version_table")` for instance).

Rolling back a database is equally easy:
```go
func Rollback(db *sql.DB, destinationVersionString string, migrations []Migration) (err error) {
```
In practice, this becomes:
```go
	rollbackTo := flag.StringVar(`rollback`, `-1`, `rollback the database version to the nth-to-last migration`)
	flag.Parse()
	if ``!=*rollbackTo {
		err := gogo.Rollback(db, *rollbackTo, myMigrations)
		if nil!=err {
			log.Fatal(err)
		}
	}
```
When providing a rollback version, you can use a positive number (e.g. 0 would roll you right back to the database before any migrations were applied), or a negative number (-1 rolls back only the last migration defined, -2 the last 2 defined migrations), and so on. *NOTE* that the migration is calculated from the migrations _defined_, not the migrations _applied_. So if you `-rollback -1` if will roll the database back to the next-to-last migration. If you `-rollback -1` a second time, nothing will happen, because the database is already at the next-to-last migration.

DEFINING MIGRATION
==================
The `gogo.Migration` methods `Apply(tx *gogo.Tx)` and `Rollback(tx *gogo.Tx)` come with an augemented `*sql.Tx`, that
adds some utility functions. See the godoc or `Tx.go` for these.

FAILING MIGRATION
=================
To fail a migration (or a rollback of one), just `panic`. gogo will catch the panic and report an error.

ERROR MESSAGES
==============
Errors that occur while gogo executes are reported with `gogo.PrintError`. This is defined in the file `Log.go` as:

```go
var PrintError = func(f string, args... interface{}) {
	fmt.Fprintf(os.Stderr, f, args...)
	log.Printf(f, args...)	
}
```
To use your own error logging, just set this to your own function:
```go
	gogo.PrintError = func(f string, args... interface{}) {
		// ... handle the error yourself
	}
```

DIFFERENT DATABASES
===================
gogo works with MySQL / Mariadb by default. However, it also supports Postgres and SQLite3.

To use a different database, you need to call `gogo.SetInterface(gogo.DbInterface)`. gogo has three defined DbInterfaces: `gogo.MysqlInterface`, `gogo.PgInterface` and `gogo.Sqlite3Interface`. So, to use gogo with PostgresQL, you need:

```go
gogo.SetInterface(&gogo.PgInterface)
```

To use gogo with Sqlite3:

```go
gogo.SetInterface(&gogo.Sqlite3Interface)
```

If you would like to use gogo with a different database, have a look at `DbInterface.go`, `MysqlInterface.go`, `PgInterface.go` and `Sqlite3Interface.go` for examples of providing the functionality gogo requires to support a different database engine.