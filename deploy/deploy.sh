#!/bin/bash

# FamBot Production Deployment Script
# This script deploys FamBot to a production environment

set -e

# Configuration
APP_NAME="fambot"
APP_USER="fambot"
APP_GROUP="fambot"
APP_DIR="/opt/fambot"
SERVICE_FILE="fambot.service"
BINARY_NAME="fambot"

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

# Print functions
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

# Check if running as root
check_root() {
    if [ "$EUID" -ne 0 ]; then
        print_error "This script must be run as root (use sudo)"
        exit 1
    fi
}

# Check prerequisites
check_prerequisites() {
    print_header "📋 Checking Prerequisites..."
    echo ""

    local all_good=true

    # Check systemd
    if command -v systemctl >/dev/null 2>&1; then
        print_status "systemd is available"
    else
        print_error "systemd is required for this deployment method"
        all_good=false
    fi

    # Check SQLite
    if command -v sqlite3 >/dev/null 2>&1; then
        print_status "SQLite is installed"
    else
        print_warning "SQLite CLI not found. Installing..."
        if command -v apt-get >/dev/null 2>&1; then
            apt-get update && apt-get install -y sqlite3
        elif command -v yum >/dev/null 2>&1; then
            yum install -y sqlite
        elif command -v dnf >/dev/null 2>&1; then
            dnf install -y sqlite
        else
            print_error "Could not install SQLite. Please install it manually."
            all_good=false
        fi
    fi

    echo ""

    if [ "$all_good" = false ]; then
        print_error "Please resolve missing prerequisites and run this script again"
        exit 1
    fi

    print_status "All prerequisites are satisfied!"
    echo ""
}

# Create application user
create_app_user() {
    print_header "👤 Creating Application User..."
    echo ""

    if id "$APP_USER" &>/dev/null; then
        print_warning "User $APP_USER already exists"
    else
        print_info "Creating user $APP_USER..."
        useradd --system --shell /bin/false --home-dir "$APP_DIR" --create-home "$APP_USER"
        print_status "User $APP_USER created"
    fi

    echo ""
}

# Setup directories
setup_directories() {
    print_header "📁 Setting Up Directories..."
    echo ""

    print_info "Creating application directories..."

    # Create main directory
    mkdir -p "$APP_DIR"
    mkdir -p "$APP_DIR/bin"
    mkdir -p "$APP_DIR/data"
    mkdir -p "$APP_DIR/logs"
    mkdir -p "$APP_DIR/backups"

    # Set ownership
    chown -R "$APP_USER:$APP_GROUP" "$APP_DIR"

    # Set permissions
    chmod 755 "$APP_DIR"
    chmod 755 "$APP_DIR/bin"
    chmod 750 "$APP_DIR/data"
    chmod 755 "$APP_DIR/logs"
    chmod 750 "$APP_DIR/backups"

    print_status "Directories created and configured"
    echo ""
}

# Build application
build_application() {
    print_header "🔨 Building Application..."
    echo ""

    if [ ! -f "cmd/main.go" ]; then
        print_error "Source code not found. Please run this script from the project root directory."
        exit 1
    fi

    print_info "Building FamBot..."

    # Build for production
    CGO_ENABLED=1 go build -ldflags="-w -s" -o "$APP_DIR/bin/$BINARY_NAME" cmd/main.go

    # Set executable permissions
    chmod +x "$APP_DIR/bin/$BINARY_NAME"
    chown "$APP_USER:$APP_GROUP" "$APP_DIR/bin/$BINARY_NAME"

    print_status "Application built and installed"
    echo ""
}

