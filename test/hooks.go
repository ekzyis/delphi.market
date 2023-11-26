package test

import (
	"fmt"
	"os"
	"path"
	"runtime"
	"testing"

	db_ "git.ekzyis.com/ekzyis/delphi.market/db"
)

var (
	dbName string = "delphi_test"
	dbUrl  string = fmt.Sprintf("postgres://delphi:delphi@localhost:5432/%s?sslmode=disable", dbName)
)

func Init(db **db_.DB) {
	// for ParseTemplates to work, cwd needs to be project root
	_, filename, _, _ := runtime.Caller(0)
	dir := path.Join(path.Dir(filename), "../")
	err := os.Chdir(dir)
	if err != nil {
		panic(err)
	}
	*db, err = db_.New(dbUrl)
	if err != nil {
		panic(err)
	}
}

func Main(m *testing.M, db *db_.DB) {
	if err := db.Reset(dbName); err != nil {
		panic(err)
	}
	retCode := m.Run()
	if err := db.Clear(dbName); err != nil {
		panic(err)
	}
	os.Exit(retCode)
}
