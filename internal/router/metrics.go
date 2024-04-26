package router

import (
	"github.com/gin-gonic/gin"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"

	"github.com/xmapst/osreapi/internal/worker"
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
		float64(worker.GetTotal()),
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

func init() {
	prometheus.MustRegister(taskMetrics)
}

func metrics(c *gin.Context) {
	h := promhttp.Handler()
	h.ServeHTTP(c.Writer, c.Request)
}
