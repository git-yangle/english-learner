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

// buildTestWords 构建指定数量的测试单词（ID 使用字符下标）
func buildTestWords(n int) []domain.Word {
	words := make([]domain.Word, n)
	for i := range words {
		words[i] = domain.Word{
			// 使用格式化字符串保证 ID 唯一
			ID:      "word-" + string(rune('a'+i%26)),
			English: "english",
			Chinese: "中文",
		}
	}
	return words
}

// TestPlanService_InitPlan_成功 验证10个词初始化14天计划，每天有词，总词数正确
func TestPlanService_InitPlan_成功(t *testing.T) {
	// given
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockPlanRepo := repomock.NewMockPlanRepository(ctrl)
	mockLibRepo := repomock.NewMockLibraryRepository(ctrl)
	mockReviewSvc := svcmock.NewMockReviewService(ctrl)

	svc := service.NewPlanService(mockPlanRepo, mockLibRepo, mockReviewSvc)

	library := &domain.WordLibrary{
		ID:    "lib-001",
		Name:  "旅行英语",
		Words: buildTestWords(10),
	}
	mockLibRepo.EXPECT().GetByID("lib-001").Return(library, nil)
	// 保存计划时期望被调用一次
	mockPlanRepo.EXPECT().Save(gomock.Any()).Return(nil)

	// when
	plan, err := svc.InitPlan("lib-001")

	// then
	assert.NoError(t, err)
	assert.NotNil(t, plan)
	assert.Equal(t, "lib-001", plan.LibraryID)
	// 必须生成14天计划
	assert.Len(t, plan.Days, 14)

	// 统计所有天的词数总量，必须等于词库单词数
	totalWords := 0
	for _, d := range plan.Days {
		totalWords += len(d.WordIDs)
		// 验证日期不为空、状态为待学习
		assert.NotEmpty(t, d.Date)
		assert.Equal(t, domain.PlanStatusPending, d.Status)
	}
	assert.Equal(t, 10, totalWords)

	// 验证日期连续：第1天为今天，第14天为今天+13天
	today := time.Now().Format("2006-01-02")
	assert.Equal(t, today, plan.Days[0].Date)
	expectedLastDay := time.Now().AddDate(0, 0, 13).Format("2006-01-02")
	assert.Equal(t, expectedLastDay, plan.Days[13].Date)
}

// TestPlanService_InitPlan_词库不存在 验证词库获取失败时，InitPlan 返回错误
func TestPlanService_InitPlan_词库不存在(t *testing.T) {
	// given
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockPlanRepo := repomock.NewMockPlanRepository(ctrl)
	mockLibRepo := repomock.NewMockLibraryRepository(ctrl)
	mockReviewSvc := svcmock.NewMockReviewService(ctrl)

	svc := service.NewPlanService(mockPlanRepo, mockLibRepo, mockReviewSvc)

	mockLibRepo.EXPECT().GetByID("lib-999").Return(nil, errors.New("词库不存在"))

	// when
	plan, err := svc.InitPlan("lib-999")

	// then
	assert.Error(t, err)
	assert.Nil(t, plan)
	assert.Contains(t, err.Error(), "词库不存在")
}

// TestPlanService_InitPlan_词库为空 验证词库存在但无单词时，返回错误
func TestPlanService_InitPlan_词库为空(t *testing.T) {
	// given
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockPlanRepo := repomock.NewMockPlanRepository(ctrl)
	mockLibRepo := repomock.NewMockLibraryRepository(ctrl)
	mockReviewSvc := svcmock.NewMockReviewService(ctrl)

	svc := service.NewPlanService(mockPlanRepo, mockLibRepo, mockReviewSvc)

	emptyLibrary := &domain.WordLibrary{
		ID:    "lib-001",
		Name:  "空词库",
		Words: []domain.Word{},
	}
	mockLibRepo.EXPECT().GetByID("lib-001").Return(emptyLibrary, nil)

	// when
	plan, err := svc.InitPlan("lib-001")

	// then
	assert.Error(t, err)
	assert.Nil(t, plan)
	assert.Contains(t, err.Error(), "没有单词")
}

// TestPlanService_HasPlan_有计划 验证 planRepo.Get 返回非 nil 时，HasPlan 返回 true
func TestPlanService_HasPlan_有计划(t *testing.T) {
	// given
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockPlanRepo := repomock.NewMockPlanRepository(ctrl)
	mockLibRepo := repomock.NewMockLibraryRepository(ctrl)
	mockReviewSvc := svcmock.NewMockReviewService(ctrl)

	svc := service.NewPlanService(mockPlanRepo, mockLibRepo, mockReviewSvc)

	existingPlan := &domain.StudyPlan{
		LibraryID: "lib-001",
		StartDate: "2026-03-01",
		Days:      []domain.DailyPlan{},
	}
	mockPlanRepo.EXPECT().Get().Return(existingPlan, nil)

	// when
	hasPlan, err := svc.HasPlan()

	// then
	assert.NoError(t, err)
	assert.True(t, hasPlan)
}

