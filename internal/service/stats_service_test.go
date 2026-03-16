package service_test

import (
	"errors"
	"testing"
	"time"

	"english-learner/internal/domain"
	repomock "english-learner/internal/repository/mock"
	"english-learner/internal/service"

	"github.com/stretchr/testify/assert"
	"go.uber.org/mock/gomock"
)

// TestStatsService_GetWordMastery_未学习 验证无学习记录时，返回未学习状态
func TestStatsService_GetWordMastery_未学习(t *testing.T) {
	// given
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockCheckInRepo := repomock.NewMockCheckInRepository(ctrl)
	mockRecordRepo := repomock.NewMockRecordRepository(ctrl)
	mockLibRepo := repomock.NewMockLibraryRepository(ctrl)

	svc := service.NewStatsService(mockCheckInRepo, mockRecordRepo, mockLibRepo)

	mockRecordRepo.EXPECT().GetByWordID("w001").Return([]*domain.LearningRecord{}, nil)

	// when
	mastery, err := svc.GetWordMastery("w001")

	// then
	assert.NoError(t, err)
	assert.Equal(t, service.MasteryStatusNotLearned, mastery)
}

// TestStatsService_GetWordMastery_已掌握 验证最近3次全对时，返回已掌握状态
func TestStatsService_GetWordMastery_已掌握(t *testing.T) {
	// given
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockCheckInRepo := repomock.NewMockCheckInRepository(ctrl)
	mockRecordRepo := repomock.NewMockRecordRepository(ctrl)
	mockLibRepo := repomock.NewMockLibraryRepository(ctrl)

	svc := service.NewStatsService(mockCheckInRepo, mockRecordRepo, mockLibRepo)

	// 最近3次全部正确
	records := []*domain.LearningRecord{
		{WordID: "w001", Correct: true, Date: "2026-03-10"},
		{WordID: "w001", Correct: true, Date: "2026-03-12"},
		{WordID: "w001", Correct: true, Date: "2026-03-15"},
	}
	mockRecordRepo.EXPECT().GetByWordID("w001").Return(records, nil)

	// when
	mastery, err := svc.GetWordMastery("w001")

	// then
	assert.NoError(t, err)
	assert.Equal(t, service.MasteryStatusMastered, mastery)
}

// TestStatsService_GetWordMastery_学习中 验证最近3次中有错误时，返回学习中状态
func TestStatsService_GetWordMastery_学习中(t *testing.T) {
	// given
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockCheckInRepo := repomock.NewMockCheckInRepository(ctrl)
	mockRecordRepo := repomock.NewMockRecordRepository(ctrl)
	mockLibRepo := repomock.NewMockLibraryRepository(ctrl)

	svc := service.NewStatsService(mockCheckInRepo, mockRecordRepo, mockLibRepo)

	// 最近3次中有1次错误
	records := []*domain.LearningRecord{
		{WordID: "w001", Correct: true, Date: "2026-03-10"},
		{WordID: "w001", Correct: false, Date: "2026-03-12"}, // 错误
		{WordID: "w001", Correct: true, Date: "2026-03-15"},
	}
	mockRecordRepo.EXPECT().GetByWordID("w001").Return(records, nil)

	// when
	mastery, err := svc.GetWordMastery("w001")

	// then
	assert.NoError(t, err)
	assert.Equal(t, service.MasteryStatusLearning, mastery)
}

// TestStatsService_GetWordMastery_超过3条记录取最近3条 验证超过3条记录时，只取最近3条判断掌握状态
func TestStatsService_GetWordMastery_超过3条记录取最近3条(t *testing.T) {
	// given
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockCheckInRepo := repomock.NewMockCheckInRepository(ctrl)
	mockRecordRepo := repomock.NewMockRecordRepository(ctrl)
	mockLibRepo := repomock.NewMockLibraryRepository(ctrl)

	svc := service.NewStatsService(mockCheckInRepo, mockRecordRepo, mockLibRepo)

	// 5条记录：前2次错误，最近3次全对 → 应返回已掌握
	records := []*domain.LearningRecord{
		{WordID: "w001", Correct: false, Date: "2026-03-01"}, // 早期错误，不在最近3条范围
		{WordID: "w001", Correct: false, Date: "2026-03-03"}, // 早期错误，不在最近3条范围
		{WordID: "w001", Correct: true, Date: "2026-03-10"},
		{WordID: "w001", Correct: true, Date: "2026-03-12"},
		{WordID: "w001", Correct: true, Date: "2026-03-15"},
	}
	mockRecordRepo.EXPECT().GetByWordID("w001").Return(records, nil)

	// when
	mastery, err := svc.GetWordMastery("w001")

	// then
	assert.NoError(t, err)
	// 最近3次全对，应为已掌握
	assert.Equal(t, service.MasteryStatusMastered, mastery)
}

