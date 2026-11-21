package template

import "testing"

func TestPublishedVersionOwnsItsTree(t *testing.T) {
	draft := []Node{{ID: "root", Name: "工程"}, {ID: "step", ParentID: "root", Name: "工序"}}
	version, err := Publish("repo", "template", 1, draft)
	if err != nil {
		t.Fatal(err)
	}
	draft[1].Name = "新草稿"
	version.Nodes()[1].Name = "外部修改"
	if got := version.Nodes()[1].Name; got != "工序" {
		t.Fatalf("published step changed to %q", got)
	}
}

func TestValidateTreeRejectsDisconnectedCycle(t *testing.T) {
	nodes := []Node{{ID: "root", Name: "工程"}, {ID: "a", ParentID: "b", Name: "A"}, {ID: "b", ParentID: "a", Name: "B"}}
	if err := ValidateTree(nodes); err != ErrInvalidTree {
		t.Fatalf("got %v, want ErrInvalidTree", err)
	}
}
