.PHONY: help build up down logs clean restart

help:
    @echo "SaferCloud Docker Commands:"
    @echo "  make build       - Build all Docker images"
    @echo "  make up          - Start all services"
    @echo "  make down        - Stop all services"
    @echo "  make logs        - Show logs (all services)"
    @echo "  make restart     - Restart all services"
    @echo "  make clean       - Remove containers, volumes and images"
    @echo "  make health      - Health check all services"

build:
    docker-compose -f docker-compose.prod.yml build

up:
    docker-compose -f docker-compose.prod.yml up -d

down:
    docker-compose -f docker-compose.prod.yml down

logs:
    docker-compose -f docker-compose.prod.yml logs -f

logs-backend:
    docker-compose -f docker-compose.prod.yml logs -f backend

logs-frontend:
    docker-compose -f docker-compose.prod.yml logs -f frontend

restart:
    docker-compose -f docker-compose.prod.yml restart

clean:
    docker-compose -f docker-compose.prod.yml down -v --rmi all

health:
    @echo "Backend API:"
    @curl -s http://localhost:8080/health || echo "Backend not responding"
    @echo "\nMetrics:"
    @curl -s http://localhost:9090/health || echo "Metrics not responding"
    @echo "\nFrontend:"
    @curl -s http://localhost/health || echo "Frontend not responding"
    @echo "\nPrometheus:"
    @curl -s http://localhost:9091/-/healthy || echo "Prometheus not responding"
