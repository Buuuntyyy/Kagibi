# SaferCloud Kubernetes Deployment Guide

Ce guide décrit comment déployer SaferCloud sur un cluster Kubernetes avec haute disponibilité (HA) et load balancing.

## Architecture

L'architecture comprend :
- **Backend** : 3+ réplicas avec autoscaling (Go)
- **Frontend** : 3+ réplicas avec autoscaling (Vue.js)
- **Website** : 2+ réplicas avec autoscaling (Site vitrine)
- **PostgreSQL** : 3 réplicas en StatefulSet
- **Redis** : 3 réplicas en StatefulSet
- **Ingress** : NGINX avec SSL/TLS et load balancing
- **Monitoring** : Prometheus + AlertManager
- **Backup** : CronJob automatique pour PostgreSQL

## Prérequis

1. **Cluster Kubernetes** (v1.24+)
   - AWS EKS, Google GKE, Azure AKS, ou on-premise
   - Au moins 3 nœuds worker pour la HA

2. **kubectl** configuré pour votre cluster

3. **Helm** (v3+) installé

4. **Cert-Manager** pour les certificats SSL/TLS
   ```bash
   kubectl apply -f https://github.com/cert-manager/cert-manager/releases/download/v1.13.0/cert-manager.yaml
   ```

5. **NGINX Ingress Controller**
   ```bash
   helm repo add ingress-nginx https://kubernetes.github.io/ingress-nginx
   helm repo update
   helm install nginx-ingress ingress-nginx/ingress-nginx \
     --namespace ingress-nginx \
     --create-namespace \
     --set controller.service.type=LoadBalancer
   ```

6. **Metrics Server** pour l'autoscaling
   ```bash
   kubectl apply -f https://github.com/kubernetes-sigs/metrics-server/releases/latest/download/components.yaml
   ```

7. **Prometheus Operator** (optionnel, pour monitoring)
   ```bash
   helm repo add prometheus-community https://prometheus-community.github.io/helm-charts
   helm install prometheus prometheus-community/kube-prometheus-stack \
     --namespace monitoring \
     --create-namespace
   ```

## Configuration

### 1. Modifier les Secrets

Éditez `k8s/security/secrets.yaml` et remplacez les valeurs encodées en base64 :

```bash
# Encoder une valeur en base64
echo -n 'votre-valeur' | base64

# Décoder une valeur base64
echo 'dmFsZXVy' | base64 -d
```

**Variables à configurer :**
- `POSTGRES_USER` et `POSTGRES_PASSWORD`
- `JWT_SECRET`
- `AWS_ACCESS_KEY_ID` et `AWS_SECRET_ACCESS_KEY`
- `SUPABASE_URL` et `SUPABASE_KEY`
- `STRIPE_SECRET_KEY` et `STRIPE_WEBHOOK_SECRET`

### 2. Modifier les ConfigMaps

Éditez `k8s/config/configmap.yaml` pour ajuster :
- `VITE_API_URL` : URL de votre API
- Autres configurations selon votre environnement

### 3. Modifier l'Ingress

Éditez `k8s/ingress/ingress.yaml` :
- Remplacez `safercloud.com` par votre domaine
- Remplacez `admin@safercloud.com` par votre email

### 4. Ajuster les StorageClass

Modifiez les `storageClassName` dans les PVC selon votre cloud provider :
- **AWS EKS** : `gp3` ou `gp2`
- **Google GKE** : `standard-rwo` ou `premium-rwo`
- **Azure AKS** : `managed-premium` ou `managed`
- **On-premise** : Créez vos propres StorageClass

## Déploiement

### Option 1 : Déploiement complet

```bash
# 1. Créer le namespace
kubectl apply -f k8s/config/namespace.yaml

# 2. Créer les ConfigMaps et Secrets
kubectl apply -f k8s/config/configmap.yaml
kubectl apply -f k8s/security/secrets.yaml

# 3. Créer les PersistentVolumeClaims
kubectl apply -f k8s/storage/persistent-volumes.yaml

# 4. Déployer les bases de données
kubectl apply -f k8s/database/postgres-statefulset.yaml
kubectl apply -f k8s/database/redis-statefulset.yaml

# Attendre que les bases de données soient prêtes
kubectl wait --for=condition=ready pod -l app=postgres -n safercloud --timeout=300s
kubectl wait --for=condition=ready pod -l app=redis -n safercloud --timeout=300s

# 5. Déployer les applications
kubectl apply -f k8s/backend/deployment.yaml
kubectl apply -f k8s/frontend/deployment.yaml
kubectl apply -f k8s/website/deployment.yaml

# 6. Configurer l'autoscaling
kubectl apply -f k8s/autoscaling/hpa.yaml

# 7. Configurer les PodDisruptionBudgets
kubectl apply -f k8s/config/pod-disruption-budgets.yaml

# 8. Configurer la sécurité réseau
kubectl apply -f k8s/security/network-policies.yaml

# 9. Déployer l'Ingress
kubectl apply -f k8s/ingress/ingress.yaml

# 10. Configurer le monitoring (optionnel)
kubectl apply -f k8s/monitoring/prometheus-rules.yaml

# 11. Configurer les backups
kubectl apply -f k8s/backup/backup-cronjob.yaml
```

