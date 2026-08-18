package op

import (
	"reflect"
	"testing"

	"github.com/bestruirui/octopus/internal/model"
)

func TestNormalizeChannelKeys(t *testing.T) {
	tests := []struct {
		name     string
		in       model.Channel
		wantKeys []model.ChannelKey
		wantKey  string
	}{
		{
			name:     "空 Keys 有 Key 回填单 Key 主",
			in:       model.Channel{Key: "k1", Model: "gpt-4o"},
			wantKeys: []model.ChannelKey{{Key: "k1", Models: "gpt-4o", IsMain: true}},
			wantKey:  "k1",
		},
		{
			name: "空 Keys 空 Key 保持不变",
			in:   model.Channel{},
		},
		{
			name: "无 IsMain 时第一个为主",
			in: model.Channel{
				Keys: []model.ChannelKey{
					{Key: "a"},
					{Key: "b"},
				},
			},
			wantKeys: []model.ChannelKey{
				{Key: "a", IsMain: true},
				{Key: "b"},
			},
			wantKey: "a",
		},
		{
			name: "多个 IsMain 只保留第一个",
			in: model.Channel{
				Keys: []model.ChannelKey{
					{Key: "a", IsMain: true},
					{Key: "b", IsMain: true},
					{Key: "c"},
				},
			},
			wantKeys: []model.ChannelKey{
				{Key: "a", IsMain: true},
				{Key: "b"},
				{Key: "c"},
			},
			wantKey: "a",
		},
		{
			name: "IsMain 在中间也归一化其余为非主",
			in: model.Channel{
				Keys: []model.ChannelKey{
					{Key: "a"},
					{Key: "b", IsMain: true},
					{Key: "c", IsMain: true},
				},
			},
			wantKeys: []model.ChannelKey{
				{Key: "a"},
				{Key: "b", IsMain: true},
				{Key: "c"},
			},
			wantKey: "b",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ch := tt.in
			normalizeChannelKeys(&ch)
			if ch.Key != tt.wantKey {
				t.Errorf("Key = %q, want %q", ch.Key, tt.wantKey)
			}
			if tt.wantKeys != nil && !reflect.DeepEqual(ch.Keys, tt.wantKeys) {
				t.Errorf("Keys = %+v, want %+v", ch.Keys, tt.wantKeys)
			}
		})
	}
}
