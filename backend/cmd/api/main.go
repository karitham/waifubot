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

	"github.com/karitham/waifubot/discord"
	"github.com/karitham/waifubot/services"
	"github.com/karitham/waifubot/storage"
	"github.com/karitham/waifubot/storage/collectionstore"
	"github.com/karitham/waifubot/storage/userstore"
	"github.com/karitham/waifubot/wishlist"
)

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

		// Use route pattern instead of URL path to avoid cardinality explosion
		// Only record metrics for routed endpoints
		if routeCtx := chi.RouteContext(r.Context()); routeCtx != nil {
			if pattern := routeCtx.RoutePattern(); pattern != "" {
				apiRequestCounter.WithLabelValues(r.Method, pattern, strconv.Itoa(rw.statusCode)).Inc()
				apiRequestDuration.WithLabelValues(r.Method, pattern).Observe(duration.Seconds())
			}
		}
	})
}

func main() {
	p := os.Getenv("PORT")
	apiPort, err := strconv.Atoi(p)
	if err != nil || apiPort == 0 {
		apiPort = 3333
	}

	logLevel := parseLogLevel(os.Getenv("LOG_LEVEL"))

	url := os.Getenv("DB_URL")
	db, err := storage.NewStore(context.Background(), url)
	if err != nil {
		slog.Error("Could not connect to database", "error", err)
		os.Exit(1)
	}

	prometheus.MustRegister(apiRequestCounter, apiRequestDuration)

	discordToken := os.Getenv("BOT_TOKEN")
	if discordToken == "" {
		slog.Warn("BOT_TOKEN not set, Discord user info will not be updated")
	}

	api := &Handler{
		db:             db,
		discordService: services.NewDiscordService(discordToken),
	}

	logger := slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{Level: logLevel}))
	slog.Info("Starting API server", "port", apiPort, "log_level", logLevel.String())
	r := chi.NewRouter()
	r.Use(middleware.Timeout(5 * time.Second))
	r.Use(loggerMiddleware(logger))
	r.Use(middleware.Compress(5))
	r.Use(prometheusMiddleware)

	// CORS
	r.Use(cors.Handler(cors.Options{
		AllowedOrigins:   []string{"https://*", "http://*"},
		AllowedMethods:   []string{"GET", "OPTIONS"},
		AllowedHeaders:   []string{"Accept", "Content-Type"},
		MaxAge:           300, // Maximum value not ignored by any of major browsers
		AllowCredentials: true,
	}))

	users := func(r chi.Router) {
		r.Use(middleware.SetHeader("Content-Type", "application/json"))
		r.Use(middleware.SetHeader("Cache-Control", "public, max-age="+cacheAge))
		r.Get("/find", api.findUser)
		r.Get("/{userID}", api.getUser)
	}

	// deprecated path
	r.Route("/user", users)
	r.Route("/api/v1", func(r chi.Router) {
		// Implement GET /user/123
		r.Route("/user", users)

		// Implement GET /wishlist/123
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
		os.Exit(1)
	}
	slog.Info("API server shutting down", "port", apiPort)
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

	// Update Discord info synchronously if needed
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