### Option 2 : Déploiement avec un seul script

Créez un fichier `deploy.sh` :

```bash
#!/bin/bash
set -e

echo "🚀 Déploiement de SaferCloud sur Kubernetes..."

# Appliquer tous les fichiers dans l'ordre
kubectl apply -f k8s/config/namespace.yaml
kubectl apply -f k8s/config/configmap.yaml
kubectl apply -f k8s/security/secrets.yaml
kubectl apply -f k8s/storage/persistent-volumes.yaml
kubectl apply -f k8s/database/
kubectl apply -f k8s/backend/
kubectl apply -f k8s/frontend/
kubectl apply -f k8s/website/
kubectl apply -f k8s/autoscaling/
kubectl apply -f k8s/config/pod-disruption-budgets.yaml
kubectl apply -f k8s/security/network-policies.yaml
kubectl apply -f k8s/ingress/
kubectl apply -f k8s/monitoring/
kubectl apply -f k8s/backup/

echo "✅ Déploiement terminé !"
echo ""
echo "Vérification des pods :"
kubectl get pods -n safercloud
```

Puis exécutez :
```bash
chmod +x deploy.sh
./deploy.sh
```

## Build et Push des Images Docker

Avant de déployer, vous devez construire et pousser vos images Docker :

### Backend

```bash
cd backend
docker build -t votre-registry/safercloud/backend:latest .
docker push votre-registry/safercloud/backend:latest
```

### Frontend

```bash
cd frontend
docker build -t votre-registry/safercloud/frontend:latest .
docker push votre-registry/safercloud/frontend:latest
```

### Website

```bash
cd website  # ou git/website selon votre structure
docker build -t votre-registry/safercloud/website:latest .
docker push votre-registry/safercloud/website:latest
```

Puis mettez à jour les images dans les fichiers de déploiement.

## Vérification

### Vérifier les pods

```bash
kubectl get pods -n safercloud
```

Tous les pods doivent être en état `Running`.

### Vérifier les services

```bash
kubectl get svc -n safercloud
```

### Vérifier l'Ingress

```bash
kubectl get ingress -n safercloud
```

Notez l'adresse IP externe de l'Ingress.

### Vérifier les certificats SSL

```bash
kubectl get certificates -n safercloud
```

Les certificats doivent être en état `Ready`.

### Vérifier l'autoscaling

```bash
kubectl get hpa -n safercloud
```

### Voir les logs

```bash
# Backend
kubectl logs -f deployment/backend -n safercloud

# Frontend
kubectl logs -f deployment/frontend -n safercloud

# Website
kubectl logs -f deployment/website -n safercloud

# PostgreSQL
kubectl logs -f statefulset/postgres -n safercloud

# Redis
kubectl logs -f statefulset/redis -n safercloud
```

## Configuration DNS

Pointez vos domaines vers l'adresse IP de l'Ingress :

```
A    safercloud.com          -> <INGRESS_IP>
A    www.safercloud.com      -> <INGRESS_IP>
A    app.safercloud.com      -> <INGRESS_IP>
A    api.safercloud.com      -> <INGRESS_IP>
```

Ou utilisez un CNAME si votre cloud provider fournit un hostname.

## Monitoring

### Prometheus

Accédez à Prometheus via port-forward :
```bash
kubectl port-forward -n monitoring svc/prometheus-kube-prometheus-prometheus 9090:9090
```

Ouvrez http://localhost:9090

### Grafana

Accédez à Grafana :
```bash
kubectl port-forward -n monitoring svc/prometheus-grafana 3000:80
```

Login par défaut : `admin` / `prom-operator`

Ouvrez http://localhost:3000

## Mise à jour

Pour mettre à jour une application :

```bash
# 1. Build et push la nouvelle image
docker build -t votre-registry/safercloud/backend:v2.0 .
docker push votre-registry/safercloud/backend:v2.0

# 2. Mettre à jour le déploiement
kubectl set image deployment/backend backend=votre-registry/safercloud/backend:v2.0 -n safercloud

# 3. Vérifier le rollout
kubectl rollout status deployment/backend -n safercloud

# 4. En cas de problème, rollback
kubectl rollout undo deployment/backend -n safercloud
```

