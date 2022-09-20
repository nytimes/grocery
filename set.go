package grocery

import (
	"sync"
)

// Set is used to represent sets stored in Redis. See CustomSetType and Store
// for more information.
type Set struct {
	sync.Map
}

// CustomSetType must be embedded by any struct that is used as a model's field
// and should be represented as a set in Redis. The parent struct must also
// embed grocery.Set. As a motivating example, let's imagine we want to store
// projects in Redis, and each project should contain a set that stores users
// who have access to that project. However, when the project is marshaled into
// JSON, we only want to reveal the number of users who have access, and not
// the identities of the users themselves.
//
//  type User struct {
//      grocery.Base
//      Name string `grocery:"name"`
//  }
//
//  type Project struct {
//      grocery.Base
//      Users *UsersSet `grocery:"users"`
//  }
//
//  type UsersSet struct {
//      grocery.CustomSetType
//      grocery.Set
//  }
//
//  func (s *UsersSet) Setup() {
//      // Setup gets called once the set's contents have already been loaded
//      // from Redis. For this example, we don't need to do anything, but this
//      // function allows for lots of flexibility. See CustomMapType for more
//      // details.
//  }
//
//  func (s *UsersSet) MarshalJSON() {
//      // Only marshal the size of the set, as opposed to all of the set's
//      // contents.
//      return json.Marshal(s.Cardinality())
//  }
type CustomSetType interface {
	Setup()
}

func NewSet(data []string) *Set {
	s := &Set{Map: sync.Map{}}

	for _, k := range data {
		s.Store(k, 0)
	}

	return s
}

// Add adds a key k to the set.
func (s *Set) Add(k any) {
	s.Store(k, 0)
}

// Contains returns true if the set contains k.
func (s *Set) Contains(k any) bool {
	_, ok := s.Load(k)
	return ok
}

// Cardinality returns the size of the set.
func (s *Set) Cardinality() int {
	num := 0

	s.Range(func(key, value any) bool {
		num += 1
		return true
	})

	return num
}

// Iter calls f once for each key in the set.
func (s *Set) Iter(f func(k any) bool) {
	s.Range(func(k, value any) bool {
		return f(k)
	})
}
