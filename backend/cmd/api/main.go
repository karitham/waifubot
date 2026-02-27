package main

import (
	"context"
	"encoding/json/v2"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/Karitham/corde"
	"github.com/Karitham/httperr"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/cors"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/urfave/cli/v2"

	"github.com/karitham/waifubot/discord"
	"github.com/karitham/waifubot/services"
	"github.com/karitham/waifubot/storage"
	"github.com/karitham/waifubot/storage/collectionstore"
	"github.com/karitham/waifubot/storage/userstore"
	"github.com/karitham/waifubot/wishlist"
)

var version = "dev"

var (
	cacheAge = strconv.Itoa(int((time.Minute * 2).Seconds()))

	userNotFound = &httperr.DefaultError{
		Message:    "user not found",
		ErrorCode:  "user_not_found",
		StatusCode: 404,
	}

	invalidIDError = &httperr.DefaultError{
		Message:    "invalid id provided",
		ErrorCode:  "invalid_id",
		StatusCode: 400,
	}

	apiRequestCounter = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: "waifubot_api_requests_total",
			Help: "Total number of API requests",
		},
		[]string{"method", "endpoint", "status"},
	)

	apiRequestDuration = prometheus.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    "waifubot_api_request_duration_seconds",
			Help:    "Duration of API requests",
			Buckets: prometheus.DefBuckets,
		},
		[]string{"method", "endpoint"},
	)
)

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

type responseWriter struct {
	http.ResponseWriter
	statusCode int
}

func (rw *responseWriter) WriteHeader(code int) {
	rw.statusCode = code
	rw.ResponseWriter.WriteHeader(code)
}

func prometheusMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		rw := &responseWriter{ResponseWriter: w, statusCode: 200}

		next.ServeHTTP(rw, r)

		duration := time.Since(start)

		if routeCtx := chi.RouteContext(r.Context()); routeCtx != nil {
			if pattern := routeCtx.RoutePattern(); pattern != "" {
				apiRequestCounter.WithLabelValues(r.Method, pattern, strconv.Itoa(rw.statusCode)).Inc()
				apiRequestDuration.WithLabelValues(r.Method, pattern).Observe(duration.Seconds())
			}
		}
	})
}

