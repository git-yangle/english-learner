package json

import (
	"path/filepath"
	"testing"

	"english-learner/internal/domain"
)

// ─────────────────────────────────────────────
// 测试辅助函数
// ─────────────────────────────────────────────

// newLearningRecord 构建一条测试用学习记录
func newLearningRecord(wordID, date string, correct bool) *domain.LearningRecord {
	return &domain.LearningRecord{
		WordID:   wordID,
		Mode:     domain.StudyModeQuiz,
		Correct:  correct,
		Attempts: 1,
		Date:     date,
	}
}

// newCheckIn 构建一条测试用打卡记录
func newCheckIn(day int, date string, completed bool) *domain.CheckIn {
	return &domain.CheckIn{
		Day:              day,
		Date:             date,
		Completed:        completed,
		Score:            0.9,
		StudyDurationSec: 600,
	}
}

// ─────────────────────────────────────────────
// RecordRepo - GetAll 测试
// ─────────────────────────────────────────────

// TestRecordRepo_GetAll_文件不存在 验证文件不存在时返回空切片而非 nil
func TestRecordRepo_GetAll_文件不存在(t *testing.T) {
	// given: 指向不存在文件的仓储
	filePath := filepath.Join(t.TempDir(), "records.json")
	repo := NewRecordRepo(filePath)

	// when: 调用 GetAll
	records, err := repo.GetAll()

	// then: 无错误，返回空切片（非 nil）
	if err != nil {
		t.Fatalf("文件不存在时 GetAll 应无错误，实际: %v", err)
	}
	if records == nil {
		t.Error("期望返回非 nil 空切片，实际为 nil")
	}
	if len(records) != 0 {
		t.Errorf("期望空切片，实际长度 %d", len(records))
	}
}

// ─────────────────────────────────────────────
// RecordRepo - Add & GetAll 测试
// ─────────────────────────────────────────────

// TestRecordRepo_Add_And_GetAll 验证 Add 后 GetAll 能正确读回所有记录
func TestRecordRepo_Add_And_GetAll(t *testing.T) {
	// given: 空仓储
	filePath := filepath.Join(t.TempDir(), "records.json")
	repo := NewRecordRepo(filePath)

	r1 := newLearningRecord("w001", "2026-03-01", true)
	r2 := newLearningRecord("w002", "2026-03-02", false)

	// when: 依次 Add 两条记录
	if err := repo.Add(r1); err != nil {
		t.Fatalf("第 1 次 Add 失败: %v", err)
	}
	if err := repo.Add(r2); err != nil {
		t.Fatalf("第 2 次 Add 失败: %v", err)
	}

	// then: GetAll 返回 2 条记录且内容正确
	records, err := repo.GetAll()
	if err != nil {
		t.Fatalf("GetAll 失败: %v", err)
	}
	if len(records) != 2 {
		t.Fatalf("期望 2 条记录，实际 %d 条", len(records))
	}
	if records[0].WordID != "w001" {
		t.Errorf("第 1 条记录 WordID 期望 'w001'，实际 '%s'", records[0].WordID)
	}
	if records[1].WordID != "w002" {
		t.Errorf("第 2 条记录 WordID 期望 'w002'，实际 '%s'", records[1].WordID)
	}
}

// TestRecordRepo_Add_追加不覆盖 验证多次 Add 是追加而非覆盖
func TestRecordRepo_Add_追加不覆盖(t *testing.T) {
	// given: 先写入 1 条记录
	filePath := filepath.Join(t.TempDir(), "records.json")
	repo := NewRecordRepo(filePath)
	repo.Add(newLearningRecord("w001", "2026-03-01", true))

	// when: 再追加 1 条
	repo.Add(newLearningRecord("w002", "2026-03-02", false))

	// then: 两条记录都在
	records, _ := repo.GetAll()
	if len(records) != 2 {
		t.Errorf("追加后期望 2 条记录，实际 %d 条", len(records))
	}
}

// ─────────────────────────────────────────────
// RecordRepo - GetByDate 测试
// ─────────────────────────────────────────────

// TestRecordRepo_GetByDate 验证按日期过滤返回正确记录
func TestRecordRepo_GetByDate(t *testing.T) {
	// given: 写入 3 条记录，其中 2 条为 2026-03-01，1 条为 2026-03-02
	filePath := filepath.Join(t.TempDir(), "records.json")
	repo := NewRecordRepo(filePath)
	repo.Add(newLearningRecord("w001", "2026-03-01", true))
	repo.Add(newLearningRecord("w002", "2026-03-01", false))
	repo.Add(newLearningRecord("w003", "2026-03-02", true))

	// when: 查询 2026-03-01 的记录
	result, err := repo.GetByDate("2026-03-01")

	// then: 返回 2 条，均属于 2026-03-01
	if err != nil {
		t.Fatalf("GetByDate 失败: %v", err)
	}
	if len(result) != 2 {
		t.Fatalf("期望 2 条，实际 %d 条", len(result))
	}
	for _, r := range result {
		if r.Date != "2026-03-01" {
			t.Errorf("过滤后记录 Date 应为 '2026-03-01'，实际 '%s'", r.Date)
		}
	}
}

