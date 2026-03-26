package main

import (
	"bufio"
	"flag"
	"fmt"
	"os"
	"strings"
	"syscall"

	"github.com/sibasismukherjee/pgview/internal/db"
	"github.com/sibasismukherjee/pgview/internal/ui"
	"golang.org/x/term"
)

var version = "dev"

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
	if *password == "" {
		fmt.Print("Password: ")
		pw, err := term.ReadPassword(int(syscall.Stdin))
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

	ui.Run(client)
}