# Setup configuration
setup_configuration() {
    print_header "⚙️ Setting Up Configuration..."
    echo ""

    # Copy environment template if .env doesn't exist
    if [ ! -f "$APP_DIR/.env" ]; then
        if [ -f ".env.example" ]; then
            print_info "Copying environment template..."
            cp .env.example "$APP_DIR/.env"
            chown "$APP_USER:$APP_GROUP" "$APP_DIR/.env"
            chmod 600 "$APP_DIR/.env"
            print_status "Environment template copied"
        else
            print_info "Creating basic environment file..."
            cat > "$APP_DIR/.env" << EOF
# Slack Bot Configuration
SLACK_BOT_TOKEN=xoxb-your-bot-token-here
SLACK_APP_TOKEN=xapp-your-app-token-here

# Database Configuration
DATABASE_PATH=$APP_DIR/data/fambot.db

# Slack Channel Configuration
PEOPLE_CHANNEL=people

# Debug Configuration
DEBUG=false
EOF
            chown "$APP_USER:$APP_GROUP" "$APP_DIR/.env"
            chmod 600 "$APP_DIR/.env"
            print_status "Basic environment file created"
        fi
    else
        print_warning ".env file already exists, skipping creation"
    fi

    echo ""
    print_step "⚠️  IMPORTANT: Please edit $APP_DIR/.env with your Slack tokens"
    echo ""
}

# Install systemd service
install_service() {
    print_header "🔧 Installing Systemd Service..."
    echo ""

    if [ ! -f "deploy/$SERVICE_FILE" ]; then
        print_error "Service file deploy/$SERVICE_FILE not found"
        exit 1
    fi

    print_info "Installing systemd service..."

    # Copy service file
    cp "deploy/$SERVICE_FILE" "/etc/systemd/system/"

    # Reload systemd
    systemctl daemon-reload

    print_status "Systemd service installed"
    echo ""
}

