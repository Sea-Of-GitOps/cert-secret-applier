package app

import (
	"bytes"
	"crypto/tls"
	"crypto/x509"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"time"

	"cert-secret-applier/config"
	"cert-secret-applier/logger"
	"go.uber.org/zap" 
)

const (
	certBasePath      = "./extra/etc/letsencrypt/live/"
	namespaceFilePath = "./extra/var/run/secrets/kubernetes.io/serviceaccount/namespace"
	caCertFilePath    = "./extra/var/run/secrets/kubernetes.io/serviceaccount/ca.crt"
	tokenFilePath     = "./extra/var/run/secrets/kubernetes.io/serviceaccount/token"
)

type Secret struct {
	APIVersion string            `json:"apiVersion"`
	Kind       string            `json:"kind"`
	Metadata   Metadata          `json:"metadata"`
	Type       string            `json:"type"`
	Data       map[string]string `json:"data"`
}

type Metadata struct {
	Name        string            `json:"name"`
	Namespace   string            `json:"namespace"`
	Annotations map[string]string `json:"annotations"`
}

func GetFileContent(path string) (string, error) {
	logger.DebugCtx(nil, fmt.Sprintf("Reading file content from path: %s", path))
	content, err := os.ReadFile(path)
	if err != nil {
		logger.ErrorCtx(nil, fmt.Sprintf("Failed to read file at %s", path), zap.Error(err))
		return "", fmt.Errorf("failed to read file at %s: %w", path, err)
	}
	return string(content), nil
}

func GetCert(domain, filename string) (string, error) {
	path := certBasePath + domain + "/" + filename + ".pem"
	logger.DebugCtx(nil, fmt.Sprintf("Fetching certificate from path: %s", path))
	return GetFileContent(path)
}

func CreateSecret(domain, namespace, fullchain, privkey string) Secret {
	logger.DebugCtx(nil, fmt.Sprintf("Creating secret for domain: %s in namespace: %s", domain, namespace))
	expiration := time.Now().AddDate(0, 0, 90).Format(time.RFC3339)
	return Secret{
		APIVersion: "v1",
		Kind:       "Secret",
		Metadata: Metadata{
			Name:      domain + "-tls",
			Namespace: namespace,
			Annotations: map[string]string{
				"certs-expiration": expiration,
			},
		},
		Type: "kubernetes.io/tls",
		Data: map[string]string{
			"tls.crt": base64.StdEncoding.EncodeToString([]byte(fullchain)),
			"tls.key": base64.StdEncoding.EncodeToString([]byte(privkey)),
		},
	}
}

func K8sHTTPRequest(client *http.Client, method, url string, jsonData []byte) (*http.Response, error) {
	logger.DebugCtx(nil, fmt.Sprintf("Creating %s request to URL: %s", method, url))
	req, err := http.NewRequest(method, url, bytes.NewBuffer(jsonData))
	if err != nil {
		logger.ErrorCtx(nil, "Failed to create HTTP request", zap.Error(err))
		return nil, fmt.Errorf("failed to create %s request to %s: %w", method, url, err)
	}
	req.Header.Set("Content-Type", "application/json")
	token, err := GetFileContent(tokenFilePath)
	if err != nil {
		logger.ErrorCtx(nil, "Failed to get token", zap.Error(err))
		return nil, fmt.Errorf("failed to get token: %w", err)
	}
	req.Header.Set("Authorization", "Bearer "+token)

	logger.InfoCtx(nil, fmt.Sprintf("Executing %s HTTP request to %s", method, url))
	return client.Do(req)
}

func Run(config *config.Config) error {
	logger.InfoCtx(nil, "Starting Run function")
	domain := config.Duckdns.Domain
	kubernetesHost := config.Kubernetes.Service.Host
	kubernetesPort := config.Kubernetes.Service.Port.HTTP

	fullchain, err := GetCert(domain, "fullchain")
	if err != nil {
		logger.ErrorCtx(nil, "Failed to get fullchain certificate", zap.Error(err))
		return fmt.Errorf("failed to get fullchain certificate: %w", err)
	}

	privkey, err := GetCert(domain, "privkey")
	if err != nil {
		logger.ErrorCtx(nil, "Failed to get private key", zap.Error(err))
		return fmt.Errorf("failed to get private key: %w", err)
	}

	namespace, err := GetFileContent(namespaceFilePath)
	if err != nil {
		logger.ErrorCtx(nil, "Failed to get namespace", zap.Error(err))
		return fmt.Errorf("failed to get namespace: %w", err)
	}

	caCertContent, err := GetFileContent(caCertFilePath)
	if err != nil {
		logger.ErrorCtx(nil, "Failed to get CA certificate", zap.Error(err))
		return fmt.Errorf("failed to get CA certificate: %w", err)
	}

	caCertPool := x509.NewCertPool()
	if !caCertPool.AppendCertsFromPEM([]byte(caCertContent)) {
		logger.ErrorCtx(nil, "Failed to append CA certificates")
		return fmt.Errorf("failed to append CA certificates")
	}

	client := &http.Client{
		Transport: &http.Transport{
			TLSClientConfig: &tls.Config{
				RootCAs: caCertPool,
			},
		},
	}

	secret := CreateSecret(domain, namespace, fullchain, privkey)
	jsonData, err := json.Marshal(secret)
	if err != nil {
		logger.ErrorCtx(nil, "Failed to marshal secret", zap.Error(err))
		return fmt.Errorf("failed to marshal secret: %w", err)
	}

	url := fmt.Sprintf("https://%s:%d/api/v1/namespaces/%s/secrets/", kubernetesHost, kubernetesPort, namespace)

	resp, err := K8sHTTPRequest(client, "GET", url+domain+"-tls", jsonData)
	if err != nil {
		logger.ErrorCtx(nil, "GET request failed", zap.Error(err))
		return fmt.Errorf("GET request failed: %w", err)
	}
	defer resp.Body.Close()

	status := resp.StatusCode
	logger.InfoCtx(nil, fmt.Sprintf("Status GET: %d", status))

	if status != http.StatusOK && status != http.StatusNotFound {
		logger.WarnCtx(nil, fmt.Sprintf("Unexpected status code %d for GET request", status))
		return fmt.Errorf("unexpected status code %d", status)
	}

	switch resp.StatusCode {
	case http.StatusOK:
		resp, err = K8sHTTPRequest(client, "PUT", url+domain+"-tls", jsonData)
		if err != nil {
			logger.ErrorCtx(nil, "PUT request failed", zap.Error(err))
			return fmt.Errorf("PUT request failed: %w", err)
		}
	case http.StatusNotFound:
		resp, err = K8sHTTPRequest(client, "POST", url, jsonData)
		if err != nil {
			logger.ErrorCtx(nil, "POST request failed", zap.Error(err))
			return fmt.Errorf("POST request failed: %w", err)
		}
	default:
		logger.WarnCtx(nil, fmt.Sprintf("Unexpected status code %d for GET", resp.StatusCode))
		return fmt.Errorf("unexpected status code %d for GET", resp.StatusCode)
	}
	defer resp.Body.Close()

	logger.InfoCtx(nil, fmt.Sprintf("Final status code: %d", resp.StatusCode))
	return nil
}