package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"strings"
	"time"

	"github.com/joho/godotenv"
	"github.com/slack-go/slack"
	"github.com/slack-go/slack/socketmode"
)

func main() {
	fmt.Println("🔍 FamBot Token Validation")
	fmt.Println("==========================")
	fmt.Println()

	// Load environment variables
	if err := godotenv.Load(); err != nil {
		fmt.Println("⚠️  Warning: Could not load .env file")
	}

	botToken := os.Getenv("SLACK_BOT_TOKEN")
	appToken := os.Getenv("SLACK_APP_TOKEN")

	if botToken == "" {
		log.Fatal("❌ SLACK_BOT_TOKEN is required")
	}
	if appToken == "" {
		log.Fatal("❌ SLACK_APP_TOKEN is required")
	}

	fmt.Printf("🤖 Bot Token: %s...\n", botToken[:20])
	fmt.Printf("📱 App Token: %s...\n", appToken[:20])
	fmt.Println()

	// Validate token formats
	validateTokenFormats(botToken, appToken)

	// Test bot token
	fmt.Println("🧪 Testing Bot Token...")
	client := slack.New(botToken, slack.OptionDebug(true))

	if err := testBotToken(client); err != nil {
		log.Fatalf("❌ Bot token validation failed: %v", err)
	}
	fmt.Println("✅ Bot token is valid!")
	fmt.Println()

	// Test app token with socket mode
	fmt.Println("🧪 Testing App Token with Socket Mode...")
	if err := testAppToken(client, appToken); err != nil {
		log.Fatalf("❌ App token validation failed: %v", err)
	}
	fmt.Println("✅ App token is valid!")
	fmt.Println()

	fmt.Println("🎉 All tokens are valid! Your configuration should work.")
}

func validateTokenFormats(botToken, appToken string) {
	fmt.Println("📋 Validating Token Formats...")

	// Check bot token format
	if !strings.HasPrefix(botToken, "xoxb-") {
		log.Fatalf("❌ Bot token should start with 'xoxb-', got: %s", botToken[:10])
	}
	fmt.Println("✅ Bot token format is correct")

	// Check app token format
	if !strings.HasPrefix(appToken, "xapp-") {
		log.Fatalf("❌ App token should start with 'xapp-', got: %s", appToken[:10])
	}
	fmt.Println("✅ App token format is correct")
	fmt.Println()
}

func testBotToken(client *slack.Client) error {
	// Test authentication
	authTest, err := client.AuthTest()
	if err != nil {
		return fmt.Errorf("auth.test failed: %w", err)
	}

	fmt.Printf("   Bot User: %s (%s)\n", authTest.User, authTest.UserID)
	fmt.Printf("   Team: %s (%s)\n", authTest.Team, authTest.TeamID)
	fmt.Printf("   URL: %s\n", authTest.URL)

	// Test basic API call
	_, err = client.GetUsers()
	if err != nil {
		return fmt.Errorf("users.list failed (check scopes): %w", err)
	}
	fmt.Println("   ✅ Basic API calls work")

	return nil
}

func testAppToken(client *slack.Client, appToken string) error {
	// Create socket mode client
	socketClient := socketmode.New(client, socketmode.OptionDebug(true))

	// Test socket mode connection with timeout
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	// Channel to capture connection result
	connResult := make(chan error, 1)

	// Start socket mode client in goroutine
	go func() {
		fmt.Println("   🔌 Attempting Socket Mode connection...")
		err := socketClient.RunContext(ctx)
		connResult <- err
	}()

	// Set up event handler to detect successful connection
	connectionEstablished := make(chan bool, 1)

	go func() {
		for evt := range socketClient.Events {
			switch evt.Type {
			case socketmode.EventTypeConnecting:
				fmt.Println("   🔄 Connecting to Socket Mode...")
			case socketmode.EventTypeConnectionError:
				fmt.Printf("   ❌ Connection error: %v\n", evt.Data)
				connectionEstablished <- false
				return
			case socketmode.EventTypeConnected:
				fmt.Println("   ✅ Socket Mode connected successfully!")
				connectionEstablished <- true
				return
			case socketmode.EventTypeInvalidAuth:
				fmt.Println("   ❌ Invalid authentication for Socket Mode")
				connectionEstablished <- false
				return
			}
		}
	}()

	// Wait for connection result or timeout
	select {
	case success := <-connectionEstablished:
		if success {
			// Give it a moment to establish fully
			time.Sleep(2 * time.Second)
			return nil
		} else {
			return fmt.Errorf("socket mode connection failed - check app-level token and Socket Mode settings")
		}
	case err := <-connResult:
		if err != nil {
			return fmt.Errorf("socket mode client error: %w", err)
		}
		return fmt.Errorf("socket mode connection ended unexpectedly")
	case <-ctx.Done():
		return fmt.Errorf("socket mode connection timed out - this usually means:\n" +
			"   1. Socket Mode is not enabled in your Slack app\n" +
			"   2. App-level token is invalid or missing 'connections:write' scope\n" +
			"   3. App is not properly installed to workspace\n" +
			"   4. Network connectivity issues")
	}
}
