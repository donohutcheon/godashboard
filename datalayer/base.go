package datalayer

import (
	"context"
	"database/sql"
	"fmt"
	"github.com/DATA-DOG/go-txdb"
	"io/ioutil"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"

	"github.com/donohutcheon/gowebserver/lib/nonce"
	_ "github.com/go-sql-driver/mysql"
	"github.com/jmoiron/sqlx"
	"github.com/stretchr/testify/require"
	"github.com/xo/dburl"
)

type Model struct {
	ID        int64        `json:"id" db:"id"`
	CreatedAt JsonNullTime `json:"createdAt" db:"created_at"`
	UpdatedAt JsonNullTime `json:"updatedAt" db:"updated_at"`
	DeletedAt JsonNullTime `json:"deletedAt" db:"deleted_at"`
}

type PersistenceDataLayer struct {
	conn *sqlx.DB
}

var (
	ErrNoData = sql.ErrNoRows
)

func NewForTesting(t *testing.T, ctx context.Context) (*PersistenceDataLayer, error)  {
	t.Helper()
	username := os.Getenv("db_user")
	password := os.Getenv("db_pass")
	dbName := os.Getenv("db_test_name")
	dbHost := os.Getenv("db_host")
	dbPort := os.Getenv("db_port")
	dbPermanent := os.Getenv("db_permanent") == "true"
	if len(username) == 0 {
		username = "root"
	}
	if len(password) == 0 {
		password = "root"
	}
	if len(dbHost) == 0 {
		dbHost = "127.0.0.1"
	}
	if len(dbPort) == 0 {
		dbPort = "3306"
	}
	if len(dbName) == 0 {
		dbName = "test_" + nonce.GenerateNonce(10)
		t.Log("ephemeral database name ", dbName)
	}

	if !dbPermanent {
		t.Cleanup(func() {
			cleanupNonPermanentDatabase(t, ctx, username, password, dbHost, dbPort, dbName)
		})
	}

	maybeCreateDatabaseForTesting(t, ctx, username, password, dbHost, dbPort, dbName)

	dbURI := fmt.Sprintf("%s:%s@tcp(%s:%s)/%s?parseTime=true", username, password, dbHost, dbPort, dbName)

	txdb.Register(dbName, "mysql", dbURI)
	/*if !isTestDriverRegistered {
		txdb.Register("txdb", "mysql", dbURI)
		isTestDriverRegistered = true
	}*/

	conn, err := sqlx.Open(dbName, dbURI)
	require.NoError(t, err)
	t.Cleanup(func(){
		conn.Close()
	})

	return &PersistenceDataLayer{
		conn: conn,
	}, nil
}

func New() (*PersistenceDataLayer, error){
	conn, err, ok := tryConnectHerokuJawsDB()
	if err != nil {
		// TODO: Use proper logger
		fmt.Printf("Could not connect to JawsDB. %s", err.Error())
		return nil, err
	} else if ok {
		return &PersistenceDataLayer{
			conn: conn,
		}, nil
	}

	username := os.Getenv("db_user")
	password := os.Getenv("db_pass")
	dbName := os.Getenv("db_name")
	dbHost := os.Getenv("db_host")
	dbPort := os.Getenv("db_port")

	conn, err = createCon("mysql", username, password, dbHost, dbPort, dbName)
	if err != nil {
		fmt.Print(err)
		return nil, err
	}
	return &PersistenceDataLayer{
		conn: conn,
	}, nil
}

func (p *PersistenceDataLayer) GetConn() *sqlx.DB {
	return p.conn
}

/*Create mysql connection*/
func createCon(driverName string, username string, password string, dbHost string, dbPort string, dbName string) (db *sqlx.DB, err error) {
	dbURI := fmt.Sprintf("%s:%s@tcp(%s:%s)/%s?parseTime=true", username, password, dbHost, dbPort, dbName)
	db, err = sqlx.Open(driverName, dbURI)
	if err != nil {
		fmt.Println(err.Error())
	} else {
		fmt.Println("database is connected")
	}

	// make sure connection is available
	err = db.Ping()
	if err != nil {
		fmt.Printf("MySQL datalayer is not connected %s", err.Error())
	}
	return db, err
}

func tryConnectHerokuJawsDB() (*sqlx.DB, error, bool){
	dbURI := os.Getenv("JAWSDB_MARIA_URL") + "?parseTime=true"
	if len(dbURI) == 0 {
		return nil, nil, false
	}

	db, err := dburl.Open( dbURI)
	if err != nil {
		return nil, err, false
	} else {
		fmt.Println("database is connected")
	}

	dbx := sqlx.NewDb(db, "mysql")

	err = dbx.Ping()
	if err != nil {
		return nil, err, false
	}

	return dbx, nil, true
}

func maybeCreateDatabaseForTesting(t *testing.T, ctx context.Context, username, password, dbHost, dbPort, dbName string) {
	t.Helper()

	_, path, _, _ := runtime.Caller(0)
	schemaFile := fmt.Sprintf("%s/../schema.mysql.sql", filepath.Dir(path))

	dbURI := fmt.Sprintf("%s:%s@tcp(%s:%s)/?parseTime=true", username, password, dbHost, dbPort)
	db, err := sql.Open("mysql", dbURI)
	require.NoError(t, err)
	defer db.Close()

	row := db.QueryRowContext(ctx, "SELECT SCHEMA_NAME FROM INFORMATION_SCHEMA.SCHEMATA WHERE SCHEMA_NAME = ?",dbName)
	var exists string
	err = row.Scan(&exists)
	if err != sql.ErrNoRows && err != nil {
		require.NoError(t, err)
	} else if err == nil {
		return
	}

	_,err = db.Exec("CREATE DATABASE IF NOT EXISTS " + dbName)
	require.NoError(t, err)

	_,err = db.Exec("USE " + dbName)
	require.NoError(t, err)

	b, err := ioutil.ReadFile(schemaFile)
	require.NoError(t, err)
	splitted := strings.Split(string(b),";")
	for _, stmt := range splitted {
		if (len(stmt) > 2 && strings.TrimSpace(stmt)[0:2] == "--") ||  len(stmt) < 2 {
			continue
		}
		_,err = db.Exec(stmt)
		require.NoError(t, err)
	}
}

func cleanupNonPermanentDatabase(t *testing.T, ctx context.Context, username, password, dbHost, dbPort, dbName string) {
	t.Helper()

	dbURI := fmt.Sprintf("%s:%s@tcp(%s:%s)/?parseTime=true", username, password, dbHost, dbPort)
	db, err := sql.Open("mysql", dbURI)
	require.NoError(t, err)
	defer db.Close()

	_, err = db.Exec("DROP DATABASE IF EXISTS " + dbName)
	require.NoError(t, err)
}