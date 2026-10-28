package fcm

import (
	"context"
	"os"
	"testing"
	"time"
)

func TestFCMClientInitialization(t *testing.T) {
	credPath := "../../config/firebase-service-account.json"
	data, err := os.ReadFile(credPath)
	if err != nil {
		t.Fatalf("Failed to read firebase credentials: %v", err)
	}

	client, err := NewClientFromJSON(data)
	if err != nil {
		t.Fatalf("Failed to initialize FCM client: %v", err)
	}

	if client.GetProjectID() != "mdtnotification" {
		t.Errorf("Expected project ID 'mdtnotification', got '%s'", client.GetProjectID())
	}

	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()

	token, err := client.getAccessToken(ctx)
	if err != nil {
		t.Fatalf("Failed to obtain OAuth2 token from Google: %v", err)
	}

	if token == "" {
		t.Fatalf("Expected non-empty OAuth2 token")
	}

	t.Logf("Successfully fetched Google OAuth2 token! Length: %d", len(token))
}
