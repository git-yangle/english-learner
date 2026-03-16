package json

import (
	"english-learner/internal/domain"
	"english-learner/internal/repository"
)

// planRepo 学习计划仓储 JSON 文件实现
// 读写 data/user/plan.json
type planRepo struct {
	// filePath plan.json 文件路径
	filePath string
}

// NewPlanRepo 创建学习计划仓储实例
func NewPlanRepo(filePath string) repository.PlanRepository {
	return &planRepo{filePath: filePath}
}

// Get 获取当前学习计划，不存在返回 nil
func (r *planRepo) Get() (*domain.StudyPlan, error) {
	// TODO: 读取 filePath JSON 文件并反序列化为 StudyPlan，文件不存在返回 nil, nil
	return nil, nil
}

// Save 保存（覆盖）学习计划
func (r *planRepo) Save(plan *domain.StudyPlan) error {
	// TODO: 将 StudyPlan 序列化并写入 filePath JSON 文件
	return nil
}

// UpdateDayStatus 更新某天计划状态
func (r *planRepo) UpdateDayStatus(day int, status domain.PlanStatus) error {
	// TODO: 读取计划 → 找到对应 day → 更新 Status → 写回文件
	return nil
}

// UpdateDayReviewWords 更新某天复习词列表
func (r *planRepo) UpdateDayReviewWords(day int, wordIDs []string) error {
	// TODO: 读取计划 → 找到对应 day → 更新 ReviewWordIDs → 写回文件
	return nil
}
