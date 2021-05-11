package monitor

import (
	"github.com/pkg/errors"
	"github.com/prometheus/client_golang/prometheus"
)

const (
	defaultMetricPath = "/metrics"
	defaultSlowTime   = int32(5)
)

var (
	defaultDuration = []float64{0.1, 0.3, 1.2, 5, 10}

	typeHandler = map[MetricType]func(metric *Metric) error{
		Counter:   counterHandler,
		Gauge:     gaugeHandler,
		Histogram: histogramHandler,
		Summary:   summaryHandler,
	}
)

type Monitor struct {
	slowTime    int32
	metricPath  string
	reqDuration []float64
	metrics     map[string]*Metric
}

// NewMonitor 初始化一个Monitor结构体对象
func NewMonitor(metricPath string, slowTime int32, reqDuration []float64) *Monitor {
	if metricPath == "" {
		metricPath = defaultMetricPath
	}
	if slowTime == 0 {
		slowTime = defaultSlowTime
	}
	if reqDuration == nil {
		reqDuration = defaultDuration
	}
	return &Monitor{
		metricPath:  metricPath,
		slowTime:    slowTime,
		reqDuration: reqDuration,
		metrics:     make(map[string]*Metric),
	}
}

// GetMetric 通过metric名获取Metric对象
func (m *Monitor) GetMetric(name string) *Metric {
	if metric, ok := m.metrics[name]; ok {
		return metric
	}
	return &Metric{}
}

// AddMetric 添加metric
func (m *Monitor) AddMetric(metric *Metric) error {
	if _, ok := m.metrics[metric.Name]; ok {
		return errors.Errorf("metric '%s' is existed", metric.Name)
	}

	if metric.Name == "" {
		return errors.Errorf("metric name cannot be empty.")
	}
	if f, ok := typeHandler[metric.Type]; ok {
		if err := f(metric); err == nil {
			prometheus.MustRegister(metric.cs)
			m.metrics[metric.Name] = metric
		}
	}
	return errors.Errorf("metric type '%d' not existed.", metric.Type)
}

func counterHandler(metric *Metric) error {
	metric.cs = prometheus.NewCounterVec(
		prometheus.CounterOpts{Name: metric.Name, Help: metric.Description},
		metric.Labels,
	)
	return nil
}

func gaugeHandler(metric *Metric) error {
	metric.cs = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{Name: metric.Name, Help: metric.Description},
		metric.Labels,
	)
	return nil
}

func histogramHandler(metric *Metric) error {
	if len(metric.Buckets) == 0 {
		return errors.Errorf("metric '%s' is histogram type, cannot lose bucket param.", metric.Name)
	}
	metric.cs = prometheus.NewHistogramVec(
		prometheus.HistogramOpts{Name: metric.Name, Help: metric.Description, Buckets: metric.Buckets},
		metric.Labels,
	)
	return nil
}

func summaryHandler(metric *Metric) error {
	if len(metric.Objectives) == 0 {
		return errors.Errorf("metric '%s' is summary type, cannot lose objectives param.", metric.Name)
	}
	prometheus.NewSummaryVec(
		prometheus.SummaryOpts{Name: metric.Name, Help: metric.Description, Objectives: metric.Objectives},
		metric.Labels,
	)
	return nil
}
