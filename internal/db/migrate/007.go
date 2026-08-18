package migrate

import (
	"encoding/json"

	"github.com/bestruirui/octopus/internal/model"
	"gorm.io/gorm"
)

func init() {
	RegisterAfterAutoMigration(Migration{
		Version: 7,
		Up:      migrateBackfillChannelKeys,
	})
}

// migrateBackfillChannelKeys 把存量单 Key 回填为 Keys=[{Key, Models, IsMain:true}]。
// AutoMigrate 已自动补充 keys 列（serializer:json → TEXT），此处仅做数据回填。
func migrateBackfillChannelKeys(db *gorm.DB) error {
	if db == nil {
		return nil
	}
	var channels []model.Channel
	if err := db.Find(&channels).Error; err != nil {
		return err
	}
	for _, ch := range channels {
		if len(ch.Keys) == 0 && ch.Key != "" {
			ch.Keys = []model.ChannelKey{{Key: ch.Key, Models: ch.Model, IsMain: true}}
			keysJSON, err := json.Marshal(ch.Keys)
			if err != nil {
				return err
			}
			// UpdateColumn 绕过 serializer 直接写入序列化值，避免 GORM serializer 在
			// 列名 Update 路径下不生效的差异。
			if err := db.Model(&model.Channel{}).Where("id = ?", ch.ID).
				UpdateColumn("keys", string(keysJSON)).Error; err != nil {
				return err
			}
		}
	}
	return nil
}
