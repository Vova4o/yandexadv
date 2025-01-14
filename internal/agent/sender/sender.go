package sender

import (
	"bytes"
	"compress/gzip"
	"crypto/hmac"
	"crypto/sha256"
	"crypto/tls"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"log"
	"time"

	"github.com/go-resty/resty/v2"
	"github.com/vova4o/yandexadv/internal/agent/flags"
	"github.com/vova4o/yandexadv/internal/agent/metrics"
)

const (
	maxRetries = 3
	retryDelay = 1 * time.Second
)

// createTLSConfig creates TLS configuration with the provided certificate
func createTLSConfig(certPath string) (*tls.Config, error) {
	return &tls.Config{
		InsecureSkipVerify: true, // For development only
		MinVersion: tls.VersionTLS12,
		CipherSuites: []uint16{
			tls.TLS_ECDHE_ECDSA_WITH_AES_128_GCM_SHA256,
			tls.TLS_ECDHE_RSA_WITH_AES_128_GCM_SHA256,
			tls.TLS_ECDHE_ECDSA_WITH_AES_256_GCM_SHA384,
			tls.TLS_ECDHE_RSA_WITH_AES_256_GCM_SHA384,
			tls.TLS_ECDHE_ECDSA_WITH_CHACHA20_POLY1305,
			tls.TLS_ECDHE_RSA_WITH_CHACHA20_POLY1305,
		},
	}, nil
}

// getProtocol returns http or https based on crypto path
func getProtocol(cryptoPath string) string {
	if cryptoPath != "" {
		return "https"
	}
	return "http"
}

// CompressData сжимает данные с использованием gzip
func CompressData(data []byte) ([]byte, error) {
	var buf bytes.Buffer
	writer := gzip.NewWriter(&buf)
	_, err := writer.Write(data)
	if err != nil {
		return nil, err
	}
	err = writer.Close()
	if err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}

// ServerSupportsGzip проверяет, поддерживает ли сервер gzip-сжатие
func ServerSupportsGzip(cfg *flags.Config) bool {
	client := resty.New()
	protocol := getProtocol(cfg.CryptoPath)

	if cfg.CryptoPath != "" {
		tlsConfig, err := createTLSConfig(cfg.CryptoPath)
		if err != nil {
			log.Printf("Failed to create TLS config: %v", err)
			return false
		}
		client.SetTLSClientConfig(tlsConfig)
	}

	resp, err := client.R().
		SetHeader("Accept-Encoding", "gzip").
		Get(fmt.Sprintf("%s://%s", protocol, cfg.ServerAddress))
	if err != nil {
		log.Printf("Failed to check gzip support: %v\n", err)
		return false
	}

	return resp.Header().Get("Content-Encoding") == "gzip"
}

// calculateHash вычисляет HMAC-SHA256 хэш из данных и ключа
func calculateHash(data, key []byte) string {
	h := hmac.New(sha256.New, key)
	h.Write(data)
	return hex.EncodeToString(h.Sum(nil))
}

// SendMetricsBatch отправляет метрики на сервер пакетом
func SendMetricsBatch(cfg *flags.Config, metricsData []metrics.Metrics) {
	client := resty.New()
	protocol := getProtocol(cfg.CryptoPath)

	// Configure TLS if crypto path is provided
	if cfg.CryptoPath != "" {
		tlsConfig, err := createTLSConfig(cfg.CryptoPath)
		if err != nil {
			log.Printf("Failed to create TLS config: %v", err)
			return
		}
		client.SetTLSClientConfig(tlsConfig)
	}

	url := fmt.Sprintf("%s://%s/updates", protocol, cfg.ServerAddress)
	log.Printf("Sending metrics to %s\n", url)	
	useGzip := ServerSupportsGzip(cfg)

	// Сериализация метрик в JSON
	jsonData, err := json.Marshal(metricsData)
	if err != nil {
		log.Printf("Failed to marshal metrics: %v\n", err)
		return
	}

	var hash string
	if cfg.SecretKey != "" {
		hash = calculateHash(jsonData, []byte(cfg.SecretKey))
	}

	request := client.R().
		SetHeader("Content-Type", "application/json").
		SetHeader("HashSHA256", hash)

	if useGzip {
		request.SetHeader("Content-Encoding", "gzip")
		compressedData, err := CompressData(jsonData)
		if err != nil {
			log.Printf("Failed to compress data for metrics: %v\n", err)
			return
		}
		request.SetBody(compressedData)
	} else {
		request.SetBody(jsonData)
	}

	if err := sendWithRetry(request, url); err != nil {
		log.Printf("Failed to send metrics: %v\n", err)
	}
}

