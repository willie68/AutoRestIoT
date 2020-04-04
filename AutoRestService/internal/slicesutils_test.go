package slicesutils_test

import (
	"testing"

	slicesutils "github.com/willie68/schematic-service-go/internal"
)

func TestContains(t *testing.T) {
	mySlice := []string{"Willie", "Arthur", "Till"}
	value := slicesutils.Contains(mySlice, "Willie")
	if value != true {
		t.Errorf("Willie was not in the slice")
	}
}
func TestRemoveString(t *testing.T) {
	mySlice := []string{"Willie", "Arthur", "Till"}
	value := slicesutils.RemoveString(mySlice, "Willie")
	if slicesutils.Contains(value, "Willie") {
		t.Errorf("Willie was not removed from the slice")
	}
	value = slicesutils.RemoveString(mySlice, "Herman")
	if len(value) != 3 {
		t.Errorf("slice not unchanged")
	}
}

func TestRemove(t *testing.T) {
	mySlice := []string{"Willie", "Arthur", "Till"}
	value := slicesutils.Remove(mySlice, 0)
	if slicesutils.Contains(value, "Willie") {
		t.Errorf("Willie was not removed from the slice")
	}
}

func TestFind(t *testing.T) {
	mySlice := []string{"Willie", "Arthur", "Till"}
	value := slicesutils.Find(mySlice, "Willie")
	if value != 0 {
		t.Errorf("Willie was not found in the slice: index: %d", value)
	}
	value = slicesutils.Find(mySlice, "Arthur")
	if value != 1 {
		t.Errorf("Arthur was not found in the slice: index: %d", value)
	}
	value = slicesutils.Find(mySlice, "Till")
	if value != 2 {
		t.Errorf("Till was not found in the slice: index: %d", value)
	}
	value = slicesutils.Find(mySlice, "till")
	if value >= 0 {
		t.Errorf("till was wrongly found in the slice: index: %d", value)
	}
	value = slicesutils.Find(mySlice, "Herman")
	if value >= 0 {
		t.Errorf("Herman was wrongly found in the slice: index: %d", value)
	}
}