// TestRecordRepo_GetByDate_无匹配 验证日期无匹配时返回空切片
func TestRecordRepo_GetByDate_无匹配(t *testing.T) {
	// given: 写入 1 条 2026-03-01 的记录
	filePath := filepath.Join(t.TempDir(), "records.json")
	repo := NewRecordRepo(filePath)
	repo.Add(newLearningRecord("w001", "2026-03-01", true))

	// when: 查询 2099-01-01（不存在的日期）
	result, err := repo.GetByDate("2099-01-01")

	// then: 返回空切片，无错误
	if err != nil {
		t.Fatalf("GetByDate 失败: %v", err)
	}
	if len(result) != 0 {
		t.Errorf("期望空切片，实际 %d 条", len(result))
	}
}

// ─────────────────────────────────────────────
// RecordRepo - GetByWordID 测试
// ─────────────────────────────────────────────

// TestRecordRepo_GetByWordID 验证按 WordID 过滤返回正确记录
func TestRecordRepo_GetByWordID(t *testing.T) {
	// given: 写入 4 条记录，w001 出现 3 次，w002 出现 1 次
	filePath := filepath.Join(t.TempDir(), "records.json")
	repo := NewRecordRepo(filePath)
	repo.Add(newLearningRecord("w001", "2026-03-01", true))
	repo.Add(newLearningRecord("w002", "2026-03-01", false))
	repo.Add(newLearningRecord("w001", "2026-03-02", false))
	repo.Add(newLearningRecord("w001", "2026-03-03", true))

	// when: 按 w001 过滤
	result, err := repo.GetByWordID("w001")

	// then: 返回 3 条，均为 w001
	if err != nil {
		t.Fatalf("GetByWordID 失败: %v", err)
	}
	if len(result) != 3 {
		t.Fatalf("期望 3 条，实际 %d 条", len(result))
	}
	for _, r := range result {
		if r.WordID != "w001" {
			t.Errorf("过滤后记录 WordID 应为 'w001'，实际 '%s'", r.WordID)
		}
	}
}

// TestRecordRepo_GetByWordID_无匹配 验证 WordID 不存在时返回空切片
func TestRecordRepo_GetByWordID_无匹配(t *testing.T) {
	// given: 写入 1 条 w001 的记录
	filePath := filepath.Join(t.TempDir(), "records.json")
	repo := NewRecordRepo(filePath)
	repo.Add(newLearningRecord("w001", "2026-03-01", true))

	// when: 查询不存在的 WordID
	result, err := repo.GetByWordID("no-such-word")

	// then: 返回空切片，无错误
	if err != nil {
		t.Fatalf("GetByWordID 失败: %v", err)
	}
	if len(result) != 0 {
		t.Errorf("期望空切片，实际 %d 条", len(result))
	}
}

// ─────────────────────────────────────────────
// CheckInRepo - Add & GetAll 测试
// ─────────────────────────────────────────────

// TestCheckInRepo_Add_And_GetAll 验证 Add 后 GetAll 能正确读回打卡记录
func TestCheckInRepo_Add_And_GetAll(t *testing.T) {
	// given: 空打卡仓储
	filePath := filepath.Join(t.TempDir(), "checkins.json")
	repo := NewCheckInRepo(filePath)

	c1 := newCheckIn(1, "2026-03-01", true)
	c2 := newCheckIn(2, "2026-03-02", false)

	// when: 依次添加两条打卡记录
	if err := repo.Add(c1); err != nil {
		t.Fatalf("第 1 次 Add 失败: %v", err)
	}
	if err := repo.Add(c2); err != nil {
		t.Fatalf("第 2 次 Add 失败: %v", err)
	}

	// then: GetAll 返回 2 条，内容正确
	all, err := repo.GetAll()
	if err != nil {
		t.Fatalf("GetAll 失败: %v", err)
	}
	if len(all) != 2 {
		t.Fatalf("期望 2 条打卡记录，实际 %d 条", len(all))
	}
	if all[0].Day != 1 {
		t.Errorf("第 1 条 Day 期望 1，实际 %d", all[0].Day)
	}
	if all[1].Day != 2 {
		t.Errorf("第 2 条 Day 期望 2，实际 %d", all[1].Day)
	}
}

// TestCheckInRepo_GetAll_文件不存在 验证文件不存在时 GetAll 返回空切片非 nil
func TestCheckInRepo_GetAll_文件不存在(t *testing.T) {
	// given: 指向不存在文件的仓储
	filePath := filepath.Join(t.TempDir(), "checkins.json")
	repo := NewCheckInRepo(filePath)

	// when: GetAll
	all, err := repo.GetAll()

	// then: 无错误，返回非 nil 空切片
	if err != nil {
		t.Fatalf("文件不存在时 GetAll 应无错误，实际: %v", err)
	}
	if all == nil {
		t.Error("期望非 nil 空切片，实际为 nil")
	}
	if len(all) != 0 {
		t.Errorf("期望空切片，实际 %d 条", len(all))
	}
}