// TestStatsService_GetWordMastery_仓储错误 验证仓储报错时，透传错误
func TestStatsService_GetWordMastery_仓储错误(t *testing.T) {
	// given
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockCheckInRepo := repomock.NewMockCheckInRepository(ctrl)
	mockRecordRepo := repomock.NewMockRecordRepository(ctrl)
	mockLibRepo := repomock.NewMockLibraryRepository(ctrl)

	svc := service.NewStatsService(mockCheckInRepo, mockRecordRepo, mockLibRepo)

	mockRecordRepo.EXPECT().GetByWordID("w001").Return(nil, errors.New("IO错误"))

	// when
	mastery, err := svc.GetWordMastery("w001")

	// then
	assert.Error(t, err)
	assert.Equal(t, service.MasteryStatusNotLearned, mastery)
}

// TestStatsService_GetOverview_无记录 验证无打卡和学习记录时，返回空统计
func TestStatsService_GetOverview_无记录(t *testing.T) {
	// given
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockCheckInRepo := repomock.NewMockCheckInRepository(ctrl)
	mockRecordRepo := repomock.NewMockRecordRepository(ctrl)
	mockLibRepo := repomock.NewMockLibraryRepository(ctrl)

	svc := service.NewStatsService(mockCheckInRepo, mockRecordRepo, mockLibRepo)

	mockCheckInRepo.EXPECT().GetAll().Return([]*domain.CheckIn{}, nil)
	mockRecordRepo.EXPECT().GetAll().Return([]*domain.LearningRecord{}, nil)
	mockLibRepo.EXPECT().GetAll().Return([]*domain.WordLibrary{}, nil)

	// when
	overview, err := svc.GetOverview()

	// then
	assert.NoError(t, err)
	assert.NotNil(t, overview)
	assert.Equal(t, 0, overview.TotalWords)
	assert.Equal(t, 0, overview.MasteredWords)
	assert.Equal(t, 0, overview.StreakDays)
	assert.Empty(t, overview.CategoryProgress)
}

// TestStatsService_GetOverview_有单词和打卡 验证有单词和打卡记录时，统计数据正确
func TestStatsService_GetOverview_有单词和打卡(t *testing.T) {
	// given
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockCheckInRepo := repomock.NewMockCheckInRepository(ctrl)
	mockRecordRepo := repomock.NewMockRecordRepository(ctrl)
	mockLibRepo := repomock.NewMockLibraryRepository(ctrl)

	svc := service.NewStatsService(mockCheckInRepo, mockRecordRepo, mockLibRepo)

	today := time.Now().Format("2006-01-02")

	// 今天的打卡记录
	checkIns := []*domain.CheckIn{
		{Day: 1, Date: today, Completed: true, Score: 0.9},
	}

	// w001 有3次正确记录（已掌握），w002 无记录（未学习）
	allRecords := []*domain.LearningRecord{
		{WordID: "w001", Correct: true, Date: "2026-03-10"},
		{WordID: "w001", Correct: true, Date: "2026-03-12"},
		{WordID: "w001", Correct: true, Date: "2026-03-15"},
	}

	library := &domain.WordLibrary{
		ID:   "lib-001",
		Name: "旅行英语",
		Words: []domain.Word{
			{ID: "w001", English: "hotel", Chinese: "酒店", Category: "hotel", LibraryID: "lib-001"},
			{ID: "w002", English: "airport", Chinese: "机场", Category: "airport", LibraryID: "lib-001"},
		},
	}

	mockCheckInRepo.EXPECT().GetAll().Return(checkIns, nil)
	mockRecordRepo.EXPECT().GetAll().Return(allRecords, nil)
	mockLibRepo.EXPECT().GetAll().Return([]*domain.WordLibrary{library}, nil)
	// GetWordMastery 内部会调用 GetByWordID
	mockRecordRepo.EXPECT().GetByWordID("w001").Return(allRecords, nil)
	mockRecordRepo.EXPECT().GetByWordID("w002").Return([]*domain.LearningRecord{}, nil)

	// when
	overview, err := svc.GetOverview()

	// then
	assert.NoError(t, err)
	assert.NotNil(t, overview)
	// 2个单词，1个已掌握
	assert.Equal(t, 2, overview.TotalWords)
	assert.Equal(t, 1, overview.MasteredWords)
	// 今天有打卡，连续天数为1
	assert.Equal(t, 1, overview.StreakDays)
	// 14天得分趋势不为空
	assert.Len(t, overview.ScoreTrend, 14)
	// 分类进度包含2个分类
	assert.Len(t, overview.CategoryProgress, 2)
}

