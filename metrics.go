package main

import (
	"net/http"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/sirupsen/logrus"
)

var (
	numAddedDesc    = prometheus.NewDesc("esproxy_indexer_added", "", []string{"index"}, nil)
	numFlushedDesc  = prometheus.NewDesc("esproxy_indexer_flushed", "", []string{"index"}, nil)
	numFailedDesc   = prometheus.NewDesc("esproxy_indexer_failed", "", []string{"index"}, nil)
	numIndexedDesc  = prometheus.NewDesc("esproxy_indexer_indexed", "", []string{"index"}, nil)
	numCreatedDesc  = prometheus.NewDesc("esproxy_indexer_created", "", []string{"index"}, nil)
	numUpdatedDesc  = prometheus.NewDesc("esproxy_indexer_updated", "", []string{"index"}, nil)
	numDeletedDesc  = prometheus.NewDesc("esproxy_indexer_deleted", "", []string{"index"}, nil)
	numRequestsDesc = prometheus.NewDesc("esproxy_indexer_requests", "", []string{"index"}, nil)

	indexerRequestsCounter = promauto.NewCounterVec(prometheus.CounterOpts{
		Name: "esproxy_indexer_requests_served",
	}, []string{"method"})

	proxyRequestsCounter = promauto.NewCounterVec(prometheus.CounterOpts{
		Name: "esproxy_proxy_requests_served",
	}, []string{"method"})
)

func newIndexerMetricsCollector(indexer *Indexer) *indexerMetricsCollector {
	return &indexerMetricsCollector{indexer: indexer}
}

type indexerMetricsCollector struct {
	indexer *Indexer
}

func (i indexerMetricsCollector) Describe(descs chan<- *prometheus.Desc) {
	prometheus.DescribeByCollect(i, descs)
}

func (i indexerMetricsCollector) Collect(metrics chan<- prometheus.Metric) {
	for name, index := range i.indexer.indexes {
		metrics <- prometheus.MustNewConstMetric(numAddedDesc, prometheus.CounterValue, float64(index.Stats().NumAdded), name)
		metrics <- prometheus.MustNewConstMetric(numFlushedDesc, prometheus.CounterValue, float64(index.Stats().NumFlushed), name)
		metrics <- prometheus.MustNewConstMetric(numFailedDesc, prometheus.CounterValue, float64(index.Stats().NumFailed), name)
		metrics <- prometheus.MustNewConstMetric(numIndexedDesc, prometheus.CounterValue, float64(index.Stats().NumIndexed), name)
		metrics <- prometheus.MustNewConstMetric(numCreatedDesc, prometheus.CounterValue, float64(index.Stats().NumCreated), name)
		metrics <- prometheus.MustNewConstMetric(numUpdatedDesc, prometheus.CounterValue, float64(index.Stats().NumUpdated), name)
		metrics <- prometheus.MustNewConstMetric(numDeletedDesc, prometheus.CounterValue, float64(index.Stats().NumDeleted), name)
		metrics <- prometheus.MustNewConstMetric(numRequestsDesc, prometheus.CounterValue, float64(index.Stats().NumRequests), name)
	}
}

func serveMetrics() {
	http.Handle("/metrics", promhttp.Handler())
	err := http.ListenAndServe(":8080", nil)
	if err != nil {
		logrus.Fatal(err)
	}
}
