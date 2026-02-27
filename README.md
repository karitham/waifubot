# Waifu Bot

**[Add it to your server](https://discord.com/oauth2/authorize?scope=bot&client_id=712332547694264341&permissions=92224)**

A Discord bot for collecting and trading anime/manga characters. Roll for random characters, claim drops, build your collection, and trade with friends.

## Features

- **Character collection**: Roll and claim characters from anime and manga
- **Trading system**: Give characters to other users or exchange them for tokens
- **Wishlist system**: Create and manage wishlists of desired characters, find trading partners
- **Web interface**: View and manage your collection and wishlist at [waifugui.karitham.dev](https://waifugui.karitham.dev)
- **API access**: Retrieve data programmatically at [waifuapi.karitham.dev](https://waifuapi.karitham.dev)
- **AniList integration**: Search for anime, manga, characters, and users

## Commands

- **claim**: claim a dropped character
- **exchange**: exchange a character for a token
- **give**: give a character to someone
- **holders**: list users in this server who have a character
- **info**: information about the bot
- **list**: view character collection
- **roll**: roll a random character
- **verify**: check if a user has a character
- **profile**:
  - **view**: view a user's profile
  - **edit**: edit your profile
    - **anilist**: set your AniList URL
    - **favorite**: set your favorite character
    - **quote**: set your quote
- **search**:
  - **anime**: search for an anime
  - **char**: search for a character
  - **manga**: search for a manga
  - **user**: search for a user
- **wishlist**:
  - **character add**: add a character to your wishlist
  - **character remove**: remove a character from your wishlist
  - **character list**: view your wishlist
  - **media add**: add all characters from an anime/manga to your wishlist
  - **holders**: find users who have characters from your wishlist
  - **wanted**: find users who want characters you own
  - **compare**: compare your wishlist with another user's collection

## Development

This project uses [nix flakes](https://nixos.org/manual/nix/stable/command-ref/new-cli/nix3-flake) for development.

### Prerequisites

- [Nix](https://nixos.org/download.html) with flakes enabled
- [Direnv](https://direnv.net/) (optional, for auto-loading the shell)

### Setup

```bash
# Clone the repository
git clone https://github.com/karitham/waifubot
cd waifubot

# Enable direnv (recommended) or enter the dev shell manually
direnv allow
# OR
nix develop
```

### Running Locally

1. **Start PostgreSQL**:

   ```bash
   docker-compose up -d
   ```

2. **Backend**:

   ```bash
   # Build binaries via nix
   nix build .#waifubot

   # Or run directly
   cd backend
   go run ./cmd/bot run
   ```

3. **Frontend**:
   ```bash
   cd frontend
   npm run dev
   ```

The backend will be available at `http://localhost:8080` (bot, metrics at `/metrics`) and `http://localhost:3333` (API).
The frontend will be available at `http://localhost:5173`.

## Self-Hosting

### Backend

The backend can be deployed via Kubernetes using the manifests in the [infra repository](https://github.com/karitham/infra/tree/main/apps/waifubot).

**Requirements:**

- Discord application (bot token, application ID, public key)
- PostgreSQL database

**Environment variables:**

- `BOT_TOKEN` - Discord bot token
- `APP_ID` - Discord application ID
- `PUBLIC_KEY` - Discord application public key
- `DB_URL` - PostgreSQL connection string

### Frontend

The frontend is a static site. Build it and deploy to any static hosting provider.

```bash
cd frontend
npm install
npm run build
```

Deploy the `dist/` directory to:

- Cloudflare Pages
- Vercel
- Netlify
- Any web server

**API URL**: Set `VITE_API_URL` at build time to point to your API instance:

```bash
VITE_API_URL=https://your-api.example.com npm run build
```
