package task

import (
	"context"
	"fmt"
	"reflect"
	"strings"
	"time"

	"github.com/bestruirui/octopus/internal/helper"
	"github.com/bestruirui/octopus/internal/model"
	"github.com/bestruirui/octopus/internal/op"
	"github.com/bestruirui/octopus/internal/utils/diff"
	"github.com/bestruirui/octopus/internal/utils/xstrings"
	"github.com/charmbracelet/log"
)

var lastSyncModelsTime = time.Now()

// SyncModelsTask 同步模型任务
func SyncModelsTask() {
	log.Debugf("sync models task started")
	startTime := time.Now()
	defer func() {
		log.Debugf("sync models task finished, sync time: %s", time.Since(startTime))
	}()
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Minute)
	defer cancel()
	channels, err := op.ChannelList(ctx)
	if err != nil {
		log.Errorf("failed to list channels: %v", err)
		return
	}
	totalNewModels := make([]string, 0, 128)
	seenTotalNewModels := make(map[string]struct{}, 128)
	for _, channel := range channels {
		if !channel.AutoSync {
			continue
		}
		merged, updatedKeys, err := syncChannelModels(ctx, channel)
		if err != nil {
			log.Warnf("failed to fetch models for channel %s: %v", channel.Name, err)
			continue
		}
		oldModels := xstrings.SplitTrimCompact(",", channel.Model)
		for _, m := range merged {
			m = strings.ToLower(m)
			if _, ok := seenTotalNewModels[m]; ok {
				continue
			}
			seenTotalNewModels[m] = struct{}{}
			totalNewModels = append(totalNewModels, m)
		}
		deletedModels, addedModels := diff.Diff(oldModels, merged)
		keysChanged := !reflect.DeepEqual(updatedKeys, keysOf(channel))
		if len(deletedModels) > 0 || len(addedModels) > 0 || keysChanged {
			req := model.ChannelUpdateRequest{ID: channel.ID}
			if len(deletedModels) > 0 || len(addedModels) > 0 {
				fetchModelStr := strings.Join(merged, ",")
				req.Model = &fetchModelStr
			}
			if keysChanged {
				req.Keys = &updatedKeys
			}
			if _, err := op.ChannelUpdate(&req, ctx); err != nil {
				log.Errorf("failed to update channel %s: %v", channel.Name, err)
				continue
			}
		}
		// 批量删除消失的模型对应的 GroupItem
		if len(deletedModels) > 0 {
			log.Infof("deleted channel %s models: %v", channel.Name, deletedModels)
			keys := make([]model.GroupIDAndLLMName, len(deletedModels))
			for i, m := range deletedModels {
				keys[i] = model.GroupIDAndLLMName{ChannelID: channel.ID, ModelName: m}
			}
			if err := op.GroupItemBatchDelByChannelAndModels(keys, ctx); err != nil {
				log.Errorf("failed to batch delete group items for channel %s: %v", channel.Name, err)
			}
		}
	}
	llmPrice, err := op.LLMList(ctx)
	if err != nil {
		log.Errorf("failed to list models price: %v", err)
		return
	}
	llmPriceNames := make([]string, 0, len(llmPrice))
	for _, price := range llmPrice {
		llmPriceNames = append(llmPriceNames, price.Name)
	}

	deletedNorm, addedNorm := diff.Diff(llmPriceNames, totalNewModels)
	if len(deletedNorm) > 0 {
		if err := helper.LLMPriceDeleteFromDBWithNoPrice(deletedNorm, ctx); err != nil {
			log.Errorf("failed to batch delete models price: %v", err)
		}
	}
	if len(addedNorm) > 0 {
		if err := helper.LLMPriceAddToDB(addedNorm, ctx); err != nil {
			log.Errorf("failed to add models price: %v", err)
		}
	}
	lastSyncModelsTime = time.Now()
}

// syncChannelModels 对渠道内每个 Key 独立拉取模型列表（复用渠道的 BaseURL/Type/MatchRegex/CustomHeader，
// 仅替换 Authorization），并将各 Key 的模型合并去重（大小写不敏感，保留首个原始写法）。
// 返回合并后的模型列表与写入了各自 Models 的 Keys 副本；所有 Key 均失败时返回错误。
func syncChannelModels(ctx context.Context, channel model.Channel) ([]string, []model.ChannelKey, error) {
	keys := channel.Keys
	if len(keys) == 0 {
		keys = []model.ChannelKey{{Key: channel.Key}}
	}
	updatedKeys := make([]model.ChannelKey, len(keys))
	copy(updatedKeys, keys)

	merged := make([]string, 0, 64)
	seenMerged := make(map[string]struct{}, 64)
	anySuccess := false
	for i := range updatedKeys {
		if updatedKeys[i].Key == "" {
			continue
		}
		keyChannel := channel
		keyChannel.Key = updatedKeys[i].Key
		fetchModels, err := helper.FetchModels(ctx, keyChannel)
		if err != nil {
			log.Warnf("failed to fetch models for channel %s key[%d]: %v", channel.Name, i, err)
			continue
		}
		anySuccess = true
		updatedKeys[i].Models = strings.Join(fetchModels, ",")
		for _, m := range fetchModels {
			lower := strings.ToLower(strings.TrimSpace(m))
			if lower == "" {
				continue
			}
			if _, ok := seenMerged[lower]; ok {
				continue
			}
			seenMerged[lower] = struct{}{}
			merged = append(merged, strings.TrimSpace(m))
		}
	}
	if !anySuccess {
		return nil, nil, fmt.Errorf("all keys failed to fetch models")
	}
	return merged, updatedKeys, nil
}

// keysOf 返回渠道 Keys 的副本；空 Keys 时返回与 syncChannelModels 相同的默认形态。
func keysOf(channel model.Channel) []model.ChannelKey {
	if len(channel.Keys) > 0 {
		return channel.Keys
	}
	return []model.ChannelKey{{Key: channel.Key}}
}

func GetLastSyncModelsTime() time.Time {
	return lastSyncModelsTime
}
