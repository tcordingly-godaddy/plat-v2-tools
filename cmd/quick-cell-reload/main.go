package main

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"flag"
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/tcordingly-godaddy/plat-v2-tools/pkg/jobmeta"
	"golang.org/x/time/rate"
)

const (
	defaultNamespace = "sites"
	defaultPattern   = "app-"
)

func main() {
	var (
		jobID      = flag.String("job", "", "Job ID to update")
		namespace  = flag.String("namespace", defaultNamespace, "Nomad namespace")
		jobPattern = flag.String("pattern", "", "Job pattern for updating multiple jobs")
		burst      = flag.Int("burst", 10, "Number of requests allowed in burst")
		limit      = flag.Int("limit", 1, "Number of requests allowed per interval")
		interval   = flag.Duration("interval", 1*time.Second, "Time interval for rate limiting")
	)
	flag.Parse()

	// Create rate limiter for multiple job updates
	var limiter *rate.Limiter
	if *jobID == "" {
		// Only create rate limiter for multiple job updates
		limiter = jobmeta.NewRateLimiter(*burst, *limit, *interval)
	}

	updater, err := jobmeta.NewUpdater(limiter)
	if err != nil {
		log.Fatalf("Failed to create job meta updater: %v", err)
	}

	// Set up signal handling for graceful shutdown
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGTERM)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Handle SIGTERM in a goroutine
	go func() {
		<-sigChan
		log.Println("Received SIGTERM, shutting down gracefully...")
		cancel()
	}()

	if err := runCommand(ctx, updater, jobID, namespace, jobPattern); err != nil {
		log.Fatalf("Command failed: %v", err)
	}

	log.Println("")
}

func runCommand(ctx context.Context, updater *jobmeta.Updater, jobID, namespace *string, jobPattern *string) error {

	// Generate a new reload hash
	reloadHash, err := generateReloadHash()
	if err != nil {
		return fmt.Errorf("failed to generate reload hash: %w", err)
	}

	metaUpdates := map[string]string{
		"reload-hash": reloadHash,
	}

	// If no job ID is provided, use the pattern (defaulting to "app-" if no pattern specified)
	if *jobID == "" {
		pattern := *jobPattern
		if pattern == "" {
			pattern = defaultPattern
		}
		return updater.UpdateMultipleJobs(ctx, *namespace, pattern, metaUpdates)
	}

	updater.UpdateJobMeta(ctx, *jobID, *namespace, metaUpdates)
	return nil
}

func generateReloadHash() (string, error) {
	// Generate 16 random bytes
	bytes := make([]byte, 16)
	if _, err := rand.Read(bytes); err != nil {
		return "", err
	}

	// Convert to hex string
	return hex.EncodeToString(bytes), nil
}