// TestPlanService_HasPlan_无计划 验证 planRepo.Get 返回 nil 时，HasPlan 返回 false
func TestPlanService_HasPlan_无计划(t *testing.T) {
	// given
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockPlanRepo := repomock.NewMockPlanRepository(ctrl)
	mockLibRepo := repomock.NewMockLibraryRepository(ctrl)
	mockReviewSvc := svcmock.NewMockReviewService(ctrl)

	svc := service.NewPlanService(mockPlanRepo, mockLibRepo, mockReviewSvc)

	mockPlanRepo.EXPECT().Get().Return(nil, nil)

	// when
	hasPlan, err := svc.HasPlan()

	// then
	assert.NoError(t, err)
	assert.False(t, hasPlan)
}

// TestPlanService_GetTodayPlan_无计划 验证计划为 nil 时，GetTodayPlan 返回 error
func TestPlanService_GetTodayPlan_无计划(t *testing.T) {
	// given
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockPlanRepo := repomock.NewMockPlanRepository(ctrl)
	mockLibRepo := repomock.NewMockLibraryRepository(ctrl)
	mockReviewSvc := svcmock.NewMockReviewService(ctrl)

	svc := service.NewPlanService(mockPlanRepo, mockLibRepo, mockReviewSvc)

	mockPlanRepo.EXPECT().Get().Return(nil, nil)

	// when
	todayPlan, err := svc.GetTodayPlan()

	// then
	assert.Error(t, err)
	assert.Nil(t, todayPlan)
	assert.Contains(t, err.Error(), "尚未初始化学习计划")
}

// TestPlanService_GetTodayPlan_今日在计划内 验证今日在计划范围内时，正常返回今日计划
func TestPlanService_GetTodayPlan_今日在计划内(t *testing.T) {
	// given
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockPlanRepo := repomock.NewMockPlanRepository(ctrl)
	mockLibRepo := repomock.NewMockLibraryRepository(ctrl)
	mockReviewSvc := svcmock.NewMockReviewService(ctrl)

	svc := service.NewPlanService(mockPlanRepo, mockLibRepo, mockReviewSvc)

	today := time.Now().Format("2006-01-02")
	plan := &domain.StudyPlan{
		LibraryID: "lib-001",
		StartDate: today,
		Days: []domain.DailyPlan{
			{
				Day:           1,
				Date:          today,
				WordIDs:       []string{"w001", "w002"},
				ReviewWordIDs: []string{},
				Status:        domain.PlanStatusPending,
			},
		},
	}
	library := &domain.WordLibrary{
		ID:   "lib-001",
		Name: "旅行英语",
		Words: []domain.Word{
			{ID: "w001", English: "hotel", Chinese: "酒店", LibraryID: "lib-001"},
			{ID: "w002", English: "airport", Chinese: "机场", LibraryID: "lib-001"},
		},
	}

	mockPlanRepo.EXPECT().Get().Return(plan, nil)
	mockLibRepo.EXPECT().GetByID("lib-001").Return(library, nil)
	// GetReviewWords 返回空列表
	mockReviewSvc.EXPECT().GetReviewWords(today).Return([]string{}, nil)

	// when
	todayPlan, err := svc.GetTodayPlan()

	// then
	assert.NoError(t, err)
	assert.NotNil(t, todayPlan)
	assert.Equal(t, 1, todayPlan.Day)
	assert.Equal(t, today, todayPlan.Date)
	assert.Len(t, todayPlan.NewWords, 2)
}

// TestPlanService_GetOverview_计划不存在 验证计划不存在时返回错误
func TestPlanService_GetOverview_计划不存在(t *testing.T) {
	// given
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockPlanRepo := repomock.NewMockPlanRepository(ctrl)
	mockLibRepo := repomock.NewMockLibraryRepository(ctrl)
	mockReviewSvc := svcmock.NewMockReviewService(ctrl)

	svc := service.NewPlanService(mockPlanRepo, mockLibRepo, mockReviewSvc)

	mockPlanRepo.EXPECT().Get().Return(nil, nil)

	// when
	overview, err := svc.GetOverview()

	// then
	assert.Error(t, err)
	assert.Nil(t, overview)
	assert.Contains(t, err.Error(), "尚未初始化学习计划")
}
