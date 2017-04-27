package gus

import (
	_ "github.com/mattn/go-sqlite3"
	"database/sql"
	"os"
	"testing"
	"runtime/debug"
	"time"
)

var orgsv *Orgs
var us *Users
var db *sql.DB

func TestMain(m *testing.M) {
	SetDebugOutput(os.Stdout)
	db = GetDb(DbOptions{Seed: true})
	orgsv = NewOrgs(db)
	us = NewUsers(db, UserOptions{MaxAuthAttempts: 5, AttemptLockDuration: time.Duration(1) * time.Second})
	code := m.Run()
	os.Exit(code)
}

func ErrIf(t *testing.T, err error) {
	if err != nil {
		t.Fatal(err, string(debug.Stack()))
	}
}
