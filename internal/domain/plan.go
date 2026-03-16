package domain

// PlanStatus 计划状态
type PlanStatus string

const (
	PlanStatusPending   PlanStatus = "pending"
	PlanStatusLearning  PlanStatus = "learning"
	PlanStatusCompleted PlanStatus = "completed"
)

// DailyPlan 每日学习计划
type DailyPlan struct {
	Day           int        `json:"day"`
	Date          string     `json:"date"`           // yyyy-MM-dd
	WordIDs       []string   `json:"word_ids"`
	ReviewWordIDs []string   `json:"review_word_ids"`
	Status        PlanStatus `json:"status"`
}

// StudyPlan 整体14天学习计划
type StudyPlan struct {
	LibraryID string      `json:"library_id"`
	StartDate string      `json:"start_date"` // yyyy-MM-dd
	Days      []DailyPlan `json:"days"`
}
