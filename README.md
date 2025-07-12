# FamBot-Go 🤖

A sassy Slack bot built in Go that brings good vibes to your workspace! FamBot helps track karma, remembers important dates, and responds with just the right amount of sass.

## ✨ Features

1. **Karma System** - When a user types `@username++` we increment their karma score with threaded responses
2. **Cross-Channel Support** - Bot listens and responds in all public channels, not just #people
3. **Threaded Responses** - All karma and thank you responses appear in message threads for better organization
4. **Grateful Channel Integration** - Track thank you messages in a dedicated channel with clickable thread links
5. **Polite Rewards** - When users mention "thank you", get a sassy response and karma
6. **Persistent Storage** - Keep track of user's karma scores using SQLite database
7. **Leaderboard** - Display the top 10 users with the highest karma scores
8. **Birthday Reminders** - Remember everyone's birthdays and send messages to #people channel
9. **Anniversary Tracking** - Remember work anniversaries and celebrate team milestones

## 🚀 Quick Start

### Prerequisites

- Go 1.24 or higher
- A Slack workspace where you have permission to install apps
- Basic knowledge of Slack app configuration

### Installation

1. **Clone the repository**

   ```bash
   git clone https://github.com/pratikgajjar/fambot-go.git
   cd fambot-go
   ```

2. **Install dependencies**

   ```bash
   go mod tidy
   ```

