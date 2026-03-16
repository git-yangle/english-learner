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

// makeWord 快速构建测试单词
func makeWord(id, english, chinese, category, libraryID string) domain.Word {
	return domain.Word{
		ID:        id,
		English:   english,
		Chinese:   chinese,
		Category:  category,
		LibraryID: libraryID,
		Phonetic:  "/" + english + "/",
		Example:   "This is a " + english,
		ExampleCN: "这是一个" + chinese,
	}
}

// makeTodayPlan 构建包含新词的今日计划
func makeTodayPlan(newWords []*domain.Word, reviewWords []*domain.Word) *service.TodayPlan {
	today := time.Now().Format("2006-01-02")
	return &service.TodayPlan{
		Day:         1,
		Date:        today,
		Status:      string(domain.PlanStatusPending),
		NewWords:    newWords,
		ReviewWords: reviewWords,
	}
}

// TestStudyService_MarkBrowse_成功 验证标记浏览后写入正确的学习记录
func TestStudyService_MarkBrowse_成功(t *testing.T) {
	// given
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockPlanSvc := svcmock.NewMockPlanService(ctrl)
	mockLibRepo := repomock.NewMockLibraryRepository(ctrl)
	mockRecordRepo := repomock.NewMockRecordRepository(ctrl)
	mockReviewSvc := svcmock.NewMockReviewService(ctrl)

	svc := service.NewStudyService(mockPlanSvc, mockLibRepo, mockRecordRepo, mockReviewSvc)

	today := time.Now().Format("2006-01-02")

	// 期望写入浏览记录，使用 gomock.Any() 匹配任意 record
	mockRecordRepo.EXPECT().Add(gomock.Any()).DoAndReturn(func(r *domain.LearningRecord) error {
		// 验证写入的记录字段
		assert.Equal(t, "w001", r.WordID)
		assert.Equal(t, domain.StudyModeBrowse, r.Mode)
		assert.True(t, r.Correct) // known=true
		assert.Equal(t, today, r.Date)
		return nil
	})

	// when
	err := svc.MarkBrowse("w001", true)

	// then
	assert.NoError(t, err)
}

// TestStudyService_MarkBrowse_不认识 验证标记不认识时 Correct==false
func TestStudyService_MarkBrowse_不认识(t *testing.T) {
	// given
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockPlanSvc := svcmock.NewMockPlanService(ctrl)
	mockLibRepo := repomock.NewMockLibraryRepository(ctrl)
	mockRecordRepo := repomock.NewMockRecordRepository(ctrl)
	mockReviewSvc := svcmock.NewMockReviewService(ctrl)

	svc := service.NewStudyService(mockPlanSvc, mockLibRepo, mockRecordRepo, mockReviewSvc)

	mockRecordRepo.EXPECT().Add(gomock.Any()).DoAndReturn(func(r *domain.LearningRecord) error {
		assert.False(t, r.Correct) // known=false
		assert.Equal(t, domain.StudyModeBrowse, r.Mode)
		return nil
	})

	// when
	err := svc.MarkBrowse("w001", false)

	// then
	assert.NoError(t, err)
}

// TestStudyService_MarkBrowse_仓储错误 验证仓储写入失败时，错误透传
func TestStudyService_MarkBrowse_仓储错误(t *testing.T) {
	// given
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockPlanSvc := svcmock.NewMockPlanService(ctrl)
	mockLibRepo := repomock.NewMockLibraryRepository(ctrl)
	mockRecordRepo := repomock.NewMockRecordRepository(ctrl)
	mockReviewSvc := svcmock.NewMockReviewService(ctrl)

	svc := service.NewStudyService(mockPlanSvc, mockLibRepo, mockRecordRepo, mockReviewSvc)

	mockRecordRepo.EXPECT().Add(gomock.Any()).Return(errors.New("写入失败"))

	// when
	err := svc.MarkBrowse("w001", true)

	// then
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "写入失败")
}

// TestStudyService_Browse_成功 验证 Browse 返回按 Category 分组的单词
func TestStudyService_Browse_成功(t *testing.T) {
	// given
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockPlanSvc := svcmock.NewMockPlanService(ctrl)
	mockLibRepo := repomock.NewMockLibraryRepository(ctrl)
	mockRecordRepo := repomock.NewMockRecordRepository(ctrl)
	mockReviewSvc := svcmock.NewMockReviewService(ctrl)

	svc := service.NewStudyService(mockPlanSvc, mockLibRepo, mockRecordRepo, mockReviewSvc)

	w1 := makeWord("w001", "hotel", "酒店", "hotel", "lib-001")
	w2 := makeWord("w002", "airport", "机场", "airport", "lib-001")
	w3 := makeWord("w003", "lobby", "大厅", "hotel", "lib-001")

	plan := makeTodayPlan([]*domain.Word{&w1, &w2, &w3}, []*domain.Word{})
	mockPlanSvc.EXPECT().GetTodayPlan().Return(plan, nil)

	// when
	grouped, err := svc.Browse()

	// then
	assert.NoError(t, err)
	assert.NotNil(t, grouped)
	// hotel 分类有 2 个单词
	assert.Len(t, grouped["hotel"], 2)
	// airport 分类有 1 个单词
	assert.Len(t, grouped["airport"], 1)
}

