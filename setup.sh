#!/bin/bash

# Automated setup script for Chat-App with vLLM
# This script will guide you through the complete setup process

set -e  # Exit on error

echo "================================================"
echo "Chat-App with vLLM Setup Script"
echo "================================================"
echo ""

# Colors for output
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
RED='\033[0;31m'
NC='\033[0m' # No Color

# Function to print colored messages
print_success() {
    echo -e "${GREEN}✓ $1${NC}"
}

print_warning() {
    echo -e "${YELLOW}⚠ $1${NC}"
}

print_error() {
    echo -e "${RED}✗ $1${NC}"
}

print_info() {
    echo -e "ℹ $1"
}

# Check if running on Linux
if [[ "$OSTYPE" != "linux-gnu"* ]]; then
    print_error "This script is designed for Linux. Detected OS: $OSTYPE"
    exit 1
fi

echo "Step 1: Checking Prerequisites"
echo "------------------------------"

# Check PostgreSQL
if command -v psql &> /dev/null; then
    print_success "PostgreSQL is installed ($(psql --version | head -1))"
else
    print_warning "PostgreSQL is not installed"
    echo ""
    echo "Would you like to install PostgreSQL? (y/n)"
    read -r install_pg
    if [[ "$install_pg" == "y" ]]; then
        print_info "Installing PostgreSQL..."
        if command -v apt &> /dev/null; then
            sudo apt update
            sudo apt install -y postgresql postgresql-contrib
            sudo systemctl start postgresql
            sudo systemctl enable postgresql
            print_success "PostgreSQL installed"
        elif command -v dnf &> /dev/null; then
            sudo dnf install -y postgresql-server postgresql-contrib
            sudo postgresql-setup --initdb
            sudo systemctl start postgresql
            sudo systemctl enable postgresql
            print_success "PostgreSQL installed"
        else
            print_error "Unsupported package manager. Please install PostgreSQL manually."
            print_info "See POSTGRESQL_SETUP.md for instructions"
            exit 1
        fi
    else
        print_error "PostgreSQL is required. Please install it manually."
        print_info "See POSTGRESQL_SETUP.md for instructions"
        exit 1
    fi
fi

# Check GPU
echo ""
if command -v nvidia-smi &> /dev/null; then
    print_success "NVIDIA GPU detected:"
    nvidia-smi --query-gpu=name,memory.total --format=csv,noheader
else
    print_error "NVIDIA GPU or drivers not found"
    print_info "vLLM requires an NVIDIA GPU. Please install NVIDIA drivers."
    exit 1
fi

# Check vLLM
echo ""
if command -v vllm &> /dev/null || pipx list 2>/dev/null | grep -q vllm; then
    print_success "vLLM is installed"
else
    print_warning "vLLM is not installed"
    echo ""
    echo "Would you like to install vLLM using pipx? (y/n)"
    read -r install_vllm
    if [[ "$install_vllm" == "y" ]]; then
        if ! command -v pipx &> /dev/null; then
            print_info "Installing pipx first..."
            python3 -m pip install --user pipx
            python3 -m pipx ensurepath
        fi
        print_info "Installing vLLM (this may take a few minutes)..."
        pipx install vllm
        print_success "vLLM installed"
    else
        print_error "vLLM is required. Please install it manually."
        print_info "See VLLM_SETUP.md for instructions"
        exit 1
    fi
fi

# Check Go
echo ""
if command -v go &> /dev/null; then
    print_success "Go is installed ($(go version | awk '{print $3}'))"
else
    print_error "Go is not installed"
    print_info "Please install Go from https://go.dev/dl/"
    exit 1
fi

echo ""
echo "================================================"
echo "Step 2: Database Setup"
echo "================================================"
echo ""

# Create database
DB_NAME="chat_app"
if psql -lqt | cut -d \| -f 1 | grep -qw "$DB_NAME"; then
    print_warning "Database '$DB_NAME' already exists"
    echo "Would you like to recreate it? This will delete all existing data! (y/n)"
    read -r recreate_db
    if [[ "$recreate_db" == "y" ]]; then
        print_info "Dropping existing database..."
        dropdb "$DB_NAME" 2>/dev/null || sudo -u postgres dropdb "$DB_NAME"
        print_info "Creating database..."
        createdb "$DB_NAME" 2>/dev/null || sudo -u postgres createdb "$DB_NAME"
        print_success "Database recreated"
    fi
else
    print_info "Creating database '$DB_NAME'..."
    if createdb "$DB_NAME" 2>/dev/null; then
        print_success "Database created"
    else
        print_info "Trying with postgres user..."
        sudo -u postgres createdb "$DB_NAME"
        print_success "Database created"
    fi
fi

# Run migrations
print_info "Running database migrations..."
if psql "$DB_NAME" < migrations/001_init.sql; then
    print_success "Migrations completed"
    psql "$DB_NAME" -c "\dt"
else
    print_error "Migration failed"
    exit 1
fi

echo ""
echo "================================================"
echo "Step 3: Environment Configuration"
echo "================================================"
echo ""

# Get current username
CURRENT_USER=$(whoami)

# Create .env file
if [ -f .env ]; then
    print_warning ".env file already exists"
    echo "Would you like to overwrite it? (y/n)"
    read -r overwrite_env
    if [[ "$overwrite_env" != "y" ]]; then
        print_info "Keeping existing .env file"
    else
        print_info "Creating .env file..."
        cat > .env << EOF
# Database Configuration
DATABASE_URL=postgresql://${CURRENT_USER}@localhost:5432/chat_app

# vLLM Configuration
VLLM_BASE_URL=http://localhost:5000

# Server Configuration
PORT=8080
EOF
        print_success ".env file created"
    fi
else
    print_info "Creating .env file..."
    cat > .env << EOF
# Database Configuration
DATABASE_URL=postgresql://${CURRENT_USER}@localhost:5432/chat_app

# vLLM Configuration
VLLM_BASE_URL=http://localhost:5000

# Server Configuration
PORT=8080
EOF
    print_success ".env file created"
fi

# Load .env
source .env

echo ""
echo "================================================"
echo "Step 4: HuggingFace Setup"
echo "================================================"
echo ""

print_info "You need a HuggingFace token to download Llama 3.1 8B"
print_info "Get your token from: https://huggingface.co/settings/tokens"
echo ""
echo "Would you like to login to HuggingFace now? (y/n)"
read -r login_hf
if [[ "$login_hf" == "y" ]]; then
    if command -v pipx &> /dev/null; then
        pipx run huggingface-cli login
    else
        huggingface-cli login
    fi
    print_success "HuggingFace login complete"
fi

echo ""
echo "================================================"
echo "Setup Complete!"
echo "================================================"
echo ""
print_success "All components are installed and configured"
echo ""
echo "Next Steps:"
echo ""
echo "1. Start vLLM server (Terminal 1):"
echo "   ${GREEN}./start-vllm-rtx5090.sh${NC}"
echo ""
echo "2. Start chat application (Terminal 2):"
echo "   ${GREEN}go run main.go${NC}"
echo ""
echo "3. Open your browser:"
echo "   ${GREEN}http://localhost:8080${NC}"
echo ""
echo "For detailed instructions, see:"
echo "  - QUICK_START.md"
echo "  - VLLM_SETUP.md (for vLLM configuration)"
echo "  - POSTGRESQL_SETUP.md (for database management)"
echo ""
echo "Happy chatting! 🚀"
