package service

import (
	"fmt"
	"math/rand"
	"sort"
	"strings"
	"time"

	"english-learner/internal/domain"
	"english-learner/internal/repository"
)

// QuizQuestion 测试题目
type QuizQuestion struct {
	WordID   string   `json:"word_id"`
	English  string   `json:"english"`
	Phonetic string   `json:"phonetic"`
	Options  []string `json:"options"` // 4个中文选项
}

// QuizResult 答题结果
type QuizResult struct {
	Correct       bool   `json:"correct"`
	CorrectAnswer string `json:"correct_answer"`
	Example       string `json:"example"`
	ExampleCN     string `json:"example_cn"`
}

// DictationQuestion 默写题目
type DictationQuestion struct {
	WordID    string `json:"word_id"`
	Chinese   string `json:"chinese"`
	Example   string `json:"example"`
	ExampleCN string `json:"example_cn"`
}

// DictationResult 默写结果
type DictationResult struct {
	Correct         bool   `json:"correct"`
	CorrectSpelling string `json:"correct_spelling"`
	Phonetic        string `json:"phonetic"`
}

// TodayProgress 今日学习进度
type TodayProgress struct {
	BrowseMarks map[string]bool `json:"browse_marks"` // wordID -> true(认识)/false(不认识)
	QuizCount   int             `json:"quiz_count"`   // 今日测试答题数
	QuizCorrect int             `json:"quiz_correct"` // 今日测试答对数
	DictCount   int             `json:"dict_count"`   // 今日默写答题数
	DictCorrect int             `json:"dict_correct"` // 今日默写答对数
}

// StudyService 学习服务接口
type StudyService interface {
	// Browse 浏览模式：获取今日单词列表（按分类分组）
	Browse() (map[string][]*domain.Word, error)
	// MarkBrowse 浏览标记（认识/不认识）
	MarkBrowse(wordID string, known bool) error
	// GenerateQuiz 生成测试题（优先出错误率高的词）
	GenerateQuiz(count int) ([]*QuizQuestion, error)
	// SubmitQuizAnswer 提交测试答案
	SubmitQuizAnswer(wordID, answer string) (*QuizResult, error)
	// GetDictationWord 获取下一个默写词
	GetDictationWord() (*DictationQuestion, error)
	// SubmitDictation 提交默写答案
	SubmitDictation(wordID, input string) (*DictationResult, error)
	// GetTodayProgress 获取今日学习进度（浏览标记状态、测试/默写完成数）
	GetTodayProgress() (*TodayProgress, error)
}

type studyService struct {
	planSvc     PlanService
	libraryRepo repository.LibraryRepository
	recordRepo  repository.RecordRepository
	reviewSvc   ReviewService
}

// NewStudyService 创建学习服务实例
func NewStudyService(planSvc PlanService, libraryRepo repository.LibraryRepository, recordRepo repository.RecordRepository, reviewSvc ReviewService) StudyService {
	return &studyService{
		planSvc:     planSvc,
		libraryRepo: libraryRepo,
		recordRepo:  recordRepo,
		reviewSvc:   reviewSvc,
	}
}

// today 返回今日日期字符串（yyyy-MM-dd）
func today() string {
	return time.Now().Format("2006-01-02")
}

// Browse 浏览模式：获取今日单词列表，按 Category 分组返回
func (s *studyService) Browse() (map[string][]*domain.Word, error) {
	// 1. 获取今日计划（含新词+复习词）
	plan, err := s.planSvc.GetTodayPlan()
	if err != nil {
		return nil, fmt.Errorf("获取今日计划失败: %w", err)
	}

	// 2. 合并新词和复习词（去重）
	seen := make(map[string]struct{})
	allWords := make([]*domain.Word, 0, len(plan.NewWords)+len(plan.ReviewWords))
	for _, w := range plan.NewWords {
		if _, ok := seen[w.ID]; !ok {
			seen[w.ID] = struct{}{}
			allWords = append(allWords, w)
		}
	}
	for _, w := range plan.ReviewWords {
		if _, ok := seen[w.ID]; !ok {
			seen[w.ID] = struct{}{}
			allWords = append(allWords, w)
		}
	}

	// 3. 按 Category 分组
	grouped := make(map[string][]*domain.Word)
	for _, w := range allWords {
		grouped[w.Category] = append(grouped[w.Category], w)
	}

	return grouped, nil
}