// TestStudyService_Browse_计划为空 验证 GetTodayPlan 失败时，Browse 返回错误
func TestStudyService_Browse_计划为空(t *testing.T) {
	// given
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockPlanSvc := svcmock.NewMockPlanService(ctrl)
	mockLibRepo := repomock.NewMockLibraryRepository(ctrl)
	mockRecordRepo := repomock.NewMockRecordRepository(ctrl)
	mockReviewSvc := svcmock.NewMockReviewService(ctrl)

	svc := service.NewStudyService(mockPlanSvc, mockLibRepo, mockRecordRepo, mockReviewSvc)

	mockPlanSvc.EXPECT().GetTodayPlan().Return(nil, errors.New("无计划"))

	// when
	grouped, err := svc.Browse()

	// then
	assert.Error(t, err)
	assert.Nil(t, grouped)
}

// TestStudyService_SubmitQuizAnswer_正确答案 验证提交正确中文选项时，Correct==true，并写入记录
func TestStudyService_SubmitQuizAnswer_正确答案(t *testing.T) {
	// given
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockPlanSvc := svcmock.NewMockPlanService(ctrl)
	mockLibRepo := repomock.NewMockLibraryRepository(ctrl)
	mockRecordRepo := repomock.NewMockRecordRepository(ctrl)
	mockReviewSvc := svcmock.NewMockReviewService(ctrl)

	svc := service.NewStudyService(mockPlanSvc, mockLibRepo, mockRecordRepo, mockReviewSvc)

	w := makeWord("w001", "hotel", "酒店", "hotel", "lib-001")
	plan := makeTodayPlan([]*domain.Word{&w}, []*domain.Word{})

	library := &domain.WordLibrary{
		ID:    "lib-001",
		Words: []domain.Word{w},
	}

	mockPlanSvc.EXPECT().GetTodayPlan().Return(plan, nil)
	mockLibRepo.EXPECT().GetByID("lib-001").Return(library, nil)
	mockRecordRepo.EXPECT().Add(gomock.Any()).DoAndReturn(func(r *domain.LearningRecord) error {
		assert.True(t, r.Correct)
		assert.Equal(t, domain.StudyModeQuiz, r.Mode)
		return nil
	})

	// when
	result, err := svc.SubmitQuizAnswer("w001", "酒店")

	// then
	assert.NoError(t, err)
	assert.True(t, result.Correct)
	assert.Equal(t, "酒店", result.CorrectAnswer)
}

// TestStudyService_SubmitQuizAnswer_错误答案 验证提交错误答案时，Correct==false
func TestStudyService_SubmitQuizAnswer_错误答案(t *testing.T) {
	// given
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockPlanSvc := svcmock.NewMockPlanService(ctrl)
	mockLibRepo := repomock.NewMockLibraryRepository(ctrl)
	mockRecordRepo := repomock.NewMockRecordRepository(ctrl)
	mockReviewSvc := svcmock.NewMockReviewService(ctrl)

	svc := service.NewStudyService(mockPlanSvc, mockLibRepo, mockRecordRepo, mockReviewSvc)

	w := makeWord("w001", "hotel", "酒店", "hotel", "lib-001")
	plan := makeTodayPlan([]*domain.Word{&w}, []*domain.Word{})

	library := &domain.WordLibrary{
		ID:    "lib-001",
		Words: []domain.Word{w},
	}

	mockPlanSvc.EXPECT().GetTodayPlan().Return(plan, nil)
	mockLibRepo.EXPECT().GetByID("lib-001").Return(library, nil)
	mockRecordRepo.EXPECT().Add(gomock.Any()).DoAndReturn(func(r *domain.LearningRecord) error {
		assert.False(t, r.Correct)
		return nil
	})

	// when
	result, err := svc.SubmitQuizAnswer("w001", "机场")

	// then
	assert.NoError(t, err)
	assert.False(t, result.Correct)
	assert.Equal(t, "酒店", result.CorrectAnswer)
}

