package handlers

import (
	"fmt"
	"log"
	"math/rand"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/slack-go/slack"
	"github.com/slack-go/slack/slackevents"
	"github.com/slack-go/slack/socketmode"

	"github.com/pratikgajjar/fambot-go/internal/database"
	"github.com/pratikgajjar/fambot-go/internal/models"
)

var (
	karmaRegex    = regexp.MustCompile(`<@([A-Z0-9]+)>\s*\+\+`)
	thankYouRegex = regexp.MustCompile(`(?i)\b(thank\s*(you|u)|thanks|thx|ty)\b`)
)

// SlackHandler handles all Slack-related events and interactions
type SlackHandler struct {
	client          *slack.Client
	db              *database.Database
	botID           string
	peopleChannel   string
	gratefulChannel string
	workspaceID     string
}

// New creates a new SlackHandler
func New(client *slack.Client, db *database.Database, peopleChannel, gratefulChannel string) *SlackHandler {
	return &SlackHandler{
		client:          client,
		db:              db,
		peopleChannel:   peopleChannel,
		gratefulChannel: gratefulChannel,
	}
}

// SetBotID sets the bot's user ID
func (h *SlackHandler) SetBotID(botID string) {
	h.botID = botID
}

// SetWorkspaceID sets the workspace ID for thread link generation
func (h *SlackHandler) SetWorkspaceID(workspaceID string) {
	h.workspaceID = workspaceID
}

// HandleSocketModeEvent handles incoming socket mode events
func (h *SlackHandler) HandleSocketModeEvent(evt socketmode.Event, client *socketmode.Client) {
	switch evt.Type {
	case socketmode.EventTypeConnecting:
		log.Println("Connecting to Slack...")
	case socketmode.EventTypeConnected:
		log.Println("Connected to Slack!")
	case socketmode.EventTypeEventsAPI:
		eventsAPIEvent, ok := evt.Data.(slackevents.EventsAPIEvent)
		if !ok {
			log.Printf("Ignored %+v\n", evt)
			return
		}

		client.Ack(*evt.Request)
		h.handleEventsAPI(eventsAPIEvent)

	case socketmode.EventTypeSlashCommand:
		cmd, ok := evt.Data.(slack.SlashCommand)
		if !ok {
			log.Printf("Ignored %+v\n", evt)
			return
		}

		client.Ack(*evt.Request)
		h.handleSlashCommand(cmd)

	default:
		log.Printf("Ignored event type: %s\n", evt.Type)
	}
}

// handleEventsAPI handles Events API events
func (h *SlackHandler) handleEventsAPI(event slackevents.EventsAPIEvent) {
	switch event.Type {
	case slackevents.CallbackEvent:
		innerEvent := event.InnerEvent
		switch ev := innerEvent.Data.(type) {
		case *slackevents.MessageEvent:
			h.handleMessage(ev)
		case *slackevents.AppMentionEvent:
			h.handleAppMention(ev)
		}
	default:
		log.Printf("Unsupported Events API event received: %v\n", event.Type)
	}
}

// handleMessage handles regular message events
func (h *SlackHandler) handleMessage(event *slackevents.MessageEvent) {
	// Skip bot messages and message subtypes we don't care about
	if event.User == h.botID || event.SubType != "" {
		return
	}

	// Handle karma increments
	h.handleKarmaIncrements(event)

	// Handle thank you responses
	h.handleThankYou(event)
}

// handleAppMention handles app mention events
func (h *SlackHandler) handleAppMention(event *slackevents.AppMentionEvent) {
	// Skip bot messages
	if event.User == h.botID {
		return
	}

	text := strings.ToLower(event.Text)

	if strings.Contains(text, "top") || strings.Contains(text, "leaderboard") {
		h.sendTopKarma(event.Channel)
	} else if strings.Contains(text, "help") {
		h.sendHelp(event.Channel)
	} else {
		// Default sassy response
		responses := []string{
			"You mentioned me! How can I sass... I mean, help you today? 😏",
			"Yes, your majesty? What do you require of this humble bot? 👑",
			"Oh, you need me? I'm flattered! What's up? 💫",
			"*clears digital throat* You rang? 🔔",
			"At your service! Though my service comes with a side of sass. 💅",
		}
		response := responses[rand.Intn(len(responses))]
		h.sendMessage(event.Channel, response)
	}
}

