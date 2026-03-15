package server

import (
	"fmt"
	"log"
	"net/http"
	"sync"
)

var mu sync.Mutex
var count int

func GetCount(w http.ResponseWriter, r *http.Request) {
	mu.Lock()
	defer mu.Unlock()
	fmt.Fprintf(w, "Count: %d", count)
}

func Increment(w http.ResponseWriter, r *http.Request) {
	mu.Lock()
	count++
	mu.Unlock()
	log.Println("incremented")
	w.Write([]byte("ok"))
}

func main() {
	http.HandleFunc("/count", GetCount)
	http.HandleFunc("/inc", Increment)
	log.Fatal(http.ListenAndServe(":8080", nil))
}
