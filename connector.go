package grocery

import (
	"context"
	"sync"
	"time"

	"github.com/go-redis/redis/v8"
)

const (
	// Redis channel to send a test message on during initialization.
	firstMessageChannel = "grocery_hello_world"
)

var (
	// The underlying Redis client powering grocery. Use this field to run your
	// own Redis commands.
	C *redis.Client

	// Context for all Redis queries, currently unused.
	ctx = context.Background()

	// Callback functions that listen for events published to Redis.
	handlers = make(map[string][]func(string, []byte))

	// Handler synchronization.
	handlersMux sync.RWMutex

	// Persistent pubsub connection that waits for published events.
	psc *redis.PubSub
)

// Init initializes the Redis client and additionally starts a pub/sub client.
func Init(config *redis.Options) error {
	C = redis.NewClient(config)

	if _, err := C.Ping(ctx).Result(); err != nil {
		return err
	}

	// Wait until PSubscribe receives its first message to return
	firstMessageSignal := make(chan bool)
	go listenForUpdates(firstMessageSignal)

	ticker := time.NewTicker(time.Millisecond * 10)
	defer ticker.Stop()

	for {
		select {
		case <-firstMessageSignal:
			// Wait until we can confirm listenForUpdates is working before returning
			close(firstMessageSignal)
			return nil
		case <-ticker.C:
			// Repeatedly send messages while we wait for listenForUpdates to
			// start listening
			C.Publish(ctx, firstMessageChannel, "")
		}
	}
}

func listenForUpdates(firstMessageSignal chan bool) {
	receivedFirstMessage := false
	psc = C.PSubscribe(ctx, "*")
	ch := psc.Channel()

	for msg := range ch {
		if !receivedFirstMessage && msg.Channel == firstMessageChannel {
			// Send signal on first message to confirm PSubscribe is ready
			firstMessageSignal <- true
			receivedFirstMessage = true
			continue
		}

		handlersMux.RLock()
		handlers, ok := handlers[msg.Channel]
		handlersMux.RUnlock()

		if !ok {
			// Received message for a channel that nobody is subscribed to
			continue
		}

		for _, handler := range handlers {
			handler(msg.Channel, []byte(msg.Payload))
		}
	}
}

// Subscribe adds a new listener function to a channel in our pub/sub
// connection. For example, if you want to listen to events on the 'reset'
// channel, and then publish a test event, you might do the following:
//
//	db.Subscribe([]string{"reset"}, func(channel string, payload []byte) {
//	    fmt.Println("receiving data from " + channel)
//	})
//
//	db.C.Publish("reset", "payload")
func Subscribe(channels []string, handler func(string, []byte)) {
	handlersMux.Lock()
	defer handlersMux.Unlock()

	for _, channel := range channels {
		if _, ok := handlers[channel]; !ok {
			handlers[channel] = []func(string, []byte){}
		}

		handlers[channel] = append(handlers[channel], handler)
	}
}

// Unsubscribe removes all listeners waiting on any channel in channels.
func Unsubscribe(channels []string) {
	handlersMux.Lock()
	defer handlersMux.Unlock()

	for _, channel := range channels {
		delete(handlers, channel)
	}
}
