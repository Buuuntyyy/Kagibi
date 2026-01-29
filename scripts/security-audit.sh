#!/bin/bash

echo "🔒 Running Security Audit..."

# Vérifier les vulnérabilités des dépendances
echo "📦 Checking npm dependencies..."
cd frontend
npm audit --audit-level=high
FRONTEND_EXIT=$?

echo "📦 Checking Go dependencies..."
cd ../backend
go list -json -m all | docker run --rm -i sonatypecommunity/nancy:latest sleuth
BACKEND_EXIT=$?

# Vérifier les secrets exposés
echo "🔍 Checking for exposed secrets..."
if grep -r "password.*=" --include="*.go" --include="*.js" --include="*.vue" . | grep -v "test" | grep -v "example"; then
    echo "⚠️  Potential hardcoded passwords found!"
    exit 1
fi

if [ $FRONTEND_EXIT -ne 0 ] || [ $BACKEND_EXIT -ne 0 ]; then
    echo "❌ Security audit failed!"
    exit 1
fi

echo "✅ Security audit passed!"
exit 0
