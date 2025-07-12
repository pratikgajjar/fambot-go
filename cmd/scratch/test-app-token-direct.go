package main

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/joho/godotenv"
)

type AuthTestResponse struct {
	Ok     bool   `json:"ok"`
	Error  string `json:"error,omitempty"`
	URL    string `json:"url,omitempty"`
	Team   string `json:"team,omitempty"`
	User   string `json:"user,omitempty"`
	TeamID string `json:"team_id,omitempty"`
	UserID string `json:"user_id,omitempty"`
	BotID  string `json:"bot_id,omitempty"`
}

type AppsConnectionsOpenResponse struct {
	Ok    bool   `json:"ok"`
	Error string `json:"error,omitempty"`
	URL   string `json:"url,omitempty"`
}

func main() {
	fmt.Println("🔬 Direct App-Level Token API Test")
	fmt.Println("===================================")
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

	// Test 1: Bot token auth.test
	fmt.Println("1️⃣ Testing Bot Token with auth.test...")
	if success := testBotTokenAuthTest(botToken); !success {
		log.Fatal("❌ Bot token test failed")
	}
	fmt.Println()

	// Test 2: App token with auth.test (should fail, but let's see how)
	fmt.Println("2️⃣ Testing App Token with auth.test...")
	testAppTokenAuthTest(appToken)
	fmt.Println()

	// Test 3: App token with apps.connections.open (the Socket Mode endpoint)
	fmt.Println("3️⃣ Testing App Token with apps.connections.open...")
	testAppTokenConnectionsOpen(appToken)
	fmt.Println()

	// Test 4: Raw HTTP request to WebSocket URL (if we got one)
	fmt.Println("4️⃣ Testing WebSocket URL retrieval...")
	testWebSocketConnection(appToken)
	fmt.Println()

	fmt.Println("🎯 Summary:")
	fmt.Println("===========")
	fmt.Println("If Test 3 (apps.connections.open) fails with 'invalid_auth',")
	fmt.Println("then your app-level token is definitely the problem.")
	fmt.Println()
	fmt.Println("Common causes:")
	fmt.Println("1. Token doesn't have 'connections:write' scope")
	fmt.Println("2. Token was created before Socket Mode was enabled")
	fmt.Println("3. Token belongs to different app/workspace")
	fmt.Println("4. App isn't properly installed in workspace")
	fmt.Println()
}

func testBotTokenAuthTest(botToken string) bool {
	url := "https://slack.com/api/auth.test"

	req, err := http.NewRequest("POST", url, nil)
	if err != nil {
		fmt.Printf("   ❌ Error creating request: %v\n", err)
		return false
	}

	req.Header.Set("Authorization", "Bearer "+botToken)
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	client := &http.Client{Timeout: 10 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		fmt.Printf("   ❌ Error making request: %v\n", err)
		return false
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		fmt.Printf("   ❌ Error reading response: %v\n", err)
		return false
	}

	var authResp AuthTestResponse
	if err := json.Unmarshal(body, &authResp); err != nil {
		fmt.Printf("   ❌ Error parsing JSON: %v\n", err)
		fmt.Printf("   📄 Raw response: %s\n", string(body))
		return false
	}

	if !authResp.Ok {
		fmt.Printf("   ❌ Auth test failed: %s\n", authResp.Error)
		return false
	}

	fmt.Printf("   ✅ Bot authenticated successfully\n")
	fmt.Printf("   📋 User: %s (%s)\n", authResp.User, authResp.UserID)
	fmt.Printf("   📋 Team: %s (%s)\n", authResp.Team, authResp.TeamID)
	fmt.Printf("   📋 Bot ID: %s\n", authResp.BotID)

	return true
}

func testAppTokenAuthTest(appToken string) {
	url := "https://slack.com/api/auth.test"

	req, err := http.NewRequest("POST", url, nil)
	if err != nil {
		fmt.Printf("   ❌ Error creating request: %v\n", err)
		return
	}

	req.Header.Set("Authorization", "Bearer "+appToken)
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	client := &http.Client{Timeout: 10 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		fmt.Printf("   ❌ Error making request: %v\n", err)
		return
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		fmt.Printf("   ❌ Error reading response: %v\n", err)
		return
	}

	var authResp AuthTestResponse
	if err := json.Unmarshal(body, &authResp); err != nil {
		fmt.Printf("   ❌ Error parsing JSON: %v\n", err)
		fmt.Printf("   📄 Raw response: %s\n", string(body))
		return
	}

	if !authResp.Ok {
		fmt.Printf("   ⚠️  Expected failure - app tokens don't work with auth.test: %s\n", authResp.Error)
	} else {
		fmt.Printf("   🤔 Unexpected success with app token on auth.test\n")
	}
}

