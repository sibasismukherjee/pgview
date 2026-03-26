package db

import (
	"strings"
	"testing"
)

func TestBuildDSN(t *testing.T) {
	tests := []struct {
		name         string
		url          string
		user         string
		pass         string
		dbname       string
		sslmode      string
		wantContains []string
	}{
		{
			name:    "full postgres URL — injects credentials and sslmode",
			url:     "postgres://host:5432/mydb",
			user:    "alice",
			pass:    "secret",
			dbname:  "",
			sslmode: "disable",
			wantContains: []string{
				"postgres://alice:secret@host:5432/mydb",
				"sslmode=disable",
			},
		},
		{
			name:    "full postgres URL — injects dbname when path is bare /",
			url:     "postgres://host:5432/",
			user:    "alice",
			pass:    "secret",
			dbname:  "mydb",
			sslmode: "prefer",
			wantContains: []string{
				"/mydb",
			},
		},
		{
			name:    "host:port — builds postgres DSN",
			url:     "localhost:5433",
			user:    "bob",
			pass:    "pw",
			dbname:  "testdb",
			sslmode: "prefer",
			wantContains: []string{
				"localhost:5433",
				"bob",
				"testdb",
				"sslmode=prefer",
			},
		},
		{
			name:    "bare host — defaults to port 5432",
			url:     "myhost",
			user:    "u",
			pass:    "p",
			dbname:  "db",
			sslmode: "require",
			wantContains: []string{
				"myhost:5432",
				"sslmode=require",
			},
		},
		{
			name:    "postgresql:// scheme treated same as postgres://",
			url:     "postgresql://host/db",
			user:    "u",
			pass:    "p",
			dbname:  "",
			sslmode: "disable",
			wantContains: []string{
				"u:p@host",
				"sslmode=disable",
			},
		},
		{
			name:    "special characters in password are URL-encoded",
			url:     "myhost",
			user:    "u",
			pass:    "p@ss!word",
			dbname:  "db",
			sslmode: "disable",
			wantContains: []string{
				"myhost:5432",
			},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			dsn := buildDSN(tc.url, tc.user, tc.pass, tc.dbname, tc.sslmode)
			for _, want := range tc.wantContains {
				if !strings.Contains(dsn, want) {
					t.Errorf("buildDSN(%q, %q, ...) = %q\n\twant it to contain %q",
						tc.url, tc.user, dsn, want)
				}
			}
		})
	}
}

func TestBuildDSN_HostPortSplit(t *testing.T) {
	// Ensure host:port is split correctly and not treated as the whole host.
	dsn := buildDSN("db-proxy.internal:6432", "user", "pass", "postgres", "require")
	if strings.Contains(dsn, "db-proxy.internal:6432:") {
		t.Errorf("port should not be doubled in DSN: %q", dsn)
	}
	if !strings.Contains(dsn, "db-proxy.internal") {
		t.Errorf("host missing from DSN: %q", dsn)
	}
	if !strings.Contains(dsn, "6432") {
		t.Errorf("custom port missing from DSN: %q", dsn)
	}
}

func TestQueryResult_IsZeroByDefault(t *testing.T) {
	var r QueryResult
	if r.Columns != nil {
		t.Error("expected nil Columns on zero value")
	}
	if r.Rows != nil {
		t.Error("expected nil Rows on zero value")
	}
	if r.Tag != "" {
		t.Error("expected empty Tag on zero value")
	}
}
