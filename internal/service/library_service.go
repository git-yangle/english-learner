package service

import (
	"fmt"
	"sort"

	"english-learner/internal/domain"
	"english-learner/internal/repository"
)

// LibraryInfo 词库摘要（不含单词详情）
type LibraryInfo struct {
	ID        string `json:"id"`
	Name      string `json:"name"`
	Desc      string `json:"desc"`
	WordCount int    `json:"word_count"`
}

// LibraryService 词库服务接口
type LibraryService interface {
	// ListLibraries 获取所有词库摘要列表
	ListLibraries() ([]*LibraryInfo, error)
	// GetLibrary 获取词库详情（含单词列表）
	GetLibrary(id string) (*domain.WordLibrary, error)
	// GetWord 获取指定词库中的某个单词
	GetWord(libraryID, wordID string) (*domain.Word, error)
	// ReloadLibraries 重新加载词库（不重启服务）
	ReloadLibraries() error
}

type libraryService struct {
	repo repository.LibraryRepository
}

// NewLibraryService 创建词库服务实例
func NewLibraryService(repo repository.LibraryRepository) LibraryService {
	return &libraryService{repo: repo}
}

// ListLibraries 获取所有词库摘要（不含单词详情），按 ID 升序排序
func (s *libraryService) ListLibraries() ([]*LibraryInfo, error) {
	libraries, err := s.repo.GetAll()
	if err != nil {
		return nil, fmt.Errorf("获取词库列表失败: %w", err)
	}

	// 按 ID 升序排序，保证返回顺序稳定
	sort.Slice(libraries, func(i, j int) bool {
		return libraries[i].ID < libraries[j].ID
	})

	// 转换为摘要列表，不暴露单词详情
	infos := make([]*LibraryInfo, 0, len(libraries))
	for _, lib := range libraries {
		infos = append(infos, &LibraryInfo{
			ID:        lib.ID,
			Name:      lib.Name,
			Desc:      lib.Desc,
			WordCount: len(lib.Words),
		})
	}
	return infos, nil
}

// GetLibrary 获取词库详情（含所有单词）
func (s *libraryService) GetLibrary(id string) (*domain.WordLibrary, error) {
	return s.repo.GetByID(id)
}

// GetWord 获取词库中的某个单词，不存在时返回错误
func (s *libraryService) GetWord(libraryID, wordID string) (*domain.Word, error) {
	library, err := s.GetLibrary(libraryID)
	if err != nil {
		return nil, fmt.Errorf("获取词库 %s 失败: %w", libraryID, err)
	}

	// 遍历单词列表查找目标单词
	for i := range library.Words {
		if library.Words[i].ID == wordID {
			return &library.Words[i], nil
		}
	}
	return nil, fmt.Errorf("词库 %s 中不存在单词 %s", libraryID, wordID)
}

// ReloadLibraries 重新加载词库（热更新，不重启服务）
func (s *libraryService) ReloadLibraries() error {
	return s.repo.Reload()
}
