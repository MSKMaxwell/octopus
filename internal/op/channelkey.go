package op

import (
	"strings"

	"github.com/bestruirui/octopus/internal/model"
	"github.com/bestruirui/octopus/internal/utils/xstrings"
)

// ResolveChannelKey 返回请求 modelName 应使用的上游 Key。
// 规则（目标1）：在 Keys 的 Models 中命中 modelName 的 Key 里优先主 Key；
// 命中多个非主 Key 取第一个；全部未命中回退主 Key。
// bool 返回是否有可用 Key（channel 为 nil、Keys 与 Key 均为空时为 false）。
func ResolveChannelKey(channel *model.Channel, modelName string) (string, bool) {
	if channel == nil {
		return "", false
	}
	target := strings.ToLower(strings.TrimSpace(modelName))
	keys := channel.Keys
	if len(keys) == 0 {
		if channel.Key == "" {
			return "", false
		}
		return channel.Key, true
	}

	mainIdx := 0
	for i := range keys {
		if keys[i].IsMain {
			mainIdx = i
			break
		}
	}
	// 主 Key 优先：先扫描主 Key 是否命中
	if keyOwnsModel(&keys[mainIdx], target) {
		return keys[mainIdx].Key, true
	}
	// 其次按顺序取第一个命中的非主 Key
	for i := range keys {
		if i == mainIdx {
			continue
		}
		if keyOwnsModel(&keys[i], target) {
			return keys[i].Key, true
		}
	}
	// 全部未命中回退主 Key
	if keys[mainIdx].Key != "" {
		return keys[mainIdx].Key, true
	}
	return "", false
}

// keyOwnsModel 判断 Key 的模型列表（逗号分隔）是否包含目标模型（大小写不敏感，精确匹配）。
func keyOwnsModel(k *model.ChannelKey, modelName string) bool {
	for _, m := range xstrings.SplitTrimCompact(",", k.Models) {
		if strings.ToLower(strings.TrimSpace(m)) == modelName {
			return true
		}
	}
	return false
}
