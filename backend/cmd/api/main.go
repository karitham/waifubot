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

	"github.com/Karitham/httperr"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/cors"

	"github.com/karitham/waifubot/storage"
	"github.com/karitham/waifubot/storage/collectionstore"
	"github.com/karitham/waifubot/storage/userstore"
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

	anilistRequiredError = &httperr.DefaultError{
		Message:    "anilist query param is required",
		ErrorCode:  "invalid_anilist",
		StatusCode: 400,
	}
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

	api := &Handler{
		db: db,
	}

	logger := slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{Level: logLevel}))
	slog.Info("Starting API server", "port", apiPort, "log_level", logLevel.String())
	r := chi.NewRouter()
	r.Use(middleware.Timeout(5 * time.Second))
	r.Use(loggerMiddleware(logger))
	r.Use(middleware.Compress(5))

	// CORS
	r.Use(cors.Handler(cors.Options{
		AllowedOrigins:   []string{"https://*", "http://*"},
		AllowedMethods:   []string{"GET", "OPTIONS"},
		AllowedHeaders:   []string{"Accept", "Content-Type"},
		MaxAge:           300, // Maximum value not ignored by any of major browsers
		AllowCredentials: true,
	}))

	// Implement GET /user/123
	r.Route("/user", func(r chi.Router) {
		r.Use(middleware.SetHeader("Content-Type", "application/json"))
		r.Use(middleware.SetHeader("Cache-Control", "public, max-age="+cacheAge))
		r.Get("/find", api.findUser)
		r.Get("/{userID}", api.getUser)
	})
	r.Get("/", func(w http.ResponseWriter, _ *http.Request) {
		_, _ = fmt.Fprint(w, "WaifuBot API - See https://github.com/karitham/waifubot for documentation")
	})

	slog.Info("API server started successfully", "port", apiPort)
	if err := http.ListenAndServe(":"+strconv.Itoa(apiPort), r); err != nil {
		slog.Error("API server crashed", "error", err, "port", apiPort)
		os.Exit(1)
	}
	slog.Info("API server shutting down", "port", apiPort)
}

type Handler struct {
	db storage.Store
}

type Profile struct {
	ID         uint64      `json:"id,string"`
	Quote      string      `json:"quote,omitempty"`
	Tokens     int32       `json:"tokens,omitempty"`
	AnilistURL string      `json:"anilist_url,omitempty"`
	Favorite   Character   `json:"favorite,omitzero"`
	Waifus     []Character `json:"waifus,omitempty"`
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

func (h *Handler) findUser(w http.ResponseWriter, r *http.Request) {
	anilist := r.URL.Query().Get("anilist")
	if anilist == "" {
		httperr.JSON(w, r, anilistRequiredError)
		return
	}

	anilistURL := normalizeAnilistURL(anilist)
	user, err := h.db.UserStore().GetByAnilist(r.Context(), anilistURL)
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
		ID:         u.UserID,
		Quote:      u.Quote,
		Tokens:     u.Tokens,
		AnilistURL: u.AnilistUrl,
		Waifus:     make([]Character, 0, len(list)),
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
