package worker

import (
	"testing"
)

func TestSimpleRule(t *testing.T) {
	jsonConfig := `[{
  "operation": "shift",
  "spec": {
    "object.id": "doc.uid",
    "gid2": "doc.guid[1]",
    "allGuids": "doc.guidObjects[*].id"
  }
}]`

	registerTransformRule("test.me", jsonConfig)

	jsonObject := `{
  "doc": {
    "uid": 12345,
    "guid": ["guid0", "guid2", "guid4"],
    "guidObjects": [{"id": "guid0"}, {"id": "guid2"}, {"id": "guid4"}]
  },
  "top-level-key": null
}`

	jsonDest, err := transformJSON("test.me", jsonObject)
	if err != nil {
		t.Errorf("can't transforn: %v", err)
		return
	}
	t.Log("transformation OK.", jsonDest)
}

func TestUnknownRule(t *testing.T) {
	jsonObject := `{
  "doc": {
    "uid": 12345,
    "guid": ["guid0", "guid2", "guid4"],
    "guidObjects": [{"id": "guid0"}, {"id": "guid2"}, {"id": "guid4"}]
  },
  "top-level-key": null
}`

	_, err := transformJSON("test.me2", jsonObject)
	if err != ErrRuleNotDefined {
		t.Errorf("something goes wrong: %v", err)
		return
	}
	t.Log("unknown rule throws error.")
}
