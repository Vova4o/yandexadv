package handler

import (
	"bytes"
	"encoding/json"
	"errors"
	"html/template"
	"net/http"
	"net/http/httptest"
	"strconv"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/vova4o/yandexadv/internal/models"
)

// MockService is a mock implementation of the Service interface
type MockService struct {
	mock.Mock
}

func (m *MockService) UpdateServ(metric models.Metric) error {
	args := m.Called(metric)
	return args.Error(0)
}

func (m *MockService) UpdateServJSON(metric *models.Metrics) error {
	args := m.Called(metric)
	return args.Error(0)
}

func (m *MockService) GetValueServ(metric models.Metrics) (string, error) {
	args := m.Called(metric)
	return args.String(0), args.Error(1)
}

func (m *MockService) GetValueServJSON(metric models.Metrics) (*models.Metrics, error) {
	args := m.Called(metric)
	return args.Get(0).(*models.Metrics), args.Error(1)
}

func (m *MockService) MetrixStatistic() (*template.Template, map[string]models.Metrics, error) {
	args := m.Called()
	return args.Get(0).(*template.Template), args.Get(1).(map[string]models.Metrics), args.Error(2)
}

func (m *MockService) UpdateBatchMetricsServ(metrics []models.Metrics) error {
	args := m.Called(metrics)
	return args.Error(0)
}

func (m *MockService) PingDB() error {
	args := m.Called()
	return args.Error(0)
}

func TestGetValueHandler(t *testing.T) {
	router := gin.Default()
	mockService := new(MockService)
	r := &Router{Service: mockService}
	router.GET("/value/:type/:name", r.GetValueHandler)

	tests := []struct {
		name           string
		metricType     string
		metricName     string
		mockReturn     string
		mockError      error
		expectedStatus int
		expectedBody   string
	}{
		{
			name:           "Metric found",
			metricType:     "gauge",
			metricName:     "metric1",
			mockReturn:     "10.5",
			mockError:      nil,
			expectedStatus: http.StatusOK,
			expectedBody:   "10.5",
		},
		{
			name:           "Metric not found",
			metricType:     "gauge",
			metricName:     "metric2",
			mockReturn:     "",
			mockError:      models.ErrMetricNotFound,
			expectedStatus: http.StatusNotFound,
			expectedBody:   models.ErrMetricNotFound.Error(),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockService.On("GetValueServ", models.Metrics{MType: tt.metricType, ID: tt.metricName}).Return(tt.mockReturn, tt.mockError)

			req, _ := http.NewRequest(http.MethodGet, "/value/"+tt.metricType+"/"+tt.metricName, nil)
			w := httptest.NewRecorder()
			router.ServeHTTP(w, req)

			assert.Equal(t, tt.expectedStatus, w.Code)
			assert.Equal(t, tt.expectedBody, w.Body.String())
		})
	}
}

func TestUpdateMetricHandler(t *testing.T) {
	router := gin.Default()
	mockService := new(MockService)
	r := &Router{Service: mockService}
	router.POST("/update/:type/:name/:value", r.UpdateMetricHandler)

	tests := []struct {
		name           string
		metricType     string
		metricName     string
		metricValue    string
		mockError      error
		expectedStatus int
		expectedBody   string
	}{
		{
			name:           "Valid gauge metric",
			metricType:     "gauge",
			metricName:     "metric1",
			metricValue:    "10.5",
			mockError:      nil,
			expectedStatus: http.StatusOK,
			expectedBody:   "",
		},
		{
			name:           "Invalid gauge value",
			metricType:     "gauge",
			metricName:     "metric1",
			metricValue:    "invalid",
			mockError:      nil,
			expectedStatus: http.StatusBadRequest,
			expectedBody:   "invalid gauge value",
		},
		{
			name:           "Valid counter metric",
			metricType:     "counter",
			metricName:     "metric2",
			metricValue:    "5",
			mockError:      nil,
			expectedStatus: http.StatusOK,
			expectedBody:   "",
		},
		{
			name:           "Invalid counter value",
			metricType:     "counter",
			metricName:     "metric2",
			metricValue:    "invalid",
			mockError:      nil,
			expectedStatus: http.StatusBadRequest,
			expectedBody:   "invalid counter value",
		},
		{
			name:           "Invalid metric type",
			metricType:     "invalid",
			metricName:     "metric3",
			metricValue:    "10",
			mockError:      nil,
			expectedStatus: http.StatusBadRequest,
			expectedBody:   "invalid metric type",
		},
		{
			name:           "Service error",
			metricType:     "gauge",
			metricName:     "metric4",
			metricValue:    "20.5",
			mockError:      errors.New("service error"),
			expectedStatus: http.StatusInternalServerError,
			expectedBody:   "failed to update metric",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var metric models.Metrics
			switch tt.metricType {
			case "gauge":
				value, err := strconv.ParseFloat(tt.metricValue, 64)
				if err == nil {
					metric = models.Metrics{
						ID:    tt.metricName,
						MType: tt.metricType,
						Value: &value,
					}
				}
			case "counter":
				delta, err := strconv.ParseInt(tt.metricValue, 10, 64)
				if err == nil {
					metric = models.Metrics{
						ID:    tt.metricName,
						MType: tt.metricType,
						Delta: &delta,
					}
				}
			}

			mockService.On("UpdateServJSON", &metric).Return(tt.mockError)

			req, _ := http.NewRequest(http.MethodPost, "/update/"+tt.metricType+"/"+tt.metricName+"/"+tt.metricValue, nil)
			w := httptest.NewRecorder()
			router.ServeHTTP(w, req)

			assert.Equal(t, tt.expectedStatus, w.Code)
			assert.Equal(t, tt.expectedBody, w.Body.String())
		})
	}
}

