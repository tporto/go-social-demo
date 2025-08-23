package main

import (
	"go-social/internal/db"
	"go-social/internal/env"
	"go-social/internal/store"
	"log"
)

const version = "1.0.0"

func main() {
	cfg := config{
		addr: env.GetString("PORT", ":8080"),
		db: dbConfig{
			addr:        env.GetString("DB_ADDR", "host=localhost port=5432 user=postgres password=postgres dbname=social sslmode=disable"),
			maxOpenConn: env.GetInt("DB_MAX_OPEN_CONN", 30),
			maxIdleConn: env.GetInt("DB_MAX_IDLE_CONN", 30),
			maxIdleTime: env.GetString("DB_MAX_IDLE_TIME", "15m"),
		},
		env: env.GetString("ENV", "development"),
	}

	log.Println("DB_ADDR:", cfg.db.addr)

	// Database configuration
	db, err := db.New(cfg.db.addr, cfg.db.maxOpenConn, cfg.db.maxIdleConn, cfg.db.maxIdleTime)
	if err != nil {
		log.Panic(err)
	}
	defer db.Close()
	log.Println("database connection pool established")

	store := store.NewStorage(db)

	app := application{
		config: cfg,
		store:  store,
	}

	mux := app.mount()
	log.Fatal(app.run(mux))
}
