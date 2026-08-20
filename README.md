# Waifu Bot

[Add to server](https://discord.com/oauth2/authorize?scope=bot&client_id=712332547694264341&permissions=92224)

Waifu Bot is a Discord bot for collecting and trading anime and manga characters. Users roll for random characters, claim drops that appear after channel activity, build collections, and trade with other users.

## Features

- Character collection via rolls and channel drops that require a name match to claim.
- Trading between users and token exchange for characters.
- Wishlist to track desired characters and find trading partners, including bulk add from anime or manga.
- Profiles with favorite character, quote, and AniList link.
- Search for anime, manga, characters, and AniList users via AniList.
- Web interface and API for browsing collections and wishlists.

The web interface is at [waifugui.karitham.dev](https://waifugui.karitham.dev). The API is at [waifuapi.karitham.dev](https://waifuapi.karitham.dev) and is defined in `openapi.yaml`; the frontend generates its client from that file.

## Development

The project uses [Nix flakes](https://nixos.org/manual/nix/stable/command-ref/new-cli/nix3-flake) for development. The dev shell provides Go, Node.js, and required tools.

### Prerequisites

- [Nix](https://nixos.org/download.html) with flakes enabled
- [Direnv](https://direnv.net/) for automatic shell loading (optional)

### Setup

```bash
git clone https://github.com/karitham/waifubot
cd waifubot
direnv allow
# OR
nix develop
```

Create environment variables for local development (`.envrc` is gitignored):

```bash
export BOT_TOKEN=...
export APP_ID=...
export PUBLIC_KEY=...
export DB_URL=postgres://postgres:postgres@localhost:5432/postgres?sslmode=disable
```

For all options, see `backend/cmd/waifubot/run.go`.

### Running locally

1. Start PostgreSQL:

   ```bash
   docker-compose up -d
   ```

2. Start the backend:

   ```bash
   cd backend
   go run ./cmd/waifubot run
   ```

   Migrations run automatically. The bot and API listen on `http://localhost:8080` (`/metrics` on the same port). The frontend expects the API at that address.

3. Start the frontend:

   ```bash
   cd frontend
   npm install
   npm run dev
   ```

   The frontend listens on `http://localhost:5173`.

For Discord slash commands, see `backend/discord/commands.go` and the in-app command list. For API details, see `openapi.yaml`.

## Self-hosting

### Backend

Build a binary with `nix build .#waifubot` or use the container from `ghcr.io/karitham/waifubot`. Provide a PostgreSQL database and the required environment variables (`BOT_TOKEN`, `APP_ID`, `PUBLIC_KEY`, `DB_URL`). Run with `waifubot run`.

For NixOS, see `nix/module.nix`. For Kubernetes, see [karitham/infra](https://github.com/karitham/infra/tree/main/apps/waifubot).

### Frontend

Build a static site:

```bash
cd frontend
npm install
VITE_API_URL=https://api.example.com npm run build
```

Deploy `dist/` to any static host such as Cloudflare Pages, Vercel, or Netlify.

Set `VITE_API_URL` at build time to the API instance.

## License

MIT — see [LICENSE](LICENSE). Copyright 2020 PERY "Karitham" Pierre-Louis.
