package monitor

import (
	"github.com/pkg/errors"
	"github.com/prometheus/client_golang/prometheus"
)

type MetricType int

const (
	None MetricType = iota
	Counter
	Gauge
	Histogram
	Summary
)

// Metric 定义度量对象, 用户可以使用它来保存度量数据,每个指标都应该是全局唯一的
type Metric struct {
	Type        MetricType
	Name        string
	Description string
	Labels      []string
	Buckets     []float64
	Objectives  map[float64]float64

	cs prometheus.Collector
}

// SetGaugeValue 设置Gauge类型的Metric值
func (m *Metric) SetGaugeValue(labelValues []string, value float64) error {
	if m.Type == None {
		return errors.Errorf("metric '%s' not existed.", m.Name)
	}

	if m.Type != Gauge {
		return errors.Errorf("metric '%s' not Gauge type", m.Name)
	}

	m.cs.(*prometheus.GaugeVec).WithLabelValues(labelValues...).Set(value)
	return nil
}

// Inc 增加Counter或者Gauge类型的Metric的值，计数器增加1
func (m *Metric) Inc(labelValues []string) error {
	if m.Type == None {
		return errors.Errorf("metric '%s' not existed.", m.Name)
	}

	if m.Type != Gauge && m.Type != Counter {
		return errors.Errorf("metric '%s' not Gauge or Counter type", m.Name)
	}
	switch m.Type {
	case Counter:
		m.cs.(*prometheus.CounterVec).WithLabelValues(labelValues...).Inc()
		break
	case Gauge:
		m.cs.(*prometheus.GaugeVec).WithLabelValues(labelValues...).Inc()
		break
	}
	return nil
}

// Add 将给定的值添加到Metric对象仅，适用于Counter/Gauge
func (m *Metric) Add(labelValues []string, value float64) error {
	if m.Type == None {
		return errors.Errorf("metric '%s' not existed.", m.Name)
	}

	if m.Type != Gauge && m.Type != Counter {
		return errors.Errorf("metric '%s' not Gauge or Counter type", m.Name)
	}
	switch m.Type {
	case Counter:
		m.cs.(*prometheus.CounterVec).WithLabelValues(labelValues...).Add(value)
		break
	case Gauge:
		m.cs.(*prometheus.GaugeVec).WithLabelValues(labelValues...).Add(value)
		break
	}
	return nil
}

// Observe 被用作 Histogram 和 Summary 类型的Metric
func (m *Metric) Observe(labelValues []string, value float64) error {
	if m.Type == 0 {
		return errors.Errorf("metric '%s' not existed.", m.Name)
	}
	if m.Type != Histogram && m.Type != Summary {
		return errors.Errorf("metric '%s' not Histogram or Summary type", m.Name)
	}
	switch m.Type {
	case Histogram:
		m.cs.(*prometheus.HistogramVec).WithLabelValues(labelValues...).Observe(value)
		break
	case Summary:
		m.cs.(*prometheus.SummaryVec).WithLabelValues(labelValues...).Observe(value)
		break
	}
	return nil
}
