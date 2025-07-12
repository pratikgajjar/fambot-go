#!/bin/bash

# FamBot-Go Setup Script
# This script helps you set up the FamBot Slack bot

set -e

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
PURPLE='\033[0;35m'
CYAN='\033[0;36m'
NC='\033[0m' # No Color

# Emojis
CHECKMARK="✅"
CROSS="❌"
ROCKET="🚀"
ROBOT="🤖"
GEAR="⚙️"
WARNING="⚠️"

# Print colored output
print_status() {
    echo -e "${GREEN}${CHECKMARK} $1${NC}"
}

print_error() {
    echo -e "${RED}${CROSS} $1${NC}"
}

print_warning() {
    echo -e "${YELLOW}${WARNING} $1${NC}"
}

print_info() {
    echo -e "${BLUE}${GEAR} $1${NC}"
}

print_header() {
    echo -e "${PURPLE}$1${NC}"
}

print_step() {
    echo -e "${CYAN}$1${NC}"
}

# Check if command exists
command_exists() {
    command -v "$1" >/dev/null 2>&1
}

# Print welcome message
print_welcome() {
    clear
    cat << "EOF"
    ███████╗ █████╗ ███╗   ███╗██████╗  ██████╗ ████████╗
    ██╔════╝██╔══██╗████╗ ████║██╔══██╗██╔═══██╗╚══██╔══╝
    █████╗  ███████║██╔████╔██║██████╔╝██║   ██║   ██║
    ██╔══╝  ██╔══██║██║╚██╔╝██║██╔══██╗██║   ██║   ██║
    ██║     ██║  ██║██║ ╚═╝ ██║██████╔╝╚██████╔╝   ██║
    ╚═╝     ╚═╝  ╚═╝╚═╝     ╚═╝╚═════╝  ╚═════╝    ╚═╝

EOF
    echo -e "${PURPLE}🤖 Welcome to FamBot-Go Setup! 🤖${NC}"
    echo -e "${CYAN}A sassy Slack bot that tracks karma and celebrations${NC}"
    echo ""
}

# Check prerequisites
check_prerequisites() {
    print_header "📋 Checking Prerequisites..."
    echo ""

    local all_good=true

    # Check Go
    if command_exists go; then
        local go_version=$(go version | awk '{print $3}' | sed 's/go//')
        print_status "Go is installed (version $go_version)"
    else
        print_error "Go is not installed. Please install Go 1.21+ from https://golang.org/"
        all_good=false
    fi

    # Check Git
    if command_exists git; then
        print_status "Git is installed"
    else
        print_error "Git is not installed. Please install Git"
        all_good=false
    fi

    # Check Make
    if command_exists make; then
        print_status "Make is installed"
    else
        print_warning "Make is not installed. You can still run Go commands directly"
    fi

    # Check SQLite
    if command_exists sqlite3; then
        print_status "SQLite is installed"
    else
        print_warning "SQLite CLI not found. The bot will still work, but you won't be able to inspect the database easily"
    fi

    echo ""

    if [ "$all_good" = false ]; then
        print_error "Please install missing prerequisites and run this script again"
        exit 1
    fi

    print_status "All prerequisites are satisfied!"
    echo ""
}

# Install dependencies
install_dependencies() {
    print_header "📦 Installing Dependencies..."
    echo ""

    print_info "Running 'go mod tidy'..."
    go mod tidy

    print_info "Downloading dependencies..."
    go mod download

    print_status "Dependencies installed successfully!"
    echo ""
}

# Setup environment
setup_environment() {
    print_header "🔧 Setting Up Environment..."
    echo ""

    if [ ! -f .env ]; then
        print_info "Creating .env file from template..."
        cp .env.example .env
        print_status ".env file created!"
    else
        print_warning ".env file already exists. Skipping creation."
    fi

    echo ""
    print_step "Please edit the .env file with your Slack tokens:"
    echo ""
    echo "1. SLACK_BOT_TOKEN=xoxb-your-bot-token-here"
    echo "2. SLACK_APP_TOKEN=xapp-your-app-token-here"
    echo "3. PEOPLE_CHANNEL=your-channel-name (default: people)"
    echo ""

    read -p "Press Enter when you've configured your .env file..."
    echo ""
}

