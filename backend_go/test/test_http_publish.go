package main
import (
  "bytes"
  "encoding/json"
  "io"
  "log"
  "net/http"
  "os"
  "time"
)
func main() {
  req := map[string]interface{}{
    "user_id":"user1","event_type":"click","url":"https://x","referrer":"https://y",
  }
  body, _ := json.Marshal(req)
  resp, err := http.Post("http://localhost:8081/events", "application/json", bytes.NewReader(body))
  if err!=nil { log.Fatalf("post error: %v", err) }
  io.Copy(os.Stdout, resp.Body)
  log.Println()
  time.Sleep(2*time.Second)
  resp2, err := http.Post("http://localhost:8080/events", "application/json", bytes.NewReader(body))
  if err!=nil { log.Fatalf("api error: %v", err) }
  io.Copy(os.Stdout, resp2.Body)
}
