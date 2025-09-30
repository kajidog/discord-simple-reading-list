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

1. `/set-bookmark-emoji` lets you select one or more emojis (comma or space separated) used to bookmark messages and optionally choose a hex embed color.
2. Reacting with any registered emoji forwards the message (and attachment URLs) to your DM inside an embed that honours your color preference.
3. The forwarded DM includes a Close button. Press it to delete the DM.

The bot registers the slash command automatically when it starts, so no additional registration command is required.

### Command usage

Use the following format when customising the bookmark behaviour:

```
/set-bookmark-emoji emoji:"📚 🔖" color:#ffcc00
```

- Provide one or more emojis separated by spaces or commas. Custom server emojis are supported as usual (e.g. `<:name:123456>`).
- The optional `color` argument accepts a 6-digit hex value with or without `#`/`0x` prefixes. Leave it out to fall back to the bot default.