# Validate environment
validate_environment() {
    print_header "✅ Validating Environment..."
    echo ""

    source .env

    local config_valid=true

    if [ -z "$SLACK_BOT_TOKEN" ] || [ "$SLACK_BOT_TOKEN" = "xoxb-your-bot-token-here" ]; then
        print_error "SLACK_BOT_TOKEN is not properly configured"
        config_valid=false
    else
        print_status "SLACK_BOT_TOKEN is configured"
    fi

    if [ -z "$SLACK_APP_TOKEN" ] || [ "$SLACK_APP_TOKEN" = "xapp-your-app-token-here" ]; then
        print_error "SLACK_APP_TOKEN is not properly configured"
        config_valid=false
    else
        print_status "SLACK_APP_TOKEN is configured"
    fi

    if [ -z "$PEOPLE_CHANNEL" ]; then
        print_warning "PEOPLE_CHANNEL not set, using default 'people'"
    else
        print_status "PEOPLE_CHANNEL is set to '$PEOPLE_CHANNEL'"
    fi

    echo ""

    if [ "$config_valid" = false ]; then
        print_error "Environment configuration is incomplete"
        echo ""
        print_info "Please ensure you have:"
        echo "  1. Created a Slack app at https://api.slack.com/apps"
        echo "  2. Enabled Socket Mode"
        echo "  3. Added the required OAuth scopes"
        echo "  4. Installed the app to your workspace"
        echo "  5. Copied the tokens to your .env file"
        echo ""
        exit 1
    fi

    print_status "Environment configuration is valid!"
    echo ""
}

# Build the application
build_application() {
    print_header "🔨 Building Application..."
    echo ""

    print_info "Building FamBot..."
    if command_exists make; then
        make build
    else
        mkdir -p bin
        go build -o bin/fambot cmd/main.go
    fi

    print_status "Build completed successfully!"
    echo ""
}

# Slack setup instructions
show_slack_instructions() {
    print_header "📱 Slack App Setup Instructions"
    echo ""
    print_info "If you haven't set up your Slack app yet, follow these steps:"
    echo ""
    echo "1. Go to https://api.slack.com/apps"
    echo "2. Click 'Create New App' → 'From scratch'"
    echo "3. Name your app (e.g., 'FamBot') and select your workspace"
    echo ""
    echo "4. Configure OAuth & Permissions:"
    echo "   Add these Bot Token Scopes:"
    echo "   - app_mentions:read"
    echo "   - channels:history"
    echo "   - channels:read"
    echo "   - chat:write"
    echo "   - commands"
    echo "   - groups:history"
    echo "   - groups:read"
    echo "   - im:history"
    echo "   - im:read"
    echo "   - mpim:history"
    echo "   - mpim:read"
    echo "   - users:read"
    echo "   - users:read.email"
    echo ""
    echo "5. Enable Socket Mode:"
    echo "   - Go to Socket Mode"
    echo "   - Enable Socket Mode"
    echo "   - Generate App-Level Token with 'connections:write' scope"
    echo ""
    echo "6. Enable Event Subscriptions:"
    echo "   Subscribe to these bot events:"
    echo "   - app_mention"
    echo "   - message.channels"
    echo "   - message.groups"
    echo "   - message.im"
    echo "   - message.mpim"
    echo ""
    echo "7. Add Slash Commands:"
    echo "   - /top-karma"
    echo "   - /my-karma"
    echo "   - /set-birthday"
    echo "   - /set-anniversary"
    echo "   - /fambot-help"
    echo ""
    echo "8. Install to Workspace and copy the tokens"
    echo ""

    read -p "Press Enter to continue..."
    echo ""
}

# Run the bot
run_bot() {
    print_header "🚀 Starting FamBot..."
    echo ""

    print_info "FamBot is starting up..."
    print_info "Press Ctrl+C to stop the bot"
    echo ""

    if [ -f bin/fambot ]; then
        ./bin/fambot
    else
        go run cmd/main.go
    fi
}

