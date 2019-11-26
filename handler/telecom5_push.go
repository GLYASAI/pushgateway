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
	"strings"
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
	"sync"

	"github.com/go-kit/kit/log"
	"github.com/go-kit/kit/log/level"
	"github.com/julienschmidt/httprouter"
	"github.com/pquerna/ffjson/ffjson"
)

var (
	// ErrEmptyMetricNotAllowed -
	ErrEmptyMetricNotAllowed = errors.New("empty labels not allowed")
)

// Telecom5Push returns an http.Handler which accepts samples over HTTP and stores them
// in the MetricStore. If replace is true, all metrics for the job and instance
// given by the request are deleted before new ones are stored.
//
// The returned handler is already instrumented for Prometheus.
//
// Customize on the original Push method
func Telecom5Push(next func(http.ResponseWriter, *http.Request, httprouter.Params), logger log.Logger) func(http.ResponseWriter, *http.Request, httprouter.Params) {
	var ps httprouter.Params
	var mtx sync.Mutex // Protects ps.

	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		identify := ps.ByName("identify")
		if identify == "" {
			http.Error(w, "identify is required", http.StatusBadRequest)
			level.Debug(logger).Log("msg", "identify is required")
			return
		}
		mtx.Unlock()

		labels := map[string]string{
			"identify": identify,
		}

		// parse body to metrics and labels
		body, err := ioutil.ReadAll(r.Body)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			level.Error(logger).Log("msg", err.Error())
			return
		}

		rbdPushReqs, err := telecom5BodyToRbdPushReq(body)
		if err != nil {
			http.Error(w, "bad request", http.StatusBadRequest)
			level.Error(logger).Log("msg", "bad request", err.Error())
			return
		}

		// convert metrics and labels to text
		var texts []string
		for _, rbdPushReq := range rbdPushReqs {
			for k, v := range labels {
				rbdPushReq.Labels = append(rbdPushReq.Labels, Label{Key: k, Value: v})
			}
			texts = append(texts, convertRbdPushReq2Text(rbdPushReq))
		}
		text := strings.Join(texts, "")

		// set a new body, which will simulate the same data we read
		r.Body = ioutil.NopCloser(bytes.NewBuffer([]byte(text)))

		r.RequestURI = "/metrics/job/telecom5"

		next(w, r, ps)
	})

	return func(w http.ResponseWriter, r *http.Request, params httprouter.Params) {
		mtx.Lock()
		ps = params
		handler.ServeHTTP(w, r)
	}
}

func telecom5BodyToRbdPushReq(body []byte) ([]*RbdPushReq, error) {
	var m map[string]interface{}
	if err := ffjson.Unmarshal(body, &m); err != nil {
		return nil, fmt.Errorf("convert body to map: %s: %v", body, err)
	}

	var metrics []Metric
	fbmetric := Metric{
		MetricName:  "foobar",
		MetricValue: -1,
	}
	var fblabels []Label
	for key, value := range m {
		if val, ok := value.(float64); ok {
			metric := Metric{
				MetricName:  key,
				MetricValue: val,
			}
			metrics = append(metrics, metric)
			continue
		}

		if val, ok := value.(string); ok {
			label := Label{
				Key:   key,
				Value: val,
			}
			fblabels = append(fblabels, label)
			continue
		}

	}

	var res []*RbdPushReq
	if len(metrics) > 0 {
		rbdPushReq := &RbdPushReq{
			Metrics: metrics,
		}
		res = append(res, rbdPushReq)
	}
	if len(fblabels) > 0 {
		rbdPushReeq := &RbdPushReq{
			Metrics: []Metric{fbmetric},
			Labels:  fblabels,
		}
		res = append(res, rbdPushReeq)
	}

	if len(res) == 0 {
		return nil, ErrEmptyMetricNotAllowed
	}

	return res, nil
}
