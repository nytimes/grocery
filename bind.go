package grocery

import (
	"errors"
	"reflect"
	"strconv"
	"strings"
	"time"
)

var mapType = reflect.TypeOf(&Map{})
var setType = reflect.TypeOf(&Set{})

// bind is used to load keys and values from data into ptr. Generally, data is
// the result of C.HGetAll(prefix + ":" + id), and ptr is a pointer to a
// struct that has been set up for usage with grocery.
func bind(prefix, id string, data map[string]string, ptr interface{}) error {
	if ptr == nil {
		return errors.New("ptr must not be nil")
	} else if reflect.TypeOf(ptr).Kind() != reflect.Ptr || reflect.TypeOf(ptr).Elem().Kind() != reflect.Struct {
		return errors.New("ptr must be a struct pointer")
	} else if len(data) == 0 {
		return errors.New("data must not be empty")
	}

	typ := reflect.TypeOf(ptr).Elem()
	val := reflect.ValueOf(ptr).Elem()
	return bindStruct(prefix, id, data, typ, val)
}

func bindStruct(prefix, id string, data map[string]string, typ reflect.Type, val reflect.Value) error {
	for i := 0; i < typ.NumField(); i++ {
		typeField := typ.Field(i)
		structField := val.Field(i)

		// Skip unexported fields
		if !structField.CanSet() {
			continue
		}

		inputFieldName := typeField.Tag.Get("grocery")

		if inputFieldName == "-" {
			// Skip values that shouldn't be stored
			continue
		} else if structField.Kind() == reflect.Struct && typeField.Anonymous {
			// Recurse on embedded structs
			bindStruct(prefix, id, data, typeField.Type, structField)
			continue
		} else if strings.Contains(inputFieldName, ",") {
			inputFieldName = strings.Split(inputFieldName, ",")[0]
		} else if inputFieldName == "" {
			// Tag was not specified, assume field name
			if len(typeField.Name) > 1 {
				inputFieldName = strings.ToLower(string(typeField.Name[0])) + string(typeField.Name[1:])
			} else {
				inputFieldName = strings.ToLower(typeField.Name)
			}
		}

		switch typeField.Type.Kind() {
		case reflect.Ptr:
			// New item to set in struct
			res := reflect.New(typeField.Type.Elem())

			if typeField.Type == mapType || (typeField.Type.Elem().Kind() == reflect.Struct && res.Elem().FieldByName("CustomMapType").IsValid()) {
				// This is a map
				m, err := C.HGetAll(ctx, prefix+":"+id+":"+inputFieldName).Result()

				if err != nil {
					return err
				}

				for k := range m {
					res.MethodByName("Store").Call([]reflect.Value{
						reflect.ValueOf(k),
						reflect.ValueOf(m[k]),
					})
				}

				if setupMethod := res.MethodByName("Setup"); setupMethod.IsValid() {
					setupMethod.Call([]reflect.Value{})
				}

				structField.Set(res)
			} else if typeField.Type == setType || (typeField.Type.Elem().Kind() == reflect.Struct && res.Elem().FieldByName("CustomSetType").IsValid()) {
				// This is a set
				s, err := C.SMembers(ctx, prefix+":"+id+":"+inputFieldName).Result()

				if err != nil {
					return err
				}

				for _, val := range s {
					res.MethodByName("Add").Call([]reflect.Value{
						reflect.ValueOf(val),
					})
				}

				if setupMethod := res.MethodByName("Setup"); setupMethod.IsValid() {
					setupMethod.Call([]reflect.Value{})
				}

				structField.Set(res)
			} else if typeField.Type.Elem().Kind() == reflect.Struct {
				if _, ok := typeField.Type.Elem().FieldByName("Base"); ok {
					// This is a reference to a model that we should load from Redis
					id := data[inputFieldName]

					// Load data from redis
					subPrefix := strings.ToLower(res.Type().Elem().Name())
					dat, err := C.HGetAll(ctx, subPrefix+":"+id).Result()

					if err != nil {
						return err
					} else if len(dat) == 0 {
						continue
					}

					err = bindStruct(subPrefix, id, dat, res.Type().Elem(), res.Elem())

					if err != nil {
						return err
					}

					// Set ID
					val := res.Elem()
					fi := reflect.Indirect(val).FieldByName("ID")
					fi.SetString(id)

					structField.Set(res)
				} else {
					return errors.New("Can't set unsupported struct with key " + inputFieldName)
				}
			} else {
				inputValue, exists := data[inputFieldName]

				if !exists {
					continue
				}

				if err := setFieldWithKind(typeField.Type.Kind(), inputValue, structField); err != nil {
					return err
				}
			}
		case reflect.Slice:
			ids, err := C.LRange(ctx, prefix+":"+id+":"+inputFieldName, 0, -1).Result()

			if err != nil {
				return err
			}

			arr := reflect.MakeSlice(typeField.Type, 0, len(ids))

			for _, itemID := range ids {
				if typeField.Type.Elem().Kind() == reflect.String {
					// This is just a string slice
					arr = reflect.Append(arr, reflect.ValueOf(itemID))
				} else if _, ok := typeField.Type.Elem().Elem().FieldByName("Base"); ok {
					// This is a reference to a model that we should load from Redis
					ptr := reflect.New(typeField.Type.Elem().Elem())

					// Load data from redis
					subPrefix := strings.ToLower(ptr.Type().Elem().Name())
					dat, err := C.HGetAll(ctx, subPrefix+":"+itemID).Result()

					if err != nil {
						return err
					}

					err = bindStruct(subPrefix, itemID, dat, ptr.Type().Elem(), ptr.Elem())

					if err != nil {
						return err
					}

					// Set ID
					val := ptr.Elem()
					fi := reflect.Indirect(val).FieldByName("ID")
					fi.SetString(itemID)

					arr = reflect.Append(arr, ptr)
				} else {
					return errors.New("Can't set unsupported struct")
				}
			}

			structField.Set(arr)
		case reflect.Bool:
			if loadFunc := structField.MethodByName("Load"); loadFunc.IsValid() {
				// Load custom boolean values
				ret := loadFunc.Call([]reflect.Value{
					reflect.ValueOf(id),
					reflect.ValueOf(inputFieldName),
				})

				val := strconv.FormatBool(ret[0].Bool())

				if !ret[1].IsNil() {
					return ret[1].Interface().(error)
				}

				if err := setFieldWithKind(structField.Kind(), val, structField); err != nil {
					return err
				}
			} else {
				inputValue, exists := data[inputFieldName]

				if !exists {
					continue
				}

				if err := setFieldWithKind(typeField.Type.Kind(), inputValue, structField); err != nil {
					return err
				}
			}
		default:
			inputValue, exists := data[inputFieldName]

			if !exists {
				continue
			}

			if err := setFieldWithKind(typeField.Type.Kind(), inputValue, structField); err != nil {
				return err
			}
		}
	}

	return nil
}

