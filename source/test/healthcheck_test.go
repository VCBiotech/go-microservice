package test

import (
	"testing"
)

func TestE2EHealthCheck(t *testing.T) {
	// Arange
	app := spawnTestApp()
	// Act
	resp, err := app.Client.Get("http://localhost:3000/auth/health")
	if err != nil {
		t.Errorf("Error %s", err)
	}
	defer resp.Body.Close()
	// Assert
	if resp.Status != "200 OK" {
		t.Errorf("Health Check did not pass: %s", resp.Body)
	}
}
