package main

import (
	"fmt"
	"log"
	"os"

	"github.com/joho/godotenv"
	"github.com/slack-go/slack"
	"github.com/slack-go/slack/socketmode"
)

func main() {
	fmt.Println("🧪 Testing socketmode.New signatures")
	fmt.Println("===================================")

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

	// Create basic Slack client with bot token
	client := slack.New(botToken)
	fmt.Println("✅ Created Slack client with bot token")

	// Test different socketmode.New signatures
	fmt.Println("\n🔬 Testing socketmode.New signatures...")

	// Method 1: Try with app token as second parameter
	fmt.Println("Method 1: socketmode.New(client, appToken)")
	socketClient1, err := trySocketMode1(client, appToken)
	if err != nil {
		fmt.Printf("   ❌ Failed: %v\n", err)
	} else {
		fmt.Printf("   ✅ Success: %T\n", socketClient1)
	}

	// Method 2: Try with options only
	fmt.Println("Method 2: socketmode.New(client, options...)")
	socketClient2, err := trySocketMode2(client, appToken)
	if err != nil {
		fmt.Printf("   ❌ Failed: %v\n", err)
	} else {
		fmt.Printf("   ✅ Success: %T\n", socketClient2)
	}

	// Method 3: Try creating client differently
	fmt.Println("Method 3: slack.New with app token directly")
	err = trySocketMode3(appToken)
	if err != nil {
		fmt.Printf("   ❌ Failed: %v\n", err)
	} else {
		fmt.Printf("   ✅ Success\n")
	}

	fmt.Println("\n📚 Documentation check...")
	fmt.Println("Check these resources:")
	fmt.Println("- https://pkg.go.dev/github.com/slack-go/slack/socketmode")
	fmt.Println("- https://github.com/slack-go/slack/tree/master/examples")
	fmt.Println("- Search for 'socketmode.New' in slack-go examples")
}

func trySocketMode1(client *slack.Client, appToken string) (*socketmode.Client, error) {
	defer func() {
		if r := recover(); r != nil {
			fmt.Printf("   🚨 Panic recovered: %v\n", r)
		}
	}()

	// This might be the correct signature
	socketClient := socketmode.New(client)
	return socketClient, nil
}

func trySocketMode2(client *slack.Client, appToken string) (*socketmode.Client, error) {
	defer func() {
		if r := recover(); r != nil {
			fmt.Printf("   🚨 Panic recovered: %v\n", r)
		}
	}()

	// Try with token as an option (if such option exists)
	socketClient := socketmode.New(client, socketmode.OptionDebug(true))
	return socketClient, nil
}

func trySocketMode3(appToken string) error {
	defer func() {
		if r := recover(); r != nil {
			fmt.Printf("   🚨 Panic recovered: %v\n", r)
		}
	}()

	// Check if we need to create the client with app token
	appClient := slack.New(appToken)
	socketClient := socketmode.New(appClient)
	_ = socketClient
	return nil
}
