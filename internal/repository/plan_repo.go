package repository

import "english-learner/internal/domain"

// PlanRepository 学习计划仓储接口
type PlanRepository interface {
	// Get 获取当前学习计划，不存在返回 nil
	Get() (*domain.StudyPlan, error)
	// Save 保存（覆盖）学习计划
	Save(plan *domain.StudyPlan) error
	// UpdateDayStatus 更新某天状态
	UpdateDayStatus(day int, status domain.PlanStatus) error
	// UpdateDayReviewWords 更新某天复习词列表
	UpdateDayReviewWords(day int, wordIDs []string) error
}