// handleKarmaIncrements processes karma increment patterns
func (h *SlackHandler) handleKarmaIncrements(event *slackevents.MessageEvent) {
	matches := karmaRegex.FindAllStringSubmatch(event.Text, -1)

	for _, match := range matches {
		if len(match) < 2 {
			continue
		}

		targetUserID := match[1]

		// Don't allow self-karma
		if targetUserID == event.User {
			h.sendThreadedMessage(event.Channel, event.TimeStamp, "Nice try! You can't give karma to yourself. That's cheating! 🚫")
			continue
		}

		// Don't allow karma to the bot
		if targetUserID == h.botID {
			h.sendThreadedMessage(event.Channel, event.TimeStamp, "Aww, trying to give me karma? I'm touched, but I'm already perfect! 😎")
			continue
		}

		// Get user info
		userInfo, err := h.client.GetUserInfo(targetUserID)
		if err != nil {
			log.Printf("Error getting user info for %s: %v", targetUserID, err)
			continue
		}

		// Store/update user in database
		user := &models.User{
			ID:       userInfo.ID,
			Username: userInfo.Name,
			RealName: userInfo.RealName,
			Email:    userInfo.Profile.Email,
		}
		h.db.UpsertUser(user)

		// Increment karma
		reason := fmt.Sprintf("Karma given in #%s", getChannelName(event.Channel))
		err = h.db.IncrementKarma(targetUserID, userInfo.Name, event.User, reason, event.Channel)
		if err != nil {
			log.Printf("Error incrementing karma: %v", err)
			h.sendThreadedMessage(event.Channel, event.TimeStamp, "Oops! Something went wrong with the karma system. 🤖💥")
			continue
		}

		// Get karma count
		karma, err := h.db.GetKarma(targetUserID)
		if err != nil {
			log.Printf("Error getting karma: %v", err)
		}

		// Send sassy response in thread
		var response string
		if karma != nil {
			response = fmt.Sprintf("Karma level up! <@%s> now has %d karma points! 📈✨", targetUserID, karma.Score)
		} else {
			response = fmt.Sprintf("Karma delivered to <@%s>! 💫", targetUserID)
		}

		// Add a random sassy comment
		sassyResponse, err := h.db.GetRandomSassyResponse("karma_given")
		if err == nil {
			response += "\n" + sassyResponse.Response
		}

		h.sendThreadedMessage(event.Channel, event.TimeStamp, response)

		// Post to grateful channel with thread link
		h.postToGratefulChannel(targetUserID, event.Channel, event.TimeStamp)
	}
}