// MarkBrowse 浏览标记认识/不认识，写入 LearningRecord
func (s *studyService) MarkBrowse(wordID string, known bool) error {
	record := &domain.LearningRecord{
		WordID:  wordID,
		Mode:    domain.StudyModeBrowse,
		Correct: known,
		Date:    today(),
	}
	if err := s.recordRepo.Add(record); err != nil {
		return fmt.Errorf("写入浏览记录失败: %w", err)
	}
	return nil
}

// GenerateQuiz 生成测试题目（优先出错误率高的词）
func (s *studyService) GenerateQuiz(count int) ([]*QuizQuestion, error) {
	// 1. 获取今日计划单词
	plan, err := s.planSvc.GetTodayPlan()
	if err != nil {
		return nil, fmt.Errorf("获取今日计划失败: %w", err)
	}

	// 合并新词和复习词（去重）
	seen := make(map[string]struct{})
	allWords := make([]*domain.Word, 0, len(plan.NewWords)+len(plan.ReviewWords))
	for _, w := range plan.NewWords {
		if _, ok := seen[w.ID]; !ok {
			seen[w.ID] = struct{}{}
			allWords = append(allWords, w)
		}
	}
	for _, w := range plan.ReviewWords {
		if _, ok := seen[w.ID]; !ok {
			seen[w.ID] = struct{}{}
			allWords = append(allWords, w)
		}
	}

	if len(allWords) == 0 {
		return nil, fmt.Errorf("今日没有需要学习的单词")
	}

	// 2. 为每个单词计算错误率并按错误率降序排序，高错误率优先
	type wordWithStat struct {
		word      *domain.Word
		errorRate float64
	}
	wordsWithStats := make([]wordWithStat, 0, len(allWords))
	for _, w := range allWords {
		stats, err := s.reviewSvc.CalcWordStats(w.ID)
		if err != nil {
			// 计算失败时错误率为 0，仍正常参与排序
			wordsWithStats = append(wordsWithStats, wordWithStat{word: w, errorRate: 0})
			continue
		}
		wordsWithStats = append(wordsWithStats, wordWithStat{word: w, errorRate: stats.ErrorRate})
	}
	sort.Slice(wordsWithStats, func(i, j int) bool {
		return wordsWithStats[i].errorRate > wordsWithStats[j].errorRate
	})

	// 3. 取前 count 个词出题（不足 count 则全部出）
	if count > len(wordsWithStats) {
		count = len(wordsWithStats)
	}
	selectedWords := wordsWithStats[:count]

	// 4. 获取词库全部单词，用于生成干扰项
	// 从今日计划的新词中获取 libraryID（新词优先，无则取复习词）
	var libraryID string
	if len(plan.NewWords) > 0 {
		libraryID = plan.NewWords[0].LibraryID
	} else if len(plan.ReviewWords) > 0 {
		libraryID = plan.ReviewWords[0].LibraryID
	}

	var allLibraryWords []domain.Word
	if libraryID != "" {
		library, err := s.libraryRepo.GetByID(libraryID)
		if err == nil && library != nil {
			allLibraryWords = library.Words
		}
	}
	// 若未能获取词库单词，退化为使用今日单词作为干扰项候选
	if len(allLibraryWords) == 0 {
		for _, w := range allWords {
			allLibraryWords = append(allLibraryWords, *w)
		}
	}

	// 5. 为每道题生成4个选项
	questions := make([]*QuizQuestion, 0, len(selectedWords))
	for _, ws := range selectedWords {
		correctWord := ws.word

		// 收集干扰项候选（排除正确答案本身）
		distractors := make([]string, 0, len(allLibraryWords))
		for _, lw := range allLibraryWords {
			if lw.ID != correctWord.ID && lw.Chinese != correctWord.Chinese {
				distractors = append(distractors, lw.Chinese)
			}
		}

		// 随机打乱干扰项候选，取前3个
		rand.Shuffle(len(distractors), func(i, j int) {
			distractors[i], distractors[j] = distractors[j], distractors[i]
		})
		if len(distractors) > 3 {
			distractors = distractors[:3]
		}

		// 组合4个选项：1个正确答案 + 最多3个干扰项
		options := append([]string{correctWord.Chinese}, distractors...)

		// 随机打乱选项顺序
		rand.Shuffle(len(options), func(i, j int) {
			options[i], options[j] = options[j], options[i]
		})

		questions = append(questions, &QuizQuestion{
			WordID:   correctWord.ID,
			English:  correctWord.English,
			Phonetic: correctWord.Phonetic,
			Options:  options,
		})
	}

	return questions, nil
}

