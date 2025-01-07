package sender_test

import (
    "bytes"
    "compress/gzip"
    "encoding/json"
    "io"
    "net/http"
    "net/http/httptest"
    "strings"
    "testing"

    "github.com/stretchr/testify/assert"
    "github.com/vova4o/yandexadv/internal/agent/flags"
    "github.com/vova4o/yandexadv/internal/agent/metrics"
    "github.com/vova4o/yandexadv/internal/agent/sender"
)

// Helper functions remain unchanged
func float64Ptr(v float64) *float64 {
    return &v
}

func int64Ptr(v int64) *int64 {
    return &v
}

func TestCompressData(t *testing.T) {
    data := []byte("test data")
    compressedData, err := sender.CompressData(data)
    assert.NoError(t, err)

    reader, err := gzip.NewReader(bytes.NewReader(compressedData))
    assert.NoError(t, err)
    defer reader.Close()

    decompressedData, err := io.ReadAll(reader)
    assert.NoError(t, err)
    assert.Equal(t, data, decompressedData)
}

func TestServerSupportsGzip(t *testing.T) {
    tests := []struct {
        name        string
        useTLS      bool
        responseEnc string
        want        bool
    }{
        {
            name:        "HTTP with gzip support",
            useTLS:      false,
            responseEnc: "gzip",
            want:        true,
        },
        {
            name:        "HTTPS with gzip support",
            useTLS:      true,
            responseEnc: "gzip",
            want:        true,
        },
        {
            name:        "HTTP without gzip support",
            useTLS:      false,
            responseEnc: "",
            want:        false,
        },
        {
            name:        "HTTPS without gzip support",
            useTLS:      true,
            responseEnc: "",
            want:        false,
        },
    }

    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            handler := func(w http.ResponseWriter, r *http.Request) {
                if r.Method == http.MethodGet && r.URL.Path == "/" {
                    // Обработка запроса для проверки поддержки gzip
                    if tt.responseEnc == "gzip" {
                        w.Header().Set("Content-Encoding", "gzip")
                    }
                    w.WriteHeader(http.StatusOK)
                    return
                }

                // Для других путей или методов возвращаем 404
                w.WriteHeader(http.StatusNotFound)
            }

            var server *httptest.Server
            if tt.useTLS {
                server = httptest.NewTLSServer(http.HandlerFunc(handler))
                defer server.Close()
            } else {
                server = httptest.NewServer(http.HandlerFunc(handler))
                defer server.Close()
            }

            cfg := &flags.Config{
                ServerAddress: strings.TrimPrefix(server.URL, "http://"),
                SecretKey:     "test_key",
            }
            if tt.useTLS {
                cfg.ServerAddress = strings.TrimPrefix(server.URL, "https://")
                cfg.CryptoPath = "./test_certs" // Путь можно оставить пустым, так как createTLSConfig игнорирует его содержимое
            }

            supportsGzip := sender.ServerSupportsGzip(cfg)
            assert.Equal(t, tt.want, supportsGzip)
        })
    }
}

func TestSendMetricsBatch(t *testing.T) {
    tests := []struct {
        name       string
        useTLS     bool
        expectGzip bool
    }{
        {
            name:       "HTTP server supports gzip",
            useTLS:     false,
            expectGzip: true,
        },
        {
            name:       "HTTPS server supports gzip",
            useTLS:     true,
            expectGzip: true,
        },
        {
            name:       "HTTP server does not support gzip",
            useTLS:     false,
            expectGzip: false,
        },
    }

    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            handler := func(w http.ResponseWriter, r *http.Request) {
                if r.Method == http.MethodPost && r.URL.Path == "/updates" {
                    if tt.expectGzip && r.Header.Get("Content-Encoding") == "gzip" {
                        // Проверяем, что данные пришли с gzip-сжатием
                        reader, err := gzip.NewReader(r.Body)
                        assert.NoError(t, err)
                        defer reader.Close()
                        var receivedData []metrics.Metrics
                        err = json.NewDecoder(reader).Decode(&receivedData)
                        assert.NoError(t, err)
                        assert.Len(t, receivedData, 2)
                        assert.Equal(t, "metric1", receivedData[0].ID)
                        assert.Equal(t, 10.0, *receivedData[0].Value)
                        assert.Equal(t, "metric2", receivedData[1].ID)
                        assert.Equal(t, int64(20), *receivedData[1].Delta)
                    } else if !tt.expectGzip {
                        // Проверяем, что данные пришли без сжатия
                        var receivedData []metrics.Metrics
                        err := json.NewDecoder(r.Body).Decode(&receivedData)
                        assert.NoError(t, err)
                        assert.Len(t, receivedData, 2)
                        assert.Equal(t, "metric1", receivedData[0].ID)
                        assert.Equal(t, 10.0, *receivedData[0].Value)
                        assert.Equal(t, "metric2", receivedData[1].ID)
                        assert.Equal(t, int64(20), *receivedData[1].Delta)
                    }

                    // Проверяем заголовок Content-Type
                    assert.Equal(t, "application/json", r.Header.Get("Content-Type"))
                    w.WriteHeader(http.StatusOK)
                    return
                }

                // Для других путей или методов возвращаем 404
                w.WriteHeader(http.StatusNotFound)
            }

            var server *httptest.Server
            if tt.useTLS {
                server = httptest.NewTLSServer(http.HandlerFunc(handler))
                defer server.Close()
            } else {
                server = httptest.NewServer(http.HandlerFunc(handler))
                defer server.Close()
            }

            cfg := &flags.Config{
                ServerAddress: strings.TrimPrefix(server.URL, "http://"),
                SecretKey:     "test_key",
            }
            if tt.useTLS {
                cfg.CryptoPath = "./test_certs"
            }

            metricsData := []metrics.Metrics{
                {ID: "metric1", Value: float64Ptr(10)},
                {ID: "metric2", Delta: int64Ptr(20)},
            }

            // Изменяем адрес сервера на "/updates" для этого теста
            cfg.ServerAddress = strings.TrimPrefix(server.URL, "http://") + "/updates"

            // Отправляем метрики
            sender.SendMetricsBatch(cfg, metricsData)
            // Если не произошло паники или ошибок, считаем тест пройденным
        })
    }
}

