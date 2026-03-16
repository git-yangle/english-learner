package service

import (
	"fmt"
	"math/rand"
	"time"

	"english-learner/internal/domain"
	"english-learner/internal/repository"
)

// PlanService 学习计划服务接口
type PlanService interface {
	// InitPlan 初始化14天学习计划
	InitPlan(libraryID string) (*domain.StudyPlan, error)
	// GetTodayPlan 获取今日计划（含新词+复习词）
	GetTodayPlan() (*TodayPlan, error)
	// GetOverview 获取14天总览
	GetOverview() (*domain.StudyPlan, error)
	// HasPlan 检查是否已有计划
	HasPlan() (bool, error)
}

// TodayPlan 今日学习计划（含完整单词信息）
type TodayPlan struct {
	Day         int            `json:"day"`
	Date        string         `json:"date"`
	Status      string         `json:"status"`
	NewWords    []*domain.Word `json:"new_words"`
	ReviewWords []*domain.Word `json:"review_words"`
}

type planService struct {
	planRepo    repository.PlanRepository
	libraryRepo repository.LibraryRepository
	reviewSvc   ReviewService
}

// NewPlanService 创建学习计划服务实例
func NewPlanService(planRepo repository.PlanRepository, libraryRepo repository.LibraryRepository, reviewSvc ReviewService) PlanService {
	return &planService{
		planRepo:    planRepo,
		libraryRepo: libraryRepo,
		reviewSvc:   reviewSvc,
	}
}

// InitPlan 初始化14天学习计划
func (s *planService) InitPlan(libraryID string) (*domain.StudyPlan, error) {
	// 1. 获取词库，词库为空返回 error
	library, err := s.libraryRepo.GetByID(libraryID)
	if err != nil {
		return nil, fmt.Errorf("获取词库失败: %w", err)
	}
	if library == nil {
		return nil, fmt.Errorf("词库 %s 不存在", libraryID)
	}
	if len(library.Words) == 0 {
		return nil, fmt.Errorf("词库 %s 没有单词", libraryID)
	}

	// 2. 复制 words 切片，随机打乱顺序
	words := make([]domain.Word, len(library.Words))
	copy(words, library.Words)
	rand.Shuffle(len(words), func(i, j int) {
		words[i], words[j] = words[j], words[i]
	})

	// 3. 平均分配到14天（每天约 len/14 词，最后一天接收余数）
	const totalDays = 14
	perDay := len(words) / totalDays
	remainder := len(words) % totalDays

	today := time.Now()
	days := make([]domain.DailyPlan, 0, totalDays)

	wordIndex := 0
	for day := 1; day <= totalDays; day++ {
		// 计算当天分配的单词数量，最后一天接收余数
		dayCount := perDay
		if day == totalDays {
			dayCount += remainder
		}

		// 收集当天单词 ID
		wordIDs := make([]string, 0, dayCount)
		for i := 0; i < dayCount && wordIndex < len(words); i++ {
			wordIDs = append(wordIDs, words[wordIndex].ID)
			wordIndex++
		}

		// 计算当天日期：今天 + (day-1) 天
		date := today.AddDate(0, 0, day-1).Format("2006-01-02")

		days = append(days, domain.DailyPlan{
			Day:           day,
			Date:          date,
			WordIDs:       wordIDs,
			ReviewWordIDs: []string{}, // 首次初始化为空
			Status:        domain.PlanStatusPending,
		})
	}

	// 4. 生成 StudyPlan
	plan := &domain.StudyPlan{
		LibraryID: libraryID,
		StartDate: today.Format("2006-01-02"),
		Days:      days,
	}

	// 5. 保存计划
	if err := s.planRepo.Save(plan); err != nil {
		return nil, fmt.Errorf("保存学习计划失败: %w", err)
	}

	// 6. 返回计划
	return plan, nil
}

// GetTodayPlan 获取今日计划（含完整单词信息）
func (s *planService) GetTodayPlan() (*TodayPlan, error) {
	// 1. 获取计划，计划为 nil 返回 error
	plan, err := s.planRepo.Get()
	if err != nil {
		return nil, fmt.Errorf("获取学习计划失败: %w", err)
	}
	if plan == nil {
		return nil, fmt.Errorf("尚未初始化学习计划")
	}

	// 2. 今日日期
	today := time.Now().Format("2006-01-02")

	// 3. 找到今日对应的 DailyPlan
	var todayPlan *domain.DailyPlan
	for i := range plan.Days {
		if plan.Days[i].Date == today {
			todayPlan = &plan.Days[i]
			break
		}
	}
	if todayPlan == nil {
		return nil, fmt.Errorf("今日不在学习计划范围内")
	}

	// 4. 获取词库
	library, err := s.libraryRepo.GetByID(plan.LibraryID)
	if err != nil {
		return nil, fmt.Errorf("获取词库失败: %w", err)
	}
	if library == nil {
		return nil, fmt.Errorf("词库 %s 不存在", plan.LibraryID)
	}

	// 5. 构建 wordMap：wordID -> *domain.Word
	wordMap := make(map[string]*domain.Word, len(library.Words))
	for i := range library.Words {
		w := &library.Words[i]
		wordMap[w.ID] = w
	}

	// 6. 获取复习词 ID 列表
	reviewWordIDs, err := s.reviewSvc.GetReviewWords(today)
	if err != nil {
		return nil, fmt.Errorf("获取复习单词失败: %w", err)
	}

	// 7. 按 WordIDs 组装新词列表
	newWords := make([]*domain.Word, 0, len(todayPlan.WordIDs))
	for _, id := range todayPlan.WordIDs {
		if w, ok := wordMap[id]; ok {
			newWords = append(newWords, w)
		}
	}

	// 8. 合并 ReviewWordIDs（计划中的）与复习服务返回的词，去重
	reviewIDSet := make(map[string]struct{})
	for _, id := range todayPlan.ReviewWordIDs {
		reviewIDSet[id] = struct{}{}
	}
	for _, id := range reviewWordIDs {
		reviewIDSet[id] = struct{}{}
	}

	reviewWords := make([]*domain.Word, 0, len(reviewIDSet))
	for id := range reviewIDSet {
		if w, ok := wordMap[id]; ok {
			reviewWords = append(reviewWords, w)
		}
	}

	// 9. 返回 TodayPlan
	return &TodayPlan{
		Day:         todayPlan.Day,
		Date:        todayPlan.Date,
		Status:      string(todayPlan.Status),
		NewWords:    newWords,
		ReviewWords: reviewWords,
	}, nil
}

// GetOverview 获取14天总览
func (s *planService) GetOverview() (*domain.StudyPlan, error) {
	plan, err := s.planRepo.Get()
	if err != nil {
		return nil, fmt.Errorf("获取学习计划失败: %w", err)
	}
	if plan == nil {
		return nil, fmt.Errorf("尚未初始化学习计划")
	}
	return plan, nil
}

// HasPlan 检查是否已有计划
func (s *planService) HasPlan() (bool, error) {
	plan, err := s.planRepo.Get()
	if err != nil {
		return false, fmt.Errorf("获取学习计划失败: %w", err)
	}
	return plan != nil, nil
}