// TestStudyService_SubmitDictation_正确拼写 验证大小写不敏感的正确拼写判断
func TestStudyService_SubmitDictation_正确拼写(t *testing.T) {
	// given
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockPlanSvc := svcmock.NewMockPlanService(ctrl)
	mockLibRepo := repomock.NewMockLibraryRepository(ctrl)
	mockRecordRepo := repomock.NewMockRecordRepository(ctrl)
	mockReviewSvc := svcmock.NewMockReviewService(ctrl)

	svc := service.NewStudyService(mockPlanSvc, mockLibRepo, mockRecordRepo, mockReviewSvc)

	w := makeWord("w001", "hotel", "酒店", "hotel", "lib-001")
	plan := makeTodayPlan([]*domain.Word{&w}, []*domain.Word{})

	library := &domain.WordLibrary{
		ID:    "lib-001",
		Words: []domain.Word{w},
	}

	mockPlanSvc.EXPECT().GetTodayPlan().Return(plan, nil)
	mockLibRepo.EXPECT().GetByID("lib-001").Return(library, nil)
	mockRecordRepo.EXPECT().Add(gomock.Any()).DoAndReturn(func(r *domain.LearningRecord) error {
		assert.True(t, r.Correct)
		assert.Equal(t, domain.StudyModeDictation, r.Mode)
		return nil
	})

	// when（大写输入，应大小写不敏感匹配）
	result, err := svc.SubmitDictation("w001", "HOTEL")

	// then
	assert.NoError(t, err)
	assert.True(t, result.Correct)
	assert.Equal(t, "hotel", result.CorrectSpelling)
}

// TestStudyService_SubmitDictation_错误拼写 验证拼写错误时 Correct==false
func TestStudyService_SubmitDictation_错误拼写(t *testing.T) {
	// given
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockPlanSvc := svcmock.NewMockPlanService(ctrl)
	mockLibRepo := repomock.NewMockLibraryRepository(ctrl)
	mockRecordRepo := repomock.NewMockRecordRepository(ctrl)
	mockReviewSvc := svcmock.NewMockReviewService(ctrl)

	svc := service.NewStudyService(mockPlanSvc, mockLibRepo, mockRecordRepo, mockReviewSvc)

	w := makeWord("w001", "hotel", "酒店", "hotel", "lib-001")
	plan := makeTodayPlan([]*domain.Word{&w}, []*domain.Word{})

	library := &domain.WordLibrary{
		ID:    "lib-001",
		Words: []domain.Word{w},
	}

	mockPlanSvc.EXPECT().GetTodayPlan().Return(plan, nil)
	mockLibRepo.EXPECT().GetByID("lib-001").Return(library, nil)
	mockRecordRepo.EXPECT().Add(gomock.Any()).Return(nil)

	// when
	result, err := svc.SubmitDictation("w001", "hotle") // 拼写错误

	// then
	assert.NoError(t, err)
	assert.False(t, result.Correct)
	assert.Equal(t, "hotel", result.CorrectSpelling)
}

// TestStudyService_GenerateQuiz_成功 验证正常生成测试题，高错误率词优先出题
func TestStudyService_GenerateQuiz_成功(t *testing.T) {
	// given
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockPlanSvc := svcmock.NewMockPlanService(ctrl)
	mockLibRepo := repomock.NewMockLibraryRepository(ctrl)
	mockRecordRepo := repomock.NewMockRecordRepository(ctrl)
	mockReviewSvc := svcmock.NewMockReviewService(ctrl)

	svc := service.NewStudyService(mockPlanSvc, mockLibRepo, mockRecordRepo, mockReviewSvc)

	w1 := makeWord("w001", "hotel", "酒店", "hotel", "lib-001")
	w2 := makeWord("w002", "airport", "机场", "airport", "lib-001")
	w3 := makeWord("w003", "taxi", "出租车", "transportation", "lib-001")
	w4 := makeWord("w004", "bus", "公共汽车", "transportation", "lib-001")
	w5 := makeWord("w005", "train", "火车", "transportation", "lib-001")

	plan := makeTodayPlan([]*domain.Word{&w1, &w2, &w3, &w4, &w5}, []*domain.Word{})

	library := &domain.WordLibrary{
		ID:    "lib-001",
		Words: []domain.Word{w1, w2, w3, w4, w5},
	}

	mockPlanSvc.EXPECT().GetTodayPlan().Return(plan, nil)
	// 为每个单词计算错误率（mock 返回各自的 stats）
	mockReviewSvc.EXPECT().CalcWordStats("w001").Return(&domain.WordStats{WordID: "w001", ErrorRate: 0.8}, nil)
	mockReviewSvc.EXPECT().CalcWordStats("w002").Return(&domain.WordStats{WordID: "w002", ErrorRate: 0.3}, nil)
	mockReviewSvc.EXPECT().CalcWordStats("w003").Return(&domain.WordStats{WordID: "w003", ErrorRate: 0.1}, nil)
	mockReviewSvc.EXPECT().CalcWordStats("w004").Return(&domain.WordStats{WordID: "w004", ErrorRate: 0.0}, nil)
	mockReviewSvc.EXPECT().CalcWordStats("w005").Return(&domain.WordStats{WordID: "w005", ErrorRate: 0.0}, nil)
	mockLibRepo.EXPECT().GetByID("lib-001").Return(library, nil)

	// when：请求2道题
	questions, err := svc.GenerateQuiz(2)

	// then
	assert.NoError(t, err)
	assert.Len(t, questions, 2)
	// 第一道题应该是错误率最高的 w001
	assert.Equal(t, "w001", questions[0].WordID)
	// 每道题必须有4个或以下选项
	assert.LessOrEqual(t, len(questions[0].Options), 4)
	assert.GreaterOrEqual(t, len(questions[0].Options), 1)
}

