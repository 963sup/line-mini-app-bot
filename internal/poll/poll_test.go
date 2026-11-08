package poll

import "testing"

func TestParseRejectsAmbiguousPolls(t *testing.T) {
	for _, text := range []string{"投票 題目｜A", "投票 題目｜A｜A", "投票 ｜A｜B", "投票 題目｜A｜"} {
		if _, err := Parse(text); err == nil {
			t.Errorf("accepted %q", text)
		}
	}
	p, err := Parse("投票 午餐｜飯｜麵")
	if err != nil || len(p.Options) != 2 || p.Question != "午餐" {
		t.Fatalf("poll=%+v err=%v", p, err)
	}
}
