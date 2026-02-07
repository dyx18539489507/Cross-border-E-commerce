package volcengine

import (
	"bytes"
	"context"
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
	DefaultOpenAPIHost       = "open.volcengineapi.com"
	DefaultSpeechSaasService = "speech_saas_prod"
	DefaultSpeechSaasRegion  = "cn-north-1"
	DefaultSpeechSaasVersion = "2025-05-20"
)

type OpenAPIClient struct {
	AccessKeyID     string
	SecretAccessKey string
	Region          string
	Service         string
	Host            string
	HTTPClient      *http.Client
}

type OpenAPIResponse struct {
	Result           json.RawMessage `json:"Result"`
	ResponseMetadata struct {
		RequestID string `json:"RequestId"`
		Error     *struct {
			Code    string `json:"Code"`
			Message string `json:"Message"`
		} `json:"Error"`
	} `json:"ResponseMetadata"`
}

func NewOpenAPIClient(accessKeyID, secretAccessKey, region, service, host string) *OpenAPIClient {
	if region == "" {
		region = DefaultSpeechSaasRegion
	}
	if service == "" {
		service = DefaultSpeechSaasService
	}
	if host == "" {
		host = DefaultOpenAPIHost
	}

	// Avoid unexpected failures from local proxy env vars (HTTP_PROXY/HTTPS_PROXY).
	// These endpoints are generally reachable directly, and a misconfigured proxy
	// can cause spurious 5xx/timeout errors.
	transport, _ := http.DefaultTransport.(*http.Transport)
	clonedTransport := transport.Clone()
	clonedTransport.Proxy = nil

	return &OpenAPIClient{
		AccessKeyID:     accessKeyID,
		SecretAccessKey: secretAccessKey,
		Region:          region,
		Service:         service,
		Host:            host,
		HTTPClient: &http.Client{
			Timeout:   30 * time.Second,
			Transport: clonedTransport,
		},
	}
}

func (c *OpenAPIClient) Do(ctx context.Context, action, version string, body any) (json.RawMessage, error) {
	if version == "" {
		version = DefaultSpeechSaasVersion
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

	var result OpenAPIResponse
	if err := json.Unmarshal(bodyBytes, &result); err != nil {
		return nil, fmt.Errorf("parse response: %w", err)
	}

	if result.ResponseMetadata.Error != nil && result.ResponseMetadata.Error.Code != "" {
		return nil, fmt.Errorf("volcengine error %s: %s", result.ResponseMetadata.Error.Code, result.ResponseMetadata.Error.Message)
	}

	return result.Result, nil
}
