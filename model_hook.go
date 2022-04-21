package grocery

import "github.com/go-redis/redis/v8"

// ModelHook provides hooks that structs may implement if you would like to
// receive callbacks on certain events.
type ModelHook interface {
	// PostStore is called after all necessary commands to store the object
	// have been passed to a pipeline. The pipeline is provided here in case
	// you would like to pass any additional commands to Redis, to be executed
	// immediately after the store completes.
	PostStore(pip redis.Pipeliner)
}
