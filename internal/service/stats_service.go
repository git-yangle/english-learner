package service

import (
	"fmt"
	"sort"
	"time"

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

// GetWordMastery 获取单词掌握状态
func (s *statsService) GetWordMastery(wordID string) (WordMasteryStatus, error) {
	// 获取该单词的所有历史学习记录
	records, err := s.recordRepo.GetByWordID(wordID)
	if err != nil {
		return MasteryStatusNotLearned, fmt.Errorf("获取单词 %s 的学习记录失败: %w", wordID, err)
	}

	// 无记录则为未学习状态
	if len(records) == 0 {
		return MasteryStatusNotLearned, nil
	}

	// 按日期升序排序，取最近3条记录
	sort.Slice(records, func(i, j int) bool {
		return records[i].Date < records[j].Date
	})

	// 取最近3条（若不足3条则取全部）
	recent := records
	if len(records) > 3 {
		recent = records[len(records)-3:]
	}

	// 最近3次（或全部记录）全部正确 → 已掌握
	allCorrect := true
	for _, r := range recent {
		if !r.Correct {
			allCorrect = false
			break
		}
	}

	if allCorrect {
		return MasteryStatusMastered, nil
	}
	return MasteryStatusLearning, nil
}

// GetOverview 获取总览统计数据
func (s *statsService) GetOverview() (*OverviewStats, error) {
	// 获取所有打卡记录
	checkIns, err := s.checkInRepo.GetAll()
	if err != nil {
		return nil, fmt.Errorf("获取打卡记录失败: %w", err)
	}

	// 按日期排序，方便后续处理
	sort.Slice(checkIns, func(i, j int) bool {
		return checkIns[i].Date < checkIns[j].Date
	})

	// 构建14天得分趋势：将打卡记录按日期映射，无打卡则得分为 0
	// 以最早打卡日期为起点，构建连续14天的得分序列
	scoreTrend := make([]float64, 14)
	if len(checkIns) > 0 {
		// 以第1天打卡日期作为基准日期
		baseDate, err := time.Parse("2006-01-02", checkIns[0].Date)
		if err == nil {
			// 构建日期到得分的映射
			dateScoreMap := make(map[string]float64)
			for _, ci := range checkIns {
				dateScoreMap[ci.Date] = ci.Score
			}
			// 填充14天得分趋势
			for i := 0; i < 14; i++ {
				dayDate := baseDate.AddDate(0, 0, i).Format("2006-01-02")
				scoreTrend[i] = dateScoreMap[dayDate]
			}
		}
	}

	// 获取所有学习记录，按 WordID 分组
	allRecords, err := s.recordRepo.GetAll()
	if err != nil {
		return nil, fmt.Errorf("获取学习记录失败: %w", err)
	}

	// 按 WordID 分组，记录哪些单词已有学习数据
	wordRecordMap := make(map[string][]*domain.LearningRecord)
	for _, r := range allRecords {
		wordRecordMap[r.WordID] = append(wordRecordMap[r.WordID], r)
	}

	// 获取所有词库
	libraries, err := s.libraryRepo.GetAll()
	if err != nil {
		return nil, fmt.Errorf("获取词库失败: %w", err)
	}

	// 按 Category 分组统计，遍历所有词库中的单词
	categoryMap := make(map[string]*CategoryProgress)
	totalWords := 0
	masteredWords := 0

	for _, lib := range libraries {
		for _, word := range lib.Words {
			totalWords++

			// 获取该单词的掌握状态
			mastery, err := s.GetWordMastery(word.ID)
			if err != nil {
				return nil, fmt.Errorf("获取单词 %s 掌握状态失败: %w", word.ID, err)
			}

			// 按 Category 分组统计
			category := word.Category
			if _, ok := categoryMap[category]; !ok {
				categoryMap[category] = &CategoryProgress{
					Category: category,
				}
			}
			cp := categoryMap[category]
			cp.Total++

			switch mastery {
			case MasteryStatusMastered:
				cp.Mastered++
				masteredWords++
			case MasteryStatusLearning:
				cp.Learning++
			case MasteryStatusNotLearned:
				cp.NotLearned++
			}
		}
	}

	// 计算各分类掌握率
	categoryProgress := make([]*CategoryProgress, 0, len(categoryMap))
	for _, cp := range categoryMap {
		if cp.Total > 0 {
			cp.MasteryRate = float64(cp.Mastered) / float64(cp.Total)
		}
		categoryProgress = append(categoryProgress, cp)
	}

	// 按分类名称排序，保证输出顺序稳定
	sort.Slice(categoryProgress, func(i, j int) bool {
		return categoryProgress[i].Category < categoryProgress[j].Category
	})

	// 计算连续打卡天数：将打卡日期放入 set，从今天往前逐日检查
	dateSet := make(map[string]bool)
	for _, ci := range checkIns {
		dateSet[ci.Date] = true
	}

	streakDays := 0
	cursor := time.Now()
	for {
		dateStr := cursor.Format("2006-01-02")
		if !dateSet[dateStr] {
			break
		}
		streakDays++
		cursor = cursor.AddDate(0, 0, -1)
	}

	return &OverviewStats{
		CheckIns:         checkIns,
		ScoreTrend:       scoreTrend,
		CategoryProgress: categoryProgress,
		TotalWords:       totalWords,
		MasteredWords:    masteredWords,
		StreakDays:       streakDays,
	}, nil
}
