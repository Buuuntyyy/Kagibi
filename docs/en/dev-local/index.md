# Local development (no Docker)

To contribute to the code or quickly try it locally, without Docker:

```bash
git clone https://github.com/Buuuntyyy/Kagibi.git
cd Kagibi
cp backend/.env.example backend/.env     # fill in S3_*, JWT_SECRET, etc.
cp frontend/.env.example frontend/.env   # VITE_API_URL=http://localhost:8080/api/v1

cd backend
go run main.go

cd frontend
npm install
npm run dev
```

Frontend: `http://localhost` — Backend: `http://localhost:8080`

This mode still needs an S3 store (external, or a MinIO/Garage instance you run yourself separately) — see `backend/.env.example` for the variable details. For a full deployment with the database and cache included, use Docker Compose below instead.
