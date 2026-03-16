package service

import (
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

// ListLibraries 获取所有词库摘要列表
func (s *libraryService) ListLibraries() ([]*LibraryInfo, error) {
	// TODO: 调用 repo.GetAll()，将词库列表转换为摘要列表返回
	return nil, nil
}

// GetLibrary 获取词库详情
func (s *libraryService) GetLibrary(id string) (*domain.WordLibrary, error) {
	// TODO: 调用 repo.GetByID(id) 返回词库详情
	return nil, nil
}

// GetWord 获取指定词库中的某个单词
func (s *libraryService) GetWord(libraryID, wordID string) (*domain.Word, error) {
	// TODO: 获取词库后遍历 Words 查找指定 wordID
	return nil, nil
}

// ReloadLibraries 重新加载词库
func (s *libraryService) ReloadLibraries() error {
	// TODO: 调用 repo.Reload() 重新加载词库
	return nil
}
