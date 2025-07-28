#!/usr/bin/env bash
set -euo pipefail

if pidof setup.sh >/dev/null; then
  echo "Script is already running"
  exit 1
fi


if [ -z "$1" ]; then
  echo "[ERROR] Missing argument: number of rows to mock"
  echo "Usage: ./setup.sh <number_of_rows>"
  exit 1
fi

ROWS=$1

# Ensure script is run from the project root
set -e

INITIAL_ROWS=100000

echo "==> Seeding $INITIAL_ROWS entries using Go + RabbitMq first..."

# Step 1: Seed with Go
cd backend_go

PIDS=$(sudo lsof -t -i :50051 || true)
if [ -n "$PIDS" ]; then
  echo "🔪 Killing process on port 50051: $PIDS"
  sudo kill -9 $PIDS
else
  echo "✅ No process running on port 50051"
fi
PIDS=$(sudo lsof -t -i :8080 || true)
if [ -n "$PIDS" ]; then
  echo "🔪 Killing process on port 8080: $PIDS"
  sudo kill -9 $PIDS
else
  echo "✅ No process running on port 8080"
fi

PIDS=$(sudo lsof -t -i :8088 || true)
if [ -n "$PIDS" ]; then
  echo "🔪 Killing process on port 8088: $PIDS"
  sudo kill -9 $PIDS
else
  echo "✅ No process running on port 8088"
fi

nohup go run cmd/consumer/main.go > logs/consumer.log 2>&1 &
nohup go run cmd/api/main.go > logs/api.log 2>&1 &
start_go=$(date +%s%3N)
go run main.go --only-seed --seed-count=$INITIAL_ROWS > ./logs/main.log
cd ..

end_go=$(date +%s%3N)
go_time=$((end_go - start_go))

echo "==> Go seeding completed in $go_time ms"

echo ""
echo "==> Seeding $INITIAL_ROWS entries using Python now..."
start_py=$(date +%s%3N)

# Step 2: Seed with Python
cd backend_python
python3 -m venv venv
source venv/bin/activate
pip install --upgrade pip >/dev/null
pip install faker clickhouse-driver >/dev/null
python seed.py "$INITIAL_ROWS"
deactivate
cd ..

end_py=$(date +%s%3N)
py_time=$((end_py - start_py))

echo "==> Python seeding completed in $py_time ms"
echo ""

# Step 3: Frontend
echo "==> Setting up frontend before"

PIDS=$(sudo lsof -t -i :5173 || true)
if [ -n "$PIDS" ]; then
  echo "🔪 Killing process on port 5173: $PIDS"
  sudo kill -9 $PIDS
else
  echo "✅ No process running on port 5173"
fi

cd frontend
mkdir -p logs
rm -rf node_modules package-lock.json
# Dev-only tools
npm install --save-dev --legacy-peer-deps svelte-preprocess typescript svelte-check dotenv
# Runtime libraries
npm install --legacy-peer-deps chart.js chartjs-adapter-date-fns date-fns
nohup npm run dev --host > ./logs/frontend.log 2>&1 &
echo "✅ Setup complete! Dashboard live at http://localhost:5173"
cd ../

# Step 4: Summary
echo "====== Seeding Time Summary ======"
echo "Go     : $go_time ms"
echo "Python : $py_time ms"
echo "=================================="
echo ""

# Step 5: Ask user
read -p "Which language do you want to use for seeding big data rows? (go/python): " choice

# Step 6: Background seeding
if [ "$choice" = "go" ]; then
  echo "==> Running full Go seed..."
  cd backend_go
  nohup go run main.go --only-seed --seed-count=$ROWS > ./logs/main.log &
  cd ..
elif [ "$choice" = "python" ]; then
  echo "==> Running full Python seed..."
  cd backend_python
  source venv/bin/activate || source venv/bin/activate
  nohup python seed.py $ROWS &
  deactivate
  cd ..
else
  echo "Invalid choice. No seeding performed."
fi

echo "🌱 Seeding in background with '$choice' for $ROWS rows... check logs in ./backend_$choice/logs/"