# Test mode
test_setup() {
    print_header "🧪 Testing Setup..."
    echo ""

    print_info "Running build test..."
    if command_exists make; then
        make build
    else
        go build -o bin/fambot cmd/main.go
    fi

    print_status "Build test passed!"

    if [ -f .env ]; then
        print_info "Checking environment configuration..."
        source .env

        if [ -n "$SLACK_BOT_TOKEN" ] && [ "$SLACK_BOT_TOKEN" != "xoxb-your-bot-token-here" ]; then
            print_status "SLACK_BOT_TOKEN is configured"
        else
            print_warning "SLACK_BOT_TOKEN needs configuration"
        fi

        if [ -n "$SLACK_APP_TOKEN" ] && [ "$SLACK_APP_TOKEN" != "xapp-your-app-token-here" ]; then
            print_status "SLACK_APP_TOKEN is configured"
        else
            print_warning "SLACK_APP_TOKEN needs configuration"
        fi
    else
        print_warning ".env file not found"
    fi

    echo ""
    print_status "Setup test completed!"
    echo ""
}

# Clean environment
clean_environment() {
    print_header "🧹 Cleaning Environment..."
    echo ""

    print_info "Removing build artifacts..."
    rm -rf bin/
    rm -rf tmp/
    rm -f *.log

    if [ -f fambot.db ]; then
        print_warning "Found database file (fambot.db)"
        read -p "Do you want to remove it? (y/N): " remove_db
        if [ "$remove_db" = "y" ] || [ "$remove_db" = "Y" ]; then
            rm -f fambot.db*
            print_status "Database removed"
        else
            print_info "Database kept"
        fi
    fi

    print_status "Environment cleaned!"
    echo ""
}

# Main menu
show_menu() {
    print_header "🎛️  FamBot Setup Menu"
    echo ""
    echo "1. Full Setup (recommended for first time)"
    echo "2. Quick Start (if already configured)"
    echo "3. Slack App Instructions"
    echo "4. Test Setup"
    echo "5. Clean Environment"
    echo "6. Exit"
    echo ""
    read -p "Choose an option (1-6): " choice
    echo ""

    case $choice in
        1)
            full_setup
            ;;
        2)
            quick_start
            ;;
        3)
            show_slack_instructions
            show_menu
            ;;
        4)
            test_setup
            show_menu
            ;;
        5)
            clean_environment
            show_menu
            ;;
        6)
            print_status "Goodbye! 👋"
            exit 0
            ;;
        *)
            print_error "Invalid option. Please choose 1-6."
            show_menu
            ;;
    esac
}

# Full setup process
full_setup() {
    check_prerequisites
    install_dependencies
    show_slack_instructions
    setup_environment
    validate_environment
    build_application

    print_header "🎉 Setup Complete!"
    echo ""
    print_status "FamBot is ready to run!"
    echo ""
    print_info "You can now:"
    echo "  • Run 'make run' or './bin/fambot' to start the bot"
    echo "  • Run 'make dev' for development mode with hot reload"
    echo "  • Run 'make help' to see all available commands"
    echo ""

    read -p "Do you want to start FamBot now? (y/N): " start_now
    if [ "$start_now" = "y" ] || [ "$start_now" = "Y" ]; then
        run_bot
    fi
}

# Quick start for already configured environments
quick_start() {
    print_info "Running quick start..."
    echo ""

    validate_environment
    build_application

    print_status "Quick start complete!"
    echo ""

    read -p "Do you want to start FamBot now? (y/N): " start_now
    if [ "$start_now" = "y" ] || [ "$start_now" = "Y" ]; then
        run_bot
    fi
}

# Main execution
main() {
    print_welcome

    # Check if running with arguments
    if [ $# -eq 0 ]; then
        show_menu
    else
        case $1 in
            --full-setup|full)
                full_setup
                ;;
            --quick-start|quick)
                quick_start
                ;;
            --test|test)
                test_setup
                ;;
            --clean|clean)
                clean_environment
                ;;
            --instructions|instructions)
                show_slack_instructions
                ;;
            --help|-h|help)
                echo "FamBot Setup Script"
                echo ""
                echo "Usage: $0 [option]"
                echo ""
                echo "Options:"
                echo "  --full-setup     Run complete setup process"
                echo "  --quick-start    Quick start for configured environments"
                echo "  --test           Test the setup"
                echo "  --clean          Clean environment"
                echo "  --instructions   Show Slack app setup instructions"
                echo "  --help           Show this help message"
                echo ""
                echo "Run without arguments for interactive menu."
                ;;
            *)
                print_error "Unknown option: $1"
                echo "Run '$0 --help' for usage information."
                exit 1
                ;;
        esac
    fi
}

# Run main function with all arguments
main "$@"
