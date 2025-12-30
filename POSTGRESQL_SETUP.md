# PostgreSQL Setup Guide for Chat-App

This guide will help you install PostgreSQL and set up the database for your chat application.

## Installation

### For Ubuntu/Debian

```bash
# Update package list
sudo apt update

# Install PostgreSQL
sudo apt install postgresql postgresql-contrib

# Start PostgreSQL service
sudo systemctl start postgresql
sudo systemctl enable postgresql

# Verify installation
sudo systemctl status postgresql
```

### For Fedora/RHEL/CentOS

```bash
# Install PostgreSQL
sudo dnf install postgresql-server postgresql-contrib

# Initialize database
sudo postgresql-setup --initdb

# Start and enable service
sudo systemctl start postgresql
sudo systemctl enable postgresql
```

### For Arch Linux

```bash
# Install PostgreSQL
sudo pacman -S postgresql

# Initialize database cluster
sudo -u postgres initdb -D /var/lib/postgres/data

# Start and enable service
sudo systemctl start postgresql
sudo systemctl enable postgresql
```

## Initial Configuration

### 1. Access PostgreSQL as postgres user

```bash
sudo -u postgres psql
```

You should see the PostgreSQL prompt: `postgres=#`

### 2. Create a database user (your username)

In the PostgreSQL prompt:

```sql
-- Create user with your username
CREATE USER jagadeesh WITH PASSWORD 'your_secure_password';

-- Grant privileges
ALTER USER jagadeesh CREATEDB;

-- Exit PostgreSQL
\q
```

### 3. Create the chat application database

Now as your regular user:

```bash
# Create database
createdb chat_app

# Verify database was created
psql -l | grep chat_app
```

### 4. Run Database Migrations

```bash
# Navigate to your project
cd /home/jagadeesh/Desktop/chat-app

# Run the migration script
psql chat_app < migrations/001_init.sql

# Verify tables were created
psql chat_app -c "\dt"
```

You should see two tables: `conversations` and `messages`

## Configure Authentication (if needed)

If you get authentication errors, you may need to configure PostgreSQL to allow local connections.

### Edit pg_hba.conf

```bash
# Find PostgreSQL config directory
sudo -u postgres psql -c "SHOW config_file"

# Edit pg_hba.conf (usually in same directory)
# For Ubuntu/Debian:
sudo nano /etc/postgresql/*/main/pg_hba.conf

# For other systems:
sudo nano /var/lib/pgsql/data/pg_hba.conf
```

Add or modify these lines for local development:

```
# TYPE  DATABASE        USER            ADDRESS                 METHOD
local   all             all                                     trust
host    all             all             127.0.0.1/32            trust
host    all             all             ::1/128                 trust
```

**Note:** Using `trust` is convenient for local development but not secure for production.

### Restart PostgreSQL

```bash
sudo systemctl restart postgresql
```

## Connection String Setup

### Option 1: Using your username (no password needed with trust)

```bash
# In your .env file
DATABASE_URL=postgresql://jagadeesh@localhost:5432/chat_app
```

### Option 2: Using password authentication

```bash
# In your .env file
DATABASE_URL=postgresql://jagadeesh:your_password@localhost:5432/chat_app
```

### Option 3: Using postgres superuser (not recommended for production)

```bash
# Set a password for postgres user first
sudo -u postgres psql -c "ALTER USER postgres PASSWORD 'your_password';"

# In your .env file
DATABASE_URL=postgresql://postgres:your_password@localhost:5432/chat_app
```

## Verify Database Setup

### 1. Test connection

```bash
# Test basic connection
psql chat_app -c "SELECT version();"

# Check tables exist
psql chat_app -c "SELECT tablename FROM pg_tables WHERE schemaname = 'public';"
```

Expected output:
```
 tablename
--------------
 conversations
 messages
```

### 2. Verify table structure

```bash
# Check conversations table
psql chat_app -c "\d conversations"

# Check messages table
psql chat_app -c "\d messages"
```

### 3. Test with your app

Create a test `.env` file:

```bash
# Create .env file
cat > .env << 'EOF'
DATABASE_URL=postgresql://jagadeesh@localhost:5432/chat_app
VLLM_BASE_URL=http://localhost:5000
PORT=8080
EOF

# Test database connection with Go app
go run main.go
```

You should see: "Connected to PostgreSQL database"

## Useful PostgreSQL Commands

### Database Management

```bash
# List all databases
psql -l

# Connect to chat_app database
psql chat_app

# Drop database (BE CAREFUL!)
dropdb chat_app

# Create database
createdb chat_app
```

### Inside PostgreSQL prompt (psql)

```sql
-- List all tables
\dt

-- Describe table structure
\d conversations
\d messages

-- View all conversations
SELECT * FROM conversations;

-- View all messages
SELECT * FROM messages;

-- Count conversations
SELECT COUNT(*) FROM conversations;

-- Count messages
SELECT COUNT(*) FROM messages;

-- Delete all data (for testing)
TRUNCATE TABLE messages;
TRUNCATE TABLE conversations CASCADE;

-- Exit psql
\q
```

## Troubleshooting

### PostgreSQL service won't start

```bash
# Check status
sudo systemctl status postgresql

# View logs
sudo journalctl -u postgresql -n 50

# Check if port 5432 is already in use
sudo lsof -i :5432
```

### Permission denied errors

```bash
# Ensure your user can create databases
sudo -u postgres psql -c "ALTER USER jagadeesh CREATEDB;"

# Or create user if doesn't exist
sudo -u postgres createuser --interactive
```

### Can't connect to database

```bash
# Verify PostgreSQL is listening
sudo netstat -plnt | grep 5432

# Check if database exists
psql -l | grep chat_app

# Test connection with verbose error
psql "postgresql://jagadeesh@localhost:5432/chat_app" -c "SELECT 1;"
```

### Reset everything and start fresh

```bash
# Drop database
dropdb chat_app

# Recreate database
createdb chat_app

# Run migrations again
psql chat_app < migrations/001_init.sql
```

## Database Backup and Restore

### Backup

```bash
# Backup entire database
pg_dump chat_app > chat_app_backup.sql

# Backup with timestamp
pg_dump chat_app > chat_app_backup_$(date +%Y%m%d_%H%M%S).sql
```

### Restore

```bash
# Restore from backup
psql chat_app < chat_app_backup.sql
```

## Security Best Practices

For production deployment:

1. **Use strong passwords** for database users
2. **Change pg_hba.conf** to use `md5` or `scram-sha-256` instead of `trust`
3. **Limit connections** to specific users and databases
4. **Enable SSL** for remote connections
5. **Regular backups** of your database

## Next Steps

After PostgreSQL is set up:

1. ✅ PostgreSQL installed and running
2. ✅ Database `chat_app` created
3. ✅ Tables created via migrations
4. ✅ Connection string configured in `.env`
5. → Continue with [QUICK_START.md](QUICK_START.md) to run the app

## Additional Resources

- PostgreSQL Documentation: https://www.postgresql.org/docs/
- pg_hba.conf explained: https://www.postgresql.org/docs/current/auth-pg-hba-conf.html
- PostgreSQL tutorial: https://www.postgresqltutorial.com/
