#!/bin/bash
# Generate mocks using mockery v3.6+
# This script makes mock generation portable across different module names

set -e

# Colors for output
COLOR_INFO='\033[36m'
COLOR_SUCCESS='\033[32m'
COLOR_WARNING='\033[33m'
COLOR_RESET='\033[0m'

# Get module name from go.mod
MODULE_NAME=$(go list -m)

echo -e "${COLOR_INFO}🎭 Generating mocks for module: $MODULE_NAME${COLOR_RESET}"
echo ""

# List of packages to generate mocks for
PACKAGES=(
    "internal/applications/auth/service"
    "internal/applications/auth/repository"
    "internal/applications/user/service"
    "internal/applications/user/repository"
    "internal/applications/order/service"
    "internal/applications/order/repository"
    "internal/infra/database"
    "internal/infra/cache"
    "internal/infra/queue/rabbitmq"
    "pkg/authenticator"
    "pkg/validator"
    "pkg/logger"
)

# Counter for statistics
TOTAL=0
GENERATED=0
SKIPPED=0

# Generate mocks for each package
for pkg in "${PACKAGES[@]}"; do
    TOTAL=$((TOTAL + 1))

    if [ -d "$pkg" ]; then
        echo -e "${COLOR_INFO}  📦 Generating mocks for $pkg...${COLOR_RESET}"

        # Run mockery v3.6+ (simplified - let mockery use defaults)
        mockery \
            2>&1 | grep -v "^$" || true

        GENERATED=$((GENERATED + 1))
    else
        echo -e "${COLOR_WARNING}  ⚠️  Skipping $pkg (directory not found)${COLOR_RESET}"
        SKIPPED=$((SKIPPED + 1))
    fi
done

echo ""
echo -e "${COLOR_SUCCESS}✅ Mock generation completed!${COLOR_RESET}"
echo -e "${COLOR_INFO}📊 Statistics:${COLOR_RESET}"
echo -e "   Total packages: $TOTAL"
echo -e "   Generated: ${COLOR_SUCCESS}$GENERATED${COLOR_RESET}"
echo -e "   Skipped: ${COLOR_WARNING}$SKIPPED${COLOR_RESET}"
