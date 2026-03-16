package repository

import "english-learner/internal/domain"

// LibraryRepository 词库仓储接口
type LibraryRepository interface {
	// GetAll 获取所有词库
	GetAll() ([]*domain.WordLibrary, error)
	// GetByID 根据ID获取词库
	GetByID(id string) (*domain.WordLibrary, error)
	// Reload 重新加载词库（不重启服务）
	Reload() error
}
