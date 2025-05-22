package metrics

import (
	"net/http"
	"sync"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

var (
	once sync.Once
)

// Metrics collects scheduler metrics
type Metrics struct {
	TasksQueued      prometheus.Gauge
	TasksInProgress  prometheus.Gauge
	TasksCompleted   *prometheus.CounterVec
	TaskDuration     *prometheus.HistogramVec
	ResourceUsage    *prometheus.GaugeVec
	TaskTypeCount    *prometheus.CounterVec
	WorkerErrors     *prometheus.CounterVec
	ResourceRequests *prometheus.GaugeVec
}

var instance *Metrics

// GetMetrics returns singleton metrics instance
func GetMetrics() *Metrics {
	once.Do(func() {
		instance = &Metrics{
			TasksQueued: prometheus.NewGauge(prometheus.GaugeOpts{
				Name: "scheduler_tasks_queued",
				Help: "Number of tasks currently queued",
			}),
			TasksInProgress: prometheus.NewGauge(prometheus.GaugeOpts{
				Name: "scheduler_tasks_in_progress",
				Help: "Number of tasks currently being processed",
			}),
			TasksCompleted: prometheus.NewCounterVec(prometheus.CounterOpts{
				Name: "scheduler_tasks_completed_total",
				Help: "Total number of completed tasks",
			}, []string{"status"}),
			TaskDuration: prometheus.NewHistogramVec(prometheus.HistogramOpts{
				Name:    "scheduler_task_duration_seconds",
				Help:    "Task processing duration distribution",
				Buckets: prometheus.DefBuckets,
			}, []string{"type"}),
			ResourceUsage: prometheus.NewGaugeVec(prometheus.GaugeOpts{
				Name: "scheduler_resource_usage",
				Help: "Current resource usage",
			}, []string{"resource", "unit"}),
			TaskTypeCount: prometheus.NewCounterVec(prometheus.CounterOpts{
				Name: "scheduler_task_type_total",
				Help: "Count of tasks by type",
			}, []string{"type"}),
			WorkerErrors: prometheus.NewCounterVec(prometheus.CounterOpts{
				Name: "scheduler_worker_errors_total",
				Help: "Count of worker errors by type",
			}, []string{"type"}),
			ResourceRequests: prometheus.NewGaugeVec(prometheus.GaugeOpts{
				Name: "scheduler_resource_requests",
				Help: "Resource requests by workers",
			}, []string{"resource", "unit"}),
		}

		prometheus.MustRegister(
			instance.TasksQueued,
			instance.TasksInProgress,
			instance.TasksCompleted,
			instance.TaskDuration,
			instance.ResourceUsage,
			instance.TaskTypeCount,
			instance.WorkerErrors,
			instance.ResourceRequests,
		)
	})
	return instance
}

// StartMetricsServer starts HTTP server for metrics
func StartMetricsServer(addr string) {
	http.Handle("/metrics", promhttp.Handler())
	go http.ListenAndServe(addr, nil)
}
