package service_test

import (
	"errors"
	"testing"
	"time"

	"english-learner/internal/domain"
	repomock "english-learner/internal/repository/mock"
	"english-learner/internal/service"
	svcmock "english-learner/internal/service/mock"

	"github.com/stretchr/testify/assert"
	"go.uber.org/mock/gomock"
)

// TestCheckInService_CalcDailyScore_无记录 验证今日无任何作答时，所有得分均为0
func TestCheckInService_CalcDailyScore_无记录(t *testing.T) {
	// given
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockCheckInRepo := repomock.NewMockCheckInRepository(ctrl)
	mockRecordRepo := repomock.NewMockRecordRepository(ctrl)
	mockPlanRepo := repomock.NewMockPlanRepository(ctrl)
	mockReviewSvc := svcmock.NewMockReviewService(ctrl)

	svc := service.NewCheckInService(mockCheckInRepo, mockRecordRepo, mockPlanRepo, mockReviewSvc)

	today := time.Now().Format("2006-01-02")
	// 今日无任何学习记录
	mockRecordRepo.EXPECT().GetByDate(today).Return([]*domain.LearningRecord{}, nil)

	// when
	score, err := svc.CalcDailyScore()

	// then
	assert.NoError(t, err)
	assert.NotNil(t, score)
	assert.Equal(t, 0.0, score.QuizScore)
	assert.Equal(t, 0.0, score.DictationScore)
	assert.Equal(t, 0.0, score.TotalScore)
	assert.False(t, score.Completed) // 0 < 0.6，未完成
}

// TestCheckInService_CalcDailyScore_仅quiz 验证4道quiz答对3道，quizScore=0.75，totalScore=0.3
func TestCheckInService_CalcDailyScore_仅quiz(t *testing.T) {
	// given
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockCheckInRepo := repomock.NewMockCheckInRepository(ctrl)
	mockRecordRepo := repomock.NewMockRecordRepository(ctrl)
	mockPlanRepo := repomock.NewMockPlanRepository(ctrl)
	mockReviewSvc := svcmock.NewMockReviewService(ctrl)

	svc := service.NewCheckInService(mockCheckInRepo, mockRecordRepo, mockPlanRepo, mockReviewSvc)

	today := time.Now().Format("2006-01-02")
	// 4道quiz，3对1错
	records := []*domain.LearningRecord{
		{WordID: "w001", Mode: domain.StudyModeQuiz, Correct: true, Date: today},
		{WordID: "w002", Mode: domain.StudyModeQuiz, Correct: true, Date: today},
		{WordID: "w003", Mode: domain.StudyModeQuiz, Correct: true, Date: today},
		{WordID: "w004", Mode: domain.StudyModeQuiz, Correct: false, Date: today},
	}
	mockRecordRepo.EXPECT().GetByDate(today).Return(records, nil)

	// when
	score, err := svc.CalcDailyScore()

	// then
	assert.NoError(t, err)
	// quizScore = 3/4 = 0.75
	assert.InDelta(t, 0.75, score.QuizScore, 1e-9)
	// 无默写记录，dictationScore = 0
	assert.Equal(t, 0.0, score.DictationScore)
	// totalScore = 0.75 * 0.4 + 0 * 0.6 = 0.3
	assert.InDelta(t, 0.3, score.TotalScore, 1e-9)
	// 0.3 < 0.6，未完成
	assert.False(t, score.Completed)
}

// TestCheckInService_CalcDailyScore_quiz和dictation 验证加权综合分计算
func TestCheckInService_CalcDailyScore_quiz和dictation(t *testing.T) {
	// given
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockCheckInRepo := repomock.NewMockCheckInRepository(ctrl)
	mockRecordRepo := repomock.NewMockRecordRepository(ctrl)
	mockPlanRepo := repomock.NewMockPlanRepository(ctrl)
	mockReviewSvc := svcmock.NewMockReviewService(ctrl)

	svc := service.NewCheckInService(mockCheckInRepo, mockRecordRepo, mockPlanRepo, mockReviewSvc)

	today := time.Now().Format("2006-01-02")
	// quiz 4道全对 → quizScore = 1.0
	// dictation 2道全对 → dictationScore = 1.0
	// totalScore = 1.0 * 0.4 + 1.0 * 0.6 = 1.0
	records := []*domain.LearningRecord{
		{WordID: "w001", Mode: domain.StudyModeQuiz, Correct: true, Date: today},
		{WordID: "w002", Mode: domain.StudyModeQuiz, Correct: true, Date: today},
		{WordID: "w003", Mode: domain.StudyModeQuiz, Correct: true, Date: today},
		{WordID: "w004", Mode: domain.StudyModeQuiz, Correct: true, Date: today},
		{WordID: "w001", Mode: domain.StudyModeDictation, Correct: true, Date: today},
		{WordID: "w002", Mode: domain.StudyModeDictation, Correct: true, Date: today},
	}
	mockRecordRepo.EXPECT().GetByDate(today).Return(records, nil)

	// when
	score, err := svc.CalcDailyScore()

	// then
	assert.NoError(t, err)
	assert.InDelta(t, 1.0, score.QuizScore, 1e-9)
	assert.InDelta(t, 1.0, score.DictationScore, 1e-9)
	assert.InDelta(t, 1.0, score.TotalScore, 1e-9)
	// 1.0 >= 0.6，已完成
	assert.True(t, score.Completed)
}

