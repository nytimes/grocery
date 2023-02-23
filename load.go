package grocery

import (
	"errors"
	"reflect"
	"strings"

	"github.com/redis/go-redis/v9"
)

// Load automates the process of loading data from Redis, binding it to a
// struct, and then setting base attributes (such as ID). See Store for more
// information on creating structs for grocery. Load can be used like so:
//
//	type Item struct {
//	    grocery.Base
//	    Name string `grocery:"name"`
//	}
//
//	itemID := "asdf"
//	item := new(Item)
//	db.Load(itemID, item)
func Load(id string, ptr interface{}) error {
	if reflect.TypeOf(ptr).Kind() != reflect.Ptr || reflect.TypeOf(ptr).Elem().Kind() != reflect.Struct {
		return errors.New("ptr must be a struct pointer")
	}

	// Get prefix for the struct (e.g. 'item:' from Item)
	prefix := strings.ToLower(reflect.TypeOf(ptr).Elem().Name())

	// Load object data
	res, _ := C.HGetAll(ctx, prefix+":"+id).Result()

	if err := bind(prefix, id, res, ptr); err != nil {
		return err
	}

	// Set the ID before returning
	val := reflect.ValueOf(ptr)
	fi := reflect.Indirect(val).FieldByName("ID")

	if fi.String() != id {
		fi.SetString(id)
	}

	// Call post-load hook
	postLoad := reflect.ValueOf(ptr).MethodByName("PostLoad")

	if postLoad.IsValid() {
		postLoad.Call([]reflect.Value{})
	}

	return nil
}

// LoadAll automates the process of running multiple Load calls through a
// pipeline. If you're fetching multiple objects from Redis, you should
// generally use LoadAll instead of calling Load multiple times. Read more
// about pipelining at https://redis.io/topics/pipelining.
func LoadAll[T any](ids []string, values *[]T) error {
	if len(ids) != len(*values) {
		return errors.New("len(ids) must equal len(*values)")
	} else if len(ids) == 0 {
		return errors.New("len(ids) must be greater than zero")
	}

	// Get prefix for the struct (e.g. 'item:' from Item)
	prefix := strings.ToLower(reflect.ValueOf(values).Elem().Index(0).Type().Name())

	// Pipeline all HGetAll commands
	pip := C.Pipeline()
	cmds := make([]*redis.MapStringStringCmd, len(ids))

	for i, id := range ids {
		cmds[i] = pip.HGetAll(ctx, prefix+":"+id)
	}

	pip.Exec(ctx)

	for i, cmd := range cmds {
		res, _ := cmd.Result()
		itemPtr := &((*values)[i])

		if err := bind(prefix, ids[i], res, itemPtr); err != nil {
			return err
		}

		// Set ID
		val := reflect.ValueOf(itemPtr).Elem()
		fi := reflect.Indirect(val).FieldByName("ID")

		if fi.String() != ids[i] {
			fi.SetString(ids[i])
		}

		// Call post-load hook
		postLoad := reflect.ValueOf(itemPtr).MethodByName("PostLoad")

		if postLoad.IsValid() {
			postLoad.Call([]reflect.Value{})
		}
	}

	return nil
}
