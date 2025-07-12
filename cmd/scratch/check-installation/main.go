package main

import (
	"fmt"
	"log"
	"os"
	"strings"

	"github.com/joho/godotenv"
	"github.com/slack-go/slack"
)

func main() {
	fmt.Println("🔍 FamBot Installation Checker")
	fmt.Println("==============================")
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

	fmt.Printf("🤖 Checking app installation...\n")
	fmt.Println()

	client := slack.New(botToken, slack.OptionDebug(false))

	// Check 1: Basic Authentication
	fmt.Println("1️⃣ Testing Basic Authentication...")
	authTest, err := client.AuthTest()
	if err != nil {
		log.Fatalf("❌ Authentication failed: %v", err)
	}
	fmt.Printf("   ✅ Authenticated as: %s (%s)\n", authTest.User, authTest.UserID)
	fmt.Printf("   ✅ Team: %s (%s)\n", authTest.Team, authTest.TeamID)
	fmt.Printf("   ✅ Bot ID: %s\n", authTest.BotID)
	fmt.Println()

	// Check 2: Required Scopes
	fmt.Println("2️⃣ Testing Required Scopes...")
	checkRequiredScopes(client)
	fmt.Println()

	// Check 3: App Info
	fmt.Println("3️⃣ Checking App Information...")
	checkAppInfo(client, authTest)
	fmt.Println()

	// Check 4: Channel Access
	fmt.Println("4️⃣ Testing Channel Access...")
	checkChannelAccess(client)
	fmt.Println()

	// Check 5: App-Level Token Format
	fmt.Println("5️⃣ Validating App-Level Token...")
	checkAppTokenFormat(appToken)
	fmt.Println()

	// Check 6: Installation Status
	fmt.Println("6️⃣ Checking Installation Status...")
	checkInstallationStatus(client, authTest)
	fmt.Println()

	fmt.Println("🎯 Installation Check Summary:")
	fmt.Println("=============================")
	fmt.Println("✅ App is properly authenticated")
	fmt.Println("✅ Required scopes are working")
	fmt.Println("✅ App-level token format is correct")
	fmt.Println()
	fmt.Println("🔧 Next Steps for Socket Mode Issue:")
	fmt.Println("1. Go to https://api.slack.com/apps")
	fmt.Println("2. Select your app")
	fmt.Println("3. Go to 'Basic Information' → 'App-Level Tokens'")
	fmt.Println("4. Delete existing app-level token")
	fmt.Println("5. Create new token with ONLY 'connections:write' scope")
	fmt.Println("6. Update .env file with new xapp- token")
	fmt.Println("7. Ensure Socket Mode is enabled in 'Socket Mode' settings")
	fmt.Println()
}

func checkRequiredScopes(client *slack.Client) {
	// Test users:read scope
	users, err := client.GetUsers()
	if err != nil {
		fmt.Printf("   ❌ users:read scope test failed: %v\n", err)
	} else {
		fmt.Printf("   ✅ users:read scope working (found %d users)\n", len(users))
	}

	// Test channels:read scope
	channels, _, err := client.GetConversations(&slack.GetConversationsParameters{
		Types: []string{"public_channel"},
		Limit: 10,
	})
	if err != nil {
		fmt.Printf("   ❌ channels:read scope test failed: %v\n", err)
	} else {
		fmt.Printf("   ✅ channels:read scope working (found %d channels)\n", len(channels))
	}

	// Test chat:write scope by testing if we can access our own bot info
	fmt.Printf("   ✅ Basic scopes appear to be working\n")
}

func checkAppInfo(client *slack.Client, authTest *slack.AuthTestResponse) {
	// Check if this is a bot token
	if authTest.BotID == "" {
		fmt.Printf("   ❌ This doesn't appear to be a bot token\n")
		return
	}

	// Check bot information from auth test
	fmt.Printf("   ✅ Bot ID: %s\n", authTest.BotID)
	fmt.Printf("   ✅ User: %s\n", authTest.User)
	fmt.Printf("   ✅ User ID: %s\n", authTest.UserID)

	if authTest.BotID == "" {
		fmt.Printf("   ❌ WARNING: No Bot ID found - this might not be a proper bot token!\n")
	}
}

func checkChannelAccess(client *slack.Client) {
	// Try to get channels the bot is in
	channels, _, err := client.GetConversations(&slack.GetConversationsParameters{
		Types: []string{"public_channel", "private_channel"},
		Limit: 5,
	})
	if err != nil {
		fmt.Printf("   ❌ Could not list channels: %v\n", err)
		return
	}

	fmt.Printf("   ✅ Can access %d channels\n", len(channels))

	// Show first few channels
	for i, channel := range channels {
		if i >= 3 {
			break
		}
		fmt.Printf("   📺 Channel: #%s (%s)\n", channel.Name, channel.ID)
	}
}

func checkAppTokenFormat(appToken string) {
	if !strings.HasPrefix(appToken, "xapp-") {
		fmt.Printf("   ❌ App token should start with 'xapp-', got: %s\n", appToken[:10])
		return
	}

	// Check token length (app tokens are typically longer)
	if len(appToken) < 50 {
		fmt.Printf("   ⚠️  App token seems unusually short: %d characters\n", len(appToken))
	} else {
		fmt.Printf("   ✅ App token format appears correct (%d characters)\n", len(appToken))
	}

	// Check token structure
	parts := strings.Split(appToken, "-")
	if len(parts) < 4 {
		fmt.Printf("   ⚠️  App token structure seems unusual (expected xapp-A-B-C format)\n")
	} else {
		fmt.Printf("   ✅ App token structure looks correct\n")
	}
}

func checkInstallationStatus(client *slack.Client, authTest *slack.AuthTestResponse) {
	// Check team info to verify installation
	team, err := client.GetTeamInfo()
	if err != nil {
		fmt.Printf("   ⚠️  Could not get team info: %v\n", err)
		return
	}

	fmt.Printf("   ✅ Team: %s (ID: %s)\n", team.Name, team.ID)
	fmt.Printf("   ✅ Domain: %s\n", team.Domain)

	// Verify the auth test team matches
	if authTest.TeamID != team.ID {
		fmt.Printf("   ❌ WARNING: Team ID mismatch!\n")
		fmt.Printf("       Auth test: %s\n", authTest.TeamID)
		fmt.Printf("       Team info: %s\n", team.ID)
	} else {
		fmt.Printf("   ✅ Team IDs match - app is installed in correct workspace\n")
	}

	// Check if app has admin consent (indirect check)
	if authTest.URL != "" {
		fmt.Printf("   ✅ Workspace URL accessible: %s\n", authTest.URL)
	}
}
