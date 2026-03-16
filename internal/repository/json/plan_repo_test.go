package json

import (
	"path/filepath"
	"testing"

	"english-learner/internal/domain"
)

// ─────────────────────────────────────────────
// 测试辅助函数
// ─────────────────────────────────────────────

// buildTestPlan 构建一个包含 3 天计划的测试用 StudyPlan
func buildTestPlan() *domain.StudyPlan {
	return &domain.StudyPlan{
		LibraryID: "travel",
		StartDate: "2026-03-01",
		Days: []domain.DailyPlan{
			{
				Day:           1,
				Date:          "2026-03-01",
				WordIDs:       []string{"w001", "w002"},
				ReviewWordIDs: []string{},
				Status:        domain.PlanStatusPending,
			},
			{
				Day:           2,
				Date:          "2026-03-02",
				WordIDs:       []string{"w003", "w004"},
				ReviewWordIDs: []string{},
				Status:        domain.PlanStatusPending,
			},
			{
				Day:           3,
				Date:          "2026-03-03",
				WordIDs:       []string{"w005"},
				ReviewWordIDs: []string{},
				Status:        domain.PlanStatusPending,
			},
		},
	}
}

// ─────────────────────────────────────────────
// Get 测试
// ─────────────────────────────────────────────

// TestPlanRepo_Get_文件不存在 验证计划文件不存在时返回 nil, nil
func TestPlanRepo_Get_文件不存在(t *testing.T) {
	// given: 指向不存在文件的仓储
	filePath := filepath.Join(t.TempDir(), "plan.json")
	repo := NewPlanRepo(filePath)

	// when: 调用 Get
	plan, err := repo.Get()

	// then: 应返回 nil, nil，不报错
	if err != nil {
		t.Fatalf("文件不存在时 Get 应返回 nil 错误，实际: %v", err)
	}
	if plan != nil {
		t.Errorf("文件不存在时 Get 应返回 nil plan，实际: %v", plan)
	}
}

// ─────────────────────────────────────────────
// Save & Get 测试
// ─────────────────────────────────────────────

// TestPlanRepo_Save_And_Get 验证 Save 后 Get 能读回完全相同的数据
func TestPlanRepo_Save_And_Get(t *testing.T) {
	// given: 一个合法的 StudyPlan
	filePath := filepath.Join(t.TempDir(), "plan.json")
	repo := NewPlanRepo(filePath)
	original := buildTestPlan()

	// when: Save 后再 Get
	if err := repo.Save(original); err != nil {
		t.Fatalf("Save 失败: %v", err)
	}
	got, err := repo.Get()

	// then: 读回数据与写入数据一致
	if err != nil {
		t.Fatalf("Get 失败: %v", err)
	}
	if got == nil {
		t.Fatal("期望读回 plan，实际为 nil")
	}
	if got.LibraryID != original.LibraryID {
		t.Errorf("LibraryID 不匹配，期望 '%s'，实际 '%s'", original.LibraryID, got.LibraryID)
	}
	if got.StartDate != original.StartDate {
		t.Errorf("StartDate 不匹配，期望 '%s'，实际 '%s'", original.StartDate, got.StartDate)
	}
	if len(got.Days) != len(original.Days) {
		t.Errorf("Days 数量不匹配，期望 %d，实际 %d", len(original.Days), len(got.Days))
	}
}

// TestPlanRepo_Save_覆盖已有计划 验证二次 Save 会覆盖旧数据
func TestPlanRepo_Save_覆盖已有计划(t *testing.T) {
	// given: 先保存一个 LibraryID = "travel" 的计划
	filePath := filepath.Join(t.TempDir(), "plan.json")
	repo := NewPlanRepo(filePath)
	repo.Save(buildTestPlan())

	// when: 再保存一个 LibraryID = "business" 的计划
	newPlan := &domain.StudyPlan{
		LibraryID: "business",
		StartDate: "2026-04-01",
		Days:      []domain.DailyPlan{},
	}
	if err := repo.Save(newPlan); err != nil {
		t.Fatalf("第二次 Save 失败: %v", err)
	}
	got, _ := repo.Get()

	// then: 读回的是新数据
	if got.LibraryID != "business" {
		t.Errorf("期望 LibraryID='business'，实际 '%s'", got.LibraryID)
	}
}

// ─────────────────────────────────────────────
// UpdateDayStatus 测试
// ─────────────────────────────────────────────

// TestPlanRepo_UpdateDayStatus 验证更新指定天状态后读取验证
func TestPlanRepo_UpdateDayStatus(t *testing.T) {
	// given: 已保存包含第 1 天的计划，初始状态为 pending
	filePath := filepath.Join(t.TempDir(), "plan.json")
	repo := NewPlanRepo(filePath)
	repo.Save(buildTestPlan())

	// when: 将第 1 天状态更新为 completed
	err := repo.UpdateDayStatus(1, domain.PlanStatusCompleted)

	// then: 无错误，读回后第 1 天状态为 completed
	if err != nil {
		t.Fatalf("UpdateDayStatus 失败: %v", err)
	}
	plan, _ := repo.Get()
	if plan.Days[0].Status != domain.PlanStatusCompleted {
		t.Errorf("第 1 天状态期望 completed，实际 '%s'", plan.Days[0].Status)
	}
}

// TestPlanRepo_UpdateDayStatus_计划不存在 验证文件不存在时 UpdateDayStatus 返回错误
func TestPlanRepo_UpdateDayStatus_计划不存在(t *testing.T) {
	// given: 指向不存在文件的仓储（未调用 Save）
	filePath := filepath.Join(t.TempDir(), "plan.json")
	repo := NewPlanRepo(filePath)

	// when: 更新不存在的计划
	err := repo.UpdateDayStatus(1, domain.PlanStatusCompleted)

	// then: 应返回错误
	if err == nil {
		t.Error("计划不存在时 UpdateDayStatus 应返回错误，实际 nil")
	}
}

