package service_test

import (
	"errors"
	"testing"

	"english-learner/internal/domain"
	repomock "english-learner/internal/repository/mock"
	"english-learner/internal/service"

	"github.com/stretchr/testify/assert"
	"go.uber.org/mock/gomock"
)

// TestLibraryService_ListLibraries_成功 验证正常获取词库列表，返回按ID升序排序的摘要
func TestLibraryService_ListLibraries_成功(t *testing.T) {
	// given
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockRepo := repomock.NewMockLibraryRepository(ctrl)
	svc := service.NewLibraryService(mockRepo)

	// 返回两个词库，ID 顺序故意倒置，验证服务层会正确排序
	libraries := []*domain.WordLibrary{
		{
			ID:   "lib-002",
			Name: "旅行英语B",
			Desc: "进阶词汇",
			Words: []domain.Word{
				{ID: "w001", English: "hotel", Chinese: "酒店"},
				{ID: "w002", English: "airport", Chinese: "机场"},
				{ID: "w003", English: "taxi", Chinese: "出租车"},
			},
		},
		{
			ID:   "lib-001",
			Name: "旅行英语A",
			Desc: "基础词汇",
			Words: []domain.Word{
				{ID: "w004", English: "hello", Chinese: "你好"},
				{ID: "w005", English: "thank you", Chinese: "谢谢"},
			},
		},
	}
	mockRepo.EXPECT().GetAll().Return(libraries, nil)

	// when
	infos, err := svc.ListLibraries()

	// then
	assert.NoError(t, err)
	assert.Len(t, infos, 2)
	// 验证按 ID 升序排列：lib-001 在前
	assert.Equal(t, "lib-001", infos[0].ID)
	assert.Equal(t, "旅行英语A", infos[0].Name)
	assert.Equal(t, 2, infos[0].WordCount)
	assert.Equal(t, "lib-002", infos[1].ID)
	assert.Equal(t, 3, infos[1].WordCount)
}

// TestLibraryService_ListLibraries_仓储错误 验证仓储层返回错误时，服务层透传错误
func TestLibraryService_ListLibraries_仓储错误(t *testing.T) {
	// given
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockRepo := repomock.NewMockLibraryRepository(ctrl)
	svc := service.NewLibraryService(mockRepo)

	mockRepo.EXPECT().GetAll().Return(nil, errors.New("数据库连接失败"))

	// when
	infos, err := svc.ListLibraries()

	// then
	assert.Error(t, err)
	assert.Nil(t, infos)
	// 验证错误信息中包含原始错误
	assert.Contains(t, err.Error(), "数据库连接失败")
}

// TestLibraryService_GetWord_存在 验证正常获取词库中存在的单词
func TestLibraryService_GetWord_存在(t *testing.T) {
	// given
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockRepo := repomock.NewMockLibraryRepository(ctrl)
	svc := service.NewLibraryService(mockRepo)

	library := &domain.WordLibrary{
		ID:   "lib-001",
		Name: "旅行英语",
		Words: []domain.Word{
			{ID: "w001", English: "hotel", Chinese: "酒店"},
			{ID: "w002", English: "airport", Chinese: "机场"},
		},
	}
	mockRepo.EXPECT().GetByID("lib-001").Return(library, nil)

	// when
	word, err := svc.GetWord("lib-001", "w001")

	// then
	assert.NoError(t, err)
	assert.NotNil(t, word)
	assert.Equal(t, "w001", word.ID)
	assert.Equal(t, "hotel", word.English)
	assert.Equal(t, "酒店", word.Chinese)
}

// TestLibraryService_GetWord_单词不存在 验证词库存在但目标 wordID 不在其中时返回错误
func TestLibraryService_GetWord_单词不存在(t *testing.T) {
	// given
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockRepo := repomock.NewMockLibraryRepository(ctrl)
	svc := service.NewLibraryService(mockRepo)

	library := &domain.WordLibrary{
		ID:   "lib-001",
		Name: "旅行英语",
		Words: []domain.Word{
			{ID: "w001", English: "hotel", Chinese: "酒店"},
		},
	}
	mockRepo.EXPECT().GetByID("lib-001").Return(library, nil)

	// when
	word, err := svc.GetWord("lib-001", "w999")

	// then
	assert.Error(t, err)
	assert.Nil(t, word)
	assert.Contains(t, err.Error(), "w999")
}

// TestLibraryService_GetWord_词库不存在 验证词库不存在时，GetWord 返回错误
func TestLibraryService_GetWord_词库不存在(t *testing.T) {
	// given
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockRepo := repomock.NewMockLibraryRepository(ctrl)
	svc := service.NewLibraryService(mockRepo)

	mockRepo.EXPECT().GetByID("lib-999").Return(nil, errors.New("词库不存在"))

	// when
	word, err := svc.GetWord("lib-999", "w001")

	// then
	assert.Error(t, err)
	assert.Nil(t, word)
	assert.Contains(t, err.Error(), "lib-999")
}
