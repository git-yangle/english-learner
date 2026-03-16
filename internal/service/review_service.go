package service

import (
	"english-learner/internal/domain"
	"english-learner/internal/repository"
)

// ReviewService 复习词筛选服务接口
type ReviewService interface {
	// GetReviewWords 根据历史错误率，计算指定日期需要复习的单词ID列表
	// 策略：错误率>60% 次日必复习；30-60% 2天后；<30% 艾宾浩斯间隔
	GetReviewWords(date string) ([]string, error)
	// CalcWordStats 计算单词的学习统计数据
	CalcWordStats(wordID string) (*domain.WordStats, error)
}

type reviewService struct {
	recordRepo repository.RecordRepository
}

// NewReviewService 创建复习服务实例
func NewReviewService(recordRepo repository.RecordRepository) ReviewService {
	return &reviewService{recordRepo: recordRepo}
}

// GetReviewWords 根据历史错误率计算需要复习的单词列表
func (s *reviewService) GetReviewWords(date string) ([]string, error) {
	// TODO:
	// 1. 读取所有 LearningRecord，按 WordID 分组统计错误率
	// 2. 错误率 > 60%  → 次日必复习（距今1天）
	// 3. 错误率 30-60% → 距上次学习 >= 2天 则加入
	// 4. 错误率 < 30%  → 艾宾浩斯间隔（1,2,4,7,15天）加入
	// 5. 返回需复习的 WordID 列表
	return nil, nil
}

// CalcWordStats 计算单词的学习统计数据
func (s *reviewService) CalcWordStats(wordID string) (*domain.WordStats, error) {
	// TODO:
	// 1. 调用 recordRepo.GetByWordID(wordID) 获取所有历史记录
	// 2. 统计总次数、正确次数、计算错误率、最后学习日期
	// 3. 返回 WordStats
	return nil, nil
}