func setFieldWithKind(valueKind reflect.Kind, val string, structField reflect.Value) error {
	switch valueKind {
	case reflect.Ptr:
		return setFieldWithKind(structField.Elem().Kind(), val, structField.Elem())
	case reflect.Int:
		return setIntField(val, 0, structField)
	case reflect.Int8:
		return setIntField(val, 8, structField)
	case reflect.Int16:
		return setIntField(val, 16, structField)
	case reflect.Int32:
		return setIntField(val, 32, structField)
	case reflect.Int64:
		return setIntField(val, 64, structField)
	case reflect.Uint:
		return setUintField(val, 0, structField)
	case reflect.Uint8:
		return setUintField(val, 8, structField)
	case reflect.Uint16:
		return setUintField(val, 16, structField)
	case reflect.Uint32:
		return setUintField(val, 32, structField)
	case reflect.Uint64:
		return setUintField(val, 64, structField)
	case reflect.Bool:
		return setBoolField(val, structField)
	case reflect.Float32:
		return setFloatField(val, 32, structField)
	case reflect.Float64:
		return setFloatField(val, 64, structField)
	case reflect.String:
		structField.SetString(val)
	case reflect.Struct:
		switch structField.Type() {
		case reflect.TypeOf(time.Now()):
			timeInt, _ := strconv.Atoi(val)
			timeVal := time.Unix(int64(timeInt), 0)
			structField.Set(reflect.ValueOf(timeVal))
		default:
			return errors.New("unknown type")
		}
	default:
		return errors.New("unknown type")
	}

	return nil
}

func setIntField(value string, bitSize int, field reflect.Value) error {
	if value == "" {
		value = "0"
	}

	val, err := strconv.ParseInt(value, 10, bitSize)

	if err != nil {
		return err
	}

	field.SetInt(val)
	return nil
}

func setUintField(value string, bitSize int, field reflect.Value) error {
	if value == "" {
		value = "0"
	}

	val, err := strconv.ParseUint(value, 10, bitSize)

	if err != nil {
		return err
	}

	field.SetUint(val)
	return nil
}

func setBoolField(value string, field reflect.Value) error {
	if value == "" {
		value = "false"
	}

	val, err := strconv.ParseBool(value)

	if err != nil {
		return err
	}

	field.SetBool(val)
	return nil
}

func setFloatField(value string, bitSize int, field reflect.Value) error {
	if value == "" {
		value = "0.0"
	}

	val, err := strconv.ParseFloat(value, bitSize)

	if err != nil {
		return err
	}

	field.SetFloat(val)
	return nil
}
