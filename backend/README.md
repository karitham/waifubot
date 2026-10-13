# Waifu Bot Backend

Discord bot for collecting and trading anime/manga characters, with a REST API.

## Binary

The `waifubot` binary runs the Discord bot and the REST API in one process; its subcommands expose maintenance tools (`waifubot --help` lists them).

Build with: `nix build .#waifubot`

## Environment Variables

### Required

| Variable     | Description                    |
| ------------ | ------------------------------ |
| `BOT_TOKEN`  | Discord bot token              |
| `APP_ID`     | Discord application ID         |
| `PUBLIC_KEY` | Discord application public key |
| `DB_URL`     | PostgreSQL connection string   |

### Optional

| Variable             | Default | Description                                   |
| -------------------- | ------- | --------------------------------------------- |
| `PORT`               | `8080`  | HTTP server port                              |
| `ROLL_TIMEOUT`       | `2h`    | Cooldown between rolls                        |
| `SERIES_ROLL_COST`   | `20`    | Token cost of a series roll                   |
| `INTERACTION_NEEDED` | `25`    | Interactions needed to trigger a channel drop |
| `SKIP_MIGRATE`       | `false` | Skip database migrations on startup           |
| `LOG_LEVEL`          | `INFO`  | Log level (DEBUG, INFO, WARN, ERROR)          |

See `cmd/waifubot/run.go` for the full list of flags.

## Deployment

See the [infra repository](https://github.com/karitham/infra/tree/main/apps/waifubot) for Kubernetes manifests.

## Development

```bash
# Enter dev shell
nix develop

# Run the bot and API
cd backend
go run ./cmd/waifubot run
```
