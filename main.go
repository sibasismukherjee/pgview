package main

import (
	"bufio"
	"flag"
	"fmt"
	"os"
	"strings"
	"syscall"

	"golang.org/x/term"

	"github.com/sibasismukherjee/pgview/internal/db"
	"github.com/sibasismukherjee/pgview/internal/tui"
)

var version = "0.4.1"

func main() {
	proxyURL := flag.String("url", "", "PostgreSQL proxy/connection URL (host:port or postgres://host:port/dbname)")
	username := flag.String("username", "", "Database username")
	password := flag.String("password", "", "Database password (prompted if omitted)")
	dbname := flag.String("dbname", "postgres", "Database name")
	sslmode := flag.String("sslmode", "prefer", "SSL mode (disable|allow|prefer|require)")
	showVersion := flag.Bool("version", false, "Print version and exit")
	flag.Parse()

	if *showVersion {
		fmt.Printf("pgview %s\n", version)
		os.Exit(0)
	}

	r := bufio.NewReader(os.Stdin)

	if *proxyURL == "" {
		fmt.Print("Connection URL (host:port or postgres://...): ")
		line, _ := r.ReadString('\n')
		*proxyURL = strings.TrimSpace(line)
	}
	if *username == "" {
		fmt.Print("Username: ")
		line, _ := r.ReadString('\n')
		*username = strings.TrimSpace(line)
	}
	// Prompt for dbname only when the URL is not a full DSN (no embedded database).
	if *dbname == "postgres" && !strings.HasPrefix(*proxyURL, "postgres://") && !strings.HasPrefix(*proxyURL, "postgresql://") {
		fmt.Print("Database [postgres]: ")
		line, _ := r.ReadString('\n')
		if name := strings.TrimSpace(line); name != "" {
			*dbname = name
		}
	}
	if *password == "" {
		fmt.Print("Password: ")
		pw, err := term.ReadPassword(syscall.Stdin)
		if err != nil {
			line, _ := r.ReadString('\n')
			*password = strings.TrimSpace(line)
		} else {
			*password = string(pw)
		}
		fmt.Println()
	}

	client, err := db.Connect(*proxyURL, *username, *password, *dbname, *sslmode)
	if err != nil {
		fmt.Fprintf(os.Stderr, "error: connection failed: %v\n", err)
		os.Exit(1)
	}
	defer client.Close()

	tui.Run(client, version)
}
