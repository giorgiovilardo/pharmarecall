package main

import (
	"context"
	"flag"
	"fmt"
	"log/slog"
	"os"

	_ "github.com/jackc/pgx/v5/stdlib"

	"github.com/giorgiovilardo/pharmarecall/internal/auth"
	"github.com/giorgiovilardo/pharmarecall/internal/config"
	"github.com/giorgiovilardo/pharmarecall/internal/db"
	"github.com/jackc/pgx/v5/pgxpool"
)

func main() {
	email := flag.String("email", "", "admin email (required)")
	password := flag.String("password", "", "admin password (required)")
	configPath := flag.String("config", "config.toml", "path to config file")
	flag.Parse()

	if *email == "" || *password == "" {
		fmt.Fprintln(os.Stderr, "usage: seed --email <email> --password <password>")
		os.Exit(1)
	}

	if err := run(*configPath, *email, *password); err != nil {
		slog.Error("seed failed", "error", err)
		os.Exit(1)
	}
}

func run(configPath, email, password string) error {
	ctx := context.Background()

	cfg, err := config.Load(configPath)
	if err != nil {
		return fmt.Errorf("loading config: %w", err)
	}

	pool, err := pgxpool.New(ctx, cfg.DB.URL)
	if err != nil {
		return fmt.Errorf("creating connection pool: %w", err)
	}
	defer pool.Close()

	queries := db.New(pool)

	user, err := auth.SeedAdmin(ctx, queries, email, password)
	if err != nil {
		return fmt.Errorf("seeding admin: %w", err)
	}

	slog.Info("admin user created", "id", user.ID, "email", user.Email)
	return nil
}