func main() {
	app := &cli.App{
		Name:    "waifubot-api",
		Usage:   "Run the waifubot API server",
		Version: version,
		Flags: []cli.Flag{
			&cli.StringFlag{
				Name:    "port",
				Aliases: []string{"p"},
				EnvVars: []string{"PORT"},
				Value:   "3333",
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
	apiPort, err := strconv.Atoi(c.String("port"))
	if err != nil || apiPort == 0 {
		apiPort = 3333
	}

	logLevel := parseLogLevel(c.String("log-level"))

	url := c.String("db-url")

	if err := storage.Migrate(url); err != nil {
		return fmt.Errorf("could not run migrations: %w", err)
	}

	db, err := storage.NewStore(context.Background(), url)
	if err != nil {
		return fmt.Errorf("could not connect to database: %w", err)
	}

	prometheus.MustRegister(apiRequestCounter, apiRequestDuration)

	discordToken := c.String("bot-token")
	var discordService *services.DiscordService
	if discordToken != "" {
		discordService = services.NewDiscordService(discordToken)
	}

	api := &Handler{
		db:             db,
		discordService: discordService,
	}

	logger := slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{Level: logLevel}))
	slog.Info("Starting API server", "port", apiPort, "log_level", logLevel.String())
	r := chi.NewRouter()
	r.Use(middleware.Timeout(5 * time.Second))
	r.Use(loggerMiddleware(logger))
	r.Use(middleware.Compress(5))
	r.Use(prometheusMiddleware)

	r.Use(cors.Handler(cors.Options{
		AllowedOrigins:   []string{"https://*", "http://*"},
		AllowedMethods:   []string{"GET", "OPTIONS"},
		AllowedHeaders:   []string{"Accept", "Content-Type"},
		MaxAge:           300,
		AllowCredentials: true,
	}))

	users := func(r chi.Router) {
		r.Use(middleware.SetHeader("Content-Type", "application/json"))
		r.Use(middleware.SetHeader("Cache-Control", "public, max-age="+cacheAge))
		r.Get("/find", api.findUser)
		r.Get("/{userID}", api.getUser)
	}

	r.Route("/user", users)
	r.Route("/api/v1", func(r chi.Router) {
		r.Route("/user", users)

		r.Route("/wishlist", func(r chi.Router) {
			r.Use(middleware.SetHeader("Content-Type", "application/json"))
			r.Use(middleware.SetHeader("Cache-Control", "public, max-age="+cacheAge))
			r.Get("/{userID}", api.getWishlist)
		})
	})

	r.Handle("/metrics", promhttp.Handler())

	slog.Info("API server started successfully", "port", apiPort)
	if err := http.ListenAndServe(":"+strconv.Itoa(apiPort), r); err != nil {
		slog.Error("API server crashed", "error", err, "port", apiPort)
		return err
	}
	slog.Info("API server shutting down", "port", apiPort)
	return nil
}

type Handler struct {
	db             storage.Store
	discordService *services.DiscordService
}

type Profile struct {
	ID              uint64      `json:"id,string"`
	Quote           string      `json:"quote,omitempty"`
	Tokens          int32       `json:"tokens,omitempty"`
	AnilistURL      string      `json:"anilist_url,omitempty"`
	DiscordUsername string      `json:"discord_username,omitempty"`
	DiscordAvatar   string      `json:"discord_avatar,omitempty"`
	Favorite        Character   `json:"favorite,omitzero"`
	Waifus          []Character `json:"waifus,omitempty"`
}

type Character struct {
	Date  time.Time `json:"date"`
	Name  string    `json:"name"`
	Image string    `json:"image"`
	Type  string    `json:"type"`
	ID    int64     `json:"id"`
}

func normalizeAnilistURL(input string) string {
	input = strings.TrimSpace(input)
	if strings.HasPrefix(input, "https://anilist.co/user/") || strings.HasPrefix(input, "http://anilist.co/user/") {
		return input
	}
	return fmt.Sprintf("https://anilist.co/user/%s", input)
}

func (h *Handler) getUser(w http.ResponseWriter, r *http.Request) {
	userID := chi.URLParam(r, "userID")

	id, err := strconv.ParseUint(userID, 10, 64)
	if err != nil || id == 0 {
		httperr.JSON(w, r, invalidIDError)
		return
	}

	if h.discordService != nil {
		ctx, cancel := context.WithTimeout(r.Context(), 10*time.Second)
		defer cancel()

		if err := h.discordService.UpdateIfNeeded(ctx, h.db, corde.Snowflake(id)); err != nil {
			slog.Debug("Failed to update Discord user info", "user_id", id, "error", err)
		}
	}

	u, err := h.db.UserStore().Get(r.Context(), id)
	if err != nil {
		httperr.JSON(w, r, userNotFound)
		return
	}

	chars, err := h.db.CollectionStore().List(r.Context(), id)
	if err != nil {
		httperr.JSON(w, r, userNotFound)
		return
	}

	if err = json.MarshalWrite(w, mapUser(u, chars)); err != nil {
		slog.Error("encoding request", "error", err)
	}
}

func (h *Handler) getWishlist(w http.ResponseWriter, r *http.Request) {
	userID := chi.URLParam(r, "userID")

	id, err := strconv.ParseUint(userID, 10, 64)
	if err != nil || id == 0 {
		httperr.JSON(w, r, invalidIDError)
		return
	}

	chars, err := wishlist.GetUserWishlist(r.Context(), wishlist.New(h.db.WishlistStore()), id)
	if err != nil {
		httperr.JSON(w, r, userNotFound)
		return
	}

	type resp struct {
		Characters []wishlist.Character `json:"characters"`
		Total      int                  `json:"total"`
	}

	if err = json.MarshalWrite(w, resp{
		Characters: chars,
		Total:      len(chars),
	}); err != nil {
		slog.Error("encoding request", "error", err)
	}
}

func (h *Handler) findUser(w http.ResponseWriter, r *http.Request) {
	anilist := r.URL.Query().Get("anilist")
	discord := r.URL.Query().Get("discord")

	if anilist == "" && discord == "" {
		httperr.JSON(w, r, &httperr.DefaultError{
			Message:    "anilist or discord query param is required",
			ErrorCode:  "missing_query_param",
			StatusCode: 400,
		})
		return
	}

	var user userstore.User
	var err error

	if anilist != "" {
		anilistURL := normalizeAnilistURL(anilist)
		user, err = h.db.UserStore().GetByAnilist(r.Context(), anilistURL)
	} else {
		user, err = h.db.UserStore().GetByDiscordUsername(r.Context(), discord)
	}

	if err != nil || user.UserID == 0 {
		httperr.JSON(w, r, userNotFound)
		return
	}

	type resp struct {
		ID uint64 `json:"id,string"`
	}

	if err = json.MarshalWrite(w, resp{
		ID: uint64(user.UserID),
	}); err != nil {
		slog.Error("encoding request", "error", err)
	}
}

func mapUser(u userstore.User, list []collectionstore.ListRow) *Profile {
	p := &Profile{
		ID:              u.UserID,
		Quote:           u.Quote,
		Tokens:          u.Tokens,
		AnilistURL:      u.AnilistUrl,
		DiscordUsername: u.DiscordUsername,
		DiscordAvatar:   discord.DiscordAvatarURL(u.UserID, u.DiscordAvatar),
		Waifus:          make([]Character, 0, len(list)),
	}

	for _, entry := range list {
		c := Character{
			ID:    entry.ID,
			Name:  entry.Name,
			Image: entry.Image,
			Type:  entry.Source,
			Date:  entry.Date.Time,
		}

		p.Waifus = append(p.Waifus, c)

		if u.Favorite.Valid && entry.ID == u.Favorite.Int64 {
			p.Favorite = c
		}
	}

	return p
}
