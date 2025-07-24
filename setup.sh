#!/usr/bin/env bash
set -euo pipefail

# Variables (override by exporting before running)
ROWS=${ROWS:-1000000000}
BATCH=${BATCH:-10000}

# 1. Purge stale ClickHouse APT entries
rm -f /etc/apt/sources.list.d/*clickhouse*.list
sed -i '/repo.clickhouse\.com/d;/packages\.clickhouse\.com\/deb/d' /etc/apt/sources.list || true

# 2. Install system prerequisites
apt-get update
DEPS=(curl gnupg lsb-release tmux git software-properties-common \
      apt-transport-https ca-certificates dirmngr python3 python3-venv)
apt-get install -y "${DEPS[@]}"

# 3. Install Go
if ! command -v go &>/dev/null; then
  add-apt-repository -y ppa:longsleep/golang-backports
  apt-get update
  apt-get install -y golang-go
fi

# 4. Install Node.js
if ! command -v node &>/dev/null; then
  curl -fsSL https://deb.nodesource.com/setup_18.x | bash -
  apt-get install -y nodejs npm
fi

# 5. Add & install ClickHouse
echo "deb [trusted=yes] https://packages.clickhouse.com/deb stable main" \
  | tee /etc/apt/sources.list.d/clickhouse.list
apt-get update
apt-get install -y clickhouse-server clickhouse-client

# 6. Reset default ClickHouse password if set
if [ -f /etc/clickhouse-server/users.d/default-password.xml ]; then
  systemctl stop clickhouse-server
  rm -f /etc/clickhouse-server/users.d/default-password.xml
  systemctl start clickhouse-server
else
  service clickhouse-server start
fi

# 7. Kafka & Zookeeper (Confluent OSS)
rm -f /etc/apt/sources.list.d/confluent*.list
echo "deb [trusted=yes] https://packages.confluent.io/deb/7.5 stable main" \
  | tee /etc/apt/sources.list.d/confluent.list
apt-get update
apt-get install -y confluent-community-2.13
nohup zookeeper-server-start /etc/kafka/zookeeper.properties > /tmp/zk.log 2>&1 &
nohup kafka-server-start     /etc/kafka/server.properties      > /tmp/kafka.log   2>&1 &

# 8. Install Redis
if ! command -v redis-server &>/dev/null; then
  apt-get install -y redis-server
fi
nohup redis-server --daemonize yes

# 9. Initialize ClickHouse schema
clickhouse-client --query="CREATE DATABASE IF NOT EXISTS analytics;"
clickhouse-client --query="
CREATE TABLE IF NOT EXISTS analytics.page_events (
    id String,
    user_id String,
    event_type String,
    url String,
    referrer String,
    ts DateTime,
    meta String
) ENGINE = MergeTree() ORDER BY ts;
"

# 10. Build & launch Go services
cd backend_go
go mod init project || true
go mod tidy
go get github.com/go-chi/chi/v5
go get github.com/go-chi/cors
go build -o seeder     ./cmd/seeder
go build -o ingest     ./cmd/ingest
go build -o api-gateway ./cmd/api-gateway
nohup ./seeder       --rows "$ROWS" --batch "$BATCH" > /tmp/seeder.log     2>&1 &
nohup ./ingest                          > /tmp/ingest.log     2>&1 &
nohup ./api-gateway                     > /tmp/api-gateway.log 2>&1 &
cd ..

# 11. Python seed in virtualenvs (ONLY TO TEST THE TIME SPEED DIFFERENCE AGAINST GO)
# 11a. Seeder: clickhouse-driver based
cd backend_python
python3 -m venv venv
source venv/bin/activate
pip install --upgrade pip
pip install faker clickhouse-driver
#python seed.py "$ROWS"
deactivate

curl -fsSL https://deb.nodesource.com/setup_22.x | sudo -E bash -
sudo apt-get install -y nodejs

# 12. Install & run SvelteKit frontend
cd frontend
rm -rf node_modules package-lock.json
sudo apt autoremove
npm install --save-dev svelte-preprocess
npm install --save-dev typescript --legacy-peer-deps
npm install --save-dev svelte-check --legacy-peer-deps
npm install chart.js --legacy-peer-deps
npm install chartjs-adapter-date-fns date-fns --legacy-peer-deps
npm install @tanstack/svelte-virtual@3.13.12 --legacy-peer-deps
npm install --legacy-peer-deps
nohup npm run dev --host > /tmp/frontend.log 2>&1 &

echo "✅ Setup complete! Dashboard live at http://localhost:3000"

