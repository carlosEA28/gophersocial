package main

import (
	"log"
	"time"

	"github.com/carlosEA28/Social/internal/auth"
	"github.com/carlosEA28/Social/internal/db"
	"github.com/carlosEA28/Social/internal/env"
	"github.com/carlosEA28/Social/internal/mail"
	"github.com/carlosEA28/Social/internal/repository"
	"github.com/joho/godotenv"
	"go.uber.org/zap"
)

const version = "0.0.1"

func main() {

	err := godotenv.Load()
	if err != nil {
		log.Fatal("Error loading .env file")
	}

	cfg := config{
		addr:        env.GetString("ADDR", ":8080"),
		frontendURL: "http://localhost:5173",
		db: dbConfig{
			addr:         env.GetString("DB_ADDR", "postgres://admin:admin@localhost:5433/go-social?sslmode=disable"),
			maxOpenConns: env.GetInt("DB_MAX_OPEN_CONNS", 30),
			maxIdleConns: env.GetInt("DB_MAX_IDLE_CONNS", 30),
			maxIdleTime:  env.GetString("DB_MAX_IDLE_TIME", "15m"),
		},
		env: env.GetString("ENV", "development"),
		mail: mailConfig{
			exp: time.Hour * 24 * 3, // 3 days
			mailtrap: mailtrapConfig{
				apiKey:    env.GetString("MAILTRAP_API_KEY", ""),
				fromEmail: env.GetString("FROM_ADDRESS", ""),
			},
		},
		auth: AuthConfig{
			token: TokenConfig{
				secret:  env.GetString("AUTH_SECRET", "example"),
				expDate: time.Hour * 24 * 3, //3 days(por nas envs)
				issuer:  "gophersocial",
			},
		},
	}

	//logger
	logger := zap.Must(zap.NewProduction()).Sugar()
	defer logger.Sync()

	//database
	db, err := db.New(cfg.db.addr, cfg.db.maxOpenConns, cfg.db.maxIdleConns, cfg.db.maxIdleTime)

	if err != nil {
		logger.Fatal(err)
	}

	defer db.Close()
	logger.Info("database running")

	store := repository.NewPostgresStorage(db)

	mailClient, err := mail.NewMailTrapClient(cfg.mail.mailtrap.apiKey, cfg.mail.mailtrap.fromEmail)
	if err != nil {
		log.Fatal(err)
	}

	JwtAuthenticator := auth.NewJWTAuthenticator(cfg.auth.token.secret, cfg.auth.token.issuer, cfg.auth.token.issuer)

	app := &app{
		config:        cfg,
		store:         store,
		logger:        logger,
		mail:          mailClient,
		authenticator: JwtAuthenticator,
	}

	mux := app.mount()
	logger.Fatal(app.run(mux))

}