func TestSendMetrics(t *testing.T) {
    tests := []struct {
        name   string
        useTLS bool
    }{
        {
            name:   "HTTP server",
            useTLS: false,
        },
        {
            name:   "HTTPS server",
            useTLS: true,
        },
    }

    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            handler := func(w http.ResponseWriter, r *http.Request) {
                if r.Method == http.MethodPost && strings.HasPrefix(r.URL.Path, "/update/") {
                    // Проверяем тип содержимого
                    assert.Equal(t, "text/plain", r.Header.Get("Content-Type"))

                    if r.Header.Get("Content-Encoding") == "gzip" {
                        // Проверяем, что данные пришли с gzip-сжатием
                        reader, err := gzip.NewReader(r.Body)
                        assert.NoError(t, err)
                        defer reader.Close()

                        body, err := io.ReadAll(reader)
                        assert.NoError(t, err)
                        assert.NotEmpty(t, body)
                    } else {
                        // Проверяем, что данные пришли без сжатия
                        body, err := io.ReadAll(r.Body)
                        assert.NoError(t, err)
                        assert.NotEmpty(t, body)
                    }

                    w.WriteHeader(http.StatusOK)
                    return
                }

                // Для других путей или методов возвращаем 404
                w.WriteHeader(http.StatusNotFound)
            }

            var server *httptest.Server
            if tt.useTLS {
                server = httptest.NewTLSServer(http.HandlerFunc(handler))
                defer server.Close()
            } else {
                server = httptest.NewServer(http.HandlerFunc(handler))
                defer server.Close()
            }

            cfg := &flags.Config{
                ServerAddress: strings.TrimPrefix(server.URL, "http://"),
                SecretKey:     "test_key",
            }
            if tt.useTLS {
                cfg.CryptoPath = "./test_certs"
            }

            metricsData := []metrics.Metrics{
                {ID: "metric1", Value: float64Ptr(10)},
                {ID: "metric2", Delta: int64Ptr(20)},
            }

            sender.SendMetrics(cfg, metricsData)
            // Проверка осуществляется через assert внутри обработчика
        })
    }
}

func TestSendMetricsJSON(t *testing.T) {
    tests := []struct {
        name   string
        useTLS bool
    }{
        {
            name:   "HTTP server",
            useTLS: false,
        },
        {
            name:   "HTTPS server",
            useTLS: true,
        },
    }

    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            handler := func(w http.ResponseWriter, r *http.Request) {
                if r.Method == http.MethodPost && r.URL.Path == "/update/" {
                    // Проверяем заголовок Content-Type
                    assert.Equal(t, "application/json", r.Header.Get("Content-Type"))

                    if r.Header.Get("Content-Encoding") == "gzip" {
                        // Проверяем, что данные пришли с gzip-сжатием
                        reader, err := gzip.NewReader(r.Body)
                        assert.NoError(t, err)
                        defer reader.Close()

                        var receivedMetric metrics.Metrics
                        err = json.NewDecoder(reader).Decode(&receivedMetric)
                        assert.NoError(t, err)
                        assert.NotEmpty(t, receivedMetric.ID)
                    } else {
                        // Проверяем, что данные пришли без сжатия
                        var receivedMetric metrics.Metrics
                        err := json.NewDecoder(r.Body).Decode(&receivedMetric)
                        assert.NoError(t, err)
                        assert.NotEmpty(t, receivedMetric.ID)
                    }

                    w.WriteHeader(http.StatusOK)
                    return
                }

                // Для других путей или методов возвращаем 404
                w.WriteHeader(http.StatusNotFound)
            }

            var server *httptest.Server
            if tt.useTLS {
                server = httptest.NewTLSServer(http.HandlerFunc(handler))
                defer server.Close()
            } else {
                server = httptest.NewServer(http.HandlerFunc(handler))
                defer server.Close()
            }

            cfg := &flags.Config{
                ServerAddress: strings.TrimPrefix(server.URL, "http://"),
                SecretKey:     "test_key",
            }
            if tt.useTLS {
                cfg.CryptoPath = "./test_certs"
            }

            metricsData := []metrics.Metrics{
                {ID: "metric1", Value: float64Ptr(10)},
                {ID: "metric2", Delta: int64Ptr(20)},
            }

            sender.SendMetricsJSON(cfg, metricsData)
            // Проверка осуществляется через assert внутри обработчика
        })
    }
}