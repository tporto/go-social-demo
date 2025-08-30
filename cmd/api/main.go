package main

import (
	"go-social/internal/auth"
	"go-social/internal/db"
	"go-social/internal/env"
	"go-social/internal/mailer"
	"go-social/internal/store"
	"go-social/internal/store/cache"
	"log"
	"time"

	"github.com/go-redis/redis/v8"
	"go.uber.org/zap"
)

const version = "1.0.0"

func main() {
	cfg := config{
		addr:        env.GetString("PORT", ":8080"),
		frontendURL: env.GetString("FRONTEND_URL", "http://localhost:4000"),
		db: dbConfig{
			addr:        env.GetString("DB_ADDR", "host=localhost port=5432 user=postgres password=postgres dbname=social sslmode=disable"),
			maxOpenConn: env.GetInt("DB_MAX_OPEN_CONN", 30),
			maxIdleConn: env.GetInt("DB_MAX_IDLE_CONN", 30),
			maxIdleTime: env.GetString("DB_MAX_IDLE_TIME", "15m"),
		},
		env: env.GetString("ENV", "development"),
		mail: mailConfig{
			exp:       time.Hour * 24 * 3, // 3 days
			fromEmail: env.GetString("FROM_EMAIL", ""),
			sendGrid: sendGridConfig{
				apiKey: env.GetString("SENDGRID_API_KEY", ""),
			},
			mailTrap: mailTrapConfig{
				apiKey: env.GetString("MAILTRAP_API_KEY", ""),
			},
		},
		auth: authConfig{
			basic: basicConfig{
				username: env.GetString("AUTH_BASIC_USER", "admin"),
				password: env.GetString("AUTH_BASIC_PASS", "admin"),
			},
			token: tokenConfig{
				secret: env.GetString("AUTH_TOKEN_SECRET", "test"),
				exp:    time.Hour * 24 * 3,
				iss:    "gophersocial",
			},
		},
		redisCfg: redisConfig{
			addr:    env.GetString("REDIS_ADDR", "localhost:6379"),
			pw:      env.GetString("REDIS_PW", ""),
			db:      env.GetInt("REDIS_DB", 0),
			enabled: env.GetBool("REDIS_ENABLED", false),
		},
	}

	// Logger
	logger := zap.Must(zap.NewProduction()).Sugar()
	defer logger.Sync()

	// Database configuration
	db, err := db.New(cfg.db.addr, cfg.db.maxOpenConn, cfg.db.maxIdleConn, cfg.db.maxIdleTime)
	if err != nil {
		logger.Fatal(err)
	}
	defer db.Close()
	logger.Info("database connection pool established")

	// Cache
	var rdb *redis.Client
	if cfg.redisCfg.enabled {
		rdb = cache.NewRedisClient(cfg.redisCfg.addr, cfg.redisCfg.pw, cfg.redisCfg.db)
		logger.Info("redis cache connection established")

		defer rdb.Close()
	}

	// Mailer
	mailer, _ := mailer.NewMailTrapClient(cfg.mail.mailTrap.apiKey, cfg.mail.fromEmail)
	// AuthJWT
	jwtAuth := auth.NewJWTAuthenticator(cfg.auth.token.secret, cfg.auth.token.iss, cfg.auth.token.iss)
	// Repository
	store := store.NewStorage(db)
	cacheStorage := cache.NewRedisStorage(rdb)

	app := application{
		config:        cfg,
		store:         store,
		logger:        logger,
		mailer:        mailer,
		authenticator: jwtAuth,
		cacheStorage:  cacheStorage,
	}

	mux := app.mount()
	log.Fatal(app.run(mux))
}
