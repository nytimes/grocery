package grocery

import (
	"errors"
	"fmt"
	"reflect"
	"strings"
	"time"

	"github.com/go-redis/redis/v8"
)

// UpdateOptions provides options that may be passed to UpdateWithOptions if
// the default behavior of Update needs to be changed.
type UpdateOptions struct {
	// Notify should be set to true if you would like a message to be published
	// to the <struct name>:<id> channel once this update completes.
	Notify bool

	// SetZeroValues should be set to true if you would like to update all zero
	// values in Redis (e.g. empty strings, 0 ints). By default, when creating
	// a struct to pass to Update, you may not set each value, which is why
	// this defaults to false. At this time, setting one zero value at a time
	// is not supported.
	SetZeroValues bool

	// If you would like to run this store/update alongside other Redis
	// updates, you may specify a pipeline.
	Pipeline redis.Pipeliner

	isStore        bool
	storeOverwrite bool
}

// Update updates an object with a given ID. By default, only non-zero values
// are updated in Redis, allowing you to update one property at a time.
//
//	type Item struct {
//	    grocery.Base
//	    Name string
//	    Price float64
//	}
//
//	item := &Item{
//	    Price: 5.64,
//	}
//
//	// Name is not updated, but price is
//	itemID := "asdf"
//	db.Update(itemID, item)
func Update(id string, ptr interface{}) error {
	return updateInternal(id, ptr, &UpdateOptions{})
}

// UpdateWithOptions updates an object in Redis, like Update, but with options.
func UpdateWithOptions(id string, ptr interface{}, opts *UpdateOptions) error {
	return updateInternal(id, ptr, opts)
}

