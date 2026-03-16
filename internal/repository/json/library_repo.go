package json

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"sync"

	"english-learner/internal/domain"
	"english-learner/internal/repository"
)

// libraryRepo 词库仓储 JSON 文件实现
// 从 data/libraries/*.json 目录扫描加载词库到内存
type libraryRepo struct {
	// dataDir 词库JSON文件所在目录
	dataDir string
	// mu 读写锁，保护并发访问内存缓存
	mu sync.RWMutex
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

// GetAll 获取所有词库，按 ID 升序排序后返回
func (r *libraryRepo) GetAll() ([]*domain.WordLibrary, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	result := make([]*domain.WordLibrary, 0, len(r.libraries))
	for _, lib := range r.libraries {
		result = append(result, lib)
	}

	// 按 ID 升序排序，保证输出顺序稳定
	sort.Slice(result, func(i, j int) bool {
		return result[i].ID < result[j].ID
	})

	return result, nil
}

// GetByID 根据 ID 从内存缓存中查找词库，不存在则返回错误
func (r *libraryRepo) GetByID(id string) (*domain.WordLibrary, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	lib, ok := r.libraries[id]
	if !ok {
		return nil, fmt.Errorf("词库不存在: %s", id)
	}
	return lib, nil
}

// Reload 扫描 dataDir 下所有 .json 文件，重新加载词库到内存 map
// 每个文件解析为一个 WordLibrary，同时为每个 Word 赋值 LibraryID 字段
func (r *libraryRepo) Reload() error {
	// 扫描目录下所有 .json 文件
	entries, err := os.ReadDir(r.dataDir)
	if err != nil {
		// 目录不存在时视为无词库，不报错
		if os.IsNotExist(err) {
			return nil
		}
		return fmt.Errorf("扫描词库目录失败: %w", err)
	}

	newLibraries := make(map[string]*domain.WordLibrary)

	for _, entry := range entries {
		// 只处理 .json 文件，跳过子目录及其他格式
		if entry.IsDir() || !strings.HasSuffix(entry.Name(), ".json") {
			continue
		}

		filePath := filepath.Join(r.dataDir, entry.Name())
		data, err := os.ReadFile(filePath)
		if err != nil {
			return fmt.Errorf("读取词库文件 %s 失败: %w", filePath, err)
		}

		var lib domain.WordLibrary
		if err := json.Unmarshal(data, &lib); err != nil {
			return fmt.Errorf("解析词库文件 %s 失败: %w", filePath, err)
		}

		// 为每个单词赋值所属词库 ID，方便上层使用
		for i := range lib.Words {
			lib.Words[i].LibraryID = lib.ID
		}

		newLibraries[lib.ID] = &lib
	}

	// 加载成功后统一替换内存缓存（加写锁）
	r.mu.Lock()
	r.libraries = newLibraries
	r.mu.Unlock()

	return nil
}
