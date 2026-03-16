package service

import (
	"errors"
	"fmt"
	"sort"
	"time"

	"english-learner/internal/domain"
	"english-learner/internal/repository"
)

// DailyScore 当日各模式得分
type DailyScore struct {
	QuizScore      float64 `json:"quiz_score"`
	DictationScore float64 `json:"dictation_score"`
	TotalScore     float64 `json:"total_score"` // quiz*0.4 + dictation*0.6
	Completed      bool    `json:"completed"`   // totalScore >= 0.6
}

// CheckInService 打卡服务接口
type CheckInService interface {
	// CalcDailyScore 计算今日得分（不打卡，仅计算）
	CalcDailyScore() (*DailyScore, error)
	// DoCheckIn 执行打卡（计算得分+保存+触发次日复习词生成）
	DoCheckIn(studyDurationSec int) (*domain.CheckIn, error)
	// GetStats 获取打卡统计（日历、连续天数等）
	GetStats() (*CheckInStats, error)
}

// CheckInStats 打卡统计汇总
type CheckInStats struct {
	CheckIns      []*domain.CheckIn `json:"check_ins"`
	TotalDays     int               `json:"total_days"`
	CompletedDays int               `json:"completed_days"`
	StreakDays    int               `json:"streak_days"` // 当前连续天数
	AvgScore      float64           `json:"avg_score"`
}

type checkInService struct {
	checkInRepo repository.CheckInRepository
	recordRepo  repository.RecordRepository
	planRepo    repository.PlanRepository
	reviewSvc   ReviewService
}

// NewCheckInService 创建打卡服务实例
func NewCheckInService(checkInRepo repository.CheckInRepository, recordRepo repository.RecordRepository, planRepo repository.PlanRepository, reviewSvc ReviewService) CheckInService {
	return &checkInService{
		checkInRepo: checkInRepo,
		recordRepo:  recordRepo,
		planRepo:    planRepo,
		reviewSvc:   reviewSvc,
	}
}

// CalcDailyScore 计算今日各模式得分（不执行打卡，仅统计）
func (s *checkInService) CalcDailyScore() (*DailyScore, error) {
	today := time.Now().Format("2006-01-02")

	// 获取今日所有学习记录
	records, err := s.recordRepo.GetByDate(today)
	if err != nil {
		return nil, fmt.Errorf("获取今日学习记录失败: %w", err)
	}

	// 分组统计 quiz 和 dictation 模式的正确率
	quizCorrect, quizTotal := 0, 0
	dictCorrect, dictTotal := 0, 0

	for _, r := range records {
		switch r.Mode {
		case domain.StudyModeQuiz:
			quizTotal++
			if r.Correct {
				quizCorrect++
			}
		case domain.StudyModeDictation:
			dictTotal++
			if r.Correct {
				dictCorrect++
			}
		}
	}

	// 无题目时得分为 0
	var quizScore float64
	if quizTotal > 0 {
		quizScore = float64(quizCorrect) / float64(quizTotal)
	}

	var dictScore float64
	if dictTotal > 0 {
		dictScore = float64(dictCorrect) / float64(dictTotal)
	}

	// 综合得分：quiz 权重 0.4，dictation 权重 0.6
	totalScore := quizScore*0.4 + dictScore*0.6
	completed := totalScore >= 0.6

	return &DailyScore{
		QuizScore:      quizScore,
		DictationScore: dictScore,
		TotalScore:     totalScore,
		Completed:      completed,
	}, nil
}

// DoCheckIn 执行打卡
func (s *checkInService) DoCheckIn(studyDurationSec int) (*domain.CheckIn, error) {
	today := time.Now().Format("2006-01-02")

	// 获取学习计划，用于计算今天是第几天
	plan, err := s.planRepo.Get()
	if err != nil {
		return nil, fmt.Errorf("获取学习计划失败: %w", err)
	}
	if plan == nil {
		return nil, errors.New("学习计划不存在，请先创建计划")
	}

	// 计算今日是第几天：今日日期与 StartDate 之差 + 1
	startDate, err := time.Parse("2006-01-02", plan.StartDate)
	if err != nil {
		return nil, fmt.Errorf("解析计划开始日期失败: %w", err)
	}
	todayDate, err := time.Parse("2006-01-02", today)
	if err != nil {
		return nil, fmt.Errorf("解析今日日期失败: %w", err)
	}
	dayDiff := int(todayDate.Truncate(24*time.Hour).Sub(startDate.Truncate(24*time.Hour)).Hours() / 24)
	day := dayDiff + 1

	// 检查今日是否已打卡
	exists, err := s.checkInRepo.Exists(day)
	if err != nil {
		return nil, fmt.Errorf("查询打卡记录失败: %w", err)
	}
	if exists {
		return nil, errors.New("今日已打卡")
	}

	// 计算今日得分
	dailyScore, err := s.CalcDailyScore()
	if err != nil {
		return nil, fmt.Errorf("计算今日得分失败: %w", err)
	}

	// 构造打卡记录
	checkIn := &domain.CheckIn{
		Day:              day,
		Date:             today,
		Completed:        dailyScore.Completed,
		Score:            dailyScore.TotalScore,
		StudyDurationSec: studyDurationSec,
	}

	// 保存打卡记录
	if err := s.checkInRepo.Add(checkIn); err != nil {
		return nil, fmt.Errorf("保存打卡记录失败: %w", err)
	}

	// 更新当天计划状态为已完成
	if err := s.planRepo.UpdateDayStatus(day, domain.PlanStatusCompleted); err != nil {
		return nil, fmt.Errorf("更新计划状态失败: %w", err)
	}

	// 计算明天日期，生成次日复习词并更新到计划中
	tomorrow := todayDate.AddDate(0, 0, 1).Format("2006-01-02")
	reviewWords, err := s.reviewSvc.GetReviewWords(tomorrow)
	if err != nil {
		return nil, fmt.Errorf("获取复习词失败: %w", err)
	}

	// 仅当 day+1 在14天计划范围内时才更新次日复习词
	nextDay := day + 1
	if nextDay <= 14 {
		if err := s.planRepo.UpdateDayReviewWords(nextDay, reviewWords); err != nil {
			return nil, fmt.Errorf("更新次日复习词失败: %w", err)
		}
	}

	return checkIn, nil
}

// GetStats 获取打卡统计汇总
func (s *checkInService) GetStats() (*CheckInStats, error) {
	// 获取所有打卡记录
	checkIns, err := s.checkInRepo.GetAll()
	if err != nil {
		return nil, fmt.Errorf("获取打卡记录失败: %w", err)
	}

	// 统计已完成天数
	completedDays := 0
	for _, ci := range checkIns {
		if ci.Completed {
			completedDays++
		}
	}

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

	// 计算已完成记录的平均得分
	var avgScore float64
	if completedDays > 0 {
		totalScore := 0.0
		for _, ci := range checkIns {
			if ci.Completed {
				totalScore += ci.Score
			}
		}
		avgScore = totalScore / float64(completedDays)
	}

	// 按日期排序，方便前端展示
	sort.Slice(checkIns, func(i, j int) bool {
		return checkIns[i].Date < checkIns[j].Date
	})

	return &CheckInStats{
		CheckIns:      checkIns,
		TotalDays:     len(checkIns),
		CompletedDays: completedDays,
		StreakDays:    streakDays,
		AvgScore:      avgScore,
	}, nil
}
