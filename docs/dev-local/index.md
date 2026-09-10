# Développement local (sans Docker)

Pour contribuer au code ou tester rapidement en local, sans Docker :

```bash
git clone https://github.com/Buuuntyyy/Kagibi.git
cd Kagibi
cp backend/.env.example backend/.env     # renseigner S3_*, JWT_SECRET, etc.
cp frontend/.env.example frontend/.env   # VITE_API_URL=http://localhost:8080/api/v1

cd backend
go run main.go

cd frontend
npm install
npm run dev
```

Frontend : `http://localhost` — Backend : `http://localhost:8080`

Ce mode nécessite malgré tout un stockage S3 (un service externe, ou une instance MinIO/Garage que vous faites tourner vous-même en parallèle) — consultez `backend/.env.example` pour le détail des variables. Pour un déploiement complet incluant base de données et cache, préférez la section [Self-hosting](/self-hosting/prerequisites), qui repose sur Docker Compose.
