package main

import (
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/cors"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/urfave/cli/v2"

	"github.com/karitham/waifubot/rest"
	"github.com/karitham/waifubot/rest/api"
	"github.com/karitham/waifubot/services"
	"github.com/karitham/waifubot/storage"
)

var version = "dev"

func parseLogLevel(levelStr string) slog.Level {
	switch strings.ToUpper(levelStr) {
	case "DEBUG":
		return slog.LevelDebug
	case "INFO":
		return slog.LevelInfo
	case "WARN", "WARNING":
		return slog.LevelWarn
	case "ERROR":
		return slog.LevelError
	default:
		return slog.LevelInfo
	}
}

func main() {
	app := &cli.App{
		Name:    "waifubot-api",
		Usage:   "Run the waifubot API server",
		Version: version,
		Flags: []cli.Flag{
			&cli.IntFlag{
				Name:    "port",
				Aliases: []string{"p"},
				EnvVars: []string{"PORT"},
				Value:   3333,
			},
			&cli.StringFlag{
				Name:    "log-level",
				Aliases: []string{"l"},
				EnvVars: []string{"LOG_LEVEL"},
				Value:   "INFO",
			},
			&cli.StringFlag{
				Name:     "db-url",
				Aliases:  []string{"d"},
				EnvVars:  []string{"DB_URL", "DB_STR"},
				Required: true,
			},
			&cli.StringFlag{
				Name:    "bot-token",
				EnvVars: []string{"BOT_TOKEN"},
			},
		},
		Action: runServer,
	}

	if err := app.Run(os.Args); err != nil {
		slog.Error("error running app", "error", err)
		os.Exit(1)
	}
}

func runServer(c *cli.Context) error {
	apiPort := c.Int("port")
	if apiPort == 0 {
		apiPort = 3333
	}

	logLevel := parseLogLevel(c.String("log-level"))

	url := c.String("db-url")

	if err := storage.Migrate(url); err != nil {
		return fmt.Errorf("could not run migrations: %w", err)
	}

	db, err := storage.NewStore(c.Context, url)
	if err != nil {
		return fmt.Errorf("could not connect to database: %w", err)
	}

	discordToken := c.String("bot-token")
	var discordService *services.DiscordService
	if discordToken != "" {
		discordService = services.NewDiscordService(discordToken)
	}

	restServer := rest.New(db, discordService)

	logger := slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{Level: logLevel}))
	slog.Info("Starting API server", "port", apiPort, "log_level", logLevel.String())

	telemetry, err := rest.SetupTelemetry(prometheus.DefaultRegisterer)
	if err != nil {
		return fmt.Errorf("failed to setup telemetry: %w", err)
	}
	defer func() {
		if err := telemetry.Shutdown(c.Context); err != nil {
			slog.Error("Error shutting down telemetry", "error", err)
		}
	}()

	r := chi.NewRouter()
	r.Use(middleware.Timeout(5 * time.Second))
	r.Use(rest.LoggerMiddleware(logger))
	r.Use(middleware.Compress(5))

	r.Use(cors.Handler(cors.Options{
		AllowedOrigins:   []string{"https://*", "http://*"},
		AllowedMethods:   []string{"GET", "OPTIONS"},
		AllowedHeaders:   []string{"Accept", "Content-Type"},
		MaxAge:           300,
		AllowCredentials: true,
	}))

	apiRouter, err := api.NewServer(
		restServer,
		api.WithMeterProvider(telemetry.MeterProvider()),
	)
	if err != nil {
		return fmt.Errorf("failed to create API router: %w", err)
	}

	r.Mount("/", apiRouter)
	r.Handle("/metrics", promhttp.Handler())

	slog.Info("API server started successfully", "port", apiPort)
	if err := http.ListenAndServe(":"+strconv.Itoa(apiPort), r); err != nil {
		slog.Error("API server crashed", "error", err, "port", apiPort)
		return err
	}

	slog.Info("API server shutting down", "port", apiPort)
	return nil
}