// handleThankYou processes thank you mentions
func (h *SlackHandler) handleThankYou(event *slackevents.MessageEvent) {
	// Check if the message contains "thank you"
	if !thankYouRegex.MatchString(event.Text) {
		return
	}

	// Get user info for the person saying thanks
	userInfo, err := h.client.GetUserInfo(event.User)
	if err != nil {
		log.Printf("Error getting user info for %s: %v", event.User, err)
		return
	}

	// Store/update user in database
	user := &models.User{
		ID:       userInfo.ID,
		Username: userInfo.Name,
		RealName: userInfo.RealName,
		Email:    userInfo.Profile.Email,
	}
	h.db.UpsertUser(user)

	// Check if someone specific is being thanked (has user mentions)
	var targetUsername string
	userMentionRegex := regexp.MustCompile(`<@([A-Z0-9]+)>`)
	mentions := userMentionRegex.FindAllStringSubmatch(event.Text, -1)

	// If there are user mentions, find who is being thanked
	for _, match := range mentions {
		if len(match) >= 2 {
			mentionedUserID := match[1]
			if mentionedUserID != h.botID && mentionedUserID != event.User {
				// Someone is thanking another user
				mentionedUser, err := h.client.GetUserInfo(mentionedUserID)
				if err == nil {
					targetUsername = mentionedUser.Name
					break
				}
			}
		}
	}

	// Give karma for being polite
	reason := fmt.Sprintf("Said thank you in #%s", getChannelName(event.Channel))
	err = h.db.IncrementKarma(event.User, userInfo.Name, h.botID, reason, event.Channel)
	if err != nil {
		log.Printf("Error incrementing karma for thank you: %v", err)
	}

	// Send sassy thank you response in thread
	sassyResponse, err := h.db.GetRandomSassyResponse("thank_you")
	var response string
	if err != nil {
		// Fallback response
		response = fmt.Sprintf("Politeness detected! <@%s> gets karma for good manners! ✨", event.User)
	} else {
		response = fmt.Sprintf("<@%s> %s", event.User, sassyResponse.Response)
	}

	h.sendThreadedMessage(event.Channel, event.TimeStamp, response)

	// Post to grateful channel with thread link only if someone specific was thanked
	if targetUsername != "" {
		// Find the user ID for the mentioned user
		gratefulUserID := ""
		for _, match := range mentions {
			if len(match) >= 2 {
				mentionedUserID := match[1]
				if mentionedUserID != h.botID && mentionedUserID != event.User {
					gratefulUserID = mentionedUserID
					break
				}
			}
		}
		if gratefulUserID != "" {
			h.postToGratefulChannel(gratefulUserID, event.Channel, event.TimeStamp)
		}
	}
}

// handleSlashCommand handles slash commands
func (h *SlackHandler) handleSlashCommand(cmd slack.SlashCommand) {
	switch cmd.Command {
	case "/top-karma":
		h.handleTopKarmaCommand(cmd)
	case "/set-birthday":
		h.handleSetBirthdayCommand(cmd)
	case "/set-anniversary":
		h.handleSetAnniversaryCommand(cmd)
	case "/my-karma":
		h.handleMyKarmaCommand(cmd)
	case "/fambot-help":
		h.handleHelpCommand(cmd)
	default:
		h.respondToSlashCommand(cmd, "Unknown command! Use `/fambot-help` to see available commands.")
	}
}

// handleTopKarmaCommand handles the /top-karma slash command
func (h *SlackHandler) handleTopKarmaCommand(cmd slack.SlashCommand) {
	karmas, err := h.db.GetTopKarma(10)
	if err != nil {
		h.respondToSlashCommand(cmd, "Error retrieving karma leaderboard! 😅")
		return
	}

	if len(karmas) == 0 {
		h.respondToSlashCommand(cmd, "No karma recorded yet! Be the first to spread some love with @username++ 💫")
		return
	}

	response := "🏆 *Karma Leaderboard* 🏆\n\n"
	emojis := []string{"🥇", "🥈", "🥉", "4️⃣", "5️⃣", "6️⃣", "7️⃣", "8️⃣", "9️⃣", "🔟"}

	for i, karma := range karmas {
		emoji := emojis[i]
		response += fmt.Sprintf("%s <@%s> - %d karma\n", emoji, karma.UserID, karma.Score)
	}

	response += "\nKeep spreading those good vibes! ✨"
	h.respondToSlashCommand(cmd, response)
}

// handleMyKarmaCommand handles the /my-karma slash command
func (h *SlackHandler) handleMyKarmaCommand(cmd slack.SlashCommand) {
	karma, err := h.db.GetKarma(cmd.UserID)
	if err != nil {
		h.respondToSlashCommand(cmd, "You don't have any karma yet! Start being awesome and ask someone to give you some with @username++ 😊")
		return
	}

	response := fmt.Sprintf("Your karma: *%d points* ✨\n", karma.Score)
	response += "Keep being awesome! 💫"
	h.respondToSlashCommand(cmd, response)
}

