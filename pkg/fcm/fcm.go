package fcm

import (
	"bytes"
	"context"
	"crypto/rsa"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/url"
	"os"
	"sync"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

const (
	FCMScope       = "https://www.googleapis.com/auth/firebase.messaging"
	DefaultAuthURL = "https://oauth2.googleapis.com/token"
)

// ServiceAccount represents Firebase Service Account JSON credentials
type ServiceAccount struct {
	Type                    string `json:"type"`
	ProjectID               string `json:"project_id"`
	PrivateKeyID            string `json:"private_key_id"`
	PrivateKey              string `json:"private_key"`
	ClientEmail             string `json:"client_email"`
	ClientID                string `json:"client_id"`
	AuthURI                 string `json:"auth_uri"`
	TokenURI                string `json:"token_uri"`
	AuthProviderX509CertURL string `json:"auth_provider_x509_cert_url"`
	ClientX509CertURL       string `json:"client_x509_cert_url"`
}

// Client manages FCM API interactions and token caching
type Client struct {
	sa         *ServiceAccount
	privateKey *rsa.PrivateKey
	httpClient *http.Client

	mu          sync.RWMutex
	accessToken string
	tokenExpiry time.Time
}

// NotificationPayload represents the notification content
type NotificationPayload struct {
	Title string            `json:"title"`
	Body  string            `json:"body"`
	Image string            `json:"image,omitempty"`
	Data  map[string]string `json:"data,omitempty"`
}

// MessageRequest represents the payload structure for FCM HTTP v1
type MessageRequest struct {
	Message MessageBody `json:"message"`
}

type MessageBody struct {
	Token        string            `json:"token"`
	Notification *FCMNotification  `json:"notification,omitempty"`
	Data         map[string]string `json:"data,omitempty"`
}

type FCMNotification struct {
	Title string `json:"title"`
	Body  string `json:"body"`
	Image string `json:"image,omitempty"`
}

type SendResult struct {
	Token     string `json:"token"`
	Success   bool   `json:"success"`
	MessageID string `json:"message_id,omitempty"`
	Error     string `json:"error,omitempty"`
}

// NewClient creates a new FCM client from a service account JSON file path or raw JSON
func NewClient(credentialsPath string) (*Client, error) {
	data, err := os.ReadFile(credentialsPath)
	if err != nil {
		return nil, fmt.Errorf("failed to read firebase service account file: %w", err)
	}

	return NewClientFromJSON(data)
}

// NewClientFromJSON creates a new FCM client from raw JSON bytes
func NewClientFromJSON(data []byte) (*Client, error) {
	var sa ServiceAccount
	if err := json.Unmarshal(data, &sa); err != nil {
		return nil, fmt.Errorf("failed to parse firebase credentials JSON: %w", err)
	}

	if sa.ProjectID == "" || sa.ClientEmail == "" || sa.PrivateKey == "" {
		return nil, errors.New("invalid service account JSON: missing project_id, client_email, or private_key")
	}

	parsedKey, err := jwt.ParseRSAPrivateKeyFromPEM([]byte(sa.PrivateKey))
	if err != nil {
		return nil, fmt.Errorf("failed to parse RSA private key: %w", err)
	}

	return &Client{
		sa:         &sa,
		privateKey: parsedKey,
		httpClient: &http.Client{
			Timeout: 20 * time.Second,
		},
	}, nil
}

// GetProjectID returns the configured Firebase Project ID
func (c *Client) GetProjectID() string {
	if c.sa == nil {
		return ""
	}
	return c.sa.ProjectID
}

// getAccessToken returns a valid cached Google OAuth2 token, or fetches a new one if expired
func (c *Client) getAccessToken(ctx context.Context) (string, error) {
	c.mu.RLock()
	if c.accessToken != "" && time.Now().Before(c.tokenExpiry.Add(-60*time.Second)) {
		token := c.accessToken
		c.mu.RUnlock()
		return token, nil
	}
	c.mu.RUnlock()

	c.mu.Lock()
	defer c.mu.Unlock()

	// Double-check after acquiring write lock
	if c.accessToken != "" && time.Now().Before(c.tokenExpiry.Add(-60*time.Second)) {
		return c.accessToken, nil
	}

	now := time.Now()
	claims := jwt.MapClaims{
		"iss":   c.sa.ClientEmail,
		"scope": FCMScope,
		"aud":   c.sa.TokenURI,
		"exp":   now.Add(1 * time.Hour).Unix(),
		"iat":   now.Unix(),
	}

	if claims["aud"] == "" {
		claims["aud"] = DefaultAuthURL
	}

	signedJWT, err := jwt.NewWithClaims(jwt.SigningMethodRS256, claims).SignedString(c.privateKey)
	if err != nil {
		return "", fmt.Errorf("failed to sign JWT assertion: %w", err)
	}

	formData := url.Values{
		"grant_type": {"urn:ietf:params:oauth:grant-type:jwt-bearer"},
		"assertion":  {signedJWT},
	}

	tokenURL := c.sa.TokenURI
	if tokenURL == "" {
		tokenURL = DefaultAuthURL
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, tokenURL, bytes.NewBufferString(formData.Encode()))
	if err != nil {
		return "", fmt.Errorf("failed to create oauth token request: %w", err)
	}
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return "", fmt.Errorf("oauth token request failed: %w", err)
	}
	defer resp.Body.Close()

	bodyBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("failed to read oauth token response: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("oauth token error (status %d): %s", resp.StatusCode, string(bodyBytes))
	}

	var tokenResp struct {
		AccessToken string `json:"access_token"`
		ExpiresIn   int    `json:"expires_in"`
		TokenType   string `json:"token_type"`
	}
	if err := json.Unmarshal(bodyBytes, &tokenResp); err != nil {
		return "", fmt.Errorf("failed to decode oauth token response: %w", err)
	}

	c.accessToken = tokenResp.AccessToken
	c.tokenExpiry = time.Now().Add(time.Duration(tokenResp.ExpiresIn) * time.Second)

	log.Printf("[INFO] [FCM] Successfully refreshed Google OAuth2 token for project '%s' (expires in %ds)", c.sa.ProjectID, tokenResp.ExpiresIn)
	return c.accessToken, nil
}

