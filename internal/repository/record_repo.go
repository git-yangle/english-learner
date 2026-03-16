package repository

import "english-learner/internal/domain"

// RecordRepository 学习记录仓储接口
type RecordRepository interface {
	// Add 追加一条学习记录
	Add(record *domain.LearningRecord) error
	// GetByDate 获取某日所有记录
	GetByDate(date string) ([]*domain.LearningRecord, error)
	// GetByWordID 获取某单词所有历史记录
	GetByWordID(wordID string) ([]*domain.LearningRecord, error)
	// GetAll 获取全部记录
	GetAll() ([]*domain.LearningRecord, error)
}

// CheckInRepository 打卡记录仓储接口
type CheckInRepository interface {
	// Add 添加打卡记录
	Add(checkIn *domain.CheckIn) error
	// GetAll 获取所有打卡记录
	GetAll() ([]*domain.CheckIn, error)
	// GetByDay 获取某天打卡记录
	GetByDay(day int) (*domain.CheckIn, error)
	// Exists 检查某天是否已打卡
	Exists(day int) (bool, error)
}
