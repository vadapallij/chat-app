#!/bin/bash

# Database setup script for chat-app
# This script creates the PostgreSQL user and database

set -e

echo "================================================"
echo "PostgreSQL Database Setup for Chat-App"
echo "================================================"
echo ""

# Colors
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
RED='\033[0;31m'
NC='\033[0m'

print_success() {
    echo -e "${GREEN}✓ $1${NC}"
}

print_error() {
    echo -e "${RED}✗ $1${NC}"
}

print_info() {
    echo -e "ℹ $1"
}

# Get current username
CURRENT_USER=$(whoami)
DB_NAME="chat_app"

echo "Current user: $CURRENT_USER"
echo "Database to create: $DB_NAME"
echo ""

# Check if PostgreSQL is running
if ! sudo systemctl is-active --quiet postgresql; then
    print_error "PostgreSQL is not running"
    echo "Starting PostgreSQL..."
    sudo systemctl start postgresql
    print_success "PostgreSQL started"
fi

# Step 1: Create PostgreSQL user if it doesn't exist
print_info "Checking if PostgreSQL user '$CURRENT_USER' exists..."

if sudo -u postgres psql -tAc "SELECT 1 FROM pg_roles WHERE rolname='$CURRENT_USER'" | grep -q 1; then
    print_success "User '$CURRENT_USER' already exists"
else
    print_info "Creating PostgreSQL user '$CURRENT_USER'..."
    sudo -u postgres psql -c "CREATE USER $CURRENT_USER WITH CREATEDB LOGIN;"
    print_success "User '$CURRENT_USER' created"
fi

# Step 2: Create database if it doesn't exist
print_info "Checking if database '$DB_NAME' exists..."

if sudo -u postgres psql -lqt | cut -d \| -f 1 | grep -qw "$DB_NAME"; then
    print_success "Database '$DB_NAME' already exists"

    echo ""
    echo "Database already exists. What would you like to do?"
    echo "1) Keep existing database (skip)"
    echo "2) Drop and recreate (WARNING: All data will be lost!)"
    read -p "Enter choice (1 or 2): " choice

    if [ "$choice" = "2" ]; then
        print_info "Dropping database '$DB_NAME'..."
        sudo -u postgres psql -c "DROP DATABASE $DB_NAME;"
        print_info "Creating database '$DB_NAME'..."
        sudo -u postgres psql -c "CREATE DATABASE $DB_NAME OWNER $CURRENT_USER;"
        print_success "Database recreated"
        RUN_MIGRATIONS=true
    else
        print_info "Keeping existing database"
        RUN_MIGRATIONS=false
    fi
else
    print_info "Creating database '$DB_NAME'..."
    sudo -u postgres psql -c "CREATE DATABASE $DB_NAME OWNER $CURRENT_USER;"
    print_success "Database created"
    RUN_MIGRATIONS=true
fi

# Step 3: Run migrations
if [ "$RUN_MIGRATIONS" = true ]; then
    echo ""
    print_info "Running database migrations..."

    if psql -d "$DB_NAME" -f migrations/001_init.sql; then
        print_success "Migrations completed successfully"
    else
        print_error "Migration failed"
        exit 1
    fi

    echo ""
    print_info "Verifying tables..."
    psql -d "$DB_NAME" -c "\dt"
fi

echo ""
echo "================================================"
print_success "Database setup complete!"
echo "================================================"
echo ""
echo "Your database connection string:"
echo "${GREEN}DATABASE_URL=postgresql://$CURRENT_USER@localhost:5432/$DB_NAME${NC}"
echo ""
echo "You can now:"
echo "1. Add this to your .env file"
echo "2. Continue with: ./setup.sh"
echo "   or manually: go run main.go"
echo ""