// TestCheckInService_CalcDailyScore_加权综合分达标 验证quiz和dictation加权后恰好达到0.6完成线
func TestCheckInService_CalcDailyScore_加权综合分达标(t *testing.T) {
	// given
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockCheckInRepo := repomock.NewMockCheckInRepository(ctrl)
	mockRecordRepo := repomock.NewMockRecordRepository(ctrl)
	mockPlanRepo := repomock.NewMockPlanRepository(ctrl)
	mockReviewSvc := svcmock.NewMockReviewService(ctrl)

	svc := service.NewCheckInService(mockCheckInRepo, mockRecordRepo, mockPlanRepo, mockReviewSvc)

	today := time.Now().Format("2006-01-02")
	// quiz 0道（quizScore=0），dictation 全对（dictationScore=1.0）
	// totalScore = 0 * 0.4 + 1.0 * 0.6 = 0.6，恰好达到完成线
	records := []*domain.LearningRecord{
		{WordID: "w001", Mode: domain.StudyModeDictation, Correct: true, Date: today},
		{WordID: "w002", Mode: domain.StudyModeDictation, Correct: true, Date: today},
	}
	mockRecordRepo.EXPECT().GetByDate(today).Return(records, nil)

	// when
	score, err := svc.CalcDailyScore()

	// then
	assert.NoError(t, err)
	assert.Equal(t, 0.0, score.QuizScore)
	assert.InDelta(t, 1.0, score.DictationScore, 1e-9)
	assert.InDelta(t, 0.6, score.TotalScore, 1e-9)
	// 0.6 >= 0.6，恰好完成
	assert.True(t, score.Completed)
}

// TestCheckInService_GetStats_连续天数 验证连续3天打卡（今天+昨天+前天），StreakDays==3
func TestCheckInService_GetStats_连续天数(t *testing.T) {
	// given
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockCheckInRepo := repomock.NewMockCheckInRepository(ctrl)
	mockRecordRepo := repomock.NewMockRecordRepository(ctrl)
	mockPlanRepo := repomock.NewMockPlanRepository(ctrl)
	mockReviewSvc := svcmock.NewMockReviewService(ctrl)

	svc := service.NewCheckInService(mockCheckInRepo, mockRecordRepo, mockPlanRepo, mockReviewSvc)

	// 构造今天、昨天、前天的打卡记录（连续3天）
	today := time.Now().Format("2006-01-02")
	yesterday := time.Now().AddDate(0, 0, -1).Format("2006-01-02")
	dayBeforeYesterday := time.Now().AddDate(0, 0, -2).Format("2006-01-02")

	checkIns := []*domain.CheckIn{
		{Day: 3, Date: today, Completed: true, Score: 0.8},
		{Day: 2, Date: yesterday, Completed: true, Score: 0.7},
		{Day: 1, Date: dayBeforeYesterday, Completed: true, Score: 0.9},
	}
	mockCheckInRepo.EXPECT().GetAll().Return(checkIns, nil)

	// when
	stats, err := svc.GetStats()

	// then
	assert.NoError(t, err)
	assert.NotNil(t, stats)
	assert.Equal(t, 3, stats.TotalDays)
	assert.Equal(t, 3, stats.CompletedDays)
	// 今天、昨天、前天均有打卡，连续天数为3
	assert.Equal(t, 3, stats.StreakDays)
	// 平均分 = (0.8 + 0.7 + 0.9) / 3 = 0.8
	assert.InDelta(t, 0.8, stats.AvgScore, 1e-9)
}

// TestCheckInService_GetStats_中断连续 验证昨天未打卡导致连续中断，今天也未打卡时 StreakDays==0
func TestCheckInService_GetStats_中断连续(t *testing.T) {
	// given
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockCheckInRepo := repomock.NewMockCheckInRepository(ctrl)
	mockRecordRepo := repomock.NewMockRecordRepository(ctrl)
	mockPlanRepo := repomock.NewMockPlanRepository(ctrl)
	mockReviewSvc := svcmock.NewMockReviewService(ctrl)

	svc := service.NewCheckInService(mockCheckInRepo, mockRecordRepo, mockPlanRepo, mockReviewSvc)

	// 只有前天的打卡记录，昨天和今天都没有 → 连续中断
	dayBeforeYesterday := time.Now().AddDate(0, 0, -2).Format("2006-01-02")

	checkIns := []*domain.CheckIn{
		{Day: 1, Date: dayBeforeYesterday, Completed: true, Score: 0.7},
	}
	mockCheckInRepo.EXPECT().GetAll().Return(checkIns, nil)

	// when
	stats, err := svc.GetStats()

	// then
	assert.NoError(t, err)
	assert.NotNil(t, stats)
	assert.Equal(t, 1, stats.TotalDays)
	// 今天没有打卡，从今天往前检查，第一天就断了，StreakDays==0
	assert.Equal(t, 0, stats.StreakDays)
}

