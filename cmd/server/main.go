package main

import (
	"context"
	"database/sql"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	_ "github.com/jackc/pgx/v5/stdlib"

	"github.com/giorgiovilardo/pharmarecall/db/migrations"
	"github.com/giorgiovilardo/pharmarecall/internal/auth"
	"github.com/giorgiovilardo/pharmarecall/internal/config"
	"github.com/giorgiovilardo/pharmarecall/internal/db"
	"github.com/giorgiovilardo/pharmarecall/internal/pharmacy"
	"github.com/giorgiovilardo/pharmarecall/internal/user"
	"github.com/giorgiovilardo/pharmarecall/internal/web"
	"github.com/giorgiovilardo/pharmarecall/internal/web/handler"
	"github.com/jackc/pgx/v5/pgxpool"
)

func main() {
	if err := run(); err != nil {
		slog.Error("fatal", "error", err)
		os.Exit(1)
	}
}

func run() error {
	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	cfg, err := config.Load("config.toml")
	if err != nil {
		return fmt.Errorf("loading config: %w", err)
	}

	// database/sql connection for goose migrations
	sqlDB, err := sql.Open("pgx", cfg.DB.URL)
	if err != nil {
		return fmt.Errorf("opening sql db: %w", err)
	}
	defer sqlDB.Close()

	if err := migrations.Run(sqlDB); err != nil {
		return fmt.Errorf("running migrations: %w", err)
	}

	// pgxpool for application use
	pool, err := pgxpool.New(ctx, cfg.DB.URL)
	if err != nil {
		return fmt.Errorf("creating connection pool: %w", err)
	}
	defer pool.Close()

	queries := db.New(pool)
	sm := auth.NewSessionManager(pool)

	// Domain services
	userRepo := user.NewPgxRepository(queries)
	userSvc := user.NewService(userRepo, auth.HashPassword, auth.VerifyPassword)

	pharmacyRepo := pharmacy.NewPgxRepository(pool, queries)
	pharmacySvc := pharmacy.NewService(pharmacyRepo, auth.HashPassword)

	// Build handlers
	mux := web.NewRouter(web.Handlers{
		LoginPage:      handler.HandleLoginPage(),
		LoginPost:      handler.HandleLoginPost(sm, userSvc),
		Logout:         handler.HandleLogout(sm),
		ChangePassPage: handler.HandleChangePasswordPage(),
		ChangePassPost: handler.HandleChangePasswordPost(sm, userSvc),
		Owner: web.OwnerHandlers{
			PersonnelList:   handler.HandleOwnerPersonnelList(pharmacySvc),
			AddPersonnel:    handler.HandleOwnerAddPersonnelPage(),
			CreatePersonnel: handler.HandleOwnerCreatePersonnel(pharmacySvc),
		},
		Admin: web.AdminHandlers{
			Dashboard:       handler.HandleAdminDashboard(pharmacySvc),
			NewPharmacy:     handler.HandleNewPharmacyPage(),
			CreatePharmacy:  handler.HandleCreatePharmacy(pharmacySvc),
			PharmacyDetail:  handler.HandlePharmacyDetail(pharmacySvc, pharmacySvc),
			UpdatePharmacy:  handler.HandleUpdatePharmacy(pharmacySvc, pharmacySvc, pharmacySvc),
			AddPersonnel:    handler.HandleAddPersonnelPage(),
			CreatePersonnel: handler.HandleCreatePersonnel(pharmacySvc),
		},
	})

	// Compose middleware: CORS → sessions → load user → router
	cop := http.NewCrossOriginProtection()
	h := cop.Handler(sm.LoadAndSave(web.LoadUser(sm)(mux)))

	server := &http.Server{
		Addr:    fmt.Sprintf(":%d", cfg.Server.Port),
		Handler: h,
	}

	go func() {
		slog.Info("server starting", "addr", server.Addr)
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			slog.Error("server error", "error", err)
			os.Exit(1)
		}
	}()

	<-ctx.Done()
	slog.Info("shutting down")

	shutdownCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if err := server.Shutdown(shutdownCtx); err != nil {
		return fmt.Errorf("shutting down server: %w", err)
	}
	slog.Info("server stopped")
	return nil
}
