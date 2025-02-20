package probes

import (
	"context"
	"fmt"
	"github.com/forkpoons/library/yamlenv"
	"net/http"
	"time"
)

type ProbeConfig struct {
	Port              *yamlenv.Env[int] `yaml:"probe_port"`
	WaitStartUpTime   *yamlenv.Env[int] `yaml:"wait_start_up_time"`
	WaitLivenessTime  *yamlenv.Env[int] `yaml:"wait_liveness_time"`
	WaitReadinessTime *yamlenv.Env[int] `yaml:"wait_readiness_time"`
}

type Probe struct {
	ctx               context.Context
	port              int
	t                 time.Time
	waitStartupTime   time.Duration
	waitLivenessTime  time.Duration
	waitReadinessTime time.Duration
	server            *http.Server
}

func NewProbe(ctx context.Context, probeConfig *ProbeConfig) *Probe {
	p := &Probe{
		ctx:               ctx,
		port:              probeConfig.Port.Value,
		waitStartupTime:   time.Duration(probeConfig.WaitStartUpTime.Value) * time.Second,
		waitLivenessTime:  time.Duration(probeConfig.WaitLivenessTime.Value) * time.Second,
		waitReadinessTime: time.Duration(probeConfig.WaitReadinessTime.Value) * time.Second,
	}

	http.HandleFunc("/startupProbe", p.startupProbe)
	http.HandleFunc("/livenessProbe", p.livenessProbe)
	http.HandleFunc("/readinessProbe", p.readinessProbe)
	srv := http.Server{
		Addr:         fmt.Sprintf(":%d", p.port),
		WriteTimeout: time.Second * 15,
		ReadTimeout:  time.Second * 15,
	}
	p.server = &srv
	return p

}
func (p *Probe) Start() error {
	p.t = time.Now()
	return p.server.ListenAndServe()
}

func (p *Probe) startupProbe(w http.ResponseWriter, r *http.Request) {
	if time.Since(p.t) > p.waitStartupTime {
		w.WriteHeader(200)
	} else {
		w.WriteHeader(503)
	}
}

func (p *Probe) livenessProbe(w http.ResponseWriter, r *http.Request) {
	if time.Since(p.t) > p.waitLivenessTime {
		w.WriteHeader(200)
	} else {
		w.WriteHeader(503)
	}
}

func (p *Probe) readinessProbe(w http.ResponseWriter, r *http.Request) {
	if time.Since(p.t) > p.waitReadinessTime {
		w.WriteHeader(200)
	} else {
		w.WriteHeader(503)
	}
}
