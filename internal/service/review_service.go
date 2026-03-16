package service

import (
	"fmt"
	"time"

	"english-learner/internal/domain"
	"english-learner/internal/repository"
)

// ReviewService 复习词筛选服务接口
type ReviewService interface {
	// GetReviewWords 根据历史错误率，计算指定日期需要复习的单词ID列表
	// 策略：错误率>60% 次日必复习；30-60% 2天后；<30% 艾宾浩斯间隔
	GetReviewWords(date string) ([]string, error)
	// CalcWordStats 计算单词的学习统计数据
	CalcWordStats(wordID string) (*domain.WordStats, error)
}

type reviewService struct {
	recordRepo repository.RecordRepository
}

// NewReviewService 创建复习服务实例
func NewReviewService(recordRepo repository.RecordRepository) ReviewService {
	return &reviewService{recordRepo: recordRepo}
}

// CalcWordStats 计算单词的学习统计数据
func (s *reviewService) CalcWordStats(wordID string) (*domain.WordStats, error) {
	records, err := s.recordRepo.GetByWordID(wordID)
	if err != nil {
		return nil, fmt.Errorf("获取单词 %s 的学习记录失败: %w", wordID, err)
	}

	stats := &domain.WordStats{
		WordID: wordID,
	}

	if len(records) == 0 {
		return stats, nil
	}

	// 统计总次数和正确次数，同时记录最后学习日期
	for _, r := range records {
		stats.TotalCount++
		if r.Correct {
			stats.CorrectCount++
		}
		// 取所有记录中最新的日期（字符串 yyyy-MM-dd 可直接比较大小）
		if r.Date > stats.LastStudied {
			stats.LastStudied = r.Date
		}
	}

	// 错误率 = 1 - 正确率
	stats.ErrorRate = 1.0 - float64(stats.CorrectCount)/float64(stats.TotalCount)

	return stats, nil
}

// GetReviewWords 根据历史错误率计算指定日期需要复习的单词列表
func (s *reviewService) GetReviewWords(date string) ([]string, error) {
	// 获取所有学习记录
	allRecords, err := s.recordRepo.GetAll()
	if err != nil {
		return nil, fmt.Errorf("获取学习记录失败: %w", err)
	}

	// 按 WordID 分组，统计每个词的正确次数、总次数、最后学习日期
	type wordStat struct {
		totalCount   int
		correctCount int
		lastStudied  string // yyyy-MM-dd，字符串可直接比较
	}
	statMap := make(map[string]*wordStat)

	for _, r := range allRecords {
		ws, ok := statMap[r.WordID]
		if !ok {
			ws = &wordStat{}
			statMap[r.WordID] = ws
		}
		ws.totalCount++
		if r.Correct {
			ws.correctCount++
		}
		if r.Date > ws.lastStudied {
			ws.lastStudied = r.Date
		}
	}

	// 艾宾浩斯复习间隔（天数）
	ebinghausIntervals := map[int]bool{
		1: true, 2: true, 4: true, 7: true, 15: true,
	}

	// 筛选需要复习的单词
	var reviewWords []string
	for wordID, ws := range statMap {
		// 从未学过的词不加入复习
		if ws.totalCount == 0 {
			continue
		}

		// 计算错误率和距今天数
		errorRate := 1.0 - float64(ws.correctCount)/float64(ws.totalCount)
		days := daysBetween(ws.lastStudied, date)

		// 按策略决定是否加入复习列表
		shouldReview := false
		switch {
		case errorRate > 0.6:
			// 高错误率：距上次学习满 1 天即复习
			shouldReview = days >= 1
		case errorRate >= 0.3:
			// 中错误率：距上次学习满 2 天才复习
			shouldReview = days >= 2
		default:
			// 低错误率：按艾宾浩斯间隔复习
			shouldReview = ebinghausIntervals[days]
		}

		if shouldReview {
			reviewWords = append(reviewWords, wordID)
		}
	}

	return reviewWords, nil
}

// daysBetween 计算两个 "yyyy-MM-dd" 日期字符串之间的天数差（date - base）
// 返回值为正数表示 date 晚于 base，负数表示 date 早于 base
func daysBetween(base, date string) int {
	const layout = "2006-01-02"

	baseTime, err := time.Parse(layout, base)
	if err != nil {
		return 0
	}
	dateTime, err := time.Parse(layout, date)
	if err != nil {
		return 0
	}

	// 截断到天，避免时区导致的小时差异影响结果
	diff := dateTime.Truncate(24 * time.Hour).Sub(baseTime.Truncate(24 * time.Hour))
	return int(diff.Hours() / 24)
}
