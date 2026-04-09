#!/bin/bash

# SaferCloud Kubernetes Deployment Script
# This script deploys the entire SaferCloud application to Kubernetes

set -e

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# Check if kubectl is installed
if ! command -v kubectl &> /dev/null; then
    print_error "kubectl is not installed. Please install kubectl first."
    exit 1
fi

# Check if cluster is accessible
if ! kubectl cluster-info &> /dev/null; then
    print_error "Cannot connect to Kubernetes cluster. Please check your kubectl configuration."
    exit 1
fi

print_info "Starting SaferCloud deployment to Kubernetes..."
echo ""

# 1. Create namespace
print_info "Creating namespace..."
kubectl apply -f config/namespace.yaml
print_success "Namespace created"
echo ""

# 2. Create ConfigMaps and Secrets
print_info "Creating ConfigMaps and Secrets..."
kubectl apply -f config/configmap.yaml
kubectl apply -f security/secrets.yaml
print_success "Configuration applied"
echo ""

# 3. Create Storage
print_info "Creating PersistentVolumeClaims..."
kubectl apply -f storage/persistent-volumes.yaml
print_success "Storage configured"
echo ""

# 4. Deploy Databases
print_info "Deploying PostgreSQL StatefulSet..."
kubectl apply -f database/postgres-statefulset.yaml

print_info "Deploying Redis StatefulSet..."
kubectl apply -f database/redis-statefulset.yaml

print_warning "Waiting for databases to be ready (this may take a few minutes)..."
kubectl wait --for=condition=ready pod -l app=postgres -n safercloud --timeout=600s || true
kubectl wait --for=condition=ready pod -l app=redis -n safercloud --timeout=600s || true
print_success "Databases are ready"
echo ""

# 5. Deploy Applications
print_info "Deploying Backend..."
kubectl apply -f backend/deployment.yaml

print_info "Deploying Frontend..."
kubectl apply -f frontend/deployment.yaml

print_info "Deploying Website..."
kubectl apply -f website/deployment.yaml

print_warning "Waiting for applications to be ready..."
sleep 30
print_success "Applications deployed"
echo ""

# 6. Configure Autoscaling
print_info "Configuring HorizontalPodAutoscalers..."
kubectl apply -f autoscaling/hpa.yaml
print_success "Autoscaling configured"
echo ""

# 7. Configure PodDisruptionBudgets
print_info "Configuring PodDisruptionBudgets..."
kubectl apply -f config/pod-disruption-budgets.yaml
print_success "PodDisruptionBudgets configured"
echo ""

# 8. Configure Network Security
print_info "Applying Network Policies..."
kubectl apply -f security/network-policies.yaml
print_success "Network security configured"
echo ""

# 9. Deploy Ingress
print_info "Deploying Ingress..."
kubectl apply -f ingress/ingress.yaml
print_success "Ingress deployed"
echo ""

# 10. Configure Monitoring (optional)
print_info "Configuring Monitoring..."
kubectl apply -f monitoring/prometheus-rules.yaml 2>/dev/null || print_warning "Monitoring configuration skipped (Prometheus may not be installed)"
echo ""

# 11. Configure Backups
print_info "Configuring Backup CronJob..."
kubectl apply -f backup/backup-cronjob.yaml
print_success "Backup configured"
echo ""

# Display deployment status
print_success "Deployment completed successfully!"
echo ""
echo "=================================================="
echo "📊 Deployment Status"
echo "=================================================="
echo ""

print_info "Pods:"
kubectl get pods -n safercloud
echo ""

print_info "Services:"
kubectl get svc -n safercloud
echo ""

print_info "Ingress:"
kubectl get ingress -n safercloud
echo ""

print_info "HorizontalPodAutoscalers:"
kubectl get hpa -n safercloud
echo ""

print_info "PersistentVolumeClaims:"
kubectl get pvc -n safercloud
echo ""

echo "=================================================="
echo "📝 Next Steps"
echo "=================================================="
echo ""
echo "1. Update your DNS records to point to the Ingress IP address"
echo "2. Wait for SSL certificates to be issued (check with: kubectl get certificates -n safercloud)"
echo "3. Monitor your deployment with: kubectl get pods -n safercloud -w"
echo "4. View logs with: kubectl logs -f deployment/<app-name> -n safercloud"
echo "5. Access Prometheus metrics (if configured)"
echo ""

# Get Ingress IP
INGRESS_IP=$(kubectl get ingress safercloud-ingress -n safercloud -o jsonpath='{.status.loadBalancer.ingress[0].ip}' 2>/dev/null || echo "pending...")
if [ "$INGRESS_IP" != "pending..." ]; then
    print_success "Ingress IP: $INGRESS_IP"
    echo ""
    echo "Add these DNS records:"
    echo "  safercloud.com     A  $INGRESS_IP"
    echo "  www.safercloud.com A  $INGRESS_IP"
    echo "  app.safercloud.com A  $INGRESS_IP"
    echo "  api.safercloud.com A  $INGRESS_IP"
else
    print_warning "Ingress IP is still pending. Check back in a few minutes."
fi

echo ""
print_success "🎉 SaferCloud is now deployed!"
