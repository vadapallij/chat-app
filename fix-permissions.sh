#!/bin/bash

# Quick fix for PostgreSQL permissions

echo "Fixing PostgreSQL permissions..."

CURRENT_USER=$(whoami)
DB_NAME="chat_app"

# Grant all privileges to the user on the database
sudo -u postgres psql -c "GRANT ALL PRIVILEGES ON DATABASE $DB_NAME TO $CURRENT_USER;"

# Grant permissions on the public schema
sudo -u postgres psql -d "$DB_NAME" -c "GRANT ALL ON SCHEMA public TO $CURRENT_USER;"

# Make user owner of public schema
sudo -u postgres psql -d "$DB_NAME" -c "ALTER SCHEMA public OWNER TO $CURRENT_USER;"

echo "✓ Permissions fixed!"
echo ""
echo "Now run the migrations:"
echo "  psql chat_app < migrations/001_init.sql"
