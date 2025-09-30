# Discord Message Saver Bot

This project is a Discord bot built with Go and [discordgo](https://github.com/bwmarrin/discordgo). Users choose an emoji via a slash command and, when they react to a message with that emoji, the bot forwards the message to their DM with a Close button. Pressing Close removes the forwarded message.

## Requirements

- Go 1.20+
- Discord Bot Token / Application ID
- Docker (optional)

## Environment variables

| Name | Description |
| --- | --- |
| `DISCORD_TOKEN` | Bot token |
| `DISCORD_APP_ID` | Application ID |
| `DISCORD_GUILD_ID` | (Optional) Guild ID to register the command. Empty registers globally |

Use `.env.example` as a reference when configuring the environment.

## Running locally

```bash
go run ./cmd/bot
```

## Running with Docker

```bash
docker compose up --build
```

## Bot features

1. `/set-bookmark-emoji` lets you select the emoji used to bookmark messages.
2. Reacting with the registered emoji forwards the message (and attachment URLs) to your DM.
3. The forwarded DM includes a Close button. Press it to delete the DM.

The bot registers the slash command automatically when it starts, so no additional registration command is required.
