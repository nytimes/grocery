package grocery

import (
	"encoding/json"
	"sync"
)

// Map is used to represent maps stored in Redis. See CustomMapType and Store
// for more information.
type Map struct {
	sync.Map
}

// CustomMapType must be embedded by any struct that is used as a model's field
// and should be represented as a map in Redis. The parent struct must also
// embed grocery.Map. As a motivating example, let's imagine we want to store
// projects in Redis, and each project should contain a map that stores user
// IDs along with each user's role for that project.
//
//  type User struct {
//      grocery.Base
//      Name string `grocery:"name"`
//
//      // Role is not stored in Redis because it is specific to each project
//      // the user is part of.
//      Role string `grocery:"-"`
//  }
//
//  type Project struct {
//      grocery.Base
//      Users *UserRolesMap `grocery:"users"`
//  }
//
//  type UserRolesMap struct {
//      grocery.CustomMapType
//      grocery.Map
//      users map[string]*User
//  }
//
//  func (m *UserRolesMap) Setup() {
//      // Setup gets called once the map's contents have already been loaded
//      // from Redis. In this function, we have the opportunity to populate the
//      // private m.users map with our own User structs, instead of using the
//      // underlying grocery.Map (which is a sync.Map internally) to just
//      // access user IDs.
//
//      userIDs := make([]string, 0, m.Count())
//
//      m.Range(func(key, value interface{}) bool {
//          userIDs = append(userIDs, key.(string))
//          return true
//      })
//
//      // Load users from Redis
//      users := make([]User, len(userIDs))
//      grocery.LoadAll(userIDs, &users)
//
//      // Populate m.users
//      m.users = make(map[string]*User)
//
//      for i := 0; i < len(userIDs); i++ {
//          // Use underlying map data to get user's role
//          users[i].Role, _ = m.Load(userIDs[i])
//
//          // Then, store retrieved user model in m.users
//          m.users[userIDs[i]] = &users[i]
//      }
//  }
type CustomMapType interface {
	// Called once the map's data is loaded from Redis, allowing us to populate
	// any custom fields.
	Setup()
}

func NewMap(data map[string]string) *Map {
	m := &Map{Map: sync.Map{}}

	for k, v := range data {
		m.Store(k, v)
	}

	return m
}

// Count returns the number of items stored in the map.
func (m *Map) Count() int {
	num := 0

	m.Range(func(key, value any) bool {
		num += 1
		return true
	})

	return num
}

func (m *Map) MarshalJSON() ([]byte, error) {
	res := make(map[string]string)

	m.Range(func(key, value interface{}) bool {
		res[key.(string)] = value.(string)
		return true
	})

	return json.Marshal(res)
}

func (m *Map) UnmarshalJSON(data []byte) error {
	res := make(map[string]string)

	if err := json.Unmarshal(data, &res); err != nil {
		return err
	}

	for k, v := range res {
		m.Store(k, v)
	}

	return nil
}
