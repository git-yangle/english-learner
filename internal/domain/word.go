package domain

// Word 单词领域模型
type Word struct {
	ID        string `json:"id"`
	English   string `json:"english"`
	Phonetic  string `json:"phonetic"`
	Chinese   string `json:"chinese"`
	Example   string `json:"example"`
	ExampleCN string `json:"example_cn"`
	Category  string `json:"category"`
	LibraryID string `json:"library_id"`
}

// WordLibrary 词库领域模型
type WordLibrary struct {
	ID    string `json:"id"`
	Name  string `json:"name"`
	Desc  string `json:"desc"`
	Words []Word `json:"words"`
}

// 场景分类常量
const (
	CategoryAirport        = "airport"
	CategoryHotel          = "hotel"
	CategoryRestaurant     = "restaurant"
	CategoryDirection      = "direction"
	CategoryShopping       = "shopping"
	CategoryEmergency      = "emergency"
	CategoryTransportation = "transportation"
	CategoryEntertainment  = "entertainment"
)
