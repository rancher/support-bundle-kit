package utils

import (
	"context"
	"fmt"

	"github.com/prometheus/client_golang/api"
	"github.com/prometheus/client_golang/api/prometheus/v1"
)

const PrometheusPort = 9090

type Prometheus struct {
	api v1.API
}

func NewPrometheus(host string) (*Prometheus, error) {
	client, err := api.NewClient(api.Config{
		Address: fmt.Sprintf("http://%s:%d", host, PrometheusPort),
	})

	if err != nil {
		return nil, err
	}

	return &Prometheus{api: v1.NewAPI(client)}, nil
}

func (p *Prometheus) GetAlerts(ctx context.Context) ([]v1.Alert, error) {
	result, err := p.api.Alerts(ctx)

	if err != nil {
		return []v1.Alert{}, err
	}

	return result.Alerts, nil
}
