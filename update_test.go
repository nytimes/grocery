package grocery

import (
	"testing"
	"time"
)

type UpdateTestModel struct {
	Base

	StringVal string
	TimeVal   time.Time
}

func TestSetZeroValues(t *testing.T) {
	m := &UpdateTestModel{
		StringVal: "asdf",
	}

	id, err := Store(m)

	if err != nil {
		t.Error(err)
	}

	// Check 0
	model := new(UpdateTestModel)
	err = Load(id, model)

	if err != nil {
		t.Error(err)
	}

	if model.StringVal != m.StringVal {
		t.Errorf("TestSetZeroValues FAILED, initial value was not set correctly")
	}

	// Update 1
	m.StringVal = ""
	err = Update(id, m)

	if err != nil {
		t.Error(err)
	}

	// Check 1
	Load(id, model)

	if model.StringVal == "" {
		t.Errorf("TestSetZeroValues FAILED, zero value was set")
	}

	// Update 2
	m.StringVal = ""
	UpdateWithOptions(id, m, &UpdateOptions{
		SetZeroValues: true,
	})

	// Check 2
	Load(id, model)

	if model.StringVal != "" {
		t.Errorf("TestSetZeroValues FAILED, zero value was not set")
	}
}

func TestSetTimeValue(t *testing.T) {
	m := &UpdateTestModel{
		TimeVal: time.Now(),
	}

	id, err := Store(m)

	if err != nil {
		t.Error(err)
	}

	model := new(UpdateTestModel)
	err = Load(id, model)

	if err != nil {
		t.Error(err)
	}

	if m.TimeVal.Before(model.TimeVal) || m.TimeVal.Sub(model.TimeVal) > time.Second {
		t.Errorf("TestSetTimeValue FAILED, initial value was not set correctly")
	}
}