// handleSetBirthdayCommand handles the /set-birthday slash command
func (h *SlackHandler) handleSetBirthdayCommand(cmd slack.SlashCommand) {
	if cmd.Text == "" {
		h.respondToSlashCommand(cmd, "Please provide your birthday in format: MM/DD or MM/DD/YYYY\nExample: `/set-birthday 03/15` or `/set-birthday 03/15/1990`")
		return
	}

	parts := strings.Split(strings.TrimSpace(cmd.Text), "/")
	if len(parts) < 2 || len(parts) > 3 {
		h.respondToSlashCommand(cmd, "Invalid format! Use MM/DD or MM/DD/YYYY\nExample: `/set-birthday 03/15` or `/set-birthday 03/15/1990`")
		return
	}

	month, err := strconv.Atoi(parts[0])
	if err != nil || month < 1 || month > 12 {
		h.respondToSlashCommand(cmd, "Invalid month! Please use MM/DD format.")
		return
	}

	day, err := strconv.Atoi(parts[1])
	if err != nil || day < 1 || day > 31 {
		h.respondToSlashCommand(cmd, "Invalid day! Please use MM/DD format.")
		return
	}

	year := 0
	if len(parts) == 3 {
		year, err = strconv.Atoi(parts[2])
		if err != nil || year < 1900 || year > time.Now().Year() {
			h.respondToSlashCommand(cmd, "Invalid year! Please use a valid year.")
			return
		}
	}

	// Get user info
	userInfo, err := h.client.GetUserInfo(cmd.UserID)
	if err != nil {
		h.respondToSlashCommand(cmd, "Error getting your user info! 😅")
		return
	}

	birthday := &models.Birthday{
		UserID:   cmd.UserID,
		Username: userInfo.Name,
		Month:    month,
		Day:      day,
		Year:     year,
		Timezone: "UTC", // Default to UTC for now
	}

	err = h.db.SetBirthday(birthday)
	if err != nil {
		h.respondToSlashCommand(cmd, "Error saving your birthday! 😅")
		return
	}

	dateStr := fmt.Sprintf("%02d/%02d", month, day)
	if year > 0 {
		dateStr = fmt.Sprintf("%02d/%02d/%d", month, day, year)
	}

	h.respondToSlashCommand(cmd, fmt.Sprintf("🎂 Birthday saved! I'll wish you happy birthday on %s! 🎉", dateStr))
}

// handleSetAnniversaryCommand handles the /set-anniversary slash command
func (h *SlackHandler) handleSetAnniversaryCommand(cmd slack.SlashCommand) {
	if cmd.Text == "" {
		h.respondToSlashCommand(cmd, "Please provide your work anniversary in format: MM/DD/YYYY\nExample: `/set-anniversary 03/15/2020`")
		return
	}

	parts := strings.Split(strings.TrimSpace(cmd.Text), "/")
	if len(parts) != 3 {
		h.respondToSlashCommand(cmd, "Invalid format! Use MM/DD/YYYY\nExample: `/set-anniversary 03/15/2020`")
		return
	}

	month, err := strconv.Atoi(parts[0])
	if err != nil || month < 1 || month > 12 {
		h.respondToSlashCommand(cmd, "Invalid month! Please use MM/DD/YYYY format.")
		return
	}

	day, err := strconv.Atoi(parts[1])
	if err != nil || day < 1 || day > 31 {
		h.respondToSlashCommand(cmd, "Invalid day! Please use MM/DD/YYYY format.")
		return
	}

	year, err := strconv.Atoi(parts[2])
	if err != nil || year < 1900 || year > time.Now().Year() {
		h.respondToSlashCommand(cmd, "Invalid year! Please use a valid year.")
		return
	}

	// Get user info
	userInfo, err := h.client.GetUserInfo(cmd.UserID)
	if err != nil {
		h.respondToSlashCommand(cmd, "Error getting your user info! 😅")
		return
	}

	anniversary := &models.Anniversary{
		UserID:   cmd.UserID,
		Username: userInfo.Name,
		Month:    month,
		Day:      day,
		Year:     year,
		Timezone: "UTC", // Default to UTC for now
	}

	err = h.db.SetAnniversary(anniversary)
	if err != nil {
		h.respondToSlashCommand(cmd, "Error saving your anniversary! 😅")
		return
	}

	yearsWorked := time.Now().Year() - year
	dateStr := fmt.Sprintf("%02d/%02d/%d", month, day, year)

	h.respondToSlashCommand(cmd, fmt.Sprintf("🎉 Work anniversary saved! You've been here for %d years as of %s! 🎊", yearsWorked, dateStr))
}

