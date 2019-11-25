// Copyright 2014 The Prometheus Authors
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
// http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package handler

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"net/http"
	"strings"
	"sync"

	"github.com/go-kit/kit/log"
	"github.com/go-kit/kit/log/level"
	"github.com/julienschmidt/httprouter"
	"github.com/pquerna/ffjson/ffjson"
)

// RbdPushReq holds the request body.
type RbdPushReq struct {
	Metrics []Metric `json:"metrics"`
	Labels  []Label  `json:"labels"`
}

// Metric metric in request body
type Metric struct {
	MetricName  string  `json:"metric_name"`
	MetricValue float64 `json:"metric_value"`
	MetricType  string  `json:"metric_type"`
}

// Label label in request body
type Label struct {
	Key   string `json:"key"`
	Value string `json:"value"`
}

// RbdPush returns an http.Handler which accepts samples over HTTP and stores them
// in the MetricStore. If replace is true, all metrics for the job and instance
// given by the request are deleted before new ones are stored.
//
// The returned handler is already instrumented for Prometheus.
//
// Customize on the original Push method
func RbdPush(next func(http.ResponseWriter, *http.Request, httprouter.Params), logger log.Logger) func(http.ResponseWriter, *http.Request, httprouter.Params) {
	var ps httprouter.Params
	var mtx sync.Mutex // Protects ps.

	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		job := ps.ByName("job")
		if job == "" {
			http.Error(w, "job name is required", http.StatusBadRequest)
			level.Debug(logger).Log("msg", "job name is required")
			return
		}
		mtx.Unlock()

		labels := map[string]string{
			"job": job,
		}

		// parse body to metrics and labels
		body, err := ioutil.ReadAll(r.Body)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			level.Error(logger).Log("msg", err.Error())
			return
		}
		req, err := parseBodyToRbdPushReq(body)
		if err != nil {
			http.Error(w, "bad request", http.StatusBadRequest)
			level.Error(logger).Log("msg", "bad request", err.Error())
			return
		}

		// convert metrics and labels to text
		for k, v := range labels {
			req.Labels = append(req.Labels, Label{Key: k, Value: v})
		}
		text := convertRbdPushReq2Text(req)

		// set a new body, which will simulate the same data we read
		r.Body = ioutil.NopCloser(bytes.NewBuffer([]byte(text)))

		next(w, r, ps)
	})

	return func(w http.ResponseWriter, r *http.Request, params httprouter.Params) {
		mtx.Lock()
		ps = params
		handler.ServeHTTP(w, r)
	}
}

func parseBodyToRbdPushReq(body []byte) (*RbdPushReq, error) {
	var res RbdPushReq
	if err := ffjson.Unmarshal(body, &res); err != nil {
		return nil, fmt.Errorf("body: %s: %v", body, err)
	}

	return &res, nil
}

func convertRbdPushReq2Text(req *RbdPushReq) string {
	var labelTexts []string
	for _, label := range req.Labels {
		labelTexts = append(labelTexts, fmt.Sprintf("%s=\"%s\"", label.Key, label.Value))
	}
	labelText := strings.Join(labelTexts, ",")

	var metricTexts []string
	for _, metric := range req.Metrics {
		var typeText string
		if metric.MetricType != "" {
			typeText = fmt.Sprintf("# TYPE %s %s\n", metric.MetricName, metric.MetricType)
		}

		var metricText string
		if len(labelText) == 0 {
			metricText = fmt.Sprintf("%s%s %v\n", typeText, metric.MetricName, metric.MetricValue)
		} else {
			metricText = fmt.Sprintf("%s%s{%s} %v\n", typeText, metric.MetricName, labelText, metric.MetricValue)
		}
		metricTexts = append(metricTexts, metricText)
	}

	res := strings.Join(metricTexts, "")

	return res
}
