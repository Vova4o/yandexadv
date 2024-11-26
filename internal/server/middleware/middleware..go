package middleware

import (
	"compress/gzip"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"io"
	"net/http"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/vova4o/yandexadv/package/logger"
	"go.uber.org/zap"
)

// Middleware структура для middleware
type Middleware struct {
	SecretKey string
	Logger    *logger.Logger
}

// New создание нового middleware
func New(log *logger.Logger, key string) *Middleware {
	return &Middleware{
		Logger:    log,
		SecretKey: key,
	}
}

// GzipReader - обертка для gzip.Reader
type GzipReader struct {
	io.ReadCloser
	reader *gzip.Reader
}

// GzipWriter - обертка для gzip.Writer
type GzipWriter struct {
	gin.ResponseWriter
	writer *gzip.Writer
}

// Пул объектов для gzip.Reader и gzip.Writer
var gzipReaderPool = sync.Pool{
	New: func() interface{} {
		return new(gzip.Reader)
	},
}

var gzipWriterPool = sync.Pool{
	New: func() interface{} {
		return gzip.NewWriter(nil)
	},
}

// Read - чтение данных из gzip.Reader
func (g *GzipReader) Read(p []byte) (int, error) {
	return g.reader.Read(p)
}

// Write - запись данных в gzip.Writer
func (g *GzipWriter) Write(data []byte) (int, error) {
	return g.writer.Write(data)
}

// CheckHash - проверка хэша
func (m Middleware) CheckHash() gin.HandlerFunc {
	return func(c *gin.Context) {
		m.Logger.Info("SecretKey", zap.String("SecretKey", m.SecretKey))
		if m.SecretKey == "" {
			c.Next()
			return
		}

		// Проверка хэша на этапе обработки запроса
		hash := c.GetHeader("HashSHA256")
		if hash == "" {
			c.AbortWithStatus(http.StatusBadRequest)
			return
		}

		// Чтение данных из тела запроса
		data, err := io.ReadAll(c.Request.Body)
		if err != nil {
			c.AbortWithStatus(http.StatusBadRequest)
			return
		}
		m.Logger.Info("data", zap.String("data", string(data)))

		c.Request.Body = io.NopCloser(strings.NewReader(string(data)))

		expectedHash := calculateHash(data, []byte(m.SecretKey))
		m.Logger.Info("Hash check", zap.String("result", fmt.Sprintf("%v", expectedHash == hash)))
		if hash != expectedHash {
			c.AbortWithStatus(http.StatusBadRequest)
			return
		}

		c.Next()

		// Добавление хэша в заголовок ответа на этапе формирования ответа
		responseData := []byte(c.Writer.Header().Get("Content-Type") + c.Request.URL.Path + c.Request.URL.RawQuery)
		responseHash := calculateHash(responseData, []byte(m.SecretKey))
		c.Writer.Header().Set("HashSHA256", responseHash)
	}
}

// calculateHash вычисляет HMAC-SHA256 хэш из данных и ключа
func calculateHash(data, key []byte) string {
	h := hmac.New(sha256.New, key)
	h.Write(data)
	return hex.EncodeToString(h.Sum(nil))
}

// GunzipMiddleware - middleware для распаковки запросов
func (m Middleware) GunzipMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		if strings.Contains(c.GetHeader("Content-Encoding"), "gzip") {
			gz := gzipReaderPool.Get().(*gzip.Reader)
			defer gzipReaderPool.Put(gz)

			if err := gz.Reset(c.Request.Body); err != nil {
				c.AbortWithStatus(http.StatusBadRequest)
				return
			}
			defer gz.Close()

			c.Request.Body = &GzipReader{c.Request.Body, gz}
		}
		c.Next()
	}
}

// GzipMiddleware - middleware для сжатия ответов
func (m Middleware) GzipMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		if strings.Contains(c.GetHeader("Accept-Encoding"), "gzip") {
			gz := gzipWriterPool.Get().(*gzip.Writer)
			defer gzipWriterPool.Put(gz)

			gz.Reset(c.Writer)
			defer gz.Close()

			c.Writer = &GzipWriter{c.Writer, gz}
			c.Header("Content-Encoding", "gzip")
		}
		c.Next()
	}
}

// GinZap возвращает middleware для логирования запросов с использованием zap
func (m Middleware) GinZap() gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()
		path := c.Request.URL.Path
		raw := c.Request.URL.RawQuery

		c.Next()

		latency := time.Since(start)

		if raw != "" {
			path = path + "?" + raw
		}

		// Получение размера содержимого ответа
		contentLength := c.Writer.Header().Get("Content-Length")
		if contentLength == "" {
			contentLength = "0"
		}

		// Преобразование размера содержимого в int
		contentLengthInt, err := strconv.Atoi(contentLength)
		if err != nil {
			m.Logger.Error("failed to parse content length", zap.String("content_length", contentLength), zap.Error(err))
			contentLengthInt = 0 // или установите значение по умолчанию
		}

		// Получение и парсинг значения заголовка X-Response-Time
		latencyStr := c.Writer.Header().Get("X-Response-Time")
		var parsedLatency time.Duration
		if latencyStr != "" {
			parsedLatency, err = time.ParseDuration(latencyStr)
			if err != nil {
				m.Logger.Error("failed to parse latency", zap.String("latency", latencyStr), zap.Error(err))
				parsedLatency = 0 // или установите значение по умолчанию
			}
		}

		m.Logger.Info("incoming request",
			zap.String("method", c.Request.Method),
			zap.String("path", path),
			zap.Duration("latency", latency),
			zap.Int("status", c.Writer.Status()),
			zap.String("client_ip", c.ClientIP()),
			zap.String("user_agent", c.Request.UserAgent()),
			zap.Int("content_length", contentLengthInt),
			zap.Duration("parsed_latency", parsedLatency),
		)
	}
}
