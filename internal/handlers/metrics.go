package handlers

import (
	"github.com/gin-gonic/gin"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/xmapst/osreapi/internal/cache"
	"github.com/xmapst/osreapi/internal/config"
	"github.com/xmapst/osreapi/internal/engine/worker"
	"github.com/xmapst/osreapi/internal/logx"
)

var taskMetrics = &Metrics{
	task: prometheus.NewDesc(
		prometheus.BuildFQName("osreapi", "", "task"),
		"Running task count",
		[]string{"mode"},
		nil,
	),
}

type Metrics struct {
	task *prometheus.Desc
}

func (m *Metrics) Describe(ch chan<- *prometheus.Desc) {
	// This only needs to output *something*
	prometheus.NewGauge(prometheus.GaugeOpts{Name: "Dummy", Help: "Dummy"}).Describe(ch)
}

func (m *Metrics) Collect(ch chan<- prometheus.Metric) {
	ch <- prometheus.MustNewConstMetric(
		m.task,
		prometheus.CounterValue,
		float64(len(cache.GetAllByBeginTime())),
		"total",
	)
	ch <- prometheus.MustNewConstMetric(
		m.task,
		prometheus.CounterValue,
		float64(worker.Running()),
		"running",
	)
	ch <- prometheus.MustNewConstMetric(
		m.task,
		prometheus.CounterValue,
		float64(worker.Waiting()),
		"waiting",
	)
}

func metrics(c *gin.Context) {
	reg := prometheus.NewRegistry()
	reg.MustRegister(taskMetrics)
	h := promhttp.HandlerFor(reg, promhttp.HandlerOpts{
		ErrorLog:      logx.GetSubLogger(),      // If an error occurs during the collection process, record the log
		ErrorHandling: promhttp.ContinueOnError, // If an error occurs during the collection process, continue to collect other data without interrupting the work of the collector
		Timeout:       config.App.WebTimeout,    // Timeout
	})
	h.ServeHTTP(c.Writer, c.Request)
}
