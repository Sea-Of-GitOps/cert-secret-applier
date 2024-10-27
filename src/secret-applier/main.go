package main

import (
    "fmt"
    "net/http"
    "crypto/tls"
    "crypto/x509"
    // "io"
    "bytes"
    "encoding/json"
    "time"
    "os"
    "encoding/base64"
)

type Secret struct {
    APIVersion string             `json:"apiVersion"`
    Kind       string             `json:"kind"`
    Metadata   Metadata           `json:"metadata"`
    Type       string             `json:"type"`
    Data       map[string]string  `json:"data"`
}

// Metadata rappresenta i metadati di un oggetto Kubernetes
type Metadata struct {
    Name        string             `json:"name"`
    Namespace   string             `json:"namespace"`
    Annotations map[string]string  `json:"annotations"`
}

func Check(e error) {
    if e != nil {
        panic(e)
    }
}

func GetCert(domain string, filename string) (string, error){
    certFilePath := "./etc/letsencrypt/live/"+domain+"/"+filename+".pem"
    cert, err := os.ReadFile(certFilePath)
    
    return string(cert), err
}

func GetNamespace()(string, error){
    namespaceFilePath := "./var/run/secrets/kubernetes.io/serviceaccount/namespace"
    namespace, err := os.ReadFile(namespaceFilePath)
    
    return string(namespace), err
}

func GetCA()(string, error){
    caCertFilePath := "./var/run/secrets/kubernetes.io/serviceaccount/ca.crt"
    caCert, err := os.ReadFile(caCertFilePath)
    
    return string(caCert), err
}

func GetServiceAccountToken()(string, error){
    tokenFilePath := "./var/run/secrets/kubernetes.io/serviceaccount/token"
    token, err := os.ReadFile(tokenFilePath)
    
    return string(token), err
}

func CreateSecret(domain string, namespace string, fullchain string, privkey string) (Secret) {
    expiration_certs := time.Now().AddDate(0, 0, 90).Format(time.RFC3339)
    return Secret{
        APIVersion: "v1",
        Kind:       "Secret",
        Metadata: Metadata{
            Name:      domain+"-tls",
            Namespace: namespace,
            Annotations: map[string]string{
                "certs-expiration": expiration_certs,
            },
        },
        Type: "kubernetes.io/tls",
        Data: map[string]string{
            "tls.crt": base64.StdEncoding.EncodeToString([]byte(fullchain)),
            "tls.key": base64.StdEncoding.EncodeToString([]byte(privkey)), 
        },
    }
}

func K8sHTTPRequest(method string, url string, jsonData []byte) (*http.Response, error) {
    req, err := http.NewRequest(method, url, bytes.NewBuffer(jsonData))
    Check(err)

    req.Header.Set("Content-Type", "application/json")
    token, err := GetServiceAccountToken()
    Check(err)
    req.Header.Set("Authorization", "Bearer "+token)

    ca, err := GetCA()
    Check(err)
    caCert := []byte(ca)
    caCertPool := x509.NewCertPool()
    caCertPool.AppendCertsFromPEM(caCert)

    client := &http.Client{
        Transport: &http.Transport{
            TLSClientConfig: &tls.Config{
                RootCAs: caCertPool,
            },
        },
    }

    resp, err := client.Do(req)
    Check(err)

    defer resp.Body.Close()

    return resp, err
}

func main() {
    os.Setenv("DOMAIN", "www.example.com")
    os.Setenv("KUBERNETES_SERVICE_HOST", "127.0.0.1")
    os.Setenv("KUBERNETES_SERVICE_PORT_HTTPS", "44093")

    domain := os.Getenv("DOMAIN")
    fullchain, err:=GetCert(domain,"fullchain")
    Check(err)
    privkey, err := GetCert(domain,"privkey")
    Check(err)
    
    namespace, err := GetNamespace();
    Check(err)

    secret := CreateSecret(domain, namespace, fullchain, privkey)

    jsonData, err := json.Marshal(secret)
    Check(err)

    kubernetesHost := os.Getenv("KUBERNETES_SERVICE_HOST")
    kubernetePort := os.Getenv("KUBERNETES_SERVICE_PORT_HTTPS")

    url := "https://"+kubernetesHost+":"+kubernetePort+"/api/v1/namespaces/"+namespace+"/secrets/"

    resp, err := K8sHTTPRequest("GET", url+domain+"-tls", jsonData)
    Check(err)
    status := resp.StatusCode
    fmt.Println(status)

    if status == 200 {
        resp, err := K8sHTTPRequest("PUT", url+domain+"-tls", jsonData)
        Check(err)
        fmt.Println(resp.StatusCode)
        fmt.Println("PUT")
    } else {
        resp, err := K8sHTTPRequest("POST", url, jsonData)
        Check(err)
        fmt.Println(resp.StatusCode)
        fmt.Println("POST")
    }
}    
