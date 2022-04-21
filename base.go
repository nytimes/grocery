package grocery

import (
	"time"
)

// Base provides default fields for every object stored with grocery.
//
// Unlike most fields, the fields contained within Base are special and should
// not be modified manually, as they are updated by grocery automatically.
type Base struct {
	// The timestamp at which the object was created.
	CreatedAt time.Time `json:"createdAt" grocery:"createdAt,immutable"`

	// The timestamp at which the object was last modified.
	UpdatedAt time.Time `json:"updatedAt" grocery:"updatedAt,immutable"`

	// The object's unique ID.
	ID string `json:"id,omitempty" grocery:"-"`
}
