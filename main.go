package main

import (
	"bufio"
	"encoding/json"
	"errors"
	"github.com/kelseyhightower/envconfig"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/sirupsen/logrus"
	"io"
	"log"
	"net/http"
	"net/http/httputil"
	"net/url"
	"time"

	"github.com/gorilla/mux"
)

var indexer *Indexer

func main() {
	var c Configuration
	err := envconfig.Process("ESPROXY", &c)
	if err != nil {
		log.Fatal(err.Error())
	}

	FlushInterval = time.Duration(c.FlushInterval) * time.Second

	logrus.SetFormatter(&logrus.JSONFormatter{})

	if c.Debug {
		logrus.SetLevel(logrus.DebugLevel)
	}

	srv := &http.Server{
		Addr:         c.ListenAddress,
		WriteTimeout: 5 * time.Minute,
		ReadTimeout:  5 * time.Minute,
	}

	tu, err := url.Parse(c.ESAddress)
	if err != nil {
		logrus.Fatal(err)
	}

	logrus.Debugf("es address: %s", c.ESAddress)

	proxy := httputil.NewSingleHostReverseProxy(tu)

	indexer, err = NewIndexer(c.ESAddress)
	if err != nil {
		logrus.Fatal(err)
	}

	prometheus.DefaultRegisterer.MustRegister(newIndexerMetricsCollector(indexer))

	if c.Debug {
		indexer.Debug()
	}

	r := mux.NewRouter()

	r.Methods("POST").Path("/_bulk").HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
		writer.Header().Set("Content-Type", "application/json; charset=UTF-8")
		_, err := writer.Write([]byte("{\"errors\": false, \"items\": []}"))
		if err != nil {
			logrus.Warning(err)
		}

		err = processBulk(request.Body)
		if err != nil {
			logrus.Error(err)
		}

		indexerRequestsCounter.With(prometheus.Labels{"method": request.Method}).Inc()
		logrus.WithFields(logrus.Fields{
			"method": request.Method,
			"url":    request.RequestURI,
		}).Debug()
	})

	r.PathPrefix("/").HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
		proxy.ServeHTTP(writer, request)

		proxyRequestsCounter.With(prometheus.Labels{"method": request.Method}).Inc()
		logrus.WithFields(logrus.Fields{
			"method": request.Method,
			"url":    request.RequestURI,
		}).Debug()
	})

	srv.Handler = r

	go serveMetrics()

	logrus.Infof("listening %s", c.ListenAddress)
	err = srv.ListenAndServe()
	if err != nil {
		log.Fatal(err)
	}
}

func processBulk(r io.Reader) error {
	scanner := bufio.NewScanner(r)

	for scanner.Scan() {
		md := scanner.Bytes()

		var meta bulkMetadata

		err := json.Unmarshal(md, &meta)
		if err != nil {
			return err
		}

		var doc []byte
		if meta.Action() != "delete" {
			// Read record
			if valid := scanner.Scan(); !valid {
				return errors.New("corrupted data")
			}

			doc = scanner.Bytes()
		}

		err = indexer.Add(&meta, doc)
		if err != nil {
			logrus.WithField("meta", meta).Error(err)
		}
	}

	return nil
}
