package domain

// StudyMode 学习模式
type StudyMode string

const (
	StudyModeBrowse    StudyMode = "browse"
	StudyModeQuiz      StudyMode = "quiz"
	StudyModeDictation StudyMode = "dictation"
)

// LearningRecord 每次作答记录
type LearningRecord struct {
	WordID   string    `json:"word_id"`
	Mode     StudyMode `json:"mode"`
	Correct  bool      `json:"correct"`
	Attempts int       `json:"attempts"`
	Date     string    `json:"date"` // yyyy-MM-dd
}

// CheckIn 每日打卡汇总
type CheckIn struct {
	Day              int     `json:"day"`
	Date             string  `json:"date"`
	Completed        bool    `json:"completed"`
	Score            float64 `json:"score"`             // 综合正确率
	StudyDurationSec int     `json:"study_duration_sec"`
}

// WordStats 单词学习统计（计算得出，不持久化）
type WordStats struct {
	WordID       string
	TotalCount   int
	CorrectCount int
	ErrorRate    float64 // 0.0-1.0
	LastStudied  string  // yyyy-MM-dd
}
