package service

import (
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

// Browse 浏览模式：获取今日单词列表，按分类分组返回
func (s *studyService) Browse() (map[string][]*domain.Word, error) {
	// TODO:
	// 1. 调用 planSvc.GetTodayPlan() 获取今日单词（新词+复习词）
	// 2. 按 Category 分组，返回 map[category][]Word
	return nil, nil
}

// MarkBrowse 浏览标记认识/不认识，写入 LearningRecord
func (s *studyService) MarkBrowse(wordID string, known bool) error {
	// TODO:
	// 1. 构造 LearningRecord{WordID, Mode: browse, Correct: known, Date: today}
	// 2. 调用 recordRepo.Add() 写入记录
	return nil
}

// GenerateQuiz 生成测试题目（优先出错误率高的词）
func (s *studyService) GenerateQuiz(count int) ([]*QuizQuestion, error) {
	// TODO:
	// 1. 获取今日单词列表
	// 2. 按错误率排序，优先选高错误率的词
	// 3. 每题生成4个选项：1个正确答案 + 3个同词库随机干扰项
	// 4. 打乱选项顺序，返回题目列表
	return nil, nil
}

// SubmitQuizAnswer 提交测试答案，返回判题结果
func (s *studyService) SubmitQuizAnswer(wordID, answer string) (*QuizResult, error) {
	// TODO:
	// 1. 根据 wordID 获取单词，对比 answer 与正确中文释义
	// 2. 写入 LearningRecord{Mode: quiz, Correct, WordID, Date: today}
	// 3. 返回 QuizResult（含是否正确、正确答案、例句）
	return nil, nil
}

// GetDictationWord 获取下一个默写词（优先返回错误率高的词）
func (s *studyService) GetDictationWord() (*DictationQuestion, error) {
	// TODO:
	// 1. 获取今日单词列表，按错误率排序
	// 2. 返回最高错误率单词的中文释义和例句（隐藏英文拼写）
	return nil, nil
}

// SubmitDictation 提交默写答案，返回判题结果
func (s *studyService) SubmitDictation(wordID, input string) (*DictationResult, error) {
	// TODO:
	// 1. 标准化处理：toLower、trim空格
	// 2. 与标准答案 exact match 对比
	// 3. 写入 LearningRecord{Mode: dictation, Correct, WordID, Date: today}
	// 4. 返回 DictationResult（含是否正确、标准拼写、音标）
	return nil, nil
}
