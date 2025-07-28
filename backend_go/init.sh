sudo lsof -t -i :8088 | xargs -r sudo kill -9
go run main.go --seed-count=$1
nohup go run cmd/api/main.go > logs/api.log 2>&1 &
