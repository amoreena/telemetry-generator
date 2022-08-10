package topology

import (
	"fmt"
	"github.com/lightstep/lightstep-partner-sdk/collector/generatorreceiver/internal/flags"
	"math/rand"
	"time"
)

type ServiceRoute struct {
	Route                 string                 `json:"route" yaml:"route"`
	DownstreamCalls       map[string]string      `json:"downstreamCalls,omitempty" yaml:"downstreamCalls,omitempty"`
	MaxLatencyMillis      int64                  `json:"maxLatencyMillis" yaml:"maxLatencyMillis"`
	LatencyPercentiles    *LatencyPercentiles    `json:"latencyPercentiles" yaml:"latencyPercentiles"`
	TagSets               []TagSet               `json:"tagSets" yaml:"tagSets"`
	ResourceAttributeSets []ResourceAttributeSet `json:"resourceAttrSets" yaml:"resourceAttrSets"`
	flags.EmbeddedFlags   `json:",inline" yaml:",inline"`
	// TODO: rename all references from `tag` to `attribute`, to follow the otel standard.

}
type LatencyPercentiles struct {
	P0Cfg     string `json:"p0" yaml:"p0"`
	P50Cfg    string `json:"p50" yaml:"p50"`
	P95Cfg    string `json:"p95" yaml:"p95"`
	P99Cfg    string `json:"p99" yaml:"p99"`
	P999Cfg   string `json:"p99.9" yaml:"p99.9"`
	P100Cfg   string `json:"p100" yaml:"p100"`
	durations struct {
		p0   time.Duration
		p50  time.Duration
		p95  time.Duration
		p99  time.Duration
		p999 time.Duration
		p100 time.Duration
	}
}

func (r *ServiceRoute) validate(t Topology) error {
	// TODO: is there better way of passing topology along
	if r.FlagSet != "" && flags.Manager.GetFlag(r.FlagSet) == nil {
		return fmt.Errorf("flag %v does not exist", r.FlagSet)
	}
	if r.FlagUnset != "" && flags.Manager.GetFlag(r.FlagUnset) == nil {
		return fmt.Errorf("flag %v does not exist", r.FlagUnset)
	}

	for service, route := range r.DownstreamCalls {
		st := t.GetServiceTier(service)
		if st == nil {
			return fmt.Errorf("downstream service %s does not exist", service)
		}
		if st.GetRoute(route) == nil {
			return fmt.Errorf("downstream service %s does not have route %s defined", service, route)
		}
		if r.MaxLatencyMillis <= 0 {
			return fmt.Errorf("must have a positive, non-zero maxLatencyMillis defined")
		}
	}
	return nil
}

func (l *LatencyPercentiles) Sample() float64 {
	uniform := func(timeA, timeB time.Duration) float64 {
		min := float64(timeA.Microseconds())
		max := float64(timeB.Microseconds())
		return (min + (max-min)*rand.Float64()) * 1000
	}
	genNumber := rand.Float64()
	switch {
	case genNumber <= 0.001:
		// 0.1% of requests
		return uniform(l.durations.p99, l.durations.p999)
	case genNumber <= 0.01:
		// 1% of requests
		return uniform(l.durations.p95, l.durations.p99)
	case genNumber <= 0.05:
		// 5% of requests
		return uniform(l.durations.p50, l.durations.p95)
	case genNumber <= 0.5:
		// 50% of requests
		return uniform(l.durations.p0, l.durations.p50)
	default:
		return uniform(l.durations.p0, l.durations.p50)
		// not sure if --> is better, seems to skew it too high generally, return uniform(percentiles.p50, percentiles.p100)
	}
	/*
		TODO: the above is still not perfect - it is a bit off on the p50m the logic for default is prob wrong, should be reorderd like below -
		Trying the below also makes it off on p50 by more (I think because its getting more skewed from the p95), so its maybe not exactly right either
		case genNumber <= 0.5:
			return uniform(percentiles.p0, percentiles.p50)
		case genNumber <= 0.95:
			return uniform(percentiles.p50, percentiles.p95)
		case genNumber <= 0.99:
			return uniform(percentiles.p95, percentiles.p99)
		default:
			return uniform(percentiles.p99, percentiles.p999)
	*/
}

func (l *LatencyPercentiles) loadDurations() error {
	// TODO/future things:
	// 		normalize function for config parsing
	// 		maybe enforce either MaxLatencyMillis or LatencyPercentiles but not both?
	//			either way which overrides which? for now LatencyPercentiles will override MaxLatencyMillis
	var err error
	l.durations.p0, err = time.ParseDuration(l.P0Cfg)
	if err != nil {
		return err
	}
	l.durations.p50, err = time.ParseDuration(l.P50Cfg)
	if err != nil {
		return err
	}
	l.durations.p95, err = time.ParseDuration(l.P95Cfg)
	if err != nil {
		return err
	}
	l.durations.p99, err = time.ParseDuration(l.P99Cfg)
	if err != nil {
		return err
	}
	l.durations.p999, err = time.ParseDuration(l.P999Cfg)
	if err != nil {
		return err
	}
	l.durations.p100, err = time.ParseDuration(l.P100Cfg)
	if err != nil {
		return err
	}
	return nil
}
