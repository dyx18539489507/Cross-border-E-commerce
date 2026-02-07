package volcengine

import (
	"bytes"
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"
)

const (
	DefaultVisualHost    = "visual.volcengineapi.com"
	DefaultVisualRegion  = "cn-north-1"
	DefaultVisualService = "cv"
	DefaultVisualVersion = "2022-08-31"
)

type VisualClient struct {
	AccessKeyID     string
	SecretAccessKey string
	Region          string
	Service         string
	Host            string
	HTTPClient      *http.Client
}

type VisualResponse struct {
	Code        int             `json:"code"`
	Message     string          `json:"message"`
	RequestID   string          `json:"request_id"`
	Status      int             `json:"status"`
	TimeElapsed string          `json:"time_elapsed"`
	Data        json.RawMessage `json:"data"`
}

func NewVisualClient(accessKeyID, secretAccessKey, region, service, host string) *VisualClient {
	if region == "" {
		region = DefaultVisualRegion
	}
	if service == "" {
		service = DefaultVisualService
	}
	if host == "" {
		host = DefaultVisualHost
	}

	// Avoid unexpected failures from local proxy env vars (HTTP_PROXY/HTTPS_PROXY).
	// Volcengine endpoints are typically reachable directly; using a misconfigured proxy
	// can cause 5xx errors like "504 Gateway Time-out".
	transport, _ := http.DefaultTransport.(*http.Transport)
	clonedTransport := transport.Clone()
	clonedTransport.Proxy = nil

	return &VisualClient{
		AccessKeyID:     accessKeyID,
		SecretAccessKey: secretAccessKey,
		Region:          region,
		Service:         service,
		Host:            host,
		HTTPClient: &http.Client{
			Timeout:   10 * time.Minute,
			Transport: clonedTransport,
		},
	}
}

func (c *VisualClient) Do(ctx context.Context, action, version string, body any) (*VisualResponse, error) {
	if version == "" {
		version = DefaultVisualVersion
	}

	payload := []byte("{}")
	if body != nil {
		data, err := json.Marshal(body)
		if err != nil {
			return nil, fmt.Errorf("marshal request: %w", err)
		}
		payload = data
	}

	queries := url.Values{}
	queries.Set("Action", action)
	queries.Set("Version", version)
	queryString := strings.ReplaceAll(queries.Encode(), "+", "%20")

	requestAddr := fmt.Sprintf("https://%s/?%s", c.Host, queryString)
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, requestAddr, bytes.NewBuffer(payload))
	if err != nil {
		return nil, fmt.Errorf("create request: %w", err)
	}

	xDate := time.Now().UTC().Format("20060102T150405Z")
	authDate := xDate[:8]

	xContentSha256 := hex.EncodeToString(hashSHA256(payload))

	req.Header.Set("X-Date", xDate)
	req.Header.Set("X-Content-Sha256", xContentSha256)
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Host", c.Host)

	signedHeaders := []string{"host", "x-date", "x-content-sha256", "content-type"}
	headerList := []string{
		"host:" + c.Host,
		"x-date:" + xDate,
		"x-content-sha256:" + xContentSha256,
		"content-type:application/json",
	}
	headerString := strings.Join(headerList, "\n")

	canonicalString := strings.Join([]string{
		http.MethodPost,
		"/",
		queryString,
		headerString + "\n",
		strings.Join(signedHeaders, ";"),
		xContentSha256,
	}, "\n")

	hashedCanonicalString := hex.EncodeToString(hashSHA256([]byte(canonicalString)))
	credentialScope := fmt.Sprintf("%s/%s/%s/request", authDate, c.Region, c.Service)
	signString := strings.Join([]string{
		"HMAC-SHA256",
		xDate,
		credentialScope,
		hashedCanonicalString,
	}, "\n")

	signedKey := getSignedKey(c.SecretAccessKey, authDate, c.Region, c.Service)
	signature := hex.EncodeToString(hmacSHA256(signedKey, signString))

	authorization := "HMAC-SHA256" +
		" Credential=" + c.AccessKeyID + "/" + credentialScope +
		", SignedHeaders=" + strings.Join(signedHeaders, ";") +
		", Signature=" + signature

	req.Header.Set("Authorization", authorization)

	resp, err := c.HTTPClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("send request: %w", err)
	}
	defer resp.Body.Close()

	bodyBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("read response: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("volcengine http status %d: %s", resp.StatusCode, string(bodyBytes))
	}

	var result VisualResponse
	if err := json.Unmarshal(bodyBytes, &result); err != nil {
		return nil, fmt.Errorf("parse response: %w", err)
	}

	if result.Code != 10000 {
		return nil, fmt.Errorf("volcengine error %d: %s", result.Code, result.Message)
	}

	return &result, nil
}

func hmacSHA256(key []byte, content string) []byte {
	mac := hmac.New(sha256.New, key)
	mac.Write([]byte(content))
	return mac.Sum(nil)
}

func getSignedKey(secretKey, date, region, service string) []byte {
	kDate := hmacSHA256([]byte(secretKey), date)
	kRegion := hmacSHA256(kDate, region)
	kService := hmacSHA256(kRegion, service)
	kSigning := hmacSHA256(kService, "request")
	return kSigning
}

func hashSHA256(data []byte) []byte {
	hash := sha256.New()
	_, _ = hash.Write(data)
	return hash.Sum(nil)
}
