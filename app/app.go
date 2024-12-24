package app

import (
	"fmt"
	"net/http"
	"crypto/tls"
	"crypto/x509"
	"bytes"
	"encoding/json"
	"time"
	"os"
	"encoding/base64"
	"cert-secret-applier/config"
)

// Definisci percorsi come costanti globali
const (
	certBasePath       = "/etc/letsencrypt/live/"
	namespaceFilePath  = "/var/run/secrets/kubernetes.io/serviceaccount/namespace"
	caCertFilePath     = "/var/run/secrets/kubernetes.io/serviceaccount/ca.crt"
	tokenFilePath      = "/var/run/secrets/kubernetes.io/serviceaccount/token"
)

type Secret struct {
	APIVersion string             `json:"apiVersion"`
	Kind       string             `json:"kind"`
	Metadata   Metadata           `json:"metadata"`
	Type       string             `json:"type"`
	Data       map[string]string  `json:"data"`
}

type Metadata struct {
	Name        string             `json:"name"`
	Namespace   string             `json:"namespace"`
	Annotations map[string]string  `json:"annotations"`
}

func GetFileContent(path string) (string, error) {
	content, err := os.ReadFile(path)
	return string(content), err
}

func GetCert(domain string, filename string) (string, error) {
	return GetFileContent(certBasePath + domain + "/" + filename + ".pem")
}

func CreateSecret(domain string, namespace string, fullchain string, privkey string) Secret {
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
	req, err := http.NewRequest(method, url, bytes.NewBuffer(jsonData))
	if err != nil {
			return nil, err
	}
	req.Header.Set("Content-Type", "application/json")
	token, err := GetFileContent(tokenFilePath)
	if err != nil {
			return nil, err
	}
	req.Header.Set("Authorization", "Bearer "+token)
	
	return client.Do(req)
}

func Run(config *config.Config) error {
	domain := config.Duckdns.Domain
	kubernetesHost := config.Kubernetes.Service.Host
	kubernetePort := config.Kubernetes.Service.Port.HTTP
	
	fullchain, err := GetCert(domain, "fullchain")
	privkey, err := GetCert(domain, "privkey")
	namespace, err := GetFileContent(namespaceFilePath)
	if err != nil {
			return err
	}

	caCertContent, err := GetFileContent(caCertFilePath)
	if err != nil {
			return err
	}

	caCertPool := x509.NewCertPool()
	caCertPool.AppendCertsFromPEM([]byte(caCertContent))

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
			return err
	}

	url := fmt.Sprintf("https://%s:%d/api/v1/namespaces/%s/secrets/", kubernetesHost, kubernetePort, namespace)

	resp, err := K8sHTTPRequest(client, "GET", url+domain+"-tls", jsonData)
	if err != nil {
			return err
	}
	defer resp.Body.Close()
	
	status := resp.StatusCode
	fmt.Println("Status GET:", status)

	if status == http.StatusOK {
			resp, err = K8sHTTPRequest(client, "PUT", url+domain+"-tls", jsonData)
			fmt.Println("PUT")
	} else {
			resp, err = K8sHTTPRequest(client, "POST", url, jsonData)
			fmt.Println("POST")
	}
	defer resp.Body.Close()

	if err != nil {
			return err
	} else {
			fmt.Println("Status code:", resp.StatusCode)
	}

	return nil
}