package relay

import (
	"testing"

	"github.com/bestruirui/octopus/internal/model"
)

func TestOrderGroupCandidates(t *testing.T) {
	items := []model.GroupItem{
		{ID: 1, ChannelID: 1, ModelName: "gpt-4o", Priority: 3},
		{ID: 2, ChannelID: 2, ModelName: "gpt-4o", Priority: 1},
		{ID: 3, ChannelID: 3, ModelName: "gpt-4o", Priority: 2},
	}

	t.Run("no active item returns sorted by priority", func(t *testing.T) {
		group := model.Group{Items: items}
		got := orderGroupCandidates(group)
		ids := idsOf(got)
		want := []int{2, 3, 1} // Priority 1, 2, 3
		if !equal(ids, want) {
			t.Errorf("got %v, want %v", ids, want)
		}
	})

	t.Run("active item first, rest sorted", func(t *testing.T) {
		group := model.Group{ActiveItemID: 1, Items: items}
		got := orderGroupCandidates(group)
		ids := idsOf(got)
		// active=1 first, then 2 (prio1), 3 (prio2)
		want := []int{1, 2, 3}
		if !equal(ids, want) {
			t.Errorf("got %v, want %v", ids, want)
		}
	})

	t.Run("active item not found in items returns sorted", func(t *testing.T) {
		group := model.Group{ActiveItemID: 99, Items: items}
		got := orderGroupCandidates(group)
		ids := idsOf(got)
		want := []int{2, 3, 1}
		if !equal(ids, want) {
			t.Errorf("got %v, want %v", ids, want)
		}
	})

	t.Run("empty items returns nil", func(t *testing.T) {
		group := model.Group{Items: nil}
		if got := orderGroupCandidates(group); got != nil {
			t.Errorf("expected nil, got %v", got)
		}
	})

	t.Run("single item", func(t *testing.T) {
		group := model.Group{Items: []model.GroupItem{{ID: 5, ModelName: "x"}}}
		got := orderGroupCandidates(group)
		if len(got) != 1 || got[0].ID != 5 {
			t.Errorf("got %v, want [5]", idsOf(got))
		}
	})

	t.Run("activie item at last position", func(t *testing.T) {
		items := []model.GroupItem{
			{ID: 10, Priority: 1},
			{ID: 20, Priority: 2},
			{ID: 30, Priority: 3},
		}
		group := model.Group{ActiveItemID: 30, Items: items}
		got := orderGroupCandidates(group)
		ids := idsOf(got)
		want := []int{30, 10, 20}
		if !equal(ids, want) {
			t.Errorf("got %v, want %v", ids, want)
		}
	})
}

func idsOf(items []model.GroupItem) []int {
	ids := make([]int, len(items))
	for i, item := range items {
		ids[i] = item.ID
	}
	return ids
}

func equal(a, b []int) bool {
	if len(a) != len(b) {
		return false
	}
	for i := range a {
		if a[i] != b[i] {
			return false
		}
	}
	return true
}