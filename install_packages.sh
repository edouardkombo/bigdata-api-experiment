#!/usr/bin/env bash
set -euo pipefail

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
  echo 'export PATH="$PATH:$(go env GOPATH)/bin"' >> ~/.bashrc
  source ~/.bashrc
fi

# 4. Install Node.js
if ! command -v node &>/dev/null; then
  apt-get remove -y nodejs npm
  apt-get autoremove -y
  rm -rf /var/lib/apt/lists/*

  apt-get update && apt-get install -y curl gnupg ca-certificates
  
  curl -fsSL https://deb.nodesource.com/setup_22.x | sudo -E bash -
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

# 7. Install Redis
if ! command -v redis-server &>/dev/null; then
  apt-get install -y redis-server
fi
nohup redis-server --daemonize yes

# 8. Install Rabbitmq
if ! command -v rabbitmq-server &>/dev/null; then
  set -e  # exit on error

  echo "[INFO] Updating system & installing prerequisites..."
  sudo apt-get update
  sudo apt-get install -y curl gnupg apt-transport-https lsb-release

  DISTRO=$(lsb_release -cs)
  # Override if using an unsupported version like 'oracular'
  if [[ "$DISTRO" == "oracular" ]]; then
    DISTRO="jammy"
    echo "[WARN] Ubuntu 'oracular' is not supported by RabbitMQ — using 'jammy' repo instead."
  fi

  echo "[INFO] Setting up Erlang repo & key..."
  curl -fsSL https://packages.erlang-solutions.com/ubuntu/erlang_solutions.asc | gpg --dearmor | sudo tee /usr/share/keyrings/erlang.gpg > /dev/null
  echo "deb [signed-by=/usr/share/keyrings/erlang.gpg] https://packages.erlang-solutions.com/ubuntu $DISTRO contrib" | sudo tee /etc/apt/sources.list.d/erlang.list

  echo "[INFO] Setting up RabbitMQ repo & key..."
  curl -fsSL https://packagecloud.io/rabbitmq/rabbitmq-server/gpgkey | gpg --dearmor | sudo tee /usr/share/keyrings/rabbitmq.gpg > /dev/null
  echo "deb [signed-by=/usr/share/keyrings/rabbitmq.gpg] https://packagecloud.io/rabbitmq/rabbitmq-server/ubuntu $DISTRO main" | sudo tee /etc/apt/sources.list.d/rabbitmq.list

  echo "[INFO] Installing Erlang & RabbitMQ..."
  sudo apt-get update
  sudo apt-get install -y erlang rabbitmq-server

  echo "[INFO] Enabling and starting RabbitMQ service..."
  sudo systemctl enable rabbitmq-server
  sudo systemctl start rabbitmq-server
  sudo systemctl status rabbitmq-server --no-pager

  echo "[INFO] Enabling RabbitMQ management plugin..."
  sudo rabbitmq-plugins enable rabbitmq_management

  echo "[DONE] RabbitMQ installed and running on port 15672"
else
  echo "[SKIP] RabbitMQ is already installed."
fi

# Prepare Go
cd backend_go
apt-get install -y protobuf-compiler >/dev/null
rm go.sum go.mod
go mod init bigdata-perf
go get github.com/ClickHouse/clickhouse-go
go get github.com/rabbitmq/amqp091-go
go get github.com/go-chi/cors
go get github.com/go-chi/chi/v5
go get github.com/go-chi/chi/v5/middleware
go get github.com/joho/godotenv
go get google.golang.org/grpc
go get google.golang.org/protobuf
go mod tidy

pkill -f "go run cmd/consumer/main.go"
nohup go run cmd/consumer/main.go > logs/consumer.log 2>&1 &
go run main.go --recreate-table

PIDS=$(sudo lsof -t -i :8088 || true)
if [ -n "$PIDS" ]; then
  echo "🔪 Killing process on port 8088: $PIDS"
  sudo kill -9 $PIDS
else
  echo "✅ No process running on port 8088"
fi

nohup go run cmd/api/main.go > logs/api.log 2>&1 &

cd ../
chmod +x ./setup.sh
