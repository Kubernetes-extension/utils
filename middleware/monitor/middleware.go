package monitor

import (
	"fmt"
	"github.com/Kubernetes-extension/utils/middleware/monitor/bloom"
	"github.com/gin-gonic/gin"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"strconv"
	"time"
)

var (
	metricRequestTotal    = "gin_request_total"
	metricRequestUV       = "gin_request_uv"
	metricURIRequestTotal = "gin_uri_request_total"
	metricRequestBody     = "gin_request_body_total"
	metricResponseBody    = "gin_response_body_total"
	metricRequestDuration = "gin_request_duration"
	metricSlowRequest     = "gin_slow_request_total"

	bloomFilter *bloom.BloomFilter // 布隆过滤器
)

// initGinMetrics used to init gin metrics
func (m *Monitor) initGinMetrics() {
	bloomFilter = bloom.NewBloomFilter()

	_ = m.AddMetric(&Metric{
		Type:        Counter,
		Name:        metricRequestTotal,
		Description: "all the server received request num.",
		Labels:      nil,
	})
	_ = m.AddMetric(&Metric{
		Type:        Counter,
		Name:        metricRequestUV,
		Description: "all the server received ip num.",
		Labels:      nil,
	})
	_ = m.AddMetric(&Metric{
		Type:        Counter,
		Name:        metricURIRequestTotal,
		Description: "all the server received request num with every uri.",
		Labels:      []string{"uri", "method", "code"},
	})
	_ = m.AddMetric(&Metric{
		Type:        Counter,
		Name:        metricRequestBody,
		Description: "the server received request body size, unit byte",
		Labels:      nil,
	})
	_ = m.AddMetric(&Metric{
		Type:        Counter,
		Name:        metricResponseBody,
		Description: "the server send response body size, unit byte",
		Labels:      nil,
	})
	_ = m.AddMetric(&Metric{
		Type:        Histogram,
		Name:        metricRequestDuration,
		Description: "the time server took to handle the request.",
		Labels:      []string{"uri"},
		Buckets:     m.reqDuration,
	})
	_ = m.AddMetric(&Metric{
		Type:        Counter,
		Name:        metricSlowRequest,
		Description: fmt.Sprintf("the server handled slow requests counter, t=%d.", m.slowTime),
		Labels:      []string{"uri", "method", "code"},
	})
}

// Use set gin metrics middleware
func (m *Monitor) Use(r *gin.Engine) {
	m.initGinMetrics()

	r.Use(m.monitorInterceptor)
	r.GET(m.metricPath, func(ctx *gin.Context) {
		promhttp.Handler().ServeHTTP(ctx.Writer, ctx.Request)
	})
}
func (m *Monitor) monitorInterceptor(ctx *gin.Context) {
	if ctx.Request.URL.Path == m.metricPath {
		ctx.Next()
		return
	}
	startTime := time.Now()

	// execute normal process.
	ctx.Next()

	// after request
	m.metricHandle(ctx, startTime)
}

func (m *Monitor) metricHandle(ctx *gin.Context, start time.Time) {
	// set request total
	_ = m.GetMetric(metricRequestTotal).Inc(nil)

	// set uv
	// 使用布隆过滤器计算UV
	if clientIP := ctx.ClientIP(); !bloomFilter.Contains(clientIP) {
		bloomFilter.Add(clientIP)
		_ = m.GetMetric(metricRequestUV).Inc(nil)
	}

	// set uri request total
	// ctx.FullPath()  "/user/:id"
	// ctx.Request.URL.Path  "user/100"
	_ = m.GetMetric(metricURIRequestTotal).Inc([]string{ctx.Request.URL.Path, ctx.Request.Method, strconv.Itoa(ctx.Writer.Status())})

	// set request body size
	_ = m.GetMetric(metricRequestBody).Add(nil, float64(ctx.Request.ContentLength))

	// set slow request
	latency := time.Since(start)
	if int32(latency.Seconds()) > m.slowTime {
		_ = m.GetMetric(metricSlowRequest).Inc([]string{ctx.Request.URL.Path, ctx.Request.Method, strconv.Itoa(ctx.Writer.Status())})
	}

	// set request duration
	_ = m.GetMetric(metricRequestDuration).Observe([]string{ctx.Request.URL.Path}, latency.Seconds())

	// set response body size
	if ctx.Writer.Size() > 0 {
		_ = m.GetMetric(metricResponseBody).Add(nil, float64(ctx.Writer.Size()))
	}
}