func TestPingHandler(t *testing.T) {
	router := gin.Default()
	mockService := new(MockService)
	r := &Router{Service: mockService}
	router.GET("/ping", r.PingHandler)

	tests := []struct {
		name           string
		mockError      error
		expectedStatus int
		expectedBody   string
	}{
		{
			name:           "Ping successful",
			mockError:      nil,
			expectedStatus: http.StatusOK,
			expectedBody:   "pong",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockService.On("PingDB").Return(tt.mockError)

			req, _ := http.NewRequest(http.MethodGet, "/ping", nil)
			w := httptest.NewRecorder()
			router.ServeHTTP(w, req)

			assert.Equal(t, tt.expectedStatus, w.Code)
			assert.Equal(t, tt.expectedBody, w.Body.String())
		})
	}
}

func TestUpdateBatchMetricsHandler(t *testing.T) {
    router := gin.Default()
    mockService := new(MockService)
    r := &Router{Service: mockService}
    router.POST("/update-batch", r.UpdateBatchMetricsHandler)

    tests := []struct {
        name           string
        requestBody    []models.Metrics
        mockError      error
        expectedStatus int
        expectedBody   string
    }{
        {
            name: "Valid batch update",
            requestBody: []models.Metrics{
                {ID: "metric1", MType: "gauge", Value: float64Ptr(10.5)},
                {ID: "metric2", MType: "counter", Delta: int64Ptr(5)},
            },
            mockError:      nil,
            expectedStatus: http.StatusOK,
            expectedBody:   "",
        },
        {
            name:           "Invalid JSON",
            requestBody:    nil,
            mockError:      nil,
            expectedStatus: http.StatusBadRequest,
            expectedBody:   "bad request",
        },
        // {
        //     name: "Service error",
        //     requestBody: []models.Metrics{
        //         {ID: "metric1", MType: "gauge", Value: float64Ptr(10.5)},
        //         {ID: "metric2", MType: "counter", Delta: int64Ptr(5)},
        //     },
        //     mockError:      errors.New("service error"),
        //     expectedStatus: http.StatusInternalServerError,
        //     expectedBody:   "internal server error",
        // },
    }

    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            var reqBody []byte
            if tt.requestBody != nil {
                reqBody, _ = json.Marshal(tt.requestBody)
            } else {
                reqBody = []byte("invalid json")
            }

            mockService.On("UpdateBatchMetricsServ", mock.Anything).Return(tt.mockError)

            req, _ := http.NewRequest(http.MethodPost, "/update-batch", bytes.NewBuffer(reqBody))
            req.Header.Set("Content-Type", "application/json")
            w := httptest.NewRecorder()
            router.ServeHTTP(w, req)

            assert.Equal(t, tt.expectedStatus, w.Code)
            assert.Equal(t, tt.expectedBody, w.Body.String())
        })
    }
}

func float64Ptr(v float64) *float64 {
    return &v
}

func int64Ptr(v int64) *int64 {
    return &v
}