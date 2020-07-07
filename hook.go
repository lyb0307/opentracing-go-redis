package otgoredis

import (
	"context"
	"github.com/go-redis/redis/v8"
	"github.com/opentracing/opentracing-go"
	"strings"
)

type OpenTracingHook struct {
	Tracker opentracing.Tracer
}

var _ redis.Hook = &OpenTracingHook{}

func NewHook() *OpenTracingHook {
	return &OpenTracingHook{
		opentracing.GlobalTracer(),
	}
}

func NewHookWithTracer(tracker opentracing.Tracer) *OpenTracingHook {
	return &OpenTracingHook{
		tracker,
	}
}

func (h *OpenTracingHook) BeforeProcess(ctx context.Context, cmd redis.Cmder) (context.Context, error) {

	b := make([]byte, 32)
	b = appendCmd(b, cmd)

	span, newContext := opentracing.StartSpanFromContextWithTracer(ctx, h.Tracker, cmd.FullName())
	span.SetTag("db.system", "redis")
	span.SetTag("redis.cmd", String(b))

	return newContext, nil
}

func (h *OpenTracingHook) AfterProcess(ctx context.Context, cmd redis.Cmder) error {
	opentracing.SpanFromContext(ctx).Finish()
	return nil
}

func (h *OpenTracingHook) BeforeProcessPipeline(ctx context.Context, cmds []redis.Cmder) (context.Context, error) {

	const numCmdLimit = 100
	const numNameLimit = 10

	seen := make(map[string]struct{}, len(cmds))
	unqNames := make([]string, 0, len(cmds))

	b := make([]byte, 0, 32*len(cmds))

	for i, cmd := range cmds {
		if i > numCmdLimit {
			break
		}

		if i > 0 {
			b = append(b, '\n')
		}
		b = appendCmd(b, cmd)

		if len(unqNames) >= numNameLimit {
			continue
		}

		name := cmd.FullName()
		if _, ok := seen[name]; !ok {
			seen[name] = struct{}{}
			unqNames = append(unqNames, name)
		}
	}

	span, newContext := opentracing.StartSpanFromContextWithTracer(ctx, h.Tracker, "pipeline "+strings.Join(unqNames, " "))
	span.SetTag("db.system", "redis")
	span.SetTag("redis.num_cmd", len(cmds))
	span.SetTag("redis.cmds", String(b))

	return newContext, nil
}

func (h *OpenTracingHook) AfterProcessPipeline(ctx context.Context, cmds []redis.Cmder) error {
	opentracing.SpanFromContext(ctx).Finish()
	return nil
}

func appendCmd(b []byte, cmd redis.Cmder) []byte {
	const lenLimit = 64

	for i, arg := range cmd.Args() {
		if i > 0 {
			b = append(b, ' ')
		}

		start := len(b)
		b = AppendArg(b, arg)
		if len(b)-start > lenLimit {
			b = append(b[:start+lenLimit], "..."...)
		}
	}

	if err := cmd.Err(); err != nil {
		b = append(b, ": "...)
		b = append(b, err.Error()...)
	}

	return b
}
