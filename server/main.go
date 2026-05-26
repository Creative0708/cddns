package main

import (
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"log"
	"net"
	"net/http"
	"os"
	"os/signal"
	"strconv"
	"strings"
	"syscall"
	"time"

	"golang.org/x/time/rate"
)

func main() {
	log.Default().SetFlags(0)

	new := flag.Bool("new", false, "add a new entry")
	address := flag.String("a", "127.0.0.1:16822", "address to bind to")
	dataFilePath := flag.String("c", "./cddns.json", "")

	flag.Parse()

	ctx := context.Background()

	entries, entriesErr := loadDnsEntries(*dataFilePath)

	if *new {
		newEntry, err := getNewEntry(ctx)
		if err != nil {
			log.Fatal("failed to add entry: ", err)
		}

		if entries == nil {
			entries = make(DnsEntryMap)
		}

		apiKey := generateApiKey()
		entries[apiKey] = *newEntry

		if err := saveDnsEntries(*dataFilePath, &entries); err != nil {
			log.Fatal("failed to save file: ", err)
		}

		fmt.Println("created DNS entry! API key: ", apiKey)
		return
	}

	if entriesErr != nil {
		log.Fatal("failed to load entries: ", entriesErr)
	}

	startServer(ctx, *address, entries)
}

const RATELIMIT_DURATION time.Duration = time.Minute

func startServer(ctx context.Context, address string, entries DnsEntryMap) {
	tokenLimiters := make(map[string]*rate.Limiter)
	for apiKey := range entries {
		tokenLimiters[apiKey] = rate.NewLimiter(rate.Every(RATELIMIT_DURATION), 1)
	}
	http.HandleFunc("/v1/update", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "POST" {
			w.WriteHeader(400)
			return
		}
		authorization := r.Header.Get("Authorization")
		if authorization == "" {
			w.WriteHeader(401)
			return
		}
		components := strings.Split(authorization, " ")
		if len(components) != 2 || components[0] != "Bearer" {
			w.WriteHeader(400)
			return
		}

		key := components[1]
		limiter, ok := tokenLimiters[key]
		if !ok {
			w.WriteHeader(400)
			w.Write([]byte("invalid api key"))
			return
		}

		var buf [128]byte

		n, err := io.ReadFull(r.Body, buf[:])
		if err != nil && err != io.ErrUnexpectedEOF {
			log.Print("warning: failed to read update request body: ", err)
			return
		}

		ip := net.ParseIP(string(buf[:n]))

		if ip == nil {
			w.WriteHeader(400)
			w.Write([]byte("invalid ip address"))
			return
		}

		if !limiter.Allow() {
			w.Header().Add("Retry-After", strconv.Itoa(int(RATELIMIT_DURATION.Seconds())))
			w.WriteHeader(429)
			return
		}

		entry := entries[key]

		w.WriteHeader(202)

		log.Println("updating", entry.Name, "to ip address", ip)

		go updateServerIp(ctx, entry.ZoneId, entry.RecordId, ip)
	})

	server := &http.Server{Addr: address}

	exitServerOnSignal(server, ctx)

	log.Printf("running on %s...", address)

	if err := server.ListenAndServe(); err != http.ErrServerClosed {
		log.Print("failed to run server:", err)
	}
}

func exitServerOnSignal(server *http.Server, ctx context.Context) {
	exitSignalChan := make(chan os.Signal, 1)
	signal.Notify(exitSignalChan, os.Interrupt, syscall.SIGTERM)

	go func() {
		<-exitSignalChan
		log.Println("Shutting down...")
		if err := server.Shutdown(ctx); err != nil {
			log.Fatal("Failed to shutdown server:", err)
		}
	}()
}

type DynamicDnsEntry struct {
	ZoneId   string
	RecordId string
	Name     string
}

// map[api key]domain
type DnsEntryMap map[string]DynamicDnsEntry

func loadDnsEntries(filePath string) (DnsEntryMap, error) {
	bytes, err := os.ReadFile(filePath)
	if err != nil {
		return nil, err
	}

	var dns DnsEntryMap
	err = json.Unmarshal(bytes, &dns)
	if err != nil {
		return nil, err
	}

	return dns, nil
}
func saveDnsEntries(filePath string, entries *DnsEntryMap) error {
	bytes, err := json.MarshalIndent(entries, "", "  ")
	if err != nil {
		return err
	}

	err = os.WriteFile(filePath, bytes, 0o600)
	if err != nil {
		log.Print("warning: failed to save file: ", err)
	} else {
		log.Print("saved file to ", filePath)
	}
	return nil
}
