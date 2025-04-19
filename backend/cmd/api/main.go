package main

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"strconv"
	"time"

	"github.com/Karitham/httperr"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/cors"
	"github.com/rs/zerolog/log"

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
		log.Fatal().Err(err).Msg("Could not connect to database")
	}
	defer db.Close(context.Background())

	api := &APIContext{
		db: db,
	}

	r := chi.NewRouter()
	r.Use(middleware.Timeout(5 * time.Second))
	r.Use(loggerMiddleware(&log.Logger))
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
		fmt.Fprint(w, "Hello user, you shouldn't be there, direct yourself to https://github.com/karitham/waifubot for docs")
	})

	log.Info().Int("API_PORT", apiPort).Msg("API started")
	if err := http.ListenAndServe(":"+strconv.Itoa(apiPort), r); err != nil {
		log.Fatal().Err(err).Int("Port", apiPort).Msg("Listen and serve crash")
	}
}

type APIContext struct {
	db *db.Store
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
		log.Debug().Err(herr).Msg("invalid ID")
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
		log.Debug().Err(herr).Msg("fetching user ID")
		return
	}

	if err = json.NewEncoder(w).Encode(user); err != nil {
		log.Err(err).Msg("encoding request")
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
		log.Debug().Err(herr).Msg("fetching user ID")
		return
	}

	user, err := a.db.UserByAnilistURL(r.Context(), fmt.Sprintf("https://anilist.co/user/%s", anilist))
	if err != nil || user.UserID == 0 {
		herr := &httperr.DefaultError{
			Message:    "user not found",
			ErrorCode:  "FU0002",
			StatusCode: 404,
		}
		httperr.JSON(w, r, herr)
		log.Debug().Err(herr).Msg("fetching user ID")
		return
	}

	type resp struct {
		ID uint64 `json:"id,string"`
	}

	if err = json.NewEncoder(w).Encode(resp{
		ID: uint64(user.UserID),
	}); err != nil {
		log.Err(err).Msg("encoding request")
	}
}
