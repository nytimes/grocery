package grocery

import (
	"testing"
)

type A struct {
	Base
	Name string `grocery:"name"`
}

type B struct {
	Base
	A *A `grocery:"ref"`
}

type Ctest struct {
	Base
	As []*A `grocery:"arrRef"`
}

func TestStoreReference(t *testing.T) {
	a := &A{Name: "bob"}
	aID, err := Store(a)

	if err != nil {
		t.Error(err)
	}

	b := &B{A: a}
	bID, err := Store(b)

	if err != nil {
		t.Error(err)
		return
	}

	loadedB := new(B)
	Load(bID, loadedB)

	if loadedB.A.ID != aID {
		t.Errorf("ref FAILED, expected %s but got %s", aID, loadedB.A.ID)
	}

	if loadedB.A.Name != a.Name {
		t.Errorf("ref name FAILED, expected %s but got %s", a.Name, loadedB.A.Name)
	}
}

func TestStoreListReference(t *testing.T) {
	a := &A{Name: "bob"}
	aID, err := Store(a)

	if err != nil {
		t.Error(err)
	}

	c := &Ctest{As: []*A{a}}
	cID, err := Store(c)

	if err != nil {
		t.Error(err)
		return
	}

	loadedC := new(Ctest)
	Load(cID, loadedC)

	if loadedC.As[0].ID != aID {
		t.Errorf("ref FAILED, expected %s but got %s", aID, loadedC.As[0].ID)
	}

	if loadedC.As[0].Name != a.Name {
		t.Errorf("ref name FAILED, expected %s but got %s", a.Name, loadedC.As[0].Name)
	}
}
