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

// TestReviewService_CalcWordStats_无记录 验证单词无任何历史记录时，TotalCount==0，ErrorRate==0
func TestReviewService_CalcWordStats_无记录(t *testing.T) {
	// given
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockRecordRepo := repomock.NewMockRecordRepository(ctrl)
	svc := service.NewReviewService(mockRecordRepo)

	mockRecordRepo.EXPECT().GetByWordID("w001").Return([]*domain.LearningRecord{}, nil)

	// when
	stats, err := svc.CalcWordStats("w001")

	// then
	assert.NoError(t, err)
	assert.NotNil(t, stats)
	assert.Equal(t, "w001", stats.WordID)
	assert.Equal(t, 0, stats.TotalCount)
	assert.Equal(t, 0, stats.CorrectCount)
	assert.Equal(t, 0.0, stats.ErrorRate)
	assert.Equal(t, "", stats.LastStudied)
}

// TestReviewService_CalcWordStats_有记录 验证3次记录2对1错时，ErrorRate==1/3
func TestReviewService_CalcWordStats_有记录(t *testing.T) {
	// given
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockRecordRepo := repomock.NewMockRecordRepository(ctrl)
	svc := service.NewReviewService(mockRecordRepo)

	records := []*domain.LearningRecord{
		{WordID: "w001", Mode: domain.StudyModeQuiz, Correct: true, Date: "2026-03-10"},
		{WordID: "w001", Mode: domain.StudyModeQuiz, Correct: true, Date: "2026-03-12"},
		{WordID: "w001", Mode: domain.StudyModeQuiz, Correct: false, Date: "2026-03-15"},
	}
	mockRecordRepo.EXPECT().GetByWordID("w001").Return(records, nil)

	// when
	stats, err := svc.CalcWordStats("w001")

	// then
	assert.NoError(t, err)
	assert.Equal(t, 3, stats.TotalCount)
	assert.Equal(t, 2, stats.CorrectCount)
	// 错误率 = 1 - 2/3 ≈ 0.3333
	assert.InDelta(t, 1.0/3.0, stats.ErrorRate, 1e-9)
	// 最后学习日期为最新记录
	assert.Equal(t, "2026-03-15", stats.LastStudied)
}

// TestReviewService_CalcWordStats_仓储错误 验证仓储层报错时，错误透传
func TestReviewService_CalcWordStats_仓储错误(t *testing.T) {
	// given
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockRecordRepo := repomock.NewMockRecordRepository(ctrl)
	svc := service.NewReviewService(mockRecordRepo)

	mockRecordRepo.EXPECT().GetByWordID("w001").Return(nil, errors.New("IO错误"))

	// when
	stats, err := svc.CalcWordStats("w001")

	// then
	assert.Error(t, err)
	assert.Nil(t, stats)
}

// TestReviewService_GetReviewWords_高错误率 验证错误率0.8的词，上次学习1天前，应加入复习
func TestReviewService_GetReviewWords_高错误率(t *testing.T) {
	// given
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockRecordRepo := repomock.NewMockRecordRepository(ctrl)
	svc := service.NewReviewService(mockRecordRepo)

	// 今日日期
	today := time.Now().Format("2006-01-02")
	// 昨天：1天前
	yesterday := time.Now().AddDate(0, 0, -1).Format("2006-01-02")

	// 5次记录中，1次对4次错 → errorRate = 0.8
	records := []*domain.LearningRecord{
		{WordID: "w001", Correct: true, Date: yesterday},
		{WordID: "w001", Correct: false, Date: yesterday},
		{WordID: "w001", Correct: false, Date: yesterday},
		{WordID: "w001", Correct: false, Date: yesterday},
		{WordID: "w001", Correct: false, Date: yesterday},
	}
	mockRecordRepo.EXPECT().GetAll().Return(records, nil)

	// when
	reviewWords, err := svc.GetReviewWords(today)

	// then
	assert.NoError(t, err)
	// 高错误率(>0.6)，距上次1天 >= 1，应加入复习
	assert.Contains(t, reviewWords, "w001")
}