// SubmitQuizAnswer 提交测试答案，返回判题结果
func (s *studyService) SubmitQuizAnswer(wordID, answer string) (*QuizResult, error) {
	// 1. 获取今日计划，取 libraryID，再从词库中查找单词
	plan, err := s.planSvc.GetTodayPlan()
	if err != nil {
		return nil, fmt.Errorf("获取今日计划失败: %w", err)
	}

	word, err := s.getWordFromPlan(plan, wordID)
	if err != nil {
		return nil, err
	}

	// 2. 判断答案是否正确（trim 后比较）
	correct := strings.TrimSpace(answer) == strings.TrimSpace(word.Chinese)

	// 3. 写入学习记录
	record := &domain.LearningRecord{
		WordID:  wordID,
		Mode:    domain.StudyModeQuiz,
		Correct: correct,
		Date:    today(),
	}
	if err := s.recordRepo.Add(record); err != nil {
		return nil, fmt.Errorf("写入测试记录失败: %w", err)
	}

	// 4. 返回判题结果
	return &QuizResult{
		Correct:       correct,
		CorrectAnswer: word.Chinese,
		Example:       word.Example,
		ExampleCN:     word.ExampleCN,
	}, nil
}

// GetDictationWord 获取下一个默写词（优先返回错误率高的词）
func (s *studyService) GetDictationWord() (*DictationQuestion, error) {
	// 1. 获取今日全部单词（新词+复习词合并）
	plan, err := s.planSvc.GetTodayPlan()
	if err != nil {
		return nil, fmt.Errorf("获取今日计划失败: %w", err)
	}

	seen := make(map[string]struct{})
	allWords := make([]*domain.Word, 0, len(plan.NewWords)+len(plan.ReviewWords))
	for _, w := range plan.NewWords {
		if _, ok := seen[w.ID]; !ok {
			seen[w.ID] = struct{}{}
			allWords = append(allWords, w)
		}
	}
	for _, w := range plan.ReviewWords {
		if _, ok := seen[w.ID]; !ok {
			seen[w.ID] = struct{}{}
			allWords = append(allWords, w)
		}
	}

	if len(allWords) == 0 {
		return nil, fmt.Errorf("今日没有需要学习的单词")
	}

	// 2. 按错误率降序排序，取第一个
	type wordWithStat struct {
		word      *domain.Word
		errorRate float64
	}
	wordsWithStats := make([]wordWithStat, 0, len(allWords))
	for _, w := range allWords {
		stats, err := s.reviewSvc.CalcWordStats(w.ID)
		if err != nil {
			wordsWithStats = append(wordsWithStats, wordWithStat{word: w, errorRate: 0})
			continue
		}
		wordsWithStats = append(wordsWithStats, wordWithStat{word: w, errorRate: stats.ErrorRate})
	}
	sort.Slice(wordsWithStats, func(i, j int) bool {
		return wordsWithStats[i].errorRate > wordsWithStats[j].errorRate
	})

	// 3. 若所有词错误率为 0（全新词），随机返回一个
	var selected *domain.Word
	if wordsWithStats[0].errorRate > 0 {
		selected = wordsWithStats[0].word
	} else {
		// 全部错误率为 0，随机选取
		idx := rand.Intn(len(wordsWithStats))
		selected = wordsWithStats[idx].word
	}

	// 4. 返回默写题目（隐藏英文，只展示中文和例句）
	return &DictationQuestion{
		WordID:    selected.ID,
		Chinese:   selected.Chinese,
		Example:   selected.Example,
		ExampleCN: selected.ExampleCN,
	}, nil
}

