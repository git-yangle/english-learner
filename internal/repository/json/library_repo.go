package json

import (
	"english-learner/internal/domain"
	"english-learner/internal/repository"
)

// libraryRepo 词库仓储 JSON 文件实现
// 从 data/libraries/*.json 目录扫描加载词库到内存
type libraryRepo struct {
	// dataDir 词库JSON文件所在目录
	dataDir string
	// libraries 词库内存缓存，key 为 libraryID
	libraries map[string]*domain.WordLibrary
}

// NewLibraryRepo 创建词库仓储实例
func NewLibraryRepo(dataDir string) repository.LibraryRepository {
	return &libraryRepo{
		dataDir:   dataDir,
		libraries: make(map[string]*domain.WordLibrary),
	}
}

// GetAll 获取所有词库
func (r *libraryRepo) GetAll() ([]*domain.WordLibrary, error) {
	// TODO: 从内存缓存返回所有词库列表
	return nil, nil
}

// GetByID 根据ID获取词库
func (r *libraryRepo) GetByID(id string) (*domain.WordLibrary, error) {
	// TODO: 从内存缓存按 ID 查找词库
	return nil, nil
}

// Reload 重新扫描目录加载词库（不重启服务）
func (r *libraryRepo) Reload() error {
	// TODO: 扫描 dataDir 下所有 .json 文件，解析并加载到内存 map
	return nil
}
