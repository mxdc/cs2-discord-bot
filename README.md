# CS2 Discord Bot

A Discord bot that automatically tracks Counter-Strike 2 match results and sends notifications to your Discord server when games finish.

<img width="490" height="180" alt="cs2-notif" src="https://github.com/user-attachments/assets/9291fca2-71df-47cd-82aa-5406a69b108d" />

## ✨ Features

- **🔄 Match Tracking**: Continuously monitors multiple Steam accounts for completed CS2 matches
- **⚡ Notifications**: Get Discord messages when matches end with detailed results
- **🏆 MVP Recognition**: Highlights the top performer with country flags (when Steam API is configured)
- **� Deduplication**: Prevents duplicate notifications when teammates play together in the same match

## Setup

### Prerequisites

1. **Discord Webhook URL**: Create a webhook in your Discord server
   - Go to Server Settings > Integrations > Webhooks
   - Create a new webhook and copy the URL

2. **Steam ID**: Your Steam ID (64-bit format)
   - You can find this from your Steam profile URL or use sites like steamid.io

3. **Steam API Key**: Steam Web API key for retrieving player country information
   - Go to [Steam Web API Key Registration](https://steamcommunity.com/dev/apikey)
   - Register for an API key (free)
   - Copy the generated key

## Design

```
┌─────────────┐
│             │
│ Leetify API │
│             │
└─────────────┘
        ▲
        │    ┌───────────────────────────────────────────────────────────┐
        │    │                     cs2-discord-bot                       │
        │    │ ┌──────────┐                                              │
        ├────┼─│Crawler 1 │──┐                     ┌───────────────────┐ │
        │    │ └──────────┘  │                     │ Notifier          │ │    ┌─────────────────┐
        │    │ ┌──────────┐  └╖  ┌───────────┐     │                   │ │    │                 │
        ├────┼─│Crawler 2 │───║─▶│  Channel  │◀────│ - Dequeue Match   │─┼───▶│ Discord webhook │
        │    │ └──────────┘  ┌╜  └───────────┘     │ - Format Message  │ │    │                 │
        │    │ ┌──────────┐  │                     │ - Send to Discord │ │    └─────────────────┘
        └────┼─│Crawler X │──┘                     └───────────────────┘ │
             │ └──────────┘                                              │
             └───────────────────────────────────────────────────────────┘
```

### Configuration

Create a `config.yml` file in your project directory:

```yaml
---

# steam
steam_api_key: "your_steam_api_key"
steam_api_url: "https://api.steampowered.com"

# leetify
leetify_api_url: "https://api.leetify.com"

# discord
discord_hook: "https://discord.com/api/webhooks/your/token"

# one crawler per tracked account
players:
- accountName: "player1"
  steamId: "76565432XXXXXXXXX"
  track: true
- accountName: "player2"
  steamId: "76561198XXXXXXXXX"
  track: false
```

### Installation

1. Clone the repository:
```bash
$ git clone https://github.com/mxdc/cs2-discord-bot
$ cd cs2-discord-bot
```

2. Edit `config.yml` with your actual values

3. Build the application:
```bash
$ go mod tidy
$ go build -o cs2-discord-bot main.go
```

4. Run the bot:
```bash
$ ./cs2-discord-bot --config.file="/path/to/config.yml"
```
