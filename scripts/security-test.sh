#!/bin/bash

# Security testing script for VPS Monitor
# This script runs various security checks as part of the CI/CD pipeline

set -e

echo "Running security tests for VPS Monitor..."

# 1. Run Go security scanner (gosec)
echo "1. Running Go security scanner (gosec)..."
if command -v gosec &> /dev/null; then
    gosec ./...
else
    echo "gosec not found, installing..."
    go install github.com/securego/gosec/v2/cmd/gosec@latest
    gosec ./...
fi

# 2. Check for hardcoded secrets
echo "2. Checking for hardcoded secrets..."
if command -v gitleaks &> /dev/null; then
    gitleaks detect --source . --no-git
else
    echo "gitleaks not found, skipping secret detection"
    echo "Note: Install gitleaks for secret detection: https://github.com/gitleaks/gitleaks"
fi

# 3. Run dependency vulnerability scanner (govulncheck)
echo "3. Running dependency vulnerability scanner (govulncheck)..."
if command -v govulncheck &> /dev/null; then
    govulncheck ./...
else
    echo "govulncheck not found, installing..."
    go install golang.org/x/vuln/cmd/govulncheck@latest
    govulncheck ./...
fi

# 4. Check for outdated dependencies
echo "4. Checking for outdated dependencies..."
if command -v go list &> /dev/null; then
    echo "Checking for outdated dependencies..."
    go list -u -m all | grep -v "github.com/username/vps-monitor"
else
    echo "go command not found, skipping dependency check"
fi

# 5. Run basic security tests
echo "5. Running basic security tests..."
go test -v ./tests/... -run "Security"

echo "Security tests completed successfully!"