// SubmitDictation 提交默写答案，返回判题结果
func (s *studyService) SubmitDictation(wordID, input string) (*DictationResult, error) {
	// 1. 获取今日计划，从词库中查找单词
	plan, err := s.planSvc.GetTodayPlan()
	if err != nil {
		return nil, fmt.Errorf("获取今日计划失败: %w", err)
	}

	word, err := s.getWordFromPlan(plan, wordID)
	if err != nil {
		return nil, err
	}

	// 2. 标准化处理：toLower 并 trim 空格
	normalizedInput := strings.ToLower(strings.TrimSpace(input))
	normalizedAnswer := strings.ToLower(word.English)

	// 3. 完全匹配判断
	correct := normalizedInput == normalizedAnswer

	// 4. 写入学习记录
	record := &domain.LearningRecord{
		WordID:  wordID,
		Mode:    domain.StudyModeDictation,
		Correct: correct,
		Date:    today(),
	}
	if err := s.recordRepo.Add(record); err != nil {
		return nil, fmt.Errorf("写入默写记录失败: %w", err)
	}

	// 5. 返回默写结果
	return &DictationResult{
		Correct:         correct,
		CorrectSpelling: word.English,
		Phonetic:        word.Phonetic,
	}, nil
}

// GetTodayProgress 获取今日三种学习模式的进度
// 浏览：返回每个单词的标记状态（同一单词以最后一条记录为准）
// 测试/默写：统计答题总数和答对数
func (s *studyService) GetTodayProgress() (*TodayProgress, error) {
	// 获取今日所有学习记录
	records, err := s.recordRepo.GetByDate(today())
	if err != nil {
		return nil, fmt.Errorf("获取今日学习记录失败: %w", err)
	}

	progress := &TodayProgress{
		BrowseMarks: make(map[string]bool),
	}

	// 遍历记录，按模式分类统计
	for _, r := range records {
		switch r.Mode {
		case domain.StudyModeBrowse:
			// 同一单词以最后一条记录为准（顺序追加，后写覆盖前写）
			progress.BrowseMarks[r.WordID] = r.Correct
		case domain.StudyModeQuiz:
			progress.QuizCount++
			if r.Correct {
				progress.QuizCorrect++
			}
		case domain.StudyModeDictation:
			progress.DictCount++
			if r.Correct {
				progress.DictCorrect++
			}
		}
	}

	return progress, nil
}

// getWordFromPlan 从今日计划中通过词库查找指定单词
// 先从今日计划的词获取 libraryID，再调用词库查询
func (s *studyService) getWordFromPlan(plan *TodayPlan, wordID string) (*domain.Word, error) {
	// 从今日计划的单词中获取 libraryID
	var libraryID string
	if len(plan.NewWords) > 0 {
		libraryID = plan.NewWords[0].LibraryID
	} else if len(plan.ReviewWords) > 0 {
		libraryID = plan.ReviewWords[0].LibraryID
	}

	if libraryID == "" {
		return nil, fmt.Errorf("无法确定词库 ID")
	}

	library, err := s.libraryRepo.GetByID(libraryID)
	if err != nil {
		return nil, fmt.Errorf("获取词库失败: %w", err)
	}
	if library == nil {
		return nil, fmt.Errorf("词库 %s 不存在", libraryID)
	}

	// 在词库中查找指定单词
	for i := range library.Words {
		if library.Words[i].ID == wordID {
			return &library.Words[i], nil
		}
	}

	return nil, fmt.Errorf("单词 %s 在词库中不存在", wordID)
}
