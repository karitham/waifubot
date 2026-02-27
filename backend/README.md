# Waifu Bot Backend

Discord bot for collecting and trading anime/manga characters, with a REST API.

## Binaries

This project produces two binaries:

- **`bot`** - Discord bot with commands for rolling, claiming, trading characters
- **`api`** - REST API for fetching user profiles and wishlists

Build with: `nix build .#waifubot`

## Environment Variables

### Required

| Variable     | Description                    |
| ------------ | ------------------------------ |
| `BOT_TOKEN`  | Discord bot token              |
| `APP_ID`     | Discord application ID         |
| `PUBLIC_KEY` | Discord application public key |
| `DB_URL`     | PostgreSQL connection string   |

### Optional (Bot)

| Variable             | Default | Description                         |
| -------------------- | ------- | ----------------------------------- |
| `PORT`               | `8080`  | HTTP server port                    |
| `ROLL_TIMEOUT`       | `2h`    | Cooldown between rolls              |
| `TOKENS_NEEDED`      | `3`     | Tokens required to exchange         |
| `INTERACTION_NEEDED` | `25`    | Interactions needed to get a token  |
| `SKIP_MIGRATE`       | `false` | Skip database migrations on startup |

### Optional (API)

| Variable    | Default | Description                          |
| ----------- | ------- | ------------------------------------ |
| `PORT`      | `3333`  | HTTP server port                     |
| `LOG_LEVEL` | `INFO`  | Log level (DEBUG, INFO, WARN, ERROR) |

## Deployment

See the [infra repository](https://github.com/karitham/infra/tree/main/apps/waifubot) for Kubernetes manifests.

## Development

```bash
# Enter dev shell
nix develop

# Run bot
cd backend
go run ./cmd/bot run

# Run API (in another terminal)
go run ./cmd/api
```
