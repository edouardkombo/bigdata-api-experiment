package main

import (
	"bufio"
	"bytes"
	"database/sql"
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"math/rand"
	"net"
	"net/http"
	"os"
	"os/exec"
	"strconv"
	"strings"
	"time"
	"github.com/joho/godotenv"

	_ "github.com/ClickHouse/clickhouse-go"
)

var (
    seedCount     int	
    onlySeed      bool
    recreateTable bool
    runServices   bool
    httpPort      string
    grpcPort      string
    requiredPorts []int
)

func init() {
    httpPort = os.Getenv("HTTP_PORT")
    if httpPort == "" {
        httpPort = "8080"
    }

    grpcPort = os.Getenv("GRPC_PORT")
    if grpcPort == "" {
        grpcPort = "50051"
    }

    requiredPorts = []int{
        parsePort(httpPort),
        parsePort(grpcPort),
    }
}

func execShell(name string, args ...string) {
	cmd := exec.Command(name, args...)
	var out bytes.Buffer
	var stderr bytes.Buffer
	cmd.Stdout = &out
	cmd.Stderr = &stderr
	err := cmd.Run()
	if err != nil {
		log.Fatalf("ERROR: %v: %s", err, stderr.String())
	}
	fmt.Print(out.String())
}

func killProcessesOnPorts(ports []int) {
	for _, port := range ports {
		conn, err := net.DialTimeout("tcp", "localhost:"+strconv.Itoa(port), time.Second)
		if err == nil {
			conn.Close()
			fmt.Printf("⚠️  Port %d is in use. Attempting to kill process...\n", port)
			out, err := exec.Command("bash", "-c", fmt.Sprintf("lsof -ti:%d | xargs kill -9", port)).CombinedOutput()
			if err != nil {
				fmt.Printf("Failed to kill process on port %d: %v\n%s\n", port, err, out)
			} else {
				fmt.Printf("✅ Killed process on port %d\n", port)
			}
		}
	}
}

func checkClickhouseConnection() {
    err := exec.Command("clickhouse-client", "--query", "SELECT 1").Run()
    if err != nil {
        log.Fatalf("❌ ClickHouse client is not working: %v", err)
    }
    fmt.Println("✅ ClickHouse client is working.")
}

func recreateTableIfConfirmed() {
	db, err := sql.Open("clickhouse", "tcp://127.0.0.1:9000?debug=false")
	if err != nil {
		log.Fatalf("❌ ClickHouse connect error: %v", err)
	}
	defer db.Close()

	row := db.QueryRow("EXISTS TABLE page_events")
	var exists uint8
	if err := row.Scan(&exists); err != nil {
		log.Fatalf("❌ Check table error: %v", err)
	}

	if exists == 1 {
		fmt.Print("⚠️  Table 'page_events' exists. Recreate? [y/N]: ")
		reader := bufio.NewReader(os.Stdin)
		response, _ := reader.ReadString('\n')

		if strings.TrimSpace(strings.ToLower(response)) == "y" {
                    execShell("clickhouse-client", "--query", "DROP TABLE IF EXISTS analytics.page_events")
                    execShell("clickhouse-client", "--multiquery", "--queries-file=clickhouse/init.sql")
                    fmt.Println("✅ Table dropped and recreated.")
                    return
                }
                fmt.Println("✅ Keeping existing table.")
                return
	}

	execShell("clickhouse-client", "--multiquery", "--queries-file=clickhouse/init.sql")
	fmt.Println("✅ Table initialized.")
}

func startAllServices() {
	go func() {
		log.Println("🚀 Starting HTTP Producer at :" + httpPort)
		execShell("go", "run", "cmd/producer/main.go", fmt.Sprintf("-port=%s", httpPort))
	}()
	time.Sleep(1 * time.Second)

	go func() {
		log.Println("🚀 Starting gRPC Server at :" + grpcPort)
		execShell("go", "run", "cmd/grpcserver/main.go", fmt.Sprintf("-port=%s", grpcPort))
	}()
	time.Sleep(1 * time.Second)

	go func() {
		log.Println("🚀 Starting RabbitMQ Consumer")
		execShell("go", "run", "cmd/consumer/main.go")
	}()
	time.Sleep(2 * time.Second)
}

func seedTestEvents(count int) {
	log.Printf("📦 Seeding %d test events...\n", count)

	users := []string{"user1", "user2", "guest", "bot", "admin"}
	types := []string{"click", "view", "scroll", "signup", "purchase"}
	urls := []string{"/home", "/shop", "/about", "/search", "/article"}
	referrers := []string{"https://google.com", "https://bing.com", "https://ads.com", "https://facebook.com"}

	batchSize := 100
	switch {
	case count < 500:
		batchSize = 20
	case count < 5000:
		batchSize = 100
	default:
		batchSize = 250
	}

	for i := 0; i < count; i++ {
		evt := map[string]string{
			"user_id":    users[rand.Intn(len(users))],
			"event_type": types[rand.Intn(len(types))],
			"url":        "https://example.com" + urls[rand.Intn(len(urls))],
			"referrer":   referrers[rand.Intn(len(referrers))],
		}
		body, _ := json.Marshal(evt)

		successCount := 0
		resp, err := http.Post("http://localhost:"+httpPort+"/events", "application/json", bytes.NewReader(body))
	        if err != nil {
		    log.Printf("❌ [%d] Error sending to RabbitMQ producer: %v", i, err)
	        } else {
		    successCount++
		    resp.Body.Close()
	        }

		// Batch log every N
	        if (i+1)%10 == 0 || i+1 == count {
		    log.Printf("📤 Sent %d/%d events so far...", i+1, count)
	        }
		if i%batchSize == 0 && i > 0 {
			log.Printf("...seeded %d/%d events", i, count)
			time.Sleep(100 * time.Millisecond)
		}
	}

	log.Println("✅ Seeding complete.")
}

func countClickhouseRows() int {
    out, err := exec.Command("clickhouse-client", "--query", "SELECT count() FROM analytics.page_events FORMAT TSV").Output()
    if err != nil {
        log.Printf("❌ Count query failed: %v", err)
        return -1
    }
    countStr := strings.TrimSpace(string(out))
    count, _ := strconv.Atoi(countStr)
    return count
}


func parsePort(portStr string) int {
    p, err := strconv.Atoi(portStr)
    if err != nil {
        log.Fatalf("Invalid port: %s", portStr)
    }
    return p
}

func main() {
	checkClickhouseConnection()

	err := godotenv.Load()
        if err != nil {
            log.Println("⚠️  No .env file found. Using default ports.")
        }

        onlySeed := flag.Bool("only-seed", false, "Only seed the database")
	recreateTable := flag.Bool("recreate-table", false, "Option to recreate the Clickhouse table if exists")
	runServices := flag.Bool("run-services", false, "Option to skip services")
        seedCount := flag.Int("seed-count", 1000, "How many rows to seed")
	flag.Parse()

        if *onlySeed {
	    startAllServices()
            before := countClickhouseRows()
            log.Printf("📊 Row count before insert: %d", before)

            seedTestEvents(*seedCount)

            after := countClickhouseRows()
            log.Printf("📊 Row count after insert: %d (Δ %d)", after, after-before)		
	    return
        }

	log.Println("📛 Checking for running services...")
	killProcessesOnPorts(requiredPorts)

        if *recreateTable {
	    recreateTableIfConfirmed()
	}

	if *runServices {
	    startAllServices()
	    log.Println("✅ All services running.")
	}
}