// Send sends a push notification to a specific FCM device token using FCM HTTP v1 API
func (c *Client) Send(ctx context.Context, token string, payload *NotificationPayload) (*SendResult, error) {
	if token == "" {
		return nil, errors.New("fcm token is required")
	}
	if payload == nil {
		return nil, errors.New("notification payload is required")
	}

	accessToken, err := c.getAccessToken(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to obtain access token: %w", err)
	}

	fcmURL := fmt.Sprintf("https://fcm.googleapis.com/v1/projects/%s/messages:send", c.sa.ProjectID)

	msgReq := MessageRequest{
		Message: MessageBody{
			Token: token,
			Notification: &FCMNotification{
				Title: payload.Title,
				Body:  payload.Body,
				Image: payload.Image,
			},
			Data: payload.Data,
		},
	}

	payloadBytes, err := json.Marshal(msgReq)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal FCM request: %w", err)
	}

	httpReq, err := http.NewRequestWithContext(ctx, http.MethodPost, fcmURL, bytes.NewBuffer(payloadBytes))
	if err != nil {
		return nil, fmt.Errorf("failed to create FCM http request: %w", err)
	}

	httpReq.Header.Set("Authorization", "Bearer "+accessToken)
	httpReq.Header.Set("Content-Type", "application/json")

	resp, err := c.httpClient.Do(httpReq)
	if err != nil {
		return &SendResult{
			Token:   token,
			Success: false,
			Error:   err.Error(),
		}, err
	}
	defer resp.Body.Close()

	respBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read FCM response: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		var errResp struct {
			Error struct {
				Code    int    `json:"code"`
				Message string `json:"message"`
				Status  string `json:"status"`
			} `json:"error"`
		}
		_ = json.Unmarshal(respBytes, &errResp)

		errMsg := errResp.Error.Message
		if errMsg == "" {
			errMsg = string(respBytes)
		}

		log.Printf("[WARN] [FCM] Notification failed for token '%s': status %d | %s", token, resp.StatusCode, errMsg)
		return &SendResult{
			Token:   token,
			Success: false,
			Error:   fmt.Sprintf("FCM Error (%d): %s", resp.StatusCode, errMsg),
		}, fmt.Errorf("fcm send error: %s", errMsg)
	}

	var okResp struct {
		Name string `json:"name"`
	}
	_ = json.Unmarshal(respBytes, &okResp)

	log.Printf("[INFO] [FCM] Notification sent successfully! MessageID: %s", okResp.Name)
	return &SendResult{
		Token:     token,
		Success:   true,
		MessageID: okResp.Name,
	}, nil
}

// SendMulticast sends the same notification to multiple FCM tokens concurrently
func (c *Client) SendMulticast(ctx context.Context, tokens []string, payload *NotificationPayload) []SendResult {
	results := make([]SendResult, len(tokens))
	var wg sync.WaitGroup

	for i, tok := range tokens {
		wg.Add(1)
		go func(idx int, t string) {
			defer wg.Done()
			res, err := c.Send(ctx, t, payload)
			if err != nil && res == nil {
				results[idx] = SendResult{
					Token:   t,
					Success: false,
					Error:   err.Error(),
				}
			} else if res != nil {
				results[idx] = *res
			}
		}(i, tok)
	}

	wg.Wait()
	return results
}
