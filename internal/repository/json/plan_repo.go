package json

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"

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

// Get 获取当前学习计划，文件不存在返回 nil, nil
func (r *planRepo) Get() (*domain.StudyPlan, error) {
	return r.read()
}

// Save 将学习计划序列化（缩进格式）写入文件（覆盖）
func (r *planRepo) Save(plan *domain.StudyPlan) error {
	return r.write(plan)
}

// UpdateDayStatus 更新指定天的计划状态
// 读取计划 → 找到对应 day → 更新 Status → 写回文件
func (r *planRepo) UpdateDayStatus(day int, status domain.PlanStatus) error {
	plan, err := r.read()
	if err != nil {
		return err
	}
	if plan == nil {
		return fmt.Errorf("学习计划不存在，无法更新第 %d 天状态", day)
	}

	// 遍历查找目标天
	found := false
	for i := range plan.Days {
		if plan.Days[i].Day == day {
			plan.Days[i].Status = status
			found = true
			break
		}
	}
	if !found {
		return fmt.Errorf("计划中不存在第 %d 天", day)
	}

	return r.write(plan)
}

// UpdateDayReviewWords 更新指定天的复习词列表
// 读取计划 → 找到对应 day → 更新 ReviewWordIDs → 写回文件
func (r *planRepo) UpdateDayReviewWords(day int, wordIDs []string) error {
	plan, err := r.read()
	if err != nil {
		return err
	}
	if plan == nil {
		return fmt.Errorf("学习计划不存在，无法更新第 %d 天复习词", day)
	}

	// 遍历查找目标天
	found := false
	for i := range plan.Days {
		if plan.Days[i].Day == day {
			plan.Days[i].ReviewWordIDs = wordIDs
			found = true
			break
		}
	}
	if !found {
		return fmt.Errorf("计划中不存在第 %d 天", day)
	}

	return r.write(plan)
}

// read 从文件读取并反序列化学习计划
// 文件不存在返回 nil, nil；其他读取或解析错误返回 error
func (r *planRepo) read() (*domain.StudyPlan, error) {
	data, err := os.ReadFile(r.filePath)
	if err != nil {
		if os.IsNotExist(err) {
			// 文件不存在，说明还没有创建计划
			return nil, nil
		}
		return nil, fmt.Errorf("读取计划文件失败: %w", err)
	}

	var plan domain.StudyPlan
	if err := json.Unmarshal(data, &plan); err != nil {
		return nil, fmt.Errorf("解析计划文件失败: %w", err)
	}
	return &plan, nil
}

// write 将学习计划序列化为缩进 JSON 并写入文件
// 写入前确保目录存在
func (r *planRepo) write(plan *domain.StudyPlan) error {
	// 确保父目录存在
	dir := filepath.Dir(r.filePath)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return fmt.Errorf("创建计划文件目录失败: %w", err)
	}

	data, err := json.MarshalIndent(plan, "", "  ")
	if err != nil {
		return fmt.Errorf("序列化计划失败: %w", err)
	}

	if err := os.WriteFile(r.filePath, data, 0644); err != nil {
		return fmt.Errorf("写入计划文件失败: %w", err)
	}
	return nil
}
