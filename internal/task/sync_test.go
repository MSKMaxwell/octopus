package task

import (
	"context"
	"net/http"
	"net/http/httptest"
	"reflect"
	"testing"

	"github.com/bestruirui/octopus/internal/model"
)

// mockUpstream 返回一个按 Authorization 区分 Key 的 /models mock 服务。
func mockUpstream(t *testing.T) *httptest.Server {
	t.Helper()
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.Header.Get("Authorization") {
		case "Bearer key-a":
			w.Write([]byte(`{"object":"list","data":[{"id":"gpt-4o"},{"id":"gpt-4o-mini"}]}`))
		case "Bearer key-b":
			w.Write([]byte(`{"object":"list","data":[{"id":"gpt-4o"},{"id":"claude-3-5-sonnet"}]}`))
		default:
			w.WriteHeader(http.StatusUnauthorized)
			w.Write([]byte(`{"error":{"message":"invalid key"}}`))
		}
	}))
	t.Cleanup(srv.Close)
	return srv
}

func TestSyncChannelModelsPerKey(t *testing.T) {
	srv := mockUpstream(t)
	ctx := context.Background()
	channel := model.Channel{
		Name:    "mock",
		Type:    model.ChannelProviderOpenAI,
		BaseURL: srv.URL,
		Keys: []model.ChannelKey{
			{Key: "key-a", IsMain: true},
			{Key: "key-b"},
		},
	}

	merged, keys, err := syncChannelModels(ctx, channel)
	if err != nil {
		t.Fatalf("syncChannelModels error: %v", err)
	}

	// 合并去重：gpt-4o 重复只保留一个；3 个唯一模型
	wantMerged := []string{"gpt-4o", "gpt-4o-mini", "claude-3-5-sonnet"}
	if !reflect.DeepEqual(merged, wantMerged) {
		t.Errorf("merged = %v, want %v", merged, wantMerged)
	}

	// per-Key 模型列表各自独立
	wantKeys := []model.ChannelKey{
		{Key: "key-a", Models: "gpt-4o,gpt-4o-mini", IsMain: true},
		{Key: "key-b", Models: "gpt-4o,claude-3-5-sonnet"},
	}
	if !reflect.DeepEqual(keys, wantKeys) {
		t.Errorf("keys = %+v, want %+v", keys, wantKeys)
	}
}

func TestSyncChannelModelsAllKeysFail(t *testing.T) {
	srv := mockUpstream(t)
	ctx := context.Background()
	channel := model.Channel{
		Name:    "mock",
		Type:    model.ChannelProviderOpenAI,
		BaseURL: srv.URL,
		Keys: []model.ChannelKey{
			{Key: "bad-key-1"},
			{Key: "bad-key-2"},
		},
	}

	if _, _, err := syncChannelModels(ctx, channel); err == nil {
		t.Error("expected error when all keys fail, got nil")
	}
}

func TestSyncChannelModelsLegacySingleKey(t *testing.T) {
	srv := mockUpstream(t)
	ctx := context.Background()
	channel := model.Channel{
		Name:    "mock",
		Type:    model.ChannelProviderOpenAI,
		BaseURL: srv.URL,
		Key:     "key-a",
	}

	merged, keys, err := syncChannelModels(ctx, channel)
	if err != nil {
		t.Fatalf("syncChannelModels error: %v", err)
	}
	if len(merged) != 2 {
		t.Errorf("merged len = %d, want 2", len(merged))
	}
	// 存量单 Key：回填默认形态且不丢 Key
	if len(keys) != 1 || keys[0].Key != "key-a" || keys[0].Models == "" {
		t.Errorf("keys = %+v, want single key-a with models", keys)
	}
}

func TestKeysOf(t *testing.T) {
	withKeys := model.Channel{Keys: []model.ChannelKey{{Key: "a"}}}
	if got := keysOf(withKeys); len(got) != 1 {
		t.Errorf("keysOf(withKeys) = %v", got)
	}
	legacy := model.Channel{Key: "legacy"}
	got := keysOf(legacy)
	if len(got) != 1 || got[0].Key != "legacy" {
		t.Errorf("keysOf(legacy) = %+v, want [legacy]", got)
	}
}