// TestCheckInService_GetStats_仅今天打卡 验证只有今天一天打卡，StreakDays==1
func TestCheckInService_GetStats_仅今天打卡(t *testing.T) {
	// given
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockCheckInRepo := repomock.NewMockCheckInRepository(ctrl)
	mockRecordRepo := repomock.NewMockRecordRepository(ctrl)
	mockPlanRepo := repomock.NewMockPlanRepository(ctrl)
	mockReviewSvc := svcmock.NewMockReviewService(ctrl)

	svc := service.NewCheckInService(mockCheckInRepo, mockRecordRepo, mockPlanRepo, mockReviewSvc)

	today := time.Now().Format("2006-01-02")
	checkIns := []*domain.CheckIn{
		{Day: 1, Date: today, Completed: true, Score: 1.0},
	}
	mockCheckInRepo.EXPECT().GetAll().Return(checkIns, nil)

	// when
	stats, err := svc.GetStats()

	// then
	assert.NoError(t, err)
	assert.Equal(t, 1, stats.StreakDays)
	assert.Equal(t, 1, stats.CompletedDays)
	assert.InDelta(t, 1.0, stats.AvgScore, 1e-9)
}

// TestCheckInService_GetStats_包含未完成打卡 验证有未完成打卡时，CompletedDays只统计已完成的
func TestCheckInService_GetStats_包含未完成打卡(t *testing.T) {
	// given
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockCheckInRepo := repomock.NewMockCheckInRepository(ctrl)
	mockRecordRepo := repomock.NewMockRecordRepository(ctrl)
	mockPlanRepo := repomock.NewMockPlanRepository(ctrl)
	mockReviewSvc := svcmock.NewMockReviewService(ctrl)

	svc := service.NewCheckInService(mockCheckInRepo, mockRecordRepo, mockPlanRepo, mockReviewSvc)

	today := time.Now().Format("2006-01-02")
	yesterday := time.Now().AddDate(0, 0, -1).Format("2006-01-02")

	// 今天完成，昨天未完成
	checkIns := []*domain.CheckIn{
		{Day: 2, Date: today, Completed: true, Score: 0.8},
		{Day: 1, Date: yesterday, Completed: false, Score: 0.2},
	}
	mockCheckInRepo.EXPECT().GetAll().Return(checkIns, nil)

	// when
	stats, err := svc.GetStats()

	// then
	assert.NoError(t, err)
	assert.Equal(t, 2, stats.TotalDays)
	// 只有1天已完成
	assert.Equal(t, 1, stats.CompletedDays)
	// 连续天数：今天有打卡记录，昨天也有 → StreakDays=2（连续性按日期判断，不管是否completed）
	assert.Equal(t, 2, stats.StreakDays)
	// 平均分只统计已完成的：0.8 / 1 = 0.8
	assert.InDelta(t, 0.8, stats.AvgScore, 1e-9)
}

// TestCheckInService_GetStats_空记录 验证无任何打卡记录时，所有统计值为0
func TestCheckInService_GetStats_空记录(t *testing.T) {
	// given
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockCheckInRepo := repomock.NewMockCheckInRepository(ctrl)
	mockRecordRepo := repomock.NewMockRecordRepository(ctrl)
	mockPlanRepo := repomock.NewMockPlanRepository(ctrl)
	mockReviewSvc := svcmock.NewMockReviewService(ctrl)

	svc := service.NewCheckInService(mockCheckInRepo, mockRecordRepo, mockPlanRepo, mockReviewSvc)

	mockCheckInRepo.EXPECT().GetAll().Return([]*domain.CheckIn{}, nil)

	// when
	stats, err := svc.GetStats()

	// then
	assert.NoError(t, err)
	assert.Equal(t, 0, stats.TotalDays)
	assert.Equal(t, 0, stats.CompletedDays)
	assert.Equal(t, 0, stats.StreakDays)
	assert.Equal(t, 0.0, stats.AvgScore)
}

// TestCheckInService_GetStats_仓储错误 验证仓储返回错误时，GetStats 透传错误
func TestCheckInService_GetStats_仓储错误(t *testing.T) {
	// given
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockCheckInRepo := repomock.NewMockCheckInRepository(ctrl)
	mockRecordRepo := repomock.NewMockRecordRepository(ctrl)
	mockPlanRepo := repomock.NewMockPlanRepository(ctrl)
	mockReviewSvc := svcmock.NewMockReviewService(ctrl)

	svc := service.NewCheckInService(mockCheckInRepo, mockRecordRepo, mockPlanRepo, mockReviewSvc)

	mockCheckInRepo.EXPECT().GetAll().Return(nil, errors.New("数据库错误"))

	// when
	stats, err := svc.GetStats()

	// then
	assert.Error(t, err)
	assert.Nil(t, stats)
	assert.Contains(t, err.Error(), "数据库错误")
}
