/*
Copyright 2020 Backup Operator Authors

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package metrics

import (
	"fmt"
	"os"
	"time"

	"github.com/kubism/backup-operator/pkg/logger"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/push"
)

type MetricsPublisher interface {
	StartTimer()
	StopTimer()
	SetSuccessfulRun()
	SetBackupSizeInBytes(sizeInBytes int64)
	PublishMetrics()
}

func NewNopMetricsPublisher() MetricsPublisher {
	return nopMetricsPublisher{}
}

type MetricsPublisherConfig struct {
	// Connection info for pushgateway
	URL      string
	Username string
	Password string
	// Labels
	Namespace string
	Pod       string
	Job       string
	App       string
}

func (c *MetricsPublisherConfig) WithURL(url string) *MetricsPublisherConfig {
	c.URL = url
	return c
}

func (c *MetricsPublisherConfig) WithUsername(username string) *MetricsPublisherConfig {
	c.Username = username
	return c
}

func (c *MetricsPublisherConfig) WithPassword(password string) *MetricsPublisherConfig {
	c.Password = password
	return c
}

func (c *MetricsPublisherConfig) WithApp(app string) *MetricsPublisherConfig {
	c.App = app
	return c
}

func (c *MetricsPublisherConfig) Validate() error {
	if c.URL == "" { // TODO: parse URL once to check for errors
		return fmt.Errorf("Invalid URL for pushgateway: %s", c.URL)
	}
	return nil
}

func DefaultConfig() *MetricsPublisherConfig {
	c := &MetricsPublisherConfig{
		URL:       "",
		Username:  "",
		Password:  "",
		Namespace: os.Getenv("K8S_NAMESPACE"),
		Pod:       os.Getenv("K8S_POD"),
		Job:       os.Getenv("K8S_JOB"),
		App:       "",
	}
	if c.Job == "" {
		c.Job = "backup_operator"
	}
	return c
}

func NewMetricsPublisher(c *MetricsPublisherConfig) MetricsPublisher {
	p := metricsPublisher{
		completionTime: prometheus.NewGauge(prometheus.GaugeOpts{
			Name: "backup_last_completion_timestamp_seconds",
			Help: "The timestamp of the last completion of a backup, successful or not.",
		}),
		successTime: prometheus.NewGauge(prometheus.GaugeOpts{
			Name: "backup_last_success_timestamp_seconds",
			Help: "The timestamp of the last successful completion of a backup.",
		}),
		duration: prometheus.NewGauge(prometheus.GaugeOpts{
			Name: "backup_duration_seconds",
			Help: "The duration of the last backup in seconds.",
		}),
		sizeInBytes: prometheus.NewGauge(prometheus.GaugeOpts{
			Name: "backup_size_in_bytes",
			Help: "The size in bytes of the last backup.",
		}),
		log: logger.WithName("metrics"),
	}

	registry := prometheus.NewRegistry()
	registry.MustRegister(p.completionTime, p.duration, p.sizeInBytes, prometheus.NewGoCollector())

	pusher := push.New(c.URL, c.Job).Gatherer(registry)

	if c.App != "" {
		pusher = pusher.Grouping("app", c.App)
	}
	if c.Namespace != "" {
		pusher = pusher.Grouping("namespace", c.Namespace)
	}

	if c.Pod != "" {
		pusher = pusher.Grouping("pod", c.Pod)
	}

	if c.Username != "" && c.Password != "" {
		pusher = pusher.BasicAuth(c.Username, c.Password)
	}

	p.pusher = pusher

	return &p
}

type metricsPublisher struct {
	pusher         *push.Pusher
	log            logger.Logger
	completionTime prometheus.Gauge
	successTime    prometheus.Gauge
	duration       prometheus.Gauge
	sizeInBytes    prometheus.Gauge
	start          time.Time
}

func (m *metricsPublisher) StartTimer() {
	m.start = time.Now()
}

func (m *metricsPublisher) StopTimer() {
	m.duration.Set(time.Since(m.start).Seconds())
	m.completionTime.SetToCurrentTime()
}

func (m *metricsPublisher) SetSuccessfulRun() {
	m.pusher.Collector(m.successTime)
	m.successTime.SetToCurrentTime()
}

func (m *metricsPublisher) SetBackupSizeInBytes(sizeInBytes int64) {
	m.sizeInBytes.Set(float64(sizeInBytes))
}

func (m *metricsPublisher) PublishMetrics() {
	err := m.pusher.Add()
	if err != nil { // TODO: should we error for real?
		m.log.Error(err, "failed to push metrics to Prometheus")
	} else {
		m.log.Info("published metrics about this run to Prometheus")
	}
}

type nopMetricsPublisher struct {
}

func (n nopMetricsPublisher) StartTimer() {
}

func (n nopMetricsPublisher) StopTimer() {
}

func (n nopMetricsPublisher) SetSuccessfulRun() {
}

func (n nopMetricsPublisher) SetBackupSizeInBytes(_ int64) {
}

func (n nopMetricsPublisher) PublishMetrics() {
}
