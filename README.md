# Discord Message Saver Bot

This project is a Discord bot built with Go and [discordgo](https://github.com/bwmarrin/discordgo). Users can associate multiple emojis with different bookmark modes and, when they react to a message with one of those emojis, the bot forwards the message to their DM with the appropriate layout and controls.

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
| `BOOKMARK_STORE_PATH` | (Optional) Path to persist user bookmark settings. Defaults to `bookmarks.json` |

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

1. `/set-bookmark` lets you choose an emoji, assign it to one of three bookmark modes, and optionally pick an embed color.
2. `/list-bookmarks` shows the emojis you have configured and their associated modes and colors.
3. `/bookmark-help` provides a quick reference for the available commands and how to use them.
4. Reacting with any registered emoji forwards the message to your DM using the configured mode (lightweight, balanced, or complete).
5. Saved messages include mode-specific action buttons such as ✅ 完了, 🗑️ 削除, and 🔗 元メッセージ.

The bot registers the slash command automatically when it starts, so no additional registration command is required.

### Command usage

Use the following format when customising the bookmark behaviour:

```
/set-bookmark emoji:👀 mode:lightweight color:#FFD700
/set-bookmark emoji:🔖 mode:balanced
/set-bookmark emoji:📌 mode:complete color:#FF6B6B
/set-bookmark emoji:⏰ mode:lightweight reminder:8:00
/set-bookmark emoji:⏰ mode:lightweight reminder:45m keep-reminder-on-complete:true
/list-bookmarks
/bookmark-help
```

- Provide exactly one emoji per command execution. Custom server emojis are supported as usual (e.g. `<:name:123456>`).
- Choose between `lightweight`, `balanced`, or `complete` for the `mode` option.
- The optional `color` argument accepts a 6-digit hex value with or without `#`/`0x` prefixes. Leave it out to fall back to the bot default.
- Use the optional `reminder` argument to schedule a reminder for each saved message. Supply either a time of day such as `08:00` or a duration like `30m`/`2h`.
- When a reminder is set, the saved DM includes the next reminder time. Reminders can be cleared with `reminder:none`.
- Add `keep-reminder-on-complete:true` if you want the reminder to remain active after pressing the ✅ 完了 button. By default the reminder is removed when the bookmark is marked as complete.
