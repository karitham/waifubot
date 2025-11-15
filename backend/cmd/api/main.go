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

	"github.com/karitham/waifubot/db"
)

var cacheAge = strconv.Itoa(int((time.Minute * 2).Seconds()))

func main() {
	p := os.Getenv("PORT")
	apiPort, err := strconv.Atoi(p)
	if err != nil || apiPort == 0 {
		apiPort = 3333
	}

	url := os.Getenv("DB_URL")
	db, err := db.NewStore(context.Background(), url)
	if err != nil {
		slog.Error("Could not connect to database", "error", err)
		os.Exit(1)
	}

	api := &APIContext{
		db: db,
	}

	logger := slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelInfo}))
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
		fmt.Fprint(w, "WaifuBot API - See https://github.com/karitham/waifubot for documentation")
	})

	slog.Info("API started", "API_PORT", apiPort)
	if err := http.ListenAndServe(":"+strconv.Itoa(apiPort), r); err != nil {
		slog.Error("Listen and serve crash", "error", err, "Port", apiPort)
		os.Exit(1)
	}
}

type APIContext struct {
	db *db.Store
}

func normalizeAnilistURL(input string) string {
	input = strings.TrimSpace(input)
	if strings.HasPrefix(input, "https://anilist.co/user/") || strings.HasPrefix(input, "http://anilist.co/user/") {
		return input
	}
	return fmt.Sprintf("https://anilist.co/user/%s", input)
}

func (a *APIContext) getUser(w http.ResponseWriter, r *http.Request) {
	userID := chi.URLParam(r, "userID")

	id, err := strconv.ParseUint(userID, 10, 64)
	if err != nil || id == 0 {
		herr := &httperr.DefaultError{
			Message:    "invalid id provided",
			ErrorCode:  "GU0002",
			StatusCode: 400,
		}
		httperr.JSON(w, r, herr)
		slog.Debug("invalid ID", "error", herr)
		return
	}

	user, err := a.db.Profile(r.Context(), id)
	if err != nil || user.ID == 0 {
		herr := &httperr.DefaultError{
			Message:    "user not found",
			ErrorCode:  "GU0001",
			StatusCode: 404,
		}
		httperr.JSON(w, r, herr)
		slog.Debug("fetching user ID", "error", herr)
		return
	}

	if err = json.MarshalWrite(w, user); err != nil {
		slog.Error("encoding request", "error", err)
	}
}

func (a *APIContext) findUser(w http.ResponseWriter, r *http.Request) {
	anilist := r.URL.Query().Get("anilist")
	if anilist == "" {
		herr := &httperr.DefaultError{
			Message:    "anilist query param is required",
			ErrorCode:  "FU0001",
			StatusCode: 400,
		}

		httperr.JSON(w, r, herr)
		slog.Debug("fetching user ID", "error", herr)
		return
	}

	anilistURL := normalizeAnilistURL(anilist)
	user, err := a.db.UserByAnilistURL(r.Context(), anilistURL)
	if err != nil || user.UserID == 0 {
		herr := &httperr.DefaultError{
			Message:    "user not found",
			ErrorCode:  "FU0002",
			StatusCode: 404,
		}
		httperr.JSON(w, r, herr)
		slog.Debug("fetching user ID", "error", herr)
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
