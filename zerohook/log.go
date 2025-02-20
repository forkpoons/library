package zerohook

import (
	"fmt"
	"github.com/rs/zerolog/pkgerrors"
	"os"
	"runtime/debug"
	"sync"
	"time"

	"github.com/ItemCloudShopping/library/yamlenv"

	"github.com/rs/zerolog"
)

var once sync.Once
var Logger = zerolog.New(os.Stdout).With().Caller().Timestamp().Logger()

type LoggerConfig struct {
	App             *yamlenv.Env[string] `yaml:"app"`
	Level           *yamlenv.Env[string] `yaml:"level"`
	Facility        *yamlenv.Env[string] `yaml:"facility"`
	CiCommitRefName *yamlenv.Env[string] `yaml:"ci_commit_ref_name"`
	Origin          *yamlenv.Env[string] `yaml:"origin"`
}

func InitLogger(cfg LoggerConfig) {
	once.Do(func() {
		zerolog.ErrorStackMarshaler = pkgerrors.MarshalStack
		zerolog.TimeFieldFormat = time.RFC3339Nano
		zerolog.TimestampFieldName = "log_time"
		zerolog.MessageFieldName = "msg"

		logCtx := Logger.With().
			Str("origin", cfg.Origin.Value).
			Str("facility", cfg.Facility.Value).
			Str("ci_commit_ref_name", cfg.CiCommitRefName.Value)

		if buildInfo, ok := debug.ReadBuildInfo(); ok {
			logCtx = logCtx.Str("go_version", buildInfo.GoVersion)
		}

		level, err := zerolog.ParseLevel(cfg.Level.Value)
		if err != nil {
			level = zerolog.TraceLevel
		}
		Logger = logCtx.Logger().Level(level).Hook(NewHook(cfg))
		if err != nil {
			Logger.Error().Msg("уровень логгирования не определен, установлен уровень trace")
		}
	})
}

type ZeroHook struct {
	cfg LoggerConfig
}

func NewHook(cfg LoggerConfig) *ZeroHook {
	return &ZeroHook{
		cfg: cfg,
	}
}

func (h *ZeroHook) Run(e *zerolog.Event, _ zerolog.Level, _ string) {
	ctx := e.GetCtx()

	if guid := ctx.Value("request-id"); guid != nil {
		e.Str("request_id", guid.(string))
	}

	if guid := ctx.Value("actor-guid"); guid != nil {
		e.Str("actor_guid", guid.(string))
	}

	if guid := ctx.Value("person-guid"); guid != nil {
		e.Str("person_guid", guid.(string))
	}

	if tag := ctx.Value("app-tag"); tag != nil {
		e.Str("app_tag", fmt.Sprintf(`%s-%s`, h.cfg.App.Value, tag))
	} else {
		e.Str("app_tag", h.cfg.App.Value)
	}
}