// handleHelpCommand handles the /fambot-help slash command
func (h *SlackHandler) handleHelpCommand(cmd slack.SlashCommand) {
	help := `🤖 *FamBot Help* 🤖

*Karma System:*
• Give karma: ` + "`@username++`" + ` - Give someone karma points
• Thank me: Mention me with "thank you" and get karma!
• ` + "`/my-karma`" + ` - Check your karma score
• ` + "`/top-karma`" + ` - See the karma leaderboard

*Birthdays & Anniversaries:*
• ` + "`/set-birthday MM/DD`" + ` or ` + "`/set-birthday MM/DD/YYYY`" + ` - Set your birthday
• ` + "`/set-anniversary MM/DD/YYYY`" + ` - Set your work anniversary

*Other:*
• Mention me for a sassy response!
• ` + "`/fambot-help`" + ` - Show this help message

I'm a sassy bot with a heart of gold! 💫✨`

	h.respondToSlashCommand(cmd, help)
}

// SendBirthdayReminder sends birthday reminders to the people channel
func (h *SlackHandler) SendBirthdayReminder() {
	birthdays, err := h.db.GetTodaysBirthdays()
	if err != nil {
		log.Printf("Error getting today's birthdays: %v", err)
		return
	}

	for _, birthday := range birthdays {
		var message string
		if birthday.Year > 1970 {
			age := time.Now().Year() - birthday.Year
			message = fmt.Sprintf("🎂 Happy Birthday <@%s>! 🎉\nAnother year older, another year wiser! Hope your %d%s year is absolutely amazing! 🎊✨",
				birthday.UserID, age, getOrdinalSuffix(age))
		} else {
			message = fmt.Sprintf("🎂 Happy Birthday <@%s>! 🎉\nHope your special day is filled with joy, laughter, and maybe some cake! 🎊✨",
				birthday.UserID)
		}

		h.sendMessage(h.peopleChannel, message)
	}
}

// SendAnniversaryReminder sends anniversary reminders to the people channel
func (h *SlackHandler) SendAnniversaryReminder() {
	anniversaries, err := h.db.GetTodaysAnniversaries()
	if err != nil {
		log.Printf("Error getting today's anniversaries: %v", err)
		return
	}

	for _, anniversary := range anniversaries {
		yearsWorked := time.Now().Year() - anniversary.Year
		message := fmt.Sprintf("🎉 Happy Work Anniversary <@%s>! 🎊\n%d years of awesomeness! Thanks for being part of our amazing team! 🚀✨",
			anniversary.UserID, yearsWorked)

		h.sendMessage(h.peopleChannel, message)
	}
}

// Helper methods
func (h *SlackHandler) sendMessage(channel, text string) {
	_, _, err := h.client.PostMessage(channel, slack.MsgOptionText(text, false))
	if err != nil {
		log.Printf("Error sending message: %v", err)
	}
}

// sendThreadedMessage sends a message as a reply in a thread
func (h *SlackHandler) sendThreadedMessage(channel, threadTS, text string) {
	_, _, err := h.client.PostMessage(channel,
		slack.MsgOptionText(text, false),
		slack.MsgOptionTS(threadTS))
	if err != nil {
		log.Printf("Error sending threaded message: %v", err)
	}
}

