package main

import (
	"context"
	"log"
	"os"
	"os/signal"
	"strings"
	"syscall"

	"github.com/robfig/cron/v3"
	"github.com/slack-go/slack"
	"github.com/slack-go/slack/socketmode"

	"github.com/pratikgajjar/fambot-go/internal/config"
	"github.com/pratikgajjar/fambot-go/internal/database"
	"github.com/pratikgajjar/fambot-go/internal/handlers"
)

func main() {
	// Load configuration
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("Failed to load configuration: %v", err)
	}

	// Initialize database
	db, err := database.New(cfg.DatabasePath)
	if err != nil {
		log.Fatalf("Failed to initialize database: %v", err)
	}
	defer db.Close()

	// Validate tokens before proceeding
	if !strings.HasPrefix(cfg.SlackBotToken, "xoxb-") {
		log.Fatalf("SLACK_BOT_TOKEN should start with 'xoxb-', got: %s", cfg.SlackBotToken[:10]+"...")
	}
	if !strings.HasPrefix(cfg.SlackAppToken, "xapp-") {
		log.Fatalf("SLACK_APP_TOKEN should start with 'xapp-', got: %s", cfg.SlackAppToken[:10]+"...")
	}
	log.Printf("Token validation passed")

	// Initialize Slack client
	client := slack.New(
		cfg.SlackBotToken,
		slack.OptionDebug(cfg.Debug),
		slack.OptionLog(log.New(os.Stdout, "api: ", log.LstdFlags|log.Lshortfile)),
		slack.OptionAppLevelToken(cfg.SlackAppToken),
	)
	socketClient := socketmode.New(
		client,
		socketmode.OptionDebug(cfg.Debug),
		socketmode.OptionLog(log.New(os.Stdout, "socketmode: ", log.LstdFlags|log.Lshortfile)),
	)

	// Get bot user info
	authTest, err := client.AuthTest()
	if err != nil {
		log.Fatalf("Failed to authenticate bot: %v", err)
	}
	log.Printf("Bot authenticated as %s (%s)", authTest.User, authTest.UserID)

	// Initialize handlers
	handler := handlers.New(client, db, cfg.PeopleChannel)
	handler.SetBotID(authTest.UserID)

	// Set up socket mode event handler
	go func() {
		for evt := range socketClient.Events {
			handler.HandleSocketModeEvent(evt, socketClient)
		}
	}()

	// Set up cron jobs for birthday and anniversary reminders
	c := cron.New()

	// Check for birthdays and anniversaries daily at 9 AM
	_, err = c.AddFunc("0 9 * * *", func() {
		log.Println("Running daily birthday check...")
		handler.SendBirthdayReminder()
	})
	if err != nil {
		log.Printf("Failed to add birthday cron job: %v", err)
	}

	_, err = c.AddFunc("0 9 * * *", func() {
		log.Println("Running daily anniversary check...")
		handler.SendAnniversaryReminder()
	})
	if err != nil {
		log.Printf("Failed to add anniversary cron job: %v", err)
	}

	// Start cron scheduler
	c.Start()
	defer c.Stop()

	// Set up graceful shutdown
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Handle interrupt signals
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM)

	// Start socket mode client in a goroutine
	go func() {
		log.Println("Starting FamBot...")
		err := socketClient.RunContext(ctx)
		if err != nil {
			log.Printf("Socket mode client error: %v", err)
		}
	}()

	// Wait for interrupt signal
	<-sigChan
	log.Println("Shutting down FamBot...")
	cancel()
}
