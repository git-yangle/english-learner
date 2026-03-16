package service

import (
	"english-learner/internal/domain"
	"english-learner/internal/repository"
)

// WordMasteryStatus 单词掌握状态
type WordMasteryStatus string

const (
	MasteryStatusMastered   WordMasteryStatus = "mastered"    // 最近3次全对
	MasteryStatusLearning   WordMasteryStatus = "learning"    // 学习中
	MasteryStatusNotLearned WordMasteryStatus = "not_learned" // 未学习
)

// CategoryProgress 场景分类进度
type CategoryProgress struct {
	Category    string  `json:"category"`
	Total       int     `json:"total"`
	Mastered    int     `json:"mastered"`
	Learning    int     `json:"learning"`
	NotLearned  int     `json:"not_learned"`
	MasteryRate float64 `json:"mastery_rate"`
}

// OverviewStats 总体统计
type OverviewStats struct {
	CheckIns         []*domain.CheckIn   `json:"check_ins"`         // 14天打卡日历
	ScoreTrend       []float64           `json:"score_trend"`       // 14天得分趋势
	CategoryProgress []*CategoryProgress `json:"category_progress"` // 各分类掌握情况
	TotalWords       int                 `json:"total_words"`
	MasteredWords    int                 `json:"mastered_words"`
	StreakDays       int                 `json:"streak_days"`
}

// StatsService 统计服务接口
type StatsService interface {
	// GetOverview 获取总览统计数据
	GetOverview() (*OverviewStats, error)
	// GetWordMastery 获取单词掌握状态
	GetWordMastery(wordID string) (WordMasteryStatus, error)
}

type statsService struct {
	checkInRepo repository.CheckInRepository
	recordRepo  repository.RecordRepository
	libraryRepo repository.LibraryRepository
}

// NewStatsService 创建统计服务实例
func NewStatsService(checkInRepo repository.CheckInRepository, recordRepo repository.RecordRepository, libraryRepo repository.LibraryRepository) StatsService {
	return &statsService{
		checkInRepo: checkInRepo,
		recordRepo:  recordRepo,
		libraryRepo: libraryRepo,
	}
}

// GetOverview 获取总览统计数据
func (s *statsService) GetOverview() (*OverviewStats, error) {
	// TODO:
	// 1. 获取所有打卡记录，提取14天日历和得分趋势
	// 2. 获取所有学习记录，判断每个单词掌握状态（最近3次全对=mastered）
	// 3. 按 Category 分组统计各分类掌握进度
	// 4. 计算连续打卡天数、总词数、已掌握词数
	return nil, nil
}

// GetWordMastery 获取单词掌握状态
func (s *statsService) GetWordMastery(wordID string) (WordMasteryStatus, error) {
	// TODO:
	// 1. 调用 recordRepo.GetByWordID(wordID) 获取历史记录
	// 2. 无记录 → not_learned
	// 3. 最近3次全部正确 → mastered
	// 4. 其他 → learning
	return MasteryStatusNotLearned, nil
}