func updateInternal(id string, ptr interface{}, opts *UpdateOptions) error {
	if reflect.TypeOf(ptr).Kind() != reflect.Ptr || reflect.TypeOf(ptr).Elem().Kind() != reflect.Struct {
		return errors.New("ptr must be a struct pointer")
	}

	if id == "" {
		return errors.New("ID must not be empty")
	}

	// Get prefix for the struct (e.g. 'answer:' from Answer)
	prefix := strings.ToLower(reflect.TypeOf(ptr).Elem().Name())

	// Make sure the object exists on an update, or not on a store
	exists, _ := C.Exists(ctx, prefix+":"+id).Result()

	if opts.isStore && exists == 1 && !opts.storeOverwrite {
		return fmt.Errorf("%s:%s already exists", prefix, id)
	} else if !opts.isStore && exists == 0 {
		return fmt.Errorf("%s:%s does not exist", prefix, id)
	}

	val := reflect.ValueOf(ptr).Elem()
	typ := val.Type()
	pip := opts.Pipeline

	if opts.Pipeline == nil {
		pip = C.Pipeline()
	}

	for i := 0; i < typ.NumField(); i++ {
		typeField := typ.Field(i)
		structField := val.Field(i)

		tagName := typeField.Tag.Get("grocery")
		k := tagName

		if tagName == "-" {
			continue
		} else if structField.Kind() == reflect.Struct && typeField.Anonymous {
			// Skip embedded structs
			continue
		} else if strings.Contains(tagName, ",") {
			tagParts := strings.Split(tagName, ",")
			k = tagParts[0]
			tagInfo := tagParts[1]

			if tagInfo == "immutable" {
				continue
			}
		} else if tagName == "" {
			// Tag was not specified, assume field name
			if len(typeField.Name) > 1 {
				k = strings.ToLower(string(typeField.Name[0])) + string(typeField.Name[1:])
			} else {
				k = strings.ToLower(typeField.Name)
			}
		}

		if !opts.SetZeroValues && structField.IsZero() {
			continue
		}

		switch typeField.Type.Kind() {
		case reflect.Ptr:
			if structField.IsNil() {
				continue
			}

			if typeField.Type == mapType || structField.Elem().FieldByName("CustomMapType").IsValid() {
				pip.Del(ctx, prefix+":"+id+":"+k)

				structField.MethodByName("Range").Call([]reflect.Value{
					reflect.ValueOf(func(key, value interface{}) bool {
						pip.HSet(ctx, prefix+":"+id+":"+k, key, value)
						return true
					}),
				})
			} else if typeField.Type == setType || structField.Elem().FieldByName("CustomSetType").IsValid() {
				pip.Del(ctx, prefix+":"+id+":"+k)

				structField.MethodByName("Range").Call([]reflect.Value{
					reflect.ValueOf(func(key, value interface{}) bool {
						pip.SAdd(ctx, prefix+":"+id+":"+k, key)
						return true
					}),
				})
			} else if !structField.Elem().FieldByName("Base").IsZero() {
				val := structField.Elem().FieldByName("Base").FieldByName("ID").String()
				pip.HSet(ctx, prefix+":"+id, k, val)
			} else {
				return fmt.Errorf("can't set unknown field '%s'", tagName)
			}
		case reflect.Slice:
			for i := 0; i < structField.Len(); i++ {
				if !structField.Index(i).Elem().FieldByName("Base").IsZero() {
					itemID := structField.Index(i).Elem().FieldByName("Base").FieldByName("ID").String()
					pip.RPush(ctx, prefix+":"+id+":"+k, itemID)
				} else {
					return fmt.Errorf("can't set unknown array item in %s", tagName)
				}
			}
		case reflect.Map:
			return fmt.Errorf("type of field '%s' must be changed to *grocery.Map", tagName)
		case reflect.Struct:
			switch structField.Type() {
			case reflect.TypeOf(time.Now()):
				timeVal := structField.MethodByName("Unix").Call([]reflect.Value{})[0].Int()
				pip.HSet(ctx, prefix+":"+id, k, timeVal)
			default:
				return fmt.Errorf("can't set unknown struct for field '%s'", tagName)
			}
		case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64, reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
			// Handle int alias types
			val := structField.Convert(reflect.TypeOf(0)).Int()
			pip.HSet(ctx, prefix+":"+id, k, val)
		case reflect.String:
			// Handle string alias types
			val := structField.Convert(reflect.TypeOf("")).String()
			pip.HSet(ctx, prefix+":"+id, k, val)
		case reflect.Bool:
			if loadFunc := structField.MethodByName("Load"); loadFunc.IsValid() {
				// Skip custom boolean values; they don't get stored
				continue
			} else {
				pip.HSet(ctx, prefix+":"+id, k, structField.Bool())
			}
		case reflect.Float64, reflect.Float32:
			val := structField.Float()
			pip.HSet(ctx, prefix+":"+id, k, val)
		case reflect.Interface:
			if structField.Type().Name() == "ModelHook" {
				// Skip ModelHook fields
				continue
			}

			return fmt.Errorf("don't know how to set interface field '%s'", k)
		default:
			return fmt.Errorf("don't know how to set field '%s'", k)
		}
	}

	// Set updatedAt timestamp
	pip.HSet(ctx, prefix+":"+id, "updatedAt", time.Now().Unix())

	if opts.isStore {
		// Set ID
		fi := reflect.Indirect(val).FieldByName("ID")

		if fi.String() != id {
			fi.SetString(id)
		}

		// Set createdAt timestamp
		pip.HSet(ctx, prefix+":"+id, "createdAt", time.Now().Unix())

		// Call hook after calling store, if the object has one
		if hook, ok := ptr.(ModelHook); ok {
			hook.PostStore(pip)
		}
	}

	if opts.Notify {
		// Publish message if notify is enabled
		pip.Publish(ctx, prefix+":"+id, "")
	}

	// Don't exec if a pipeline was provided to us
	if opts.Pipeline == nil {
		if _, err := pip.Exec(ctx); err != nil {
			return err
		}

		if err := Load(id, ptr); err != nil {
			return err
		}
	}

	return nil
}
