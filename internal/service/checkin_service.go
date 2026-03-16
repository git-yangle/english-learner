package service

import (
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

// CalcDailyScore 计算今日得分（不执行打卡，仅统计）
func (s *checkInService) CalcDailyScore() (*DailyScore, error) {
	// TODO:
	// 1. 调用 recordRepo.GetByDate(today) 获取今日所有记录
	// 2. 分别统计 quiz 和 dictation 模式的正确率
	// 3. 计算综合得分：quiz*0.4 + dictation*0.6
	// 4. 判断是否达标（totalScore >= 0.6）
	return nil, nil
}

// DoCheckIn 执行打卡
func (s *checkInService) DoCheckIn(studyDurationSec int) (*domain.CheckIn, error) {
	// TODO:
	// 1. 调用 CalcDailyScore() 计算今日得分
	// 2. 构造 CheckIn{Day, Date, Completed, Score, StudyDurationSec}
	// 3. 调用 checkInRepo.Add() 保存打卡记录
	// 4. 调用 planRepo.UpdateDayStatus() 更新当天计划为 completed
	// 5. 调用 reviewSvc.GetReviewWords(tomorrow) 生成次日复习词
	// 6. 调用 planRepo.UpdateDayReviewWords() 更新次日复习词列表
	return nil, nil
}

// GetStats 获取打卡统计汇总
func (s *checkInService) GetStats() (*CheckInStats, error) {
	// TODO:
	// 1. 调用 checkInRepo.GetAll() 获取所有打卡记录
	// 2. 统计总天数、完成天数、当前连续天数、平均得分
	// 3. 返回 CheckInStats
	return nil, nil
}