// postToGratefulChannel posts a thread link to the grateful channel
func (h *SlackHandler) postToGratefulChannel(userID, originalChannel, threadTS string) {
	// Skip if grateful channel is not configured
	if h.gratefulChannel == "" {
		return
	}

	// Get grateful channel ID by name
	gratefulChannelID, err := h.getChannelIDByName(h.gratefulChannel)
	if err != nil {
		log.Printf("Error getting grateful channel ID: %v", err)
		return
	}

	// Build the thread link using Slack's permalink format
	threadLink := fmt.Sprintf("https://slack.com/archives/%s/p%s", originalChannel, strings.Replace(threadTS, ".", "", 1))

	// Format the message with proper user tagging and Slack hyperlink format
	message := fmt.Sprintf("<@%s> received <%s|thanks>!", userID, threadLink)

	// Send to grateful channel
	h.sendMessage(gratefulChannelID, message)
}

// getChannelIDByName resolves a channel name to its ID
func (h *SlackHandler) getChannelIDByName(channelName string) (string, error) {
	// If it's already a channel ID (starts with C), return as-is
	if strings.HasPrefix(channelName, "C") {
		return channelName, nil
	}

	// Remove # prefix if present
	channelName = strings.TrimPrefix(channelName, "#")

	// Get list of channels
	channels, _, err := h.client.GetConversationsForUser(&slack.GetConversationsForUserParameters{
		Types: []string{"public_channel"},
		Limit: 1000,
	})
	if err != nil {
		return "", fmt.Errorf("failed to get channels: %w", err)
	}

	// Find channel by name
	for _, channel := range channels {
		if channel.Name == channelName {
			return channel.ID, nil
		}
	}

	return "", fmt.Errorf("channel #%s not found", channelName)
}

func (h *SlackHandler) sendTopKarma(channel string) {
	karmas, err := h.db.GetTopKarma(10)
	if err != nil {
		h.sendMessage(channel, "Error retrieving karma leaderboard! 😅")
		return
	}

	if len(karmas) == 0 {
		h.sendMessage(channel, "No karma recorded yet! Be the first to spread some love with @username++ 💫")
		return
	}

	response := "🏆 *Karma Leaderboard* 🏆\n\n"
	emojis := []string{"🥇", "🥈", "🥉", "4️⃣", "5️⃣", "6️⃣", "7️⃣", "8️⃣", "9️⃣", "🔟"}

	for i, karma := range karmas {
		emoji := emojis[i]
		response += fmt.Sprintf("%s <@%s> - %d karma\n", emoji, karma.UserID, karma.Score)
	}

	response += "\nKeep spreading those good vibes! ✨"
	h.sendMessage(channel, response)
}

func (h *SlackHandler) sendHelp(channel string) {
	help := `🤖 *FamBot Help* 🤖

*Karma System:*
• Give karma: ` + "`@username++`" + ` - Give someone karma points
• Thank me: Mention me with "thank you" and get karma!
• Ask for leaderboard: mention me with "top" or "leaderboard"

*Commands:*
• ` + "`/my-karma`" + ` - Check your karma score
• ` + "`/top-karma`" + ` - See the karma leaderboard
• ` + "`/set-birthday MM/DD`" + ` - Set your birthday
• ` + "`/set-anniversary MM/DD/YYYY`" + ` - Set your work anniversary
• ` + "`/fambot-help`" + ` - Show detailed help

I'm here to spread good vibes and sass! 💫✨`

	h.sendMessage(channel, help)
}

func (h *SlackHandler) respondToSlashCommand(cmd slack.SlashCommand, text string) {
	_, _, err := h.client.PostMessage(cmd.ChannelID, slack.MsgOptionText(text, false))
	if err != nil {
		log.Printf("Error responding to slash command: %v", err)
	}
}

// Utility functions
func getChannelName(channelID string) string {
	// This is a simplified version. In a real implementation,
	// you might want to cache channel names or fetch them from Slack API
	return channelID
}

func getOrdinalSuffix(n int) string {
	if n%100 >= 11 && n%100 <= 13 {
		return "th"
	}
	switch n % 10 {
	case 1:
		return "st"
	case 2:
		return "nd"
	case 3:
		return "rd"
	default:
		return "th"
	}
}