// TestStudyService_GenerateQuiz_今日无单词 验证今日无单词时返回错误
func TestStudyService_GenerateQuiz_今日无单词(t *testing.T) {
	// given
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockPlanSvc := svcmock.NewMockPlanService(ctrl)
	mockLibRepo := repomock.NewMockLibraryRepository(ctrl)
	mockRecordRepo := repomock.NewMockRecordRepository(ctrl)
	mockReviewSvc := svcmock.NewMockReviewService(ctrl)

	svc := service.NewStudyService(mockPlanSvc, mockLibRepo, mockRecordRepo, mockReviewSvc)

	// 返回空词列表的今日计划
	emptyPlan := makeTodayPlan([]*domain.Word{}, []*domain.Word{})
	mockPlanSvc.EXPECT().GetTodayPlan().Return(emptyPlan, nil)

	// when
	questions, err := svc.GenerateQuiz(5)

	// then
	assert.Error(t, err)
	assert.Nil(t, questions)
	assert.Contains(t, err.Error(), "今日没有需要学习的单词")
}

// TestStudyService_GetDictationWord_成功 验证获取默写词，高错误率优先
func TestStudyService_GetDictationWord_成功(t *testing.T) {
	// given
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockPlanSvc := svcmock.NewMockPlanService(ctrl)
	mockLibRepo := repomock.NewMockLibraryRepository(ctrl)
	mockRecordRepo := repomock.NewMockRecordRepository(ctrl)
	mockReviewSvc := svcmock.NewMockReviewService(ctrl)

	svc := service.NewStudyService(mockPlanSvc, mockLibRepo, mockRecordRepo, mockReviewSvc)

	w1 := makeWord("w001", "hotel", "酒店", "hotel", "lib-001")
	w2 := makeWord("w002", "airport", "机场", "airport", "lib-001")

	plan := makeTodayPlan([]*domain.Word{&w1, &w2}, []*domain.Word{})
	mockPlanSvc.EXPECT().GetTodayPlan().Return(plan, nil)

	// w001 错误率 0.9（最高），应优先返回
	mockReviewSvc.EXPECT().CalcWordStats("w001").Return(&domain.WordStats{WordID: "w001", ErrorRate: 0.9}, nil)
	mockReviewSvc.EXPECT().CalcWordStats("w002").Return(&domain.WordStats{WordID: "w002", ErrorRate: 0.1}, nil)

	// when
	question, err := svc.GetDictationWord()

	// then
	assert.NoError(t, err)
	assert.NotNil(t, question)
	// 高错误率的 w001 应被优先返回
	assert.Equal(t, "w001", question.WordID)
	// 默写题只展示中文，不展示英文
	assert.Equal(t, "酒店", question.Chinese)
	assert.Empty(t, "") // 不暴露英文拼写（字段本身不存在于 DictationQuestion）
}

// TestStudyService_GetDictationWord_今日无单词 验证今日无单词时返回错误
func TestStudyService_GetDictationWord_今日无单词(t *testing.T) {
	// given
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockPlanSvc := svcmock.NewMockPlanService(ctrl)
	mockLibRepo := repomock.NewMockLibraryRepository(ctrl)
	mockRecordRepo := repomock.NewMockRecordRepository(ctrl)
	mockReviewSvc := svcmock.NewMockReviewService(ctrl)

	svc := service.NewStudyService(mockPlanSvc, mockLibRepo, mockRecordRepo, mockReviewSvc)

	emptyPlan := makeTodayPlan([]*domain.Word{}, []*domain.Word{})
	mockPlanSvc.EXPECT().GetTodayPlan().Return(emptyPlan, nil)

	// when
	question, err := svc.GetDictationWord()

	// then
	assert.Error(t, err)
	assert.Nil(t, question)
	assert.Contains(t, err.Error(), "今日没有需要学习的单词")
}