func testAppTokenConnectionsOpen(appToken string) {
	url := "https://slack.com/api/apps.connections.open"

	req, err := http.NewRequest("POST", url, nil)
	if err != nil {
		fmt.Printf("   ❌ Error creating request: %v\n", err)
		return
	}

	req.Header.Set("Authorization", "Bearer "+appToken)
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	client := &http.Client{Timeout: 10 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		fmt.Printf("   ❌ Error making request: %v\n", err)
		return
	}
	defer resp.Body.Close()

	fmt.Printf("   📊 HTTP Status: %d\n", resp.StatusCode)

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		fmt.Printf("   ❌ Error reading response: %v\n", err)
		return
	}

	var connResp AppsConnectionsOpenResponse
	if err := json.Unmarshal(body, &connResp); err != nil {
		fmt.Printf("   ❌ Error parsing JSON: %v\n", err)
		fmt.Printf("   📄 Raw response: %s\n", string(body))
		return
	}

	if !connResp.Ok {
		fmt.Printf("   ❌ Apps.connections.open failed: %s\n", connResp.Error)

		switch connResp.Error {
		case "invalid_auth":
			fmt.Printf("   🔍 DIAGNOSIS: Your app-level token is invalid or lacks 'connections:write' scope\n")
		case "not_authed":
			fmt.Printf("   🔍 DIAGNOSIS: Authentication failed - check token format and scopes\n")
		case "account_inactive":
			fmt.Printf("   🔍 DIAGNOSIS: Account/app is inactive\n")
		case "invalid_arg_name":
			fmt.Printf("   🔍 DIAGNOSIS: API argument issue\n")
		case "missing_scope":
			fmt.Printf("   🔍 DIAGNOSIS: App-level token missing 'connections:write' scope\n")
		default:
			fmt.Printf("   🔍 DIAGNOSIS: Unknown error - check Slack API docs\n")
		}
		return
	}

	fmt.Printf("   ✅ Apps.connections.open succeeded!\n")
	fmt.Printf("   🔗 WebSocket URL: %s\n", connResp.URL)

	if connResp.URL != "" {
		fmt.Printf("   🎉 Your app-level token is VALID and Socket Mode should work!\n")
	}
}

func testWebSocketConnection(appToken string) {
	// First get the WebSocket URL
	url := "https://slack.com/api/apps.connections.open"

	req, err := http.NewRequest("POST", url, nil)
	if err != nil {
		fmt.Printf("   ❌ Error creating request: %v\n", err)
		return
	}

	req.Header.Set("Authorization", "Bearer "+appToken)
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	client := &http.Client{Timeout: 10 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		fmt.Printf("   ❌ Error making request: %v\n", err)
		return
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		fmt.Printf("   ❌ Error reading response: %v\n", err)
		return
	}

	var connResp AppsConnectionsOpenResponse
	if err := json.Unmarshal(body, &connResp); err != nil {
		fmt.Printf("   ❌ Error parsing JSON: %v\n", err)
		return
	}

	if !connResp.Ok {
		fmt.Printf("   ⚠️  Cannot test WebSocket - apps.connections.open failed: %s\n", connResp.Error)
		return
	}

	if connResp.URL == "" {
		fmt.Printf("   ⚠️  No WebSocket URL returned\n")
		return
	}

	fmt.Printf("   🔗 WebSocket URL obtained: %s\n", connResp.URL[:50]+"...")

	// Test if URL is reachable (basic connectivity test)
	wsURL := strings.Replace(connResp.URL, "wss://", "https://", 1)
	wsURL = strings.Split(wsURL, "?")[0] // Remove query params for basic test

	testReq, err := http.NewRequest("GET", wsURL, nil)
	if err != nil {
		fmt.Printf("   ❌ Error creating WebSocket test request: %v\n", err)
		return
	}

	testClient := &http.Client{Timeout: 5 * time.Second}
	testResp, err := testClient.Do(testReq)
	if err != nil {
		fmt.Printf("   ⚠️  WebSocket endpoint test failed: %v\n", err)
		fmt.Printf("   📝 This might be normal - WebSocket endpoints often reject HTTP requests\n")
		return
	}
	defer testResp.Body.Close()

	fmt.Printf("   📊 WebSocket endpoint HTTP status: %d\n", testResp.StatusCode)
	fmt.Printf("   ✅ WebSocket endpoint is reachable\n")
}
