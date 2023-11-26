package handler_test

import (
	"testing"

	db_ "git.ekzyis.com/ekzyis/delphi.market/db"
	"git.ekzyis.com/ekzyis/delphi.market/test"
)

var (
	db *db_.DB
)

func TestMain(m *testing.M) {
	test.Main(m, db)
}