# Setup log rotation
setup_log_rotation() {
    print_header "📝 Setting Up Log Rotation..."
    echo ""

    print_info "Creating logrotate configuration..."

    cat > "/etc/logrotate.d/fambot" << EOF
$APP_DIR/logs/*.log {
    daily
    missingok
    rotate 30
    compress
    delaycompress
    notifempty
    create 644 $APP_USER $APP_GROUP
    postrotate
        systemctl reload $APP_NAME > /dev/null 2>&1 || true
    endscript
}
EOF

    print_status "Log rotation configured"
    echo ""
}

# Setup firewall (if needed)
setup_firewall() {
    print_header "🔥 Configuring Firewall..."
    echo ""

    # Check if ufw is available
    if command -v ufw >/dev/null 2>&1; then
        print_info "UFW detected, ensuring SSH access..."
        ufw allow ssh
        print_status "Firewall configured"
    elif command -v firewall-cmd >/dev/null 2>&1; then
        print_info "Firewalld detected, ensuring SSH access..."
        firewall-cmd --permanent --add-service=ssh
        firewall-cmd --reload
        print_status "Firewall configured"
    else
        print_warning "No supported firewall detected, skipping firewall configuration"
    fi

    echo ""
}

# Start service
start_service() {
    print_header "🚀 Starting Service..."
    echo ""

    print_info "Enabling and starting $APP_NAME service..."

    # Enable service
    systemctl enable "$APP_NAME"

    # Start service
    systemctl start "$APP_NAME"

    # Wait a moment for startup
    sleep 3

    # Check status
    if systemctl is-active --quiet "$APP_NAME"; then
        print_status "Service started successfully"

        print_info "Service status:"
        systemctl status "$APP_NAME" --no-pager -l
    else
        print_error "Service failed to start"
        print_info "Checking logs..."
        journalctl -u "$APP_NAME" --no-pager -l -n 20
        exit 1
    fi

    echo ""
}

# Create backup script
create_backup_script() {
    print_header "💾 Creating Backup Script..."
    echo ""

    print_info "Creating backup script..."

    cat > "$APP_DIR/backup.sh" << 'EOF'
#!/bin/bash

# FamBot Backup Script
APP_DIR="/opt/fambot"
BACKUP_DIR="$APP_DIR/backups"
DATE=$(date +%Y%m%d_%H%M%S)
DB_FILE="$APP_DIR/data/fambot.db"

# Create backup
if [ -f "$DB_FILE" ]; then
    cp "$DB_FILE" "$BACKUP_DIR/fambot_backup_$DATE.db"
    echo "Database backed up to $BACKUP_DIR/fambot_backup_$DATE.db"

    # Keep only last 10 backups
    cd "$BACKUP_DIR"
    ls -t fambot_backup_*.db | tail -n +11 | xargs -r rm
    echo "Old backups cleaned up"
else
    echo "Database file not found: $DB_FILE"
    exit 1
fi
EOF

    chmod +x "$APP_DIR/backup.sh"
    chown "$APP_USER:$APP_GROUP" "$APP_DIR/backup.sh"

    # Create daily backup cron job
    cat > "/etc/cron.d/fambot-backup" << EOF
# FamBot daily backup
0 2 * * * $APP_USER $APP_DIR/backup.sh >> $APP_DIR/logs/backup.log 2>&1
EOF

    print_status "Backup script and cron job created"
    echo ""
}

# Create maintenance script
create_maintenance_script() {
    print_header "🔧 Creating Maintenance Script..."
    echo ""

    print_info "Creating maintenance script..."

    cat > "$APP_DIR/maintenance.sh" << 'EOF'
#!/bin/bash

# FamBot Maintenance Script
APP_NAME="fambot"
APP_DIR="/opt/fambot"

case "$1" in
    start)
        echo "Starting FamBot..."
        sudo systemctl start $APP_NAME
        ;;
    stop)
        echo "Stopping FamBot..."
        sudo systemctl stop $APP_NAME
        ;;
    restart)
        echo "Restarting FamBot..."
        sudo systemctl restart $APP_NAME
        ;;
    status)
        sudo systemctl status $APP_NAME --no-pager
        ;;
    logs)
        echo "Recent logs:"
        sudo journalctl -u $APP_NAME --no-pager -l -n 50
        ;;
    tail)
        echo "Tailing logs (Ctrl+C to exit):"
        sudo journalctl -u $APP_NAME -f
        ;;
    backup)
        echo "Creating backup..."
        $APP_DIR/backup.sh
        ;;
    db-info)
        if [ -f "$APP_DIR/data/fambot.db" ]; then
            echo "Database info:"
            echo "Size: $(du -h $APP_DIR/data/fambot.db | cut -f1)"
            echo "Tables:"
            sqlite3 $APP_DIR/data/fambot.db ".tables"
        else
            echo "Database not found"
        fi
        ;;
    update)
        echo "Updating FamBot..."
        echo "Please deploy new version manually"
        ;;
    *)
        echo "Usage: $0 {start|stop|restart|status|logs|tail|backup|db-info|update}"
        exit 1
        ;;
esac
EOF

    chmod +x "$APP_DIR/maintenance.sh"
    chown "$APP_USER:$APP_GROUP" "$APP_DIR/maintenance.sh"

    print_status "Maintenance script created"
    echo ""
}

# Validate deployment
validate_deployment() {
    print_header "✅ Validating Deployment..."
    echo ""

    local validation_passed=true

    # Check if binary exists
    if [ -f "$APP_DIR/bin/$BINARY_NAME" ]; then
        print_status "Binary is installed"
    else
        print_error "Binary not found"
        validation_passed=false
    fi

    # Check if service is enabled
    if systemctl is-enabled --quiet "$APP_NAME"; then
        print_status "Service is enabled"
    else
        print_error "Service is not enabled"
        validation_passed=false
    fi

    # Check if service is running
    if systemctl is-active --quiet "$APP_NAME"; then
        print_status "Service is running"
    else
        print_error "Service is not running"
        validation_passed=false
    fi

    # Check if directories exist with correct permissions
    if [ -d "$APP_DIR" ] && [ "$(stat -c '%U' $APP_DIR)" = "$APP_USER" ]; then
        print_status "Directories are properly configured"
    else
        print_error "Directory permissions are incorrect"
        validation_passed=false
    fi

    # Check if .env file exists
    if [ -f "$APP_DIR/.env" ]; then
        print_status "Configuration file exists"

        # Check if tokens are configured
        if grep -q "xoxb-your-bot-token-here" "$APP_DIR/.env" || grep -q "xapp-your-app-token-here" "$APP_DIR/.env"; then
            print_warning "Slack tokens need to be configured in $APP_DIR/.env"
        else
            print_status "Slack tokens appear to be configured"
        fi
    else
        print_error "Configuration file not found"
        validation_passed=false
    fi

    echo ""

    if [ "$validation_passed" = true ]; then
        print_status "Deployment validation passed!"
    else
        print_error "Deployment validation failed!"
        return 1
    fi

    echo ""
}

# Show deployment summary
show_summary() {
    print_header "🎉 Deployment Complete!"
    echo ""
    print_info "FamBot has been successfully deployed!"
    echo ""
    print_step "Next steps:"
    echo "1. Edit $APP_DIR/.env with your Slack tokens"
    echo "2. Restart the service: sudo systemctl restart $APP_NAME"
    echo "3. Check the status: sudo systemctl status $APP_NAME"
    echo ""
    print_step "Useful commands:"
    echo "• View logs: sudo journalctl -u $APP_NAME -f"
    echo "• Restart service: sudo systemctl restart $APP_NAME"
    echo "• Backup database: $APP_DIR/backup.sh"
    echo "• Maintenance: $APP_DIR/maintenance.sh {start|stop|restart|status|logs}"
    echo ""
    print_step "Files and directories:"
    echo "• Application: $APP_DIR"
    echo "• Binary: $APP_DIR/bin/$BINARY_NAME"
    echo "• Configuration: $APP_DIR/.env"
    echo "• Database: $APP_DIR/data/fambot.db"
    echo "• Logs: $APP_DIR/logs/"
    echo "• Backups: $APP_DIR/backups/"
    echo ""
}

# Update deployment
update_deployment() {
    print_header "🔄 Updating Deployment..."
    echo ""

    # Stop service
    print_info "Stopping service..."
    systemctl stop "$APP_NAME"

    # Backup current binary
    if [ -f "$APP_DIR/bin/$BINARY_NAME" ]; then
        print_info "Backing up current binary..."
        cp "$APP_DIR/bin/$BINARY_NAME" "$APP_DIR/bin/${BINARY_NAME}.backup.$(date +%Y%m%d_%H%M%S)"
    fi

    # Build and install new binary
    build_application

    # Start service
    print_info "Starting service..."
    systemctl start "$APP_NAME"

    # Validate
    if systemctl is-active --quiet "$APP_NAME"; then
        print_status "Update completed successfully"
    else
        print_error "Update failed, service is not running"
        print_info "Check logs: journalctl -u $APP_NAME"
        exit 1
    fi

    echo ""
}

# Uninstall
uninstall() {
    print_header "🗑️  Uninstalling FamBot..."
    echo ""

    read -p "Are you sure you want to uninstall FamBot? This will remove all data! (y/N): " confirm
    if [ "$confirm" != "y" ] && [ "$confirm" != "Y" ]; then
        print_info "Uninstall cancelled"
        exit 0
    fi

    # Stop and disable service
    print_info "Stopping and disabling service..."
    systemctl stop "$APP_NAME" || true
    systemctl disable "$APP_NAME" || true

    # Remove service file
    print_info "Removing service file..."
    rm -f "/etc/systemd/system/$SERVICE_FILE"
    systemctl daemon-reload

    # Remove application directory
    print_info "Removing application directory..."
    rm -rf "$APP_DIR"

    # Remove user
    print_info "Removing user..."
    userdel "$APP_USER" || true

    # Remove logrotate config
    print_info "Removing logrotate config..."
    rm -f "/etc/logrotate.d/fambot"

    # Remove cron job
    print_info "Removing cron job..."
    rm -f "/etc/cron.d/fambot-backup"

    print_status "FamBot has been uninstalled"
    echo ""
}

# Main function
main() {
    case "${1:-install}" in
        install)
            print_header "🤖 FamBot Production Deployment"
            echo ""
            check_root
            check_prerequisites
            create_app_user
            setup_directories
            build_application
            setup_configuration
            install_service
            setup_log_rotation
            setup_firewall
            create_backup_script
            create_maintenance_script
            start_service
            validate_deployment
            show_summary
            ;;
        update)
            check_root
            update_deployment
            ;;
        uninstall)
            check_root
            uninstall
            ;;
        validate)
            validate_deployment
            ;;
        *)
            echo "Usage: $0 {install|update|uninstall|validate}"
            echo ""
            echo "Commands:"
            echo "  install    - Full installation (default)"
            echo "  update     - Update existing installation"
            echo "  uninstall  - Remove FamBot completely"
            echo "  validate   - Validate current installation"
            exit 1
            ;;
    esac
}

# Run main function
main "$@"
