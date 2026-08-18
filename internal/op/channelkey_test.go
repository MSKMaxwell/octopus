package op

import (
	"testing"

	"github.com/bestruirui/octopus/internal/model"
)

func TestResolveChannelKey(t *testing.T) {
	dualKey := &model.Channel{
		Key: "main",
		Keys: []model.ChannelKey{
			{Key: "main", Models: "gpt-4o, gpt-4o-mini", IsMain: true},
			{Key: "sub", Models: "gpt-4o, claude-3-5-sonnet"},
		},
	}
	legacySingle := &model.Channel{Key: "legacy", Model: "gpt-4o"}
	empty := &model.Channel{}
	nilChannel := (*model.Channel)(nil)

	tests := []struct {
		name      string
		channel   *model.Channel
		modelName string
		wantKey   string
		wantOK    bool
	}{
		{name: "主 Key 命中", channel: dualKey, modelName: "gpt-4o-mini", wantKey: "main", wantOK: true},
		{name: "重复模型主 Key 优先", channel: dualKey, modelName: "gpt-4o", wantKey: "main", wantOK: true},
		{name: "仅非主 Key 命中", channel: dualKey, modelName: "claude-3-5-sonnet", wantKey: "sub", wantOK: true},
		{name: "无命中回退主 Key", channel: dualKey, modelName: "unknown-model", wantKey: "main", wantOK: true},
		{name: "大小写不敏感", channel: dualKey, modelName: "GPT-4O", wantKey: "main", wantOK: true},
		{name: "精确匹配不误命中前缀", channel: dualKey, modelName: "gpt-4", wantKey: "main", wantOK: true},
		{name: "存量单 Key 形态", channel: legacySingle, modelName: "anything", wantKey: "legacy", wantOK: true},
		{name: "无 Key 不可用", channel: empty, modelName: "x", wantOK: false},
		{name: "nil 渠道不可用", channel: nilChannel, modelName: "x", wantOK: false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, ok := ResolveChannelKey(tt.channel, tt.modelName)
			if ok != tt.wantOK {
				t.Errorf("ok = %v, want %v", ok, tt.wantOK)
			}
			if ok && got != tt.wantKey {
				t.Errorf("key = %q, want %q", got, tt.wantKey)
			}
		})
	}
}
