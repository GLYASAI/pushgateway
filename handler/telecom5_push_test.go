package handler

import (
	"testing"
	"strings"
)

func TestBody2Labels(t *testing.T) {
	tests := []struct {
		name, req string
		want      string
		wantErr   error
	}{
		{
			name:    "empty",
			req:     "{}",
			wantErr: ErrEmptyMetricNotAllowed,
		},
		{
			name: "some metric",
			req:  "{\"some_metric\": 3.14}",
			want: "some_metric 3.14\n",
		},
		{
			name: "another metirc",
			want: "some_metric 3.14\nanother_metric 2398.283\n",
			req:  "{\"some_metric\":3.14,\"another_metric\":2398.283}",
		},
		{
			name: "version",
			want: "some_metric 3.14\nanother_metric 2398.283\nfoobar{version=\"1.0.5.1| yyyy.mm.dd-hh24:mi:ss\"} -1\n",
			req:  "{\"some_metric\":3.14,\"another_metric\":2398.283,\"version\": \"1.0.5.1| yyyy.mm.dd-hh24:mi:ss\"}",
		},
	}

	for idx := range tests {
		tc := tests[idx]
		t.Run(tc.name, func(t *testing.T) {
			rbdPushReqs, err := telecom5BodyToRbdPushReq([]byte(tc.req))
			if tc.wantErr != err {
				t.Errorf("want error %v, but got %v", tc.wantErr, err)
				return
			}
			if tc.wantErr != nil {
				return
			}

			var texts []string
			for _, rbdPushReq := range rbdPushReqs {
				texts = append(texts, convertRbdPushReq2Text(rbdPushReq))
			}
			got := strings.Join(texts, "")
			if tc.want != got {
				t.Errorf("want %v, but got %v", tc.want, got)
			}
		})
	}
}
