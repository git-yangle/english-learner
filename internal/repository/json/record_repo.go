package json

import (
	"english-learner/internal/domain"
	"english-learner/internal/repository"
)

// recordRepo 学习记录仓储 JSON 文件实现
// 读写 data/user/records.json
type recordRepo struct {
	// filePath records.json 文件路径
	filePath string
}

// NewRecordRepo 创建学习记录仓储实例
func NewRecordRepo(filePath string) repository.RecordRepository {
	return &recordRepo{filePath: filePath}
}

// Add 追加一条学习记录
func (r *recordRepo) Add(record *domain.LearningRecord) error {
	// TODO: 读取现有记录列表 → 追加新记录 → 写回文件
	return nil
}

// GetByDate 获取某日所有记录
func (r *recordRepo) GetByDate(date string) ([]*domain.LearningRecord, error) {
	// TODO: 读取所有记录，过滤出 date 字段匹配的记录返回
	return nil, nil
}

// GetByWordID 获取某单词所有历史记录
func (r *recordRepo) GetByWordID(wordID string) ([]*domain.LearningRecord, error) {
	// TODO: 读取所有记录，过滤出 word_id 字段匹配的记录返回
	return nil, nil
}

// GetAll 获取全部学习记录
func (r *recordRepo) GetAll() ([]*domain.LearningRecord, error) {
	// TODO: 读取并反序列化 filePath JSON 文件，返回全部记录
	return nil, nil
}

// checkInRepo 打卡记录仓储 JSON 文件实现
// 读写 data/user/checkins.json
type checkInRepo struct {
	// filePath checkins.json 文件路径
	filePath string
}

// NewCheckInRepo 创建打卡记录仓储实例
func NewCheckInRepo(filePath string) repository.CheckInRepository {
	return &checkInRepo{filePath: filePath}
}

// Add 添加打卡记录
func (c *checkInRepo) Add(checkIn *domain.CheckIn) error {
	// TODO: 读取现有打卡列表 → 追加新打卡 → 写回文件
	return nil
}

// GetAll 获取所有打卡记录
func (c *checkInRepo) GetAll() ([]*domain.CheckIn, error) {
	// TODO: 读取并反序列化 filePath JSON 文件，返回全部打卡记录
	return nil, nil
}

// GetByDay 获取某天打卡记录
func (c *checkInRepo) GetByDay(day int) (*domain.CheckIn, error) {
	// TODO: 读取所有打卡记录，过滤出 day 字段匹配的记录返回
	return nil, nil
}

// Exists 检查某天是否已打卡
func (c *checkInRepo) Exists(day int) (bool, error) {
	// TODO: 读取所有打卡记录，判断是否存在 day 字段匹配的记录
	return false, nil
}