## Scaling manuel

```bash
# Scale le backend à 5 réplicas
kubectl scale deployment/backend --replicas=5 -n safercloud

# Scale le frontend à 4 réplicas
kubectl scale deployment/frontend --replicas=4 -n safercloud
```

## Backup et Restore

### Backup manuel

```bash
# Créer un backup
kubectl create job --from=cronjob/postgres-backup manual-backup-$(date +%Y%m%d) -n safercloud
```

### Restore depuis un backup

```bash
# 1. Copier le fichier de backup
kubectl cp safercloud/postgres-0:/backups/backup-20260210-020000.sql.gz ./backup.sql.gz

# 2. Décompresser
gunzip backup.sql.gz

# 3. Restore
kubectl exec -it postgres-0 -n safercloud -- psql -U user -d mydb < backup.sql
```

## Troubleshooting

### Pod en CrashLoopBackOff

```bash
# Voir les logs
kubectl logs <pod-name> -n safercloud --previous

# Décrire le pod
kubectl describe pod <pod-name> -n safercloud
```

### Pods ne démarrent pas

```bash
# Vérifier les events
kubectl get events -n safercloud --sort-by='.lastTimestamp'

# Vérifier les ressources
kubectl top nodes
kubectl top pods -n safercloud
```

### Ingress ne fonctionne pas

```bash
# Vérifier l'Ingress Controller
kubectl get pods -n ingress-nginx

# Vérifier les logs
kubectl logs -n ingress-nginx deployment/nginx-ingress-controller
```

### Certificats SSL ne sont pas créés

```bash
# Vérifier cert-manager
kubectl get pods -n cert-manager

# Vérifier les logs
kubectl logs -n cert-manager deployment/cert-manager

# Vérifier les certificats
kubectl describe certificate <cert-name> -n safercloud
```

## Nettoyage

Pour supprimer tout le déploiement :

```bash
# Supprimer le namespace (supprime tout)
kubectl delete namespace safercloud

# Ou supprimer individuellement
kubectl delete -f k8s/ingress/
kubectl delete -f k8s/monitoring/
kubectl delete -f k8s/backup/
kubectl delete -f k8s/security/network-policies.yaml
kubectl delete -f k8s/config/pod-disruption-budgets.yaml
kubectl delete -f k8s/autoscaling/
kubectl delete -f k8s/website/
kubectl delete -f k8s/frontend/
kubectl delete -f k8s/backend/
kubectl delete -f k8s/database/
kubectl delete -f k8s/storage/
kubectl delete -f k8s/security/secrets.yaml
kubectl delete -f k8s/config/configmap.yaml
kubectl delete -f k8s/config/namespace.yaml
```

## Sécurité

### Best Practices

1. **Secrets** : Utilisez un gestionnaire de secrets externe (HashiCorp Vault, AWS Secrets Manager, etc.)
2. **RBAC** : Configurez des rôles et permissions appropriés
3. **Network Policies** : Activées par défaut dans ce déploiement
4. **Pod Security Standards** : Appliquez les standards de sécurité
5. **Image Scanning** : Scannez vos images pour les vulnérabilités
6. **TLS** : Utilisez toujours HTTPS avec certificats valides

## Performance

### Optimisations

1. **Resource Requests/Limits** : Ajustez selon vos besoins réels
2. **HPA** : Configurez l'autoscaling selon votre charge
3. **Node Affinity** : Distribuez les pods sur différents nœuds
4. **PodDisruptionBudgets** : Configurés pour maintenir la disponibilité
5. **ReadinessProbes** : Assurent que seuls les pods prêts reçoivent du trafic

## Coûts

### Estimation des ressources

**Minimum (développement)** :
- 3 nœuds : 2 vCPU, 8 GB RAM chacun
- Storage : ~200 GB

**Recommandé (production)** :
- 5+ nœuds : 4 vCPU, 16 GB RAM chacun
- Storage : ~500 GB

Ajustez selon votre charge et utilisez l'autoscaling de cluster (Cluster Autoscaler).

## Support

Pour toute question ou problème :
1. Consultez les logs : `kubectl logs <pod> -n safercloud`
2. Vérifiez les events : `kubectl get events -n safercloud`
3. Consultez la documentation Kubernetes : https://kubernetes.io/docs/
4. Contactez votre équipe DevOps

---

**Note** : Ce déploiement est configuré pour la haute disponibilité en production. Pour un environnement de développement, vous pouvez réduire le nombre de réplicas et les ressources allouées.