// ─────────────────────────────────────────────
// CheckInRepo - GetByDay 测试
// ─────────────────────────────────────────────

// TestCheckInRepo_GetByDay_存在 验证查询已存在的打卡记录返回正确数据
func TestCheckInRepo_GetByDay_存在(t *testing.T) {
	// given: 已添加 day 1 的打卡记录
	filePath := filepath.Join(t.TempDir(), "checkins.json")
	repo := NewCheckInRepo(filePath)
	repo.Add(newCheckIn(1, "2026-03-01", true))
	repo.Add(newCheckIn(2, "2026-03-02", false))

	// when: 查询 day 1
	ci, err := repo.GetByDay(1)

	// then: 返回 day 1 的记录，无错误
	if err != nil {
		t.Fatalf("GetByDay 失败: %v", err)
	}
	if ci == nil {
		t.Fatal("期望返回打卡记录，实际为 nil")
	}
	if ci.Day != 1 {
		t.Errorf("期望 Day=1，实际 %d", ci.Day)
	}
	if ci.Date != "2026-03-01" {
		t.Errorf("期望 Date='2026-03-01'，实际 '%s'", ci.Date)
	}
	if !ci.Completed {
		t.Error("期望 Completed=true，实际 false")
	}
}

// TestCheckInRepo_GetByDay_不存在 验证查询不存在的天数返回 nil, nil
func TestCheckInRepo_GetByDay_不存在(t *testing.T) {
	// given: 只有 day 1 的打卡记录
	filePath := filepath.Join(t.TempDir(), "checkins.json")
	repo := NewCheckInRepo(filePath)
	repo.Add(newCheckIn(1, "2026-03-01", true))

	// when: 查询 day 99（不存在）
	ci, err := repo.GetByDay(99)

	// then: 返回 nil, nil
	if err != nil {
		t.Fatalf("GetByDay 不存在时应无错误，实际: %v", err)
	}
	if ci != nil {
		t.Errorf("期望 nil，实际 %v", ci)
	}
}

// TestCheckInRepo_GetByDay_空仓储 验证空仓储中查询任何天都返回 nil, nil
func TestCheckInRepo_GetByDay_空仓储(t *testing.T) {
	// given: 没有任何记录
	filePath := filepath.Join(t.TempDir(), "checkins.json")
	repo := NewCheckInRepo(filePath)

	// when: 查询 day 1
	ci, err := repo.GetByDay(1)

	// then: nil, nil
	if err != nil {
		t.Fatalf("空仓储 GetByDay 应无错误，实际: %v", err)
	}
	if ci != nil {
		t.Errorf("空仓储 GetByDay 期望 nil，实际 %v", ci)
	}
}

// ─────────────────────────────────────────────
// CheckInRepo - Exists 测试
// ─────────────────────────────────────────────

// TestCheckInRepo_Exists 验证 Exists 在存在和不存在两种情况下的返回值
func TestCheckInRepo_Exists(t *testing.T) {
	// given: 已添加 day 3 的打卡记录
	filePath := filepath.Join(t.TempDir(), "checkins.json")
	repo := NewCheckInRepo(filePath)
	repo.Add(newCheckIn(3, "2026-03-03", true))

	// when/then: day 3 存在
	exists, err := repo.Exists(3)
	if err != nil {
		t.Fatalf("Exists(3) 失败: %v", err)
	}
	if !exists {
		t.Error("day 3 已添加，Exists 应返回 true")
	}

	// when/then: day 5 不存在
	exists, err = repo.Exists(5)
	if err != nil {
		t.Fatalf("Exists(5) 失败: %v", err)
	}
	if exists {
		t.Error("day 5 未添加，Exists 应返回 false")
	}
}

// TestCheckInRepo_Exists_空仓储 验证空仓储中 Exists 始终返回 false
func TestCheckInRepo_Exists_空仓储(t *testing.T) {
	// given: 无任何打卡记录
	filePath := filepath.Join(t.TempDir(), "checkins.json")
	repo := NewCheckInRepo(filePath)

	// when: 查询 day 1
	exists, err := repo.Exists(1)

	// then: false, nil
	if err != nil {
		t.Fatalf("空仓储 Exists 应无错误，实际: %v", err)
	}
	if exists {
		t.Error("空仓储 Exists 应返回 false")
	}
}

// TestCheckInRepo_Exists_边界_day为零 验证 day=0 的边界情况
func TestCheckInRepo_Exists_边界_day为零(t *testing.T) {
	// given: 无任何打卡记录
	filePath := filepath.Join(t.TempDir(), "checkins.json")
	repo := NewCheckInRepo(filePath)

	// when: 查询 day=0
	exists, err := repo.Exists(0)

	// then: false, nil（不报错）
	if err != nil {
		t.Fatalf("Exists(0) 应无错误，实际: %v", err)
	}
	if exists {
		t.Error("不存在 day=0 的记录，Exists 应返回 false")
	}
}