3. **Set up your Slack app** (see [Slack App Setup](#slack-app-setup) below)

4. **Configure environment variables**

   ```bash
   cp .env.example .env
   # Edit .env with your Slack tokens
   ```

5. **Run the bot**
   ```bash
   go run cmd/main.go
   ```

## 🔧 Slack App Setup

### 1. Create a Slack App

1. Go to [Slack API Apps](https://api.slack.com/apps)
2. Click "Create New App" → "From scratch"
3. Name your app (e.g., "FamBot") and select your workspace

### 2. Configure Bot Token Scopes

In your app settings, go to **OAuth & Permissions** and add these Bot Token Scopes:

```
app_mentions:read
channels:history
channels:read
chat:write
commands
groups:history
groups:read
im:history
im:read
mpim:history
mpim:read
users:read
users:read.email
```

### 3. Enable Socket Mode

1. Go to **Socket Mode** in your app settings
2. Enable Socket Mode
3. Generate an App-Level Token with `connections:write` scope
4. Save this token (starts with `xapp-`)

### 4. Enable Events API

1. Go to **Event Subscriptions**
2. Enable Events
3. Subscribe to these bot events:
   - `app_mention`
   - `message.channels`
   - `message.groups`
   - `message.im`
   - `message.mpim`

### 5. Add Slash Commands

Go to **Slash Commands** and create these commands:

| Command            | Description            | Usage Hint                   |
| ------------------ | ---------------------- | ---------------------------- |
| `/top-karma`       | Show karma leaderboard | Show top 10 karma holders    |
| `/my-karma`        | Check your karma       | See your current karma score |
| `/set-birthday`    | Set your birthday      | MM/DD or MM/DD/YYYY          |
| `/set-anniversary` | Set work anniversary   | MM/DD/YYYY                   |
| `/fambot-help`     | Show help message      | Get bot usage instructions   |

### 6. Install to Workspace

1. Go to **Install App** in your app settings
2. Click "Install to Workspace"
3. Copy the Bot User OAuth Token (starts with `xoxb-`)

### 7. Configure Environment

Update your `.env` file:

```env
SLACK_BOT_TOKEN=xoxb-your-bot-token-here
SLACK_APP_TOKEN=xapp-your-app-level-token-here
DATABASE_PATH=fambot.db
PEOPLE_CHANNEL=people
GRATEFUL_CHANNEL=thankyou
DEBUG=false
```

## 🎯 Usage

### Karma System

**Give karma to someone (works in any public channel):**

```
@username++ Great job on that presentation!
@alice++ Thanks for helping with the bug fix
```

_Bot responds in thread:_ "Karma level up! @username now has 15 karma points! 📈✨"

**Thank anyone (and get karma):**

```
Thank you @alice for the help!
Thanks everyone for the great meeting!
@bob thank you so much!
```

_Bot responds in thread and posts to grateful channel with thread link_

**Check karma leaderboard:**

```
/top-karma
```

or mention the bot:

```
@fambot show me the leaderboard
@fambot top karma
```

**Check your karma:**

```
/my-karma
```

### Birthday & Anniversary Management

**Set your birthday:**

```
/set-birthday 03/15
/set-birthday 03/15/1990
```

**Set your work anniversary:**

```
/set-anniversary 03/15/2020
```

### Getting Help

**Show help:**

```
/fambot-help
```

or mention the bot:

```
@fambot help
```

## 🗂️ Project Structure

```
fambot-go/
├── cmd/
│   └── main.go              # Application entry point
├── internal/
│   ├── config/
│   │   └── config.go        # Configuration management
│   ├── database/
│   │   └── database.go      # Database operations
│   ├── handlers/
│   │   └── slack.go         # Slack event handlers
│   └── models/
│       └── models.go        # Data models
├── .env.example             # Environment configuration template
├── go.mod                   # Go module definition
└── README.md               # This file
```

## 🎭 Sassy Responses

FamBot comes with built-in sassy responses that make interactions more fun. All responses now appear in message threads to keep channels organized:

**Thank you responses:**

- "Oh, you're being polite now? How refreshing! Here's some karma for good manners. 💫"
- "Look who remembered their manners! Take some karma, you well-behaved human. ✨"
- "Gratitude detected! Don't get used to this generosity though... 😏"

**Karma given responses:**

- "Karma delivered with a side of sass! You're welcome. 💅"
- "Another karma point hits the bank! Keep spreading those good vibes. 🏦"
- "Ding! Karma deposited. Your account is looking mighty fine! 💰"

## 📝 Grateful Channel Integration

When someone receives karma or thanks, FamBot automatically posts a summary to the configured grateful channel:

```
📝 @username received thanks! Check it out: [clickable thread link]
```

This feature helps teams:

- Track positive interactions across all channels
- Celebrate team members in a centralized location
- Build a culture of appreciation and recognition

Configure with the `GRATEFUL_CHANNEL` environment variable (defaults to "thankyou").

## 🕒 Automated Reminders

FamBot automatically sends reminders at 9 AM daily:

- **Birthday reminders** in the configured people channel
- **Work anniversary celebrations** with years of service

## 🔧 Development

### Running in Development

```bash
# Enable debug logging
export DEBUG=true

# Run with hot reload (requires air)
air

# Or run directly
go run cmd/main.go
```

### Database

The bot uses SQLite for data persistence. The database file will be created automatically when you first run the bot. Tables include:

- `users` - Slack user information
- `karma` - User karma scores
- `karma_log` - Individual karma transactions
- `birthdays` - User birthday information
- `anniversaries` - Work anniversary dates
- `sassy_responses` - Bot response messages

### Adding New Features

1. Add models to `internal/models/models.go`
2. Add database operations to `internal/database/database.go`
3. Add handlers to `internal/handlers/slack.go`
4. Update the main application if needed

## 🤝 Contributing

1. Fork the repository
2. Create a feature branch (`git checkout -b feature/amazing-feature`)
3. Commit your changes (`git commit -m 'Add some amazing feature'`)
4. Push to the branch (`git push origin feature/amazing-feature`)
5. Open a Pull Request

## 📝 License

This project is licensed under the MIT License - see the [LICENSE](LICENSE) file for details.

## 🆘 Troubleshooting

### Common Issues

**Bot doesn't respond to messages:**

- Ensure the bot is added to the channel
- Check that Event Subscriptions are properly configured
- Verify Socket Mode is enabled

**Slash commands don't work:**

- Make sure all slash commands are created in your Slack app
- Verify the app is installed to your workspace

**Database errors:**

- Ensure the bot has write permissions in the directory
- Check that SQLite is properly installed

**Permission errors:**

- Review the OAuth scopes in your Slack app
- Reinstall the app to your workspace if you added new scopes

### Debug Mode

Enable debug logging by setting `DEBUG=true` in your environment to see detailed logs of all Slack events and bot responses.

---

Made with ❤️ and a healthy dose of sass! 🤖✨
