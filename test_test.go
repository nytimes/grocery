package grocery

import (
	"context"
	"os"
	"testing"

	"github.com/go-redis/redis/v8"
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

func (testHook) BeforeProcess(ctx context.Context, cmd redis.Cmder) (context.Context, error) {
	return ctx, nil
}

func (testHook) AfterProcess(ctx context.Context, cmd redis.Cmder) error {
	keys[cmd.Args()[1].(string)] = true
	return nil
}

func (testHook) BeforeProcessPipeline(ctx context.Context, cmds []redis.Cmder) (context.Context, error) {
	return ctx, nil
}

func (testHook) AfterProcessPipeline(ctx context.Context, cmds []redis.Cmder) error {
	for _, cmd := range cmds {
		keys[cmd.Args()[1].(string)] = true
	}

	return nil
}
