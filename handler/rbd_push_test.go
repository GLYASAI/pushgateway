package handler

import (
	"testing"
)

func TestConvertRbdPushReq2Text(t *testing.T) {
	tests := []struct {
		name, want string
		req        *RbdPushReq
	}{
		{
			name: "empty metrics",
			want: "",
			req: &RbdPushReq{
				Metrics: []Metric{},
			},
		},
		{
			name: "single metirc",
			want: "some_metric 3.14\n",
			req: &RbdPushReq{
				Metrics: []Metric{
					{
						MetricName:  "some_metric",
						MetricValue: 3.14,
					},
				},
			},
		},
		{
			name: "another metirc",
			want: "some_metric 3.14\nanother_metric 2398.283\n",
			req: &RbdPushReq{
				Metrics: []Metric{
					{
						MetricName:  "some_metric",
						MetricValue: 3.14,
					},
					{
						MetricName:  "another_metric",
						MetricValue: 2398.283,
					},
				},
			},
		},
		{
			name: "another metirc with labels",
			want: "some_metric{tenant_id=\"my_tenant_id\",tenant_id=\"my_service_id\"} 3.14\nanother_metric{tenant_id=\"my_tenant_id\",tenant_id=\"my_service_id\"} 2398.283\n",
			req: &RbdPushReq{
				Metrics: []Metric{
					{
						MetricName:  "some_metric",
						MetricValue: 3.14,
					},
					{
						MetricName:  "another_metric",
						MetricValue: 2398.283,
					},
				},
				Labels: []Label{
					{
						Key:   "tenant_id",
						Value: "my_tenant_id",
					},
					{
						Key:   "tenant_id",
						Value: "my_service_id",
					},
				},
			},
		},
		{
			name: "another metirc with type",
			want: "# TYPE some_metric gauge\nsome_metric 3.14\n# TYPE another_metric counter\nanother_metric 2398.283\n",
			req: &RbdPushReq{
				Metrics: []Metric{
					{
						MetricName:  "some_metric",
						MetricValue: 3.14,
						MetricType:  "gauge",
					},
					{
						MetricName:  "another_metric",
						MetricValue: 2398.283,
						MetricType:  "counter",
					},
				},
			},
		},
		{
			name: "another metirc with labels and type",
			want: "# TYPE some_metric gauge\nsome_metric{tenant_id=\"my_tenant_id\",tenant_id=\"my_service_id\"} 3.14\n# TYPE another_metric counter\nanother_metric{tenant_id=\"my_tenant_id\",tenant_id=\"my_service_id\"} 2398.283\n",
			req: &RbdPushReq{
				Metrics: []Metric{
					{
						MetricName:  "some_metric",
						MetricValue: 3.14,
						MetricType:  "gauge",
					},
					{
						MetricName:  "another_metric",
						MetricValue: 2398.283,
						MetricType:  "counter",
					},
				},
				Labels: []Label{
					{
						Key:   "tenant_id",
						Value: "my_tenant_id",
					},
					{
						Key:   "tenant_id",
						Value: "my_service_id",
					},
				},
			},
		},
	}

	for idx := range tests {
		tc := tests[idx]
		t.Run(tc.name, func(t *testing.T) {
			want, got := tc.want, convertRbdPushReq2Text(tc.req)
			if want != got {
				t.Errorf("want `%s`, but got `%s`", want, got)
			}
		})
	}
}

func TestConvertRbdPushReq2TextWithoutTableDrivenTests(t *testing.T) {
	emptyMetrics := &RbdPushReq{
		Metrics: []Metric{},
	}
	if res := convertRbdPushReq2Text(emptyMetrics); res != "" {
		t.Errorf("empty metrics; want \"\", but got %s", res)
	}

	singleMetric := &RbdPushReq{
		Metrics: []Metric{
			{
				MetricName:  "some_metric",
				MetricValue: 3.14,
			},
		},
	}
	if res := convertRbdPushReq2Text(singleMetric); res != "some_metric 3.14\n" {
		t.Errorf("single metrics; want \"some_metric 3.14\n\", but got %s", res)
	}

	anotherMetrics := &RbdPushReq{
		Metrics: []Metric{
			{
				MetricName:  "some_metric",
				MetricValue: 3.14,
			},
			{
				MetricName:  "another_metric",
				MetricValue: 2398.283,
			},
		},
	}
	if res := convertRbdPushReq2Text(anotherMetrics); res != "some_metric 3.14\nanother_metric 2398.283\n" {
		t.Errorf("another metrics; want \"some_metric 3.14\nanother_metric 2398.283\n\", but got %s", res)
	}
}