// TestStatsService_GetOverview_打卡仓储错误 验证打卡仓储报错时，GetOverview 返回错误
func TestStatsService_GetOverview_打卡仓储错误(t *testing.T) {
	// given
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockCheckInRepo := repomock.NewMockCheckInRepository(ctrl)
	mockRecordRepo := repomock.NewMockRecordRepository(ctrl)
	mockLibRepo := repomock.NewMockLibraryRepository(ctrl)

	svc := service.NewStatsService(mockCheckInRepo, mockRecordRepo, mockLibRepo)

	mockCheckInRepo.EXPECT().GetAll().Return(nil, errors.New("打卡数据库错误"))

	// when
	overview, err := svc.GetOverview()

	// then
	assert.Error(t, err)
	assert.Nil(t, overview)
	assert.Contains(t, err.Error(), "打卡数据库错误")
}

// TestStatsService_GetOverview_分类掌握率 验证分类掌握率计算正确
func TestStatsService_GetOverview_分类掌握率(t *testing.T) {
	// given
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockCheckInRepo := repomock.NewMockCheckInRepository(ctrl)
	mockRecordRepo := repomock.NewMockRecordRepository(ctrl)
	mockLibRepo := repomock.NewMockLibraryRepository(ctrl)

	svc := service.NewStatsService(mockCheckInRepo, mockRecordRepo, mockLibRepo)

	// hotel 分类有 2 个单词，都已掌握（3次全对）
	w1Records := []*domain.LearningRecord{
		{WordID: "w001", Correct: true, Date: "2026-03-10"},
		{WordID: "w001", Correct: true, Date: "2026-03-12"},
		{WordID: "w001", Correct: true, Date: "2026-03-15"},
	}
	w2Records := []*domain.LearningRecord{
		{WordID: "w002", Correct: true, Date: "2026-03-10"},
		{WordID: "w002", Correct: true, Date: "2026-03-12"},
		{WordID: "w002", Correct: true, Date: "2026-03-15"},
	}

	library := &domain.WordLibrary{
		ID:   "lib-001",
		Name: "旅行英语",
		Words: []domain.Word{
			{ID: "w001", English: "hotel", Chinese: "酒店", Category: "hotel"},
			{ID: "w002", English: "motel", Chinese: "汽车旅馆", Category: "hotel"},
		},
	}

	allRecords := append(w1Records, w2Records...)

	mockCheckInRepo.EXPECT().GetAll().Return([]*domain.CheckIn{}, nil)
	mockRecordRepo.EXPECT().GetAll().Return(allRecords, nil)
	mockLibRepo.EXPECT().GetAll().Return([]*domain.WordLibrary{library}, nil)
	mockRecordRepo.EXPECT().GetByWordID("w001").Return(w1Records, nil)
	mockRecordRepo.EXPECT().GetByWordID("w002").Return(w2Records, nil)

	// when
	overview, err := svc.GetOverview()

	// then
	assert.NoError(t, err)
	assert.Len(t, overview.CategoryProgress, 1)
	cp := overview.CategoryProgress[0]
	assert.Equal(t, "hotel", cp.Category)
	assert.Equal(t, 2, cp.Total)
	assert.Equal(t, 2, cp.Mastered)
	assert.Equal(t, 0, cp.Learning)
	assert.Equal(t, 0, cp.NotLearned)
	// 掌握率 = 2/2 = 1.0
	assert.InDelta(t, 1.0, cp.MasteryRate, 1e-9)
}
