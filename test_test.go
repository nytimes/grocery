package grocery

import (
	"context"
	"net"
	"os"
	"testing"

	"github.com/redis/go-redis/v9"
)

var keys = map[string]bool{}

func TestMain(m *testing.M) {
	// Create grocery client
	Init(&redis.Options{
		Addr: "localhost:6379",
	})

	// Add hook for saving keys
	C.AddHook(testHook{})

	// Run unit test
	code := m.Run()

	// Delete all keys saved to Redis while running unit tests
	pip := C.Pipeline()

	for id := range keys {
		pip.Del(ctx, id)
	}

	pip.Exec(ctx)

	// Exit from test
	os.Exit(code)
}

// Hooks used to save keys stored during tests so we can delete them once tests complete
type testHook struct{}

func (testHook) DialHook(next redis.DialHook) redis.DialHook {
	return func(ctx context.Context, network, addr string) (net.Conn, error) {
		return next(ctx, network, addr)
	}
}

func (testHook) ProcessHook(next redis.ProcessHook) redis.ProcessHook {
	return func(ctx context.Context, cmd redis.Cmder) error {
		if err := next(ctx, cmd); err != nil {
			return err
		}

		keys[cmd.Args()[1].(string)] = true
		return nil
	}
}

func (testHook) ProcessPipelineHook(next redis.ProcessPipelineHook) redis.ProcessPipelineHook {
	return func(ctx context.Context, cmds []redis.Cmder) error {
		if err := next(ctx, cmds); err != nil {
			return err
		}

		for _, cmd := range cmds {
			keys[cmd.Args()[1].(string)] = true
		}

		return nil
	}
}
