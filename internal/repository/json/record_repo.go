package json

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"

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

// Add 追加一条学习记录：读取现有列表 → 追加 → 写回
func (r *recordRepo) Add(record *domain.LearningRecord) error {
	records, err := r.readRecords()
	if err != nil {
		return err
	}
	records = append(records, record)
	return r.writeRecords(records)
}

// GetByDate 获取某日所有学习记录
func (r *recordRepo) GetByDate(date string) ([]*domain.LearningRecord, error) {
	all, err := r.GetAll()
	if err != nil {
		return nil, err
	}

	result := make([]*domain.LearningRecord, 0)
	for _, rec := range all {
		if rec.Date == date {
			result = append(result, rec)
		}
	}
	return result, nil
}

// GetByWordID 获取某单词的所有历史学习记录
func (r *recordRepo) GetByWordID(wordID string) ([]*domain.LearningRecord, error) {
	all, err := r.GetAll()
	if err != nil {
		return nil, err
	}

	result := make([]*domain.LearningRecord, 0)
	for _, rec := range all {
		if rec.WordID == wordID {
			result = append(result, rec)
		}
	}
	return result, nil
}

// GetAll 获取全部学习记录，文件不存在返回空切片（不报错）
func (r *recordRepo) GetAll() ([]*domain.LearningRecord, error) {
	return r.readRecords()
}

// readRecords 从文件读取学习记录列表
// 文件不存在返回空切片；其他错误返回 error
func (r *recordRepo) readRecords() ([]*domain.LearningRecord, error) {
	data, err := os.ReadFile(r.filePath)
	if err != nil {
		if os.IsNotExist(err) {
			return make([]*domain.LearningRecord, 0), nil
		}
		return nil, fmt.Errorf("读取学习记录文件失败: %w", err)
	}

	var records []*domain.LearningRecord
	if err := json.Unmarshal(data, &records); err != nil {
		return nil, fmt.Errorf("解析学习记录文件失败: %w", err)
	}
	return records, nil
}

// writeRecords 将学习记录列表序列化为缩进 JSON 并写入文件
// 写入前确保目录存在
func (r *recordRepo) writeRecords(records []*domain.LearningRecord) error {
	dir := filepath.Dir(r.filePath)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return fmt.Errorf("创建学习记录目录失败: %w", err)
	}

	data, err := json.MarshalIndent(records, "", "  ")
	if err != nil {
		return fmt.Errorf("序列化学习记录失败: %w", err)
	}

	if err := os.WriteFile(r.filePath, data, 0644); err != nil {
		return fmt.Errorf("写入学习记录文件失败: %w", err)
	}
	return nil
}

// ─────────────────────────────────────────────

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

// Add 追加一条打卡记录：读取现有列表 → 追加 → 写回
func (c *checkInRepo) Add(checkIn *domain.CheckIn) error {
	checkIns, err := c.readCheckIns()
	if err != nil {
		return err
	}
	checkIns = append(checkIns, checkIn)
	return c.writeCheckIns(checkIns)
}

// GetAll 获取所有打卡记录，文件不存在返回空切片（不报错）
func (c *checkInRepo) GetAll() ([]*domain.CheckIn, error) {
	return c.readCheckIns()
}

// GetByDay 获取指定天的打卡记录，不存在返回 nil, nil
func (c *checkInRepo) GetByDay(day int) (*domain.CheckIn, error) {
	all, err := c.readCheckIns()
	if err != nil {
		return nil, err
	}

	for _, ci := range all {
		if ci.Day == day {
			return ci, nil
		}
	}
	// 不存在时返回 nil, nil，由上层判断
	return nil, nil
}

// Exists 检查指定天是否已有打卡记录
func (c *checkInRepo) Exists(day int) (bool, error) {
	ci, err := c.GetByDay(day)
	if err != nil {
		return false, err
	}
	return ci != nil, nil
}

// readCheckIns 从文件读取打卡记录列表
// 文件不存在返回空切片；其他错误返回 error
func (c *checkInRepo) readCheckIns() ([]*domain.CheckIn, error) {
	data, err := os.ReadFile(c.filePath)
	if err != nil {
		if os.IsNotExist(err) {
			return make([]*domain.CheckIn, 0), nil
		}
		return nil, fmt.Errorf("读取打卡记录文件失败: %w", err)
	}

	var checkIns []*domain.CheckIn
	if err := json.Unmarshal(data, &checkIns); err != nil {
		return nil, fmt.Errorf("解析打卡记录文件失败: %w", err)
	}
	return checkIns, nil
}

// writeCheckIns 将打卡记录列表序列化为缩进 JSON 并写入文件
// 写入前确保目录存在
func (c *checkInRepo) writeCheckIns(checkIns []*domain.CheckIn) error {
	dir := filepath.Dir(c.filePath)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return fmt.Errorf("创建打卡记录目录失败: %w", err)
	}

	data, err := json.MarshalIndent(checkIns, "", "  ")
	if err != nil {
		return fmt.Errorf("序列化打卡记录失败: %w", err)
	}

	if err := os.WriteFile(c.filePath, data, 0644); err != nil {
		return fmt.Errorf("写入打卡记录文件失败: %w", err)
	}
	return nil
}
