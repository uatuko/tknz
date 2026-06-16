package db

import (
	"log"
	"os"
	"strconv"
	"testing"
	"time"
)

var (
	timeout time.Duration
)

func TestMain(m *testing.M) {
	tSecs, err := strconv.Atoi(os.Getenv("TEST_TIMEOUT_S"))
	if err != nil || tSecs < 1 {
		tSecs = 2
	}

	timeout = time.Duration(tSecs) * time.Second

	// Override configs with test values
	if err := Setup(); err != nil {
		log.Fatalf("failed to setup database: %v", err)
	}

	// Run
	code := m.Run()

	// Teardown
	Teardown()

	os.Exit(code)
}