// TestReviewService_GetReviewWords_低错误率艾宾浩斯 验证错误率0.1的词，距今4天，在艾宾浩斯间隔[1,2,4,7,15]中，应加入
func TestReviewService_GetReviewWords_低错误率艾宾浩斯(t *testing.T) {
	// given
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockRecordRepo := repomock.NewMockRecordRepository(ctrl)
	svc := service.NewReviewService(mockRecordRepo)

	today := time.Now().Format("2006-01-02")
	// 4天前学习
	fourDaysAgo := time.Now().AddDate(0, 0, -4).Format("2006-01-02")

	// 10次记录，9次正确1次错误 → errorRate = 0.1（低错误率，走艾宾浩斯）
	records := []*domain.LearningRecord{
		{WordID: "w002", Correct: true, Date: fourDaysAgo},
		{WordID: "w002", Correct: true, Date: fourDaysAgo},
		{WordID: "w002", Correct: true, Date: fourDaysAgo},
		{WordID: "w002", Correct: true, Date: fourDaysAgo},
		{WordID: "w002", Correct: true, Date: fourDaysAgo},
		{WordID: "w002", Correct: true, Date: fourDaysAgo},
		{WordID: "w002", Correct: true, Date: fourDaysAgo},
		{WordID: "w002", Correct: true, Date: fourDaysAgo},
		{WordID: "w002", Correct: true, Date: fourDaysAgo},
		{WordID: "w002", Correct: false, Date: fourDaysAgo},
	}
	mockRecordRepo.EXPECT().GetAll().Return(records, nil)

	// when
	reviewWords, err := svc.GetReviewWords(today)

	// then
	assert.NoError(t, err)
	// 低错误率(<0.3)，距今4天，4 在艾宾浩斯间隔集合{1,2,4,7,15}中，应加入
	assert.Contains(t, reviewWords, "w002")
}

// TestReviewService_GetReviewWords_未学过 验证无记录的词不加入复习列表
func TestReviewService_GetReviewWords_未学过(t *testing.T) {
	// given
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockRecordRepo := repomock.NewMockRecordRepository(ctrl)
	svc := service.NewReviewService(mockRecordRepo)

	today := time.Now().Format("2006-01-02")

	// 返回空记录集合，模拟从未学过任何单词
	mockRecordRepo.EXPECT().GetAll().Return([]*domain.LearningRecord{}, nil)

	// when
	reviewWords, err := svc.GetReviewWords(today)

	// then
	assert.NoError(t, err)
	assert.Empty(t, reviewWords)
}

// TestReviewService_GetReviewWords_中错误率满2天 验证错误率0.4的词，距今恰好2天，应加入复习
func TestReviewService_GetReviewWords_中错误率满2天(t *testing.T) {
	// given
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockRecordRepo := repomock.NewMockRecordRepository(ctrl)
	svc := service.NewReviewService(mockRecordRepo)

	today := time.Now().Format("2006-01-02")
	twoDaysAgo := time.Now().AddDate(0, 0, -2).Format("2006-01-02")

	// 5次记录，3次正确2次错误 → errorRate = 0.4（中错误率，满2天复习）
	records := []*domain.LearningRecord{
		{WordID: "w003", Correct: true, Date: twoDaysAgo},
		{WordID: "w003", Correct: true, Date: twoDaysAgo},
		{WordID: "w003", Correct: true, Date: twoDaysAgo},
		{WordID: "w003", Correct: false, Date: twoDaysAgo},
		{WordID: "w003", Correct: false, Date: twoDaysAgo},
	}
	mockRecordRepo.EXPECT().GetAll().Return(records, nil)

	// when
	reviewWords, err := svc.GetReviewWords(today)

	// then
	assert.NoError(t, err)
	// 中错误率(0.3<=errorRate<=0.6)，距今2天 >= 2，应加入
	assert.Contains(t, reviewWords, "w003")
}

// TestReviewService_GetReviewWords_低错误率非艾宾浩斯间隔 验证错误率低的词，在非艾宾浩斯间隔天数时，不加入复习
func TestReviewService_GetReviewWords_低错误率非艾宾浩斯间隔(t *testing.T) {
	// given
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockRecordRepo := repomock.NewMockRecordRepository(ctrl)
	svc := service.NewReviewService(mockRecordRepo)

	today := time.Now().Format("2006-01-02")
	// 3天前学习（3不在艾宾浩斯间隔集合{1,2,4,7,15}中）
	threeDaysAgo := time.Now().AddDate(0, 0, -3).Format("2006-01-02")

	// 全部正确 → errorRate = 0（低错误率）
	records := []*domain.LearningRecord{
		{WordID: "w004", Correct: true, Date: threeDaysAgo},
		{WordID: "w004", Correct: true, Date: threeDaysAgo},
	}
	mockRecordRepo.EXPECT().GetAll().Return(records, nil)

	// when
	reviewWords, err := svc.GetReviewWords(today)

	// then
	assert.NoError(t, err)
	// 3 不在艾宾浩斯间隔集合中，不应加入复习
	assert.NotContains(t, reviewWords, "w004")
}