// TestPlanRepo_UpdateDayStatus_天数不存在 验证更新计划中不存在的天数时返回错误
func TestPlanRepo_UpdateDayStatus_天数不存在(t *testing.T) {
	// given: 计划只有 day 1~3
	filePath := filepath.Join(t.TempDir(), "plan.json")
	repo := NewPlanRepo(filePath)
	repo.Save(buildTestPlan())

	// when: 更新 day 99
	err := repo.UpdateDayStatus(99, domain.PlanStatusCompleted)

	// then: 应返回错误
	if err == nil {
		t.Error("天数不存在时 UpdateDayStatus 应返回错误，实际 nil")
	}
}

// TestPlanRepo_UpdateDayStatus_边界_第一天和最后一天 验证边界天数都能正确更新
func TestPlanRepo_UpdateDayStatus_边界_第一天和最后一天(t *testing.T) {
	// given: 包含 day 1/2/3 的计划
	filePath := filepath.Join(t.TempDir(), "plan.json")
	repo := NewPlanRepo(filePath)
	repo.Save(buildTestPlan())

	// when/then: 更新第 1 天
	if err := repo.UpdateDayStatus(1, domain.PlanStatusLearning); err != nil {
		t.Errorf("更新第 1 天失败: %v", err)
	}

	// when/then: 更新最后一天（第 3 天）
	if err := repo.UpdateDayStatus(3, domain.PlanStatusCompleted); err != nil {
		t.Errorf("更新第 3 天失败: %v", err)
	}

	plan, _ := repo.Get()
	if plan.Days[0].Status != domain.PlanStatusLearning {
		t.Errorf("第 1 天状态期望 learning，实际 '%s'", plan.Days[0].Status)
	}
	if plan.Days[2].Status != domain.PlanStatusCompleted {
		t.Errorf("第 3 天状态期望 completed，实际 '%s'", plan.Days[2].Status)
	}
}

// ─────────────────────────────────────────────
// UpdateDayReviewWords 测试
// ─────────────────────────────────────────────

// TestPlanRepo_UpdateDayReviewWords 验证更新复习词后读取验证
func TestPlanRepo_UpdateDayReviewWords(t *testing.T) {
	// given: 已保存计划，第 2 天 ReviewWordIDs 初始为空
	filePath := filepath.Join(t.TempDir(), "plan.json")
	repo := NewPlanRepo(filePath)
	repo.Save(buildTestPlan())

	reviewWords := []string{"w001", "w003", "w005"}

	// when: 更新第 2 天的复习词列表
	err := repo.UpdateDayReviewWords(2, reviewWords)

	// then: 无错误，读回后第 2 天复习词与写入一致
	if err != nil {
		t.Fatalf("UpdateDayReviewWords 失败: %v", err)
	}
	plan, _ := repo.Get()
	day2 := plan.Days[1]
	if len(day2.ReviewWordIDs) != len(reviewWords) {
		t.Fatalf("ReviewWordIDs 数量不匹配，期望 %d，实际 %d", len(reviewWords), len(day2.ReviewWordIDs))
	}
	for i, id := range reviewWords {
		if day2.ReviewWordIDs[i] != id {
			t.Errorf("ReviewWordIDs[%d] 期望 '%s'，实际 '%s'", i, id, day2.ReviewWordIDs[i])
		}
	}
}

// TestPlanRepo_UpdateDayReviewWords_空列表 验证更新为空复习词列表时正常持久化
func TestPlanRepo_UpdateDayReviewWords_空列表(t *testing.T) {
	// given: 先更新复习词，再清空
	filePath := filepath.Join(t.TempDir(), "plan.json")
	repo := NewPlanRepo(filePath)
	repo.Save(buildTestPlan())
	repo.UpdateDayReviewWords(1, []string{"w001", "w002"})

	// when: 将复习词清空
	err := repo.UpdateDayReviewWords(1, []string{})

	// then: 无错误，读回后复习词为空
	if err != nil {
		t.Fatalf("UpdateDayReviewWords 清空失败: %v", err)
	}
	plan, _ := repo.Get()
	if len(plan.Days[0].ReviewWordIDs) != 0 {
		t.Errorf("期望 ReviewWordIDs 为空，实际 %v", plan.Days[0].ReviewWordIDs)
	}
}

// TestPlanRepo_UpdateDayReviewWords_计划不存在 验证计划文件不存在时返回错误
func TestPlanRepo_UpdateDayReviewWords_计划不存在(t *testing.T) {
	// given: 指向不存在文件的仓储
	filePath := filepath.Join(t.TempDir(), "plan.json")
	repo := NewPlanRepo(filePath)

	// when: 尝试更新复习词
	err := repo.UpdateDayReviewWords(1, []string{"w001"})

	// then: 应返回错误
	if err == nil {
		t.Error("计划不存在时 UpdateDayReviewWords 应返回错误，实际 nil")
	}
}

// TestPlanRepo_UpdateDayReviewWords_天数不存在 验证天数不存在时返回错误
func TestPlanRepo_UpdateDayReviewWords_天数不存在(t *testing.T) {
	// given: 计划只有 day 1~3
	filePath := filepath.Join(t.TempDir(), "plan.json")
	repo := NewPlanRepo(filePath)
	repo.Save(buildTestPlan())

	// when: 更新 day 100 的复习词
	err := repo.UpdateDayReviewWords(100, []string{"w001"})

	// then: 应返回错误
	if err == nil {
		t.Error("天数不存在时 UpdateDayReviewWords 应返回错误，实际 nil")
	}
}
