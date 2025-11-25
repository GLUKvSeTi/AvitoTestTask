package main

import (
	"AvitoTestTask/internal/adapters/api"
	"AvitoTestTask/internal/adapters/postgres"
	"AvitoTestTask/internal/infra"
	pruc "AvitoTestTask/internal/usecases/pullrequest"
	teamuc "AvitoTestTask/internal/usecases/team"
	useruc "AvitoTestTask/internal/usecases/user"
	"context"
	"flag"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"
)

func main() {

	addr := flag.String("addr", ":8080", "listen addr")
	dsn := flag.String("dsn", getEnv("DATABASE_URL", "postgres://postgres:postgres@localhost:5432/AvitoTestTask?sslmode=disable"), "postgres dsn")
	flag.Parse()

	ctx := context.Background()
	pool, err := infra.NewPool(ctx, *dsn)
	if err != nil {
		log.Fatalf("pg connect: %v", err)
	}
	defer pool.Close()

	if err := infra.Migrate(ctx, pool); err != nil {
		log.Fatalf("migrate: %v", err)
	}

	teamRepo := postgres.NewTeamRepo(pool)
	userRepo := postgres.NewUserRepo(pool)
	prRepo := postgres.NewPRRepo(pool)

	teamSvc := teamuc.NewService(teamRepo)
	userSvc := useruc.NewService(userRepo)
	prSvc := pruc.NewService(prRepo, teamRepo, userRepo)

	server := api.NewServer(teamSvc, userSvc, prSvc)

	go func() {
		log.Printf("listening on %s", *addr)
		if err := server.ListenAndServe(*addr); err != nil {
			log.Printf("server stopped: %v", err)
		}
	}()

	stop := make(chan os.Signal, 1)
	signal.Notify(stop, syscall.SIGINT, syscall.SIGTERM)
	<-stop
	ctxSh, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()
	_ = server.Shutdown(ctxSh)
}

func getEnv(k, def string) string {
	if v := os.Getenv(k); v != "" {
		return v
	}
	return def
}
