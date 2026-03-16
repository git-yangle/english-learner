package json

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"

	"english-learner/internal/domain"
)

// ─────────────────────────────────────────────
// 测试辅助函数
// ─────────────────────────────────────────────

// createTestLibraryDir 在临时目录中写入一个测试词库 JSON 文件，返回目录路径
func createTestLibraryDir(t *testing.T) string {
	t.Helper()
	dir := t.TempDir()

	lib := domain.WordLibrary{
		ID:   "test",
		Name: "测试词库",
		Desc: "单元测试用词库",
		Words: []domain.Word{
			{ID: "w001", English: "hello", Chinese: "你好", Category: "airport"},
			{ID: "w002", English: "bye", Chinese: "再见", Category: "hotel"},
		},
	}
	data, err := json.Marshal(lib)
	if err != nil {
		t.Fatalf("序列化测试词库失败: %v", err)
	}
	if err := os.WriteFile(filepath.Join(dir, "test.json"), data, 0644); err != nil {
		t.Fatalf("写入测试词库文件失败: %v", err)
	}
	return dir
}

// ─────────────────────────────────────────────
// Reload 测试
// ─────────────────────────────────────────────

// TestLibraryRepo_Reload_成功加载词库 验证 Reload 能正确解析词库文件并回填 LibraryID
func TestLibraryRepo_Reload_成功加载词库(t *testing.T) {
	// given: 临时目录中有一个合法词库 JSON
	dir := createTestLibraryDir(t)
	repo := NewLibraryRepo(dir)

	// when: 执行 Reload
	err := repo.Reload()

	// then: 无错误，且内存中有一个词库
	if err != nil {
		t.Fatalf("Reload 应成功，得到错误: %v", err)
	}

	libs, err := repo.GetAll()
	if err != nil {
		t.Fatalf("GetAll 失败: %v", err)
	}
	if len(libs) != 1 {
		t.Fatalf("期望 1 个词库，实际 %d 个", len(libs))
	}

	lib := libs[0]
	if lib.ID != "test" {
		t.Errorf("词库 ID 期望 'test'，实际 '%s'", lib.ID)
	}
	if len(lib.Words) != 2 {
		t.Fatalf("期望 2 个单词，实际 %d 个", len(lib.Words))
	}

	// 验证 LibraryID 已被自动回填
	for _, w := range lib.Words {
		if w.LibraryID != "test" {
			t.Errorf("单词 %s 的 LibraryID 期望 'test'，实际 '%s'", w.ID, w.LibraryID)
		}
	}
}

// TestLibraryRepo_Reload_目录不存在 验证传入不存在的目录时 Reload 返回 nil（无报错）
func TestLibraryRepo_Reload_目录不存在(t *testing.T) {
	// given: 一个不存在的目录路径
	repo := NewLibraryRepo("/tmp/no_such_dir_for_test_12345")

	// when: 执行 Reload
	err := repo.Reload()

	// then: 应返回 nil，不报错
	if err != nil {
		t.Errorf("目录不存在时 Reload 应返回 nil，实际: %v", err)
	}
}

// TestLibraryRepo_Reload_非JSON文件被跳过 验证目录中的非 .json 文件不会被解析
func TestLibraryRepo_Reload_非JSON文件被跳过(t *testing.T) {
	// given: 临时目录中有一个 .json 和一个 .txt 文件
	dir := createTestLibraryDir(t)
	// 写入一个 .txt 文件（内容非 JSON）
	if err := os.WriteFile(filepath.Join(dir, "ignore.txt"), []byte("not json"), 0644); err != nil {
		t.Fatalf("写入辅助文件失败: %v", err)
	}
	repo := NewLibraryRepo(dir)

	// when: Reload
	err := repo.Reload()

	// then: 只加载了 1 个词库，.txt 被忽略，无错误
	if err != nil {
		t.Fatalf("Reload 应成功，得到: %v", err)
	}
	libs, _ := repo.GetAll()
	if len(libs) != 1 {
		t.Errorf("期望跳过非 JSON 文件，只加载 1 个词库，实际 %d 个", len(libs))
	}
}

// TestLibraryRepo_Reload_JSON解析失败 验证词库文件内容非法时 Reload 返回错误
func TestLibraryRepo_Reload_JSON解析失败(t *testing.T) {
	// given: 临时目录中有一个内容损坏的 .json 文件
	dir := t.TempDir()
	if err := os.WriteFile(filepath.Join(dir, "bad.json"), []byte("not valid json"), 0644); err != nil {
		t.Fatalf("写入损坏文件失败: %v", err)
	}
	repo := NewLibraryRepo(dir)

	// when: Reload
	err := repo.Reload()

	// then: 应返回解析错误
	if err == nil {
		t.Error("JSON 解析失败时 Reload 应返回错误，实际返回 nil")
	}
}

// ─────────────────────────────────────────────
// GetAll 测试
// ─────────────────────────────────────────────

