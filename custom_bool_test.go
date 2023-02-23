package grocery

import (
	"testing"

	"github.com/redis/go-redis/v9"
)

type ZContains bool

func (ZContains) Load(itemID, fieldName string) (bool, error) {
	_, err := C.ZRank(ctx, fieldName, itemID).Result()
	return err == nil, nil
}

type ZContainsTestModel struct {
	Base

	Name     string    `grocery:"name"`
	IsMember ZContains `grocery:"sortedSetTest"`
}

func TestZContains(t *testing.T) {
	model := &ZContainsTestModel{
		Name: "hello world",
	}

	model2 := &ZContainsTestModel{
		Name: "hello world 2",
	}

	id, err := Store(model)
	id2, _ := Store(model2)

	if err != nil {
		t.Error(err)
	}

	C.ZAdd(ctx, "sortedSetTest", redis.Z{Score: 1, Member: id}).Result()

	loadedModel := new(ZContainsTestModel)
	err = Load(id, loadedModel)

	loadedModel2 := new(ZContainsTestModel)
	Load(id2, loadedModel2)

	if err != nil {
		t.Error(err)
	}

	if !loadedModel.IsMember {
		t.Errorf("zcontains FAILED, expected %t but got %t", true, loadedModel.IsMember)
	}

	if loadedModel2.IsMember {
		t.Errorf("zcontains FAILED, expected %t but got %t", false, loadedModel2.IsMember)
	}

	if loadedModel.Name != model.Name {
		t.Errorf("zcontains name FAILED, expected %s but got %s", model.Name, loadedModel.Name)
	}
}