// SendMetrics отправляет метрики на сервер
func SendMetrics(cfg *flags.Config, metricsData []metrics.Metrics) {
	client := resty.New()
	protocol := getProtocol(cfg.CryptoPath)

	if cfg.CryptoPath != "" {
		tlsConfig, err := createTLSConfig(cfg.CryptoPath)
		if err != nil {
			log.Printf("Failed to create TLS config: %v", err)
			return
		}
		client.SetTLSClientConfig(tlsConfig)
	}

	useGzip := ServerSupportsGzip(cfg)

	for _, metric := range metricsData {
		var url string
		if metric.Value == nil {
			url = fmt.Sprintf("%s://%s/update/%s/%s/%v", protocol, cfg.ServerAddress, metric.MType, metric.ID, *metric.Delta)
		} else {
			url = fmt.Sprintf("%s://%s/update/%s/%s/%v", protocol, cfg.ServerAddress, metric.MType, metric.ID, *metric.Value)
		}

		request := client.R().SetHeader("Content-Type", "text/plain")

		if useGzip {
			request.SetHeader("Content-Encoding", "gzip")
			compressedData, err := CompressData([]byte(url))
			if err != nil {
				log.Printf("Failed to compress data for metric %s: %v\n", metric.ID, err)
				continue
			}
			request.SetBody(compressedData)
		} else {
			request.SetBody(url)
		}

		if err := sendWithRetry(request, url); err != nil {
			log.Printf("Failed to send metric %s: %v\n", metric.ID, err)
		}
	}
}

// SendMetricsJSON отправляет метрики на сервер в формате JSON
func SendMetricsJSON(cfg *flags.Config, metricsData []metrics.Metrics) {
	client := resty.New()
	protocol := getProtocol(cfg.CryptoPath)

	if cfg.CryptoPath != "" {
		tlsConfig, err := createTLSConfig(cfg.CryptoPath)
		if err != nil {
			log.Printf("Failed to create TLS config: %v", err)
			return
		}
		client.SetTLSClientConfig(tlsConfig)
	}

	useGzip := ServerSupportsGzip(cfg)

	for _, metric := range metricsData {
		url := fmt.Sprintf("%s://%s/update/", protocol, cfg.ServerAddress)

		// Сериализация метрики в JSON
		jsonData, err := json.Marshal(metric)
		if err != nil {
			log.Printf("Failed to marshal metric %s: %v\n", metric.ID, err)
			continue
		}

		request := client.R().SetHeader("Content-Type", "application/json")

		if useGzip {
			request.SetHeader("Content-Encoding", "gzip")
			compressedData, err := CompressData(jsonData)
			if err != nil {
				log.Printf("Failed to compress data for metric %s: %v\n", metric.ID, err)
				continue
			}
			request.SetBody(compressedData)
		} else {
			request.SetBody(jsonData)
		}

		if err := sendWithRetry(request, url); err != nil {
			log.Printf("Failed to send metric %s: %v\n", metric.ID, err)
		}
	}
}

// sendWithRetry отправляет запрос с повторными попытками в случае ошибки
func sendWithRetry(request *resty.Request, url string) error {
	delay := retryDelay
	for i := 0; i < maxRetries; i++ {
		resp, err := request.Post(url)
		if err != nil {
			log.Printf("Failed to send request: %v\n", err)
		} else if resp.StatusCode() == 200 {
			return nil
		} else {
			log.Printf("Failed to send request: status code %d\n", resp.StatusCode())
			log.Printf("Response body: %s\n", resp.String())
		}

		time.Sleep(delay)
		delay += 2 * time.Second
	}
	return fmt.Errorf("failed to send request after %d attempts", maxRetries)
}