// TestLibraryRepo_GetAll_有数据 验证 Reload 后 GetAll 返回词库列表
func TestLibraryRepo_GetAll_有数据(t *testing.T) {
	// given: 已经 Reload 的仓储
	dir := createTestLibraryDir(t)
	repo := NewLibraryRepo(dir)
	if err := repo.Reload(); err != nil {
		t.Fatalf("前置 Reload 失败: %v", err)
	}

	// when: 调用 GetAll
	libs, err := repo.GetAll()

	// then: 返回非空列表，无错误
	if err != nil {
		t.Fatalf("GetAll 失败: %v", err)
	}
	if len(libs) == 0 {
		t.Error("GetAll 期望返回至少一个词库，实际为空")
	}
}

// TestLibraryRepo_GetAll_空词库 验证未调用 Reload 时 GetAll 返回空列表
func TestLibraryRepo_GetAll_空词库(t *testing.T) {
	// given: 新建仓储，未执行 Reload
	repo := NewLibraryRepo(t.TempDir())

	// when: 调用 GetAll
	libs, err := repo.GetAll()

	// then: 返回空列表，无错误
	if err != nil {
		t.Fatalf("GetAll 失败: %v", err)
	}
	if len(libs) != 0 {
		t.Errorf("未 Reload 时 GetAll 应返回空列表，实际 %d 个", len(libs))
	}
}

// TestLibraryRepo_GetAll_排序稳定 验证多词库时按 ID 升序返回
func TestLibraryRepo_GetAll_排序稳定(t *testing.T) {
	// given: 目录中有两个词库，ID 分别为 "b-lib" 和 "a-lib"
	dir := t.TempDir()
	for _, id := range []string{"b-lib", "a-lib"} {
		lib := domain.WordLibrary{ID: id, Name: id}
		data, _ := json.Marshal(lib)
		os.WriteFile(filepath.Join(dir, id+".json"), data, 0644)
	}
	repo := NewLibraryRepo(dir)
	repo.Reload()

	// when: GetAll
	libs, err := repo.GetAll()

	// then: 返回顺序为 a-lib, b-lib
	if err != nil {
		t.Fatalf("GetAll 失败: %v", err)
	}
	if len(libs) != 2 {
		t.Fatalf("期望 2 个词库，实际 %d 个", len(libs))
	}
	if libs[0].ID != "a-lib" || libs[1].ID != "b-lib" {
		t.Errorf("排序不正确，期望 [a-lib, b-lib]，实际 [%s, %s]", libs[0].ID, libs[1].ID)
	}
}

// ─────────────────────────────────────────────
// GetByID 测试
// ─────────────────────────────────────────────

// TestLibraryRepo_GetByID_存在 验证 Reload 后 GetByID 返回正确词库
func TestLibraryRepo_GetByID_存在(t *testing.T) {
	// given: 已 Reload 的仓储
	dir := createTestLibraryDir(t)
	repo := NewLibraryRepo(dir)
	if err := repo.Reload(); err != nil {
		t.Fatalf("前置 Reload 失败: %v", err)
	}

	// when: 查询存在的词库 ID
	lib, err := repo.GetByID("test")

	// then: 返回正确词库，无错误
	if err != nil {
		t.Fatalf("GetByID 失败: %v", err)
	}
	if lib == nil {
		t.Fatal("期望返回词库，实际为 nil")
	}
	if lib.ID != "test" {
		t.Errorf("词库 ID 期望 'test'，实际 '%s'", lib.ID)
	}
	if lib.Name != "测试词库" {
		t.Errorf("词库名称期望 '测试词库'，实际 '%s'", lib.Name)
	}
}

// TestLibraryRepo_GetByID_不存在 验证查询不存在 ID 时返回错误
func TestLibraryRepo_GetByID_不存在(t *testing.T) {
	// given: 已 Reload 但只有 "test" 词库
	dir := createTestLibraryDir(t)
	repo := NewLibraryRepo(dir)
	repo.Reload()

	// when: 查询不存在的 ID
	lib, err := repo.GetByID("no-such-id")

	// then: 返回 nil 和非 nil 错误
	if err == nil {
		t.Error("期望返回错误，实际 err == nil")
	}
	if lib != nil {
		t.Errorf("期望 lib == nil，实际 %v", lib)
	}
}

// TestLibraryRepo_GetByID_空ID 验证空字符串 ID 也返回错误
func TestLibraryRepo_GetByID_空ID(t *testing.T) {
	// given: 已 Reload 的仓储
	dir := createTestLibraryDir(t)
	repo := NewLibraryRepo(dir)
	repo.Reload()

	// when: 传入空 ID
	lib, err := repo.GetByID("")

	// then: 返回错误
	if err == nil {
		t.Error("空 ID 时期望返回错误，实际 err == nil")
	}
	if lib != nil {
		t.Errorf("期望 lib == nil，实际 %v", lib)
	}
}
