package tracing

import (
	"fmt"
	"io"
	"sync"

	"github.com/forkpoons/library/yamlenv"
	"github.com/forkpoonsg/library/zerohook"
	"github.com/opentracing/opentracing-go"
	jaeger "github.com/uber/jaeger-client-go"
	traceconfig "github.com/uber/jaeger-client-go/config"
	"github.com/uber/jaeger-lib/metrics/prometheus"
)

type JaegerConfig struct {
	AgentHost   *yamlenv.Env[string] `yaml:"agent_host"`
	AgentPort   *yamlenv.Env[int]    `yaml:"agent_port"`
	ServiceName *yamlenv.Env[string] `yaml:"service_name"`
}

var (
	Tracer opentracing.Tracer
	closer io.Closer
	once   sync.Once
)

func InitTracer(cfg JaegerConfig) error {
	var initErr error
	once.Do(func() {
		jaegerCfg := traceconfig.Configuration{
			ServiceName: cfg.ServiceName.Value,
			Sampler: &traceconfig.SamplerConfig{
				Type:  jaeger.SamplerTypeConst,
				Param: 1, // 1 включает полную выборку
			},
			Reporter: &traceconfig.ReporterConfig{
				LogSpans:           true,
				LocalAgentHostPort: fmt.Sprintf("%s:%d", cfg.AgentHost.Value, cfg.AgentPort.Value),
			},
		}

		metricsFactory := prometheus.New()

		var err error
		Tracer, closer, err = jaegerCfg.NewTracer(
			traceconfig.Logger(jaeger.StdLogger),
			traceconfig.Metrics(metricsFactory),
		)
		if err != nil {
			zerohook.Logger.Error().Err(err).Msg("Не удалось инициализировать трейсер")
			initErr = err
			return
		}
		opentracing.SetGlobalTracer(Tracer)
		zerohook.Logger.Info().Msg("Трейсер успешно инициализирован")
	})

	return initErr
}

// CloseTracer закрывает соединение с JaegerConfig
func CloseTracer() error {
	if closer != nil {
		return closer.Close()
	}
	return nil
}
