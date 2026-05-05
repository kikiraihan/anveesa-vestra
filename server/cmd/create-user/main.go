package main

import (
	"database/sql"
	"flag"
	"fmt"
	"os"
	"strings"
	"time"

	"golang.org/x/crypto/bcrypt"
	_ "modernc.org/sqlite"
)

func main() {
	username := flag.String("username", "", "username (min 3 chars)")
	password := flag.String("password", "", "password (min 8 chars)")
	role := flag.String("role", "admin", "role: admin, editor, viewer")
	dbPath := flag.String("db", "server/data.db", "path to data.db")
	flag.Parse()

	*username = strings.TrimSpace(*username)

	if *username == "" || *password == "" {
		fmt.Fprintln(os.Stderr, "usage: create-user -username <name> -password <pass> [-role admin|editor|viewer] [-db path/to/data.db]")
		os.Exit(1)
	}
	if len(*username) < 3 {
		fmt.Fprintln(os.Stderr, "error: username must be at least 3 characters")
		os.Exit(1)
	}
	if len(*password) < 8 {
		fmt.Fprintln(os.Stderr, "error: password must be at least 8 characters")
		os.Exit(1)
	}
	validRoles := map[string]bool{"admin": true, "editor": true, "viewer": true}
	if !validRoles[*role] {
		fmt.Fprintln(os.Stderr, "error: role must be one of: admin, editor, viewer")
		os.Exit(1)
	}

	hash, err := bcrypt.GenerateFromPassword([]byte(*password), bcrypt.DefaultCost)
	if err != nil {
		fmt.Fprintf(os.Stderr, "error hashing password: %v\n", err)
		os.Exit(1)
	}

	db, err := sql.Open("sqlite", "file:"+*dbPath+"?_foreign_keys=1&_journal_mode=WAL&_busy_timeout=5000")
	if err != nil {
		fmt.Fprintf(os.Stderr, "error opening database: %v\n", err)
		os.Exit(1)
	}
	defer db.Close()

	result, err := db.Exec(
		"INSERT INTO users (username, password_hash, role, created_at) VALUES (?, ?, ?, ?)",
		*username, string(hash), *role, time.Now().UTC(),
	)
	if err != nil {
		fmt.Fprintf(os.Stderr, "error creating user (username may already exist): %v\n", err)
		os.Exit(1)
	}

	id, _ := result.LastInsertId()
	fmt.Printf("user created: id=%d username=%s role=%s\n", id, *username, *role)
}
