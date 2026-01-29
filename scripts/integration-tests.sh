#!/bin/bash

echo "🧪 Running Integration Tests..."

# Démarrer les services de test
echo "🚀 Starting test services..."
docker-compose -f docker-compose.test.yml up -d postgres redis

# Attendre que les services soient prêts
echo "⏳ Waiting for services to be ready..."
sleep 10

# Exécuter les tests d'intégration
echo "📝 Running backend integration tests..."
cd backend
export TEST_DB_HOST=localhost
export TEST_DB_PORT=5433
export TEST_REDIS_HOST=localhost
export TEST_REDIS_PORT=6380

go test -v ./handlers/... -tags=integration
TEST_EXIT=$?

# Nettoyer
echo "🧹 Cleaning up test services..."
cd ..
docker-compose -f docker-compose.test.yml down

if [ $TEST_EXIT -ne 0 ]; then
    echo "❌ Integration tests failed!"
    exit 1
fi

echo "✅ Integration tests passed!"
exit 0
