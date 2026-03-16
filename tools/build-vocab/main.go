// build-vocab 词库构建工具
// 数据来源：
//   - Oxford 3000 词频基础（BNC + Oxford Corpus 语料库统计）
//   - ECDICT 英汉词典（中文翻译、IPA音标、BNC词频）
//   - 内置旅游专项高频短语（覆盖 Oxford 3000 未收录的旅游复合词）
//
// 用法：go run tools/build-vocab/main.go -oxford /tmp/oxford3000.csv -ecdict /tmp/ecdict.csv -out data/libraries/travel.json

package main

import (
	"encoding/csv"
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"log"
	"os"
	"strings"
)

// WordEntry ECDICT 词条结构
type WordEntry struct {
	Word        string
	Phonetic    string
	Translation string
	Oxford      string // 是否 Oxford 词典收录
	BNC         string // BNC 词频
}

// OutputWord 最终输出词条结构
type OutputWord struct {
	ID        string `json:"id"`
	English   string `json:"english"`
	Phonetic  string `json:"phonetic"`
	Chinese   string `json:"chinese"`
	Example   string `json:"example"`
	ExampleCN string `json:"example_cn"`
	Category  string `json:"category"`
	Source    string `json:"source"` // oxford3000 / travel_phrase
}

// OutputLibrary 输出词库结构
type OutputLibrary struct {
	ID    string       `json:"id"`
	Name  string       `json:"name"`
	Desc  string       `json:"desc"`
	Words []OutputWord `json:"words"`
}

// categoryRules 已废弃，仅保留结构供后续扩展
// 当前使用精确手工标注（oxford3000TravelCategory），不再依赖宽泛关键词匹配
var categoryRules = map[string][]string{}

// oxford3000TravelCategory Oxford 3000 旅游相关词精确手工标注表
// 只收录真正与旅游场景相关的词，确保分类准确
var oxford3000TravelCategory = map[string]string{
	// ===== airport 机场 =====
	"abroad":        "airport",
	"air":           "airport",
	"aircraft":      "airport",
	"airline":       "airport",
	"airport":       "airport",
	"arrival":       "airport",
	"arrive":        "airport",
	"bag":           "airport",
	"baggage":       "airport",
	"border":        "airport",
	"cancel":        "airport",
	"carrier":       "airport",
	"connect":       "airport",
	"customs":       "airport",
	"delay":         "airport",
	"depart":        "airport",
	"departure":     "airport",
	"destination":   "airport",
	"duty":          "airport",
	"fare":          "airport",
	"flight":        "airport",
	"foreign":       "airport",
	"gate":          "airport",
	"immigration":   "airport",
	"international": "airport",
	"jet":           "airport",
	"journey":       "airport",
	"land":          "airport",
	"luggage":       "airport",
	"miss":          "airport",
	"nationality":   "airport",
	"pack":          "airport",
	"passport":      "airport",
	"pilot":         "airport",
	"plane":         "airport",
	"runway":        "airport",
	"suitcase":      "airport",
	"terminal":      "airport",
	"travel":        "airport",
	"trip":          "airport",
	"visa":          "airport",
	"voyage":        "airport",

	// ===== hotel 酒店 =====
	"accommodation": "hotel",
	"apartment":     "hotel",
	"bed":           "hotel",
	"book":          "hotel",
	"clean":         "hotel",
	"complaint":     "hotel",
	"confirm":       "hotel",
	"dirty":         "hotel",
	"elevator":      "hotel",
	"floor":         "hotel",
	"hotel":         "hotel",
	"inn":           "hotel",
	"internet":      "hotel",
	"key":           "hotel",
	"lift":          "hotel",
	"lobby":         "hotel",
	"lodge":         "hotel",
	"overnight":     "hotel",
	"reception":     "hotel",
	"reserve":       "hotel",
	"resort":        "hotel",
	"room":          "hotel",
	"service":       "hotel",
	"shower":        "hotel",
	"single":        "hotel",
	"stay":          "hotel",
	"suite":         "hotel",
	"temperature":   "hotel",
	"towel":         "hotel",
	"welcome":       "hotel",
	"wifi":          "hotel",

	// ===== restaurant 餐厅 =====
	"alcohol":    "restaurant",
	"bake":       "restaurant",
	"bar":        "restaurant",
	"beef":       "restaurant",
	"beer":       "restaurant",
	"bill":       "restaurant",
	"bottle":     "restaurant",
	"bread":      "restaurant",
	"breakfast":  "restaurant",
	"cake":       "restaurant",
	"cheese":     "restaurant",
	"chef":       "restaurant",
	"chicken":    "restaurant",
	"chocolate":  "restaurant",
	"coffee":     "restaurant",
	"cold":       "restaurant",
	"cook":       "restaurant",
	"cuisine":    "restaurant",
	"dessert":    "restaurant",
	"diet":       "restaurant",
	"dine":       "restaurant",
	"dinner":     "restaurant",
	"dish":       "restaurant",
	"drink":      "restaurant",
	"eat":        "restaurant",
	"egg":        "restaurant",
	"fish":       "restaurant",
	"food":       "restaurant",
	"fresh":      "restaurant",
	"fruit":      "restaurant",
	"grill":      "restaurant",
	"hungry":     "restaurant",
	"juice":      "restaurant",
	"lunch":      "restaurant",
	"meal":       "restaurant",
	"meat":       "restaurant",
	"menu":       "restaurant",
	"milk":       "restaurant",
	"oil":        "restaurant",
	"order":      "restaurant",
	"potato":     "restaurant",
	"recommend":  "restaurant",
	"restaurant": "restaurant",
	"rice":       "restaurant",
	"salad":      "restaurant",
	"salt":       "restaurant",
	"sauce":      "restaurant",
	"snack":      "restaurant",
	"soup":       "restaurant",
	"spicy":      "restaurant",
	"sugar":      "restaurant",
	"sweet":      "restaurant",
	"taste":      "restaurant",
	"tea":        "restaurant",
	"thirsty":    "restaurant",
	"tip":        "restaurant",
	"vegetable":  "restaurant",
	"waiter":     "restaurant",
	"water":      "restaurant",
	"wine":       "restaurant",

	// ===== direction 问路/方向 =====
	"across":    "direction",
	"address":   "direction",
	"along":     "direction",
	"avenue":    "direction",
	"behind":    "direction",
	"beside":    "direction",
	"between":   "direction",
	"bridge":    "direction",
	"centre":    "direction",
	"city":      "direction",
	"corner":    "direction",
	"cross":     "direction",
	"direction": "direction",
	"distance":  "direction",
	"downtown":  "direction",
	"east":      "direction",
	"entrance":  "direction",
	"exit":      "direction",
	"far":       "direction",
	"go":        "direction",
	"highway":   "direction",
	"hurry":     "direction",
	"left":      "direction",
	"locate":    "direction",
	"location":  "direction",
	"map":       "direction",
	"nearby":    "direction",
	"north":     "direction",
	"opposite":  "direction",
	"path":      "direction",
	"right":     "direction",
	"road":      "direction",
	"route":     "direction",
	"sign":      "direction",
	"south":     "direction",
	"straight":  "direction",
	"street":    "direction",
	"tunnel":    "direction",
	"turn":      "direction",
	"walk":      "direction",
	"way":       "direction",
	"west":      "direction",

	// ===== shopping 购物 =====
	"afford":   "shopping",
	"amount":   "shopping",
	"bank":     "shopping",
	"bargain":  "shopping",
	"buy":      "shopping",
	"cash":     "shopping",
	"charge":   "shopping",
	"cheap":    "shopping",
	"coin":     "shopping",
	"cost":     "shopping",
	"credit":   "shopping",
	"currency": "shopping",
	"deal":     "shopping",
	"deliver":  "shopping",
	"discount": "shopping",
	"exchange": "shopping",
	"expensive": "shopping",
	"free":     "shopping",
	"gift":     "shopping",
	"market":   "shopping",
	"money":    "shopping",
	"pay":      "shopping",
	"payment":  "shopping",
	"price":    "shopping",
	"purchase": "shopping",
	"receipt":  "shopping",
	"refund":   "shopping",
	"sale":     "shopping",
	"sell":     "shopping",
	"shop":     "shopping",
	"shopping": "shopping",
	"souvenir": "shopping",
	"spend":    "shopping",
	"store":    "shopping",
	"tax":      "shopping",

	// ===== emergency 紧急情况 =====
	"accident":  "emergency",
	"aid":       "emergency",
	"allergy":   "emergency",
	"ambulance": "emergency",
	"bleed":     "emergency",
	"broken":    "emergency",
	"danger":    "emergency",
	"disease":   "emergency",
	"doctor":    "emergency",
	"emergency": "emergency",
	"fever":     "emergency",
	"health":    "emergency",
	"help":      "emergency",
	"hospital":  "emergency",
	"hurt":      "emergency",
	"ill":       "emergency",
	"illness":   "emergency",
	"injection": "emergency",
	"injury":    "emergency",
	"insurance": "emergency",
	"lost":      "emergency",
	"medicine":  "emergency",
	"nurse":     "emergency",
	"pain":      "emergency",
	"pharmacy":  "emergency",
	"police":    "emergency",
	"protect":   "emergency",
	"safe":      "emergency",
	"safety":    "emergency",
	"sick":      "emergency",
	"steal":     "emergency",
	"theft":     "emergency",

	// ===== transportation 交通 =====
	"bicycle":       "transportation",
	"boat":          "transportation",
	"bus":           "transportation",
	"cab":           "transportation",
	"car":           "transportation",
	"coach":         "transportation",
	"cycle":         "transportation",
	"drive":         "transportation",
	"driver":        "transportation",
	"ferry":         "transportation",
	"harbour":       "transportation",
	"hire":          "transportation",
	"metro":         "transportation",
	"motor":         "transportation",
	"motorcycle":    "transportation",
	"passenger":     "transportation",
	"platform":      "transportation",
	"port":          "transportation",
	"rent":          "transportation",
	"schedule":      "transportation",
	"seat":          "transportation",
	"ship":          "transportation",
	"station":       "transportation",
	"subway":        "transportation",
	"taxi":          "transportation",
	"ticket":        "transportation",
	"timetable":     "transportation",
	"train":         "transportation",
	"transfer":      "transportation",
	"tram":          "transportation",
	"transport":     "transportation",
	"underground":   "transportation",
	"vehicle":       "transportation",

	// ===== entertainment 景点娱乐 =====
	"art":         "entertainment",
	"attraction":  "entertainment",
	"camera":      "entertainment",
	"cathedral":   "entertainment",
	"cinema":      "entertainment",
	"climb":       "entertainment",
	"concert":     "entertainment",
	"crowded":     "entertainment",
	"culture":     "entertainment",
	"dance":       "entertainment",
	"desert":      "entertainment",
	"exhibit":     "entertainment",
	"exhibition":  "entertainment",
	"explore":     "entertainment",
	"festival":    "entertainment",
	"forest":      "entertainment",
	"gallery":     "entertainment",
	"guide":       "entertainment",
	"hike":        "entertainment",
	"hill":        "entertainment",
	"historic":    "entertainment",
	"history":     "entertainment",
	"island":      "entertainment",
	"lake":        "entertainment",
	"language":    "entertainment",
	"local":       "entertainment",
	"monument":    "entertainment",
	"mountain":    "entertainment",
	"museum":      "entertainment",
	"music":       "entertainment",
	"nature":      "entertainment",
	"ocean":       "entertainment",
	"park":        "entertainment",
	"photo":       "entertainment",
	"photograph":  "entertainment",
	"picnic":      "entertainment",
	"river":       "entertainment",
	"scenery":     "entertainment",
	"sea":         "entertainment",
	"sightseeing": "entertainment",
	"sport":       "entertainment",
	"swim":        "entertainment",
	"temple":      "entertainment",
	"theater":     "entertainment",
	"tour":        "entertainment",
	"tourist":     "entertainment",
	"tradition":   "entertainment",
	"valley":      "entertainment",
	"view":        "entertainment",
	"visit":       "entertainment",
	"waterfall":   "entertainment",
	"weather":     "entertainment",
}

// 旅游专项高频短语（Oxford 3000 未收录，但旅途必用）
// 每条：英文、音标、中文、例句、例句翻译、场景
var travelPhrases = [][]string{
	// 机场场景
	{"boarding pass", "/ˈbɔːrdɪŋ pæs/", "登机牌", "Please show your boarding pass at the gate.", "请在登机口出示您的登机牌。", "airport"},
	{"carry-on baggage", "/ˈkæri ɒn ˈbægɪdʒ/", "随身行李", "Carry-on baggage must fit in the overhead bin.", "随身行李必须能放入头顶行李架。", "airport"},
	{"checked baggage", "/tʃekt ˈbægɪdʒ/", "托运行李", "There is a fee for checked baggage over 23kg.", "超过23公斤的托运行李需要收费。", "airport"},
	{"baggage claim", "/ˈbægɪdʒ kleɪm/", "行李认领处", "Please proceed to baggage claim area 3.", "请前往第三号行李认领处。", "airport"},
	{"duty-free shop", "/ˌdjuːti ˈfriː ʃɒp/", "免税店", "You can buy perfume at the duty-free shop.", "你可以在免税店购买香水。", "airport"},
	{"connecting flight", "/kəˈnektɪŋ flaɪt/", "转机航班", "My connecting flight departs in two hours.", "我的转机航班两小时后起飞。", "airport"},
	{"window seat", "/ˈwɪndəʊ siːt/", "靠窗座位", "I prefer a window seat on long flights.", "长途飞行我更喜欢靠窗座位。", "airport"},
	{"aisle seat", "/ˈaɪl siːt/", "靠通道座位", "Can I have an aisle seat, please?", "请问我可以要靠过道的座位吗？", "airport"},
	{"overhead bin", "/ˌəʊvəhed ˈbɪn/", "头顶行李架", "Please place your bag in the overhead bin.", "请将您的包放入头顶行李架。", "airport"},
	{"security check", "/sɪˈkjʊərɪti tʃek/", "安全检查", "Please remove your shoes for the security check.", "安全检查时请脱下鞋子。", "airport"},
	{"baggage allowance", "/ˈbægɪdʒ əˈlaʊəns/", "行李限额", "The baggage allowance is 20 kilograms.", "行李限额是20公斤。", "airport"},
	{"flight number", "/flaɪt ˈnʌmbər/", "航班号", "What is your flight number?", "您的航班号是多少？", "airport"},
	{"departure lounge", "/dɪˈpɑːtʃər laʊndʒ/", "候机室", "Please wait in the departure lounge.", "请在候机室等候。", "airport"},
	{"entry card", "/ˈentri kɑːrd/", "入境卡", "Please fill in the entry card before landing.", "请在降落前填写入境卡。", "airport"},
	{"jet lag", "/ˈdʒet læɡ/", "时差反应", "I'm suffering from jet lag after my long flight.", "长途飞行后我有时差反应。", "airport"},
	{"transit visa", "/ˈtrænsɪt ˈviːzə/", "过境签证", "Do I need a transit visa for this stopover?", "这次中转我需要过境签证吗？", "airport"},
	{"customs declaration", "/ˈkʌstəmz ˌdekləˈreɪʃn/", "海关申报", "Please complete the customs declaration form.", "请填写海关申报表。", "airport"},
	{"boarding time", "/ˈbɔːrdɪŋ taɪm/", "登机时间", "Boarding time is 30 minutes before departure.", "登机时间在起飞前30分钟。", "airport"},
	{"excess baggage", "/ˈekses ˈbægɪdʒ/", "超重行李", "You have to pay for excess baggage.", "您需要为超重行李付费。", "airport"},
	{"passport control", "/ˈpɑːspɔːrt kənˈtrəʊl/", "护照检查", "Please queue at passport control.", "请在护照检查处排队。", "airport"},
	{"round trip", "/ˌraʊnd ˈtrɪp/", "往返行程", "I'd like to book a round trip ticket.", "我想预订一张往返机票。", "airport"},
	{"one way", "/ˌwʌn ˈweɪ/", "单程", "Is this a one-way ticket or round trip?", "这是单程票还是往返票？", "airport"},
	{"hand luggage", "/hænd ˈlʌɡɪdʒ/", "手提行李", "Only one piece of hand luggage is allowed.", "只允许携带一件手提行李。", "airport"},
	{"direct flight", "/dəˈrekt flaɪt/", "直飞航班", "Is there a direct flight to Paris?", "有直飞巴黎的航班吗？", "airport"},
	{"layover", "/ˈleɪəʊvər/", "中转停留", "We have a three-hour layover in Dubai.", "我们在迪拜有三小时中转停留。", "airport"},

	// 酒店场景
	{"check-in", "/ˌtʃek ˈɪn/", "入住登记", "Check-in time is 3 PM.", "入住时间是下午3点。", "hotel"},
	{"check-out", "/ˌtʃek ˈaʊt/", "退房", "Check-out time is 12 noon.", "退房时间是中午12点。", "hotel"},
	{"room service", "/ˈruːm ˌsɜːrvɪs/", "客房服务", "I'd like to order room service, please.", "我想点客房服务。", "hotel"},
	{"front desk", "/frʌnt desk/", "前台", "Please ask at the front desk.", "请在前台询问。", "hotel"},
	{"single room", "/ˈsɪŋɡl ruːm/", "单人房", "I need a single room for two nights.", "我需要一间单人房住两晚。", "hotel"},
	{"double room", "/ˈdʌbl ruːm/", "双人房（大床）", "A double room with a sea view, please.", "请给我一间海景双人大床房。", "hotel"},
	{"twin room", "/twɪn ruːm/", "双人房（双床）", "We'd like a twin room with two single beds.", "我们想要一间有两张单人床的双床房。", "hotel"},
	{"suite", "/swiːt/", "套房", "The presidential suite on the top floor.", "顶层总统套房。", "hotel"},
	{"keycard", "/ˈkiːkɑːrd/", "房卡", "Here is your keycard for room 302.", "这是您302房间的房卡。", "hotel"},
	{"late checkout", "/leɪt ˈtʃekaʊt/", "延迟退房", "Can I request a late checkout?", "我可以申请延迟退房吗？", "hotel"},
	{"do not disturb", "/duː nɒt dɪˈstɜːrb/", "请勿打扰", "Please hang the do not disturb sign on the door.", "请把请勿打扰的牌子挂在门上。", "hotel"},
	{"minibar", "/ˈmɪnibɑːr/", "小冰箱（迷你吧）", "Drinks from the minibar will be charged.", "迷你吧里的饮品需要收费。", "hotel"},
	{"complimentary breakfast", "/ˌkɒmplɪˈmentri ˈbrekfəst/", "免费早餐", "Complimentary breakfast is served from 7 to 10.", "免费早餐供应时间是7点到10点。", "hotel"},
	{"housekeeping", "/ˈhaʊskiːpɪŋ/", "客房清洁", "Can housekeeping clean my room now?", "客房服务现在可以打扫我的房间吗？", "hotel"},
	{"wake-up call", "/ˈweɪk ʌp kɔːl/", "叫醒服务", "Can I have a wake-up call at 7 AM?", "我可以安排早上7点的叫醒服务吗？", "hotel"},
	{"safe deposit box", "/seɪf dɪˈpɒzɪt bɒks/", "保险箱", "I'd like to use the safe deposit box.", "我想使用保险箱。", "hotel"},
	{"laundry service", "/ˈlɔːndri ˌsɜːrvɪs/", "洗衣服务", "Does the hotel offer laundry service?", "酒店提供洗衣服务吗？", "hotel"},
	{"continental breakfast", "/ˌkɒntɪˈnentl ˈbrekfəst/", "欧式早餐", "The room rate includes continental breakfast.", "房费包含欧式早餐。", "hotel"},
	{"air conditioning", "/ˌeər kənˈdɪʃənɪŋ/", "空调", "How do I adjust the air conditioning?", "怎么调节空调？", "hotel"},
	{"room upgrade", "/ruːm ˈʌpɡreɪd/", "升级房间", "Could you upgrade my room, please?", "请问可以帮我升级房间吗？", "hotel"},
	{"non-smoking room", "/nɒn ˈsməʊkɪŋ ruːm/", "禁烟房", "I'd like a non-smoking room on a high floor.", "我想要高楼层的禁烟房。", "hotel"},
	{"fitness center", "/ˈfɪtnəs ˈsentər/", "健身中心", "Where is the fitness center?", "健身中心在哪里？", "hotel"},
	{"concierge", "/ˈkɒnsɪerʒ/", "礼宾员", "Please ask the concierge for restaurant recommendations.", "请向礼宾员询问餐厅推荐。", "hotel"},
	{"bellhop", "/ˈbelhɒp/", "行李员", "The bellhop will bring your luggage up.", "行李员会把您的行李送上来。", "hotel"},
	{"valet parking", "/ˈvæleɪ ˈpɑːrkɪŋ/", "代客泊车", "Does the hotel offer valet parking?", "酒店提供代客泊车服务吗？", "hotel"},

	// 餐厅场景
	{"set menu", "/set ˈmenjuː/", "套餐", "I'll have the set menu for lunch.", "我午餐点套餐。", "restaurant"},
	{"à la carte", "/ˌɑː lə ˈkɑːrt/", "单点（单品菜单）", "Would you like to order à la carte?", "您想单点还是套餐？", "restaurant"},
	{"daily special", "/ˈdeɪli ˈspeʃl/", "每日特供", "What is today's daily special?", "今天的每日特供是什么？", "restaurant"},
	{"house specialty", "/haʊs spəˈʃælɪti/", "招牌菜", "What is the house specialty?", "这里有什么招牌菜？", "restaurant"},
	{"appetizer", "/ˈæpɪtaɪzər/", "开胃菜", "I'd like the soup as an appetizer.", "我想点汤作为开胃菜。", "restaurant"},
	{"main course", "/ˌmeɪn ˈkɔːrs/", "主菜", "What do you recommend for the main course?", "主菜您有什么推荐？", "restaurant"},
	{"side dish", "/ˈsaɪd dɪʃ/", "配菜", "Can I substitute the side dish?", "我可以换一个配菜吗？", "restaurant"},
	{"wine list", "/ˈwaɪn lɪst/", "酒单", "May I see the wine list, please?", "请问可以看看酒单吗？", "restaurant"},
	{"sparkling water", "/ˌspɑːklɪŋ ˈwɔːtər/", "气泡水", "Still or sparkling water?", "要普通水还是气泡水？", "restaurant"},
	{"still water", "/stɪl ˈwɔːtər/", "矿泉水（非气泡）", "I'll have still water, thank you.", "我要普通水，谢谢。", "restaurant"},
	{"dietary restriction", "/ˈdaɪɪtri rɪˈstrɪkʃn/", "饮食限制", "I have some dietary restrictions.", "我有一些饮食限制。", "restaurant"},
	{"food allergy", "/fuːd ˈælərdʒi/", "食物过敏", "I have a nut food allergy.", "我对坚果食物过敏。", "restaurant"},
	{"vegetarian", "/ˌvedʒɪˈteəriən/", "素食者", "Do you have any vegetarian options?", "你们有素食选项吗？", "restaurant"},
	{"well done", "/ˌwel ˈdʌn/", "全熟（牛排）", "I'd like my steak well done.", "我的牛排要全熟的。", "restaurant"},
	{"medium rare", "/ˌmiːdiəm ˈreər/", "五分熟", "I prefer my steak medium rare.", "我喜欢五分熟的牛排。", "restaurant"},
	{"to go", "/tə ˈɡəʊ/", "打包带走", "Can I get this to go, please?", "这个可以打包吗？", "restaurant"},
	{"split the bill", "/splɪt ðə bɪl/", "AA制", "Can we split the bill?", "我们可以各付各的吗？", "restaurant"},
	{"service charge", "/ˈsɜːrvɪs tʃɑːrdʒ/", "服务费", "Is the service charge included?", "包含服务费吗？", "restaurant"},
	{"table for two", "/ˈteɪbl fər tuː/", "两人桌", "A table for two, please.", "请给我安排一张两人桌。", "restaurant"},
	{"waiting list", "/ˈweɪtɪŋ lɪst/", "等位名单", "How long is the waiting list?", "等位名单还有多久？", "restaurant"},
	{"doggy bag", "/ˈdɒɡi bæɡ/", "打包袋", "Can I have a doggy bag for the leftovers?", "剩下的菜可以帮我打包吗？", "restaurant"},
	{"all-you-can-eat", "/ˌɔːl jə kən ˈiːt/", "自助餐（无限量）", "This is an all-you-can-eat buffet.", "这是无限量自助餐。", "restaurant"},
	{"tap water", "/tæp ˈwɔːtər/", "自来水", "Can I have tap water please?", "请给我一杯自来水好吗？", "restaurant"},
	{"seafood", "/ˈsiːfuːd/", "海鲜", "I love fresh seafood by the beach.", "我喜欢海边的新鲜海鲜。", "restaurant"},
	{"gluten-free", "/ˌɡluːtn ˈfriː/", "无麸质", "Do you have any gluten-free options?", "你们有无麸质的选项吗？", "restaurant"},

	// 问路/方向
	{"excuse me", "/ɪkˈskjuːz miː/", "打扰一下", "Excuse me, how do I get to the station?", "打扰一下，请问怎么去车站？", "direction"},
	{"turn left", "/tɜːrn left/", "左转", "Turn left at the traffic lights.", "在红绿灯处左转。", "direction"},
	{"turn right", "/tɜːrn raɪt/", "右转", "Turn right at the next corner.", "在下一个路口右转。", "direction"},
	{"go straight", "/ɡəʊ streɪt/", "直走", "Go straight for two blocks.", "直走两个街区。", "direction"},
	{"traffic light", "/ˈtræfɪk laɪt/", "红绿灯", "Wait until the traffic light turns green.", "等红绿灯变绿再走。", "direction"},
	{"pedestrian crossing", "/pəˈdestriən ˈkrɒsɪŋ/", "人行横道", "Use the pedestrian crossing to cross the road.", "走人行横道过马路。", "direction"},
	{"bus stop", "/ˈbʌs stɒp/", "公交站", "The bus stop is just around the corner.", "公交站就在拐角处。", "direction"},
	{"subway station", "/ˈsʌbweɪ ˈsteɪʃn/", "地铁站", "Is there a subway station nearby?", "附近有地铁站吗？", "direction"},
	{"walking distance", "/ˈwɔːkɪŋ ˈdɪstəns/", "步行距离", "The hotel is within walking distance.", "酒店在步行范围内。", "direction"},
	{"GPS navigation", "/ˌdʒiː piː ˈes ˌnævɪˈɡeɪʃn/", "GPS导航", "Let me check the GPS navigation.", "让我查一下GPS导航。", "direction"},
	{"roundabout", "/ˈraʊndəbaʊt/", "环形交叉路口", "Take the second exit at the roundabout.", "在环形交叉路口走第二个出口。", "direction"},
	{"city center", "/ˈsɪti ˌsentər/", "市中心", "How far is it to the city center?", "离市中心有多远？", "direction"},
	{"on foot", "/ɒn fʊt/", "步行", "Can we get there on foot?", "我们可以步行到那里吗？", "direction"},
	{"local bus", "/ˈləʊkl bʌs/", "本地公交", "Which local bus goes to the old town?", "哪路本地公交去老城区？", "direction"},
	{"how far", "/haʊ fɑːr/", "有多远", "How far is it from here?", "离这里有多远？", "direction"},

	// 购物场景
	{"shopping mall", "/ˈʃɒpɪŋ mɔːl/", "购物中心", "The shopping mall is open until 10 PM.", "购物中心营业到晚上10点。", "shopping"},
	{"department store", "/dɪˈpɑːrtmənt stɔːr/", "百货公司", "There's a big department store downtown.", "市中心有一家大型百货公司。", "shopping"},
	{"fitting room", "/ˈfɪtɪŋ ruːm/", "试衣间", "Where is the fitting room?", "试衣间在哪里？", "shopping"},
	{"on sale", "/ɒn seɪl/", "打折中", "Everything in this store is on sale.", "这家店所有商品都在打折。", "shopping"},
	{"price tag", "/ˈpraɪs tæɡ/", "价格标签", "The price tag says 50 dollars.", "价格标签上写着50美元。", "shopping"},
	{"tax-free shopping", "/ˌtæks friː ˈʃɒpɪŋ/", "免税购物", "Can I get a tax refund here?", "这里可以退税吗？", "shopping"},
	{"VAT refund", "/ˌviː eɪ ˈtiː ˈriːfʌnd/", "增值税退税", "How do I claim a VAT refund?", "如何申请增值税退税？", "shopping"},
	{"cash register", "/kæʃ ˈredʒɪstər/", "收银台", "Please pay at the cash register.", "请到收银台付款。", "shopping"},
	{"contactless payment", "/ˈkɒntæktləs ˈpeɪmənt/", "非接触式支付", "Do you accept contactless payment?", "你们接受非接触式支付吗？", "shopping"},
	{"exchange rate", "/ɪksˈtʃeɪndʒ reɪt/", "汇率", "What is the current exchange rate?", "当前汇率是多少？", "shopping"},
	{"currency exchange", "/ˈkʌrənsi ɪksˈtʃeɪndʒ/", "货币兑换", "Where is the currency exchange counter?", "货币兑换处在哪里？", "shopping"},
	{"credit card", "/ˈkredɪt kɑːrd/", "信用卡", "Do you accept credit cards?", "你们接受信用卡吗？", "shopping"},
	{"out of stock", "/ˌaʊt əv ˈstɒk/", "缺货", "I'm sorry, this item is out of stock.", "很抱歉，这件商品缺货了。", "shopping"},
	{"gift shop", "/ˈɡɪft ʃɒp/", "礼品店", "I'll buy some souvenirs at the gift shop.", "我在礼品店买些纪念品。", "shopping"},
	{"return policy", "/rɪˈtɜːrn ˈpɒlɪsi/", "退货政策", "What is your return policy?", "你们的退货政策是什么？", "shopping"},

	// 紧急情况
	{"call the police", "/kɔːl ðə pəˈliːs/", "报警", "I need to call the police.", "我需要报警。", "emergency"},
	{"emergency room", "/ɪˈmɜːrdʒənsi ruːm/", "急诊室", "Please take me to the emergency room.", "请带我去急诊室。", "emergency"},
	{"first aid", "/ˌfɜːrst ˈeɪd/", "急救", "Is there a first aid kit nearby?", "附近有急救箱吗？", "emergency"},
	{"travel insurance", "/ˈtrævl ɪnˈʃʊərəns/", "旅行保险", "I have travel insurance for medical emergencies.", "我有旅行保险可以应对医疗紧急情况。", "emergency"},
	{"lost passport", "/lɒst ˈpɑːspɔːrt/", "护照丢失", "I've lost my passport.", "我的护照丢失了。", "emergency"},
	{"stolen wallet", "/ˈstəʊlən ˈwɒlɪt/", "钱包被盗", "My wallet has been stolen.", "我的钱包被偷了。", "emergency"},
	{"nearest hospital", "/ˈnɪərɪst ˈhɒspɪtl/", "最近的医院", "Where is the nearest hospital?", "最近的医院在哪里？", "emergency"},
	{"allergy medication", "/ˈælərdʒi ˌmedɪˈkeɪʃn/", "过敏药物", "I need my allergy medication.", "我需要我的过敏药物。", "emergency"},
	{"embassy", "/ˈembəsi/", "大使馆", "I need to contact my country's embassy.", "我需要联系我国大使馆。", "emergency"},
	{"police report", "/pəˈliːs rɪˈpɔːrt/", "报案记录", "I need to file a police report.", "我需要填写报案记录。", "emergency"},
	{"medical history", "/ˈmedɪkl ˈhɪstri/", "病史", "The doctor asked about my medical history.", "医生询问了我的病史。", "emergency"},
	{"blood type", "/blʌd taɪp/", "血型", "My blood type is O positive.", "我的血型是O型阳性。", "emergency"},
	{"pain killer", "/ˈpeɪn ˌkɪlər/", "止痛药", "Do you have any pain killers?", "你们有止痛药吗？", "emergency"},
	{"prescription", "/prɪˈskrɪpʃn/", "处方", "I need a prescription for this medicine.", "我需要一张处方购买这种药。", "emergency"},
	{"emergency contact", "/ɪˈmɜːrdʒənsi ˈkɒntækt/", "紧急联系人", "Please call my emergency contact.", "请联系我的紧急联系人。", "emergency"},

	// 交通出行
	{"one-way ticket", "/ˌwʌn weɪ ˈtɪkɪt/", "单程票", "One one-way ticket to the city center, please.", "请给我一张去市中心的单程票。", "transportation"},
	{"return ticket", "/rɪˈtɜːrn ˈtɪkɪt/", "往返票", "I'd like a return ticket to the airport.", "我想要一张去机场的往返票。", "transportation"},
	{"transit card", "/ˈtrænsɪt kɑːrd/", "交通卡", "You can use a transit card on all buses.", "所有公交都可以使用交通卡。", "transportation"},
	{"rush hour", "/ˈrʌʃ aʊər/", "高峰时段", "Avoid traveling during rush hour.", "避免在高峰时段出行。", "transportation"},
	{"car rental", "/kɑːr ˈrentl/", "租车", "Where is the car rental counter?", "租车柜台在哪里？", "transportation"},
	{"GPS device", "/ˌdʒiː piː ˈes dɪˌvaɪs/", "GPS设备", "The rental car includes a GPS device.", "租来的车包含GPS设备。", "transportation"},
	{"toll road", "/ˈtəʊl rəʊd/", "收费公路", "There is a toll on this road.", "这条路需要交过路费。", "transportation"},
	{"parking lot", "/ˈpɑːrkɪŋ lɒt/", "停车场", "Is there a parking lot nearby?", "附近有停车场吗？", "transportation"},
	{"high-speed rail", "/ˌhaɪ spiːd ˈreɪl/", "高铁", "The high-speed rail is the fastest option.", "高铁是最快的选择。", "transportation"},
	{"left luggage", "/left ˈlʌɡɪdʒ/", "行李寄存", "Is there a left luggage office at the station?", "车站有行李寄存处吗？", "transportation"},
	{"timetable", "/ˈtaɪmteɪbl/", "时刻表", "Let me check the timetable.", "让我查一下时刻表。", "transportation"},
	{"platform number", "/ˈplætfɔːrm ˈnʌmbər/", "站台号", "Which platform number is this train?", "这列火车是几号站台？", "transportation"},
	{"harbor", "/ˈhɑːrbər/", "港口", "The ferry departs from the harbor.", "渡轮从港口出发。", "transportation"},
	{"rideshare", "/ˈraɪdʃeər/", "拼车/网约车", "Let's use a rideshare app to get there.", "我们用网约车APP去那里吧。", "transportation"},
	{"bike rental", "/baɪk ˈrentl/", "自行车租赁", "Is there a bike rental service nearby?", "附近有自行车租赁服务吗？", "transportation"},

	// 景点娱乐
	{"tourist attraction", "/ˈtʊərɪst əˈtrækʃn/", "旅游景点", "The Eiffel Tower is a famous tourist attraction.", "埃菲尔铁塔是著名的旅游景点。", "entertainment"},
	{"guided tour", "/ˌɡaɪdɪd ˈtʊər/", "导游带队游览", "I'd like to join a guided tour.", "我想参加导游带队的游览。", "entertainment"},
	{"audio guide", "/ˈɔːdiəʊ ɡaɪd/", "语音导览", "An audio guide is available at the entrance.", "入口处有语音导览可用。", "entertainment"},
	{"opening hours", "/ˈəʊpənɪŋ ˈaʊərz/", "营业/开放时间", "What are the opening hours of the museum?", "博物馆的开放时间是什么？", "entertainment"},
	{"entrance fee", "/ˈentrəns fiː/", "门票费", "What is the entrance fee?", "门票费是多少？", "entertainment"},
	{"photo opportunity", "/ˈfəʊtəʊ ˌɒpəˈtjuːnɪti/", "拍照机会", "This spot offers a great photo opportunity.", "这个地方是很好的拍照机会。", "entertainment"},
	{"local cuisine", "/ˈləʊkl kwɪˈziːn/", "当地美食", "I want to try the local cuisine.", "我想尝试当地美食。", "entertainment"},
	{"cultural experience", "/ˈkʌltʃərəl ɪkˈspɪəriəns/", "文化体验", "This is a unique cultural experience.", "这是一次独特的文化体验。", "entertainment"},
	{"historical site", "/hɪˈstɒrɪkl saɪt/", "历史遗址", "We visited several historical sites today.", "今天我们参观了几处历史遗址。", "entertainment"},
	{"scenic spot", "/ˈsiːnɪk spɒt/", "风景区", "This is one of the most popular scenic spots.", "这是最受欢迎的风景区之一。", "entertainment"},
	{"national park", "/ˌnæʃnəl ˈpɑːrk/", "国家公园", "Yellowstone is a famous national park.", "黄石是著名的国家公园。", "entertainment"},
	{"street food", "/ˈstriːt fuːd/", "街头小吃", "I love trying street food when I travel.", "旅行时我喜欢尝试街头小吃。", "entertainment"},
	{"night market", "/ˈnaɪt ˌmɑːrkɪt/", "夜市", "The night market opens after 6 PM.", "夜市在下午6点后开放。", "entertainment"},
	{"sunset view", "/ˈsʌnset vjuː/", "日落景色", "The sunset view here is spectacular.", "这里的日落景色非常壮观。", "entertainment"},
	{"hiking trail", "/ˈhaɪkɪŋ treɪl/", "徒步小径", "The hiking trail starts near the park entrance.", "徒步小径从公园入口附近开始。", "entertainment"},

	// ===== 扩充：机场更多短语 =====
	{"immigration officer", "/ˌɪmɪˈɡreɪʃn ˈɒfɪsər/", "移民官员", "The immigration officer checked my passport.", "移民官检查了我的护照。", "airport"},
	{"boarding gate", "/ˈbɔːrdɪŋ ɡeɪt/", "登机门", "Your boarding gate is B12.", "您的登机门是B12。", "airport"},
	{"overweight baggage", "/ˌəʊvəˈweɪt ˈbægɪdʒ/", "超重行李", "My baggage is overweight by 3 kilograms.", "我的行李超重了3公斤。", "airport"},
	{"departure time", "/dɪˈpɑːtʃər taɪm/", "出发时间", "What is the departure time for this flight?", "这次航班的出发时间是什么？", "airport"},
	{"arrival time", "/əˈraɪvl taɪm/", "到达时间", "The estimated arrival time is 3 PM.", "预计到达时间是下午3点。", "airport"},
	{"domestic flight", "/dəˈmestɪk flaɪt/", "国内航班", "I need to check in for a domestic flight.", "我需要为国内航班办理登机手续。", "airport"},
	{"international flight", "/ˌɪntəˈnæʃnəl flaɪt/", "国际航班", "Gate 5 is for international flights.", "5号门用于国际航班。", "airport"},
	{"seat belt", "/ˈsiːt belt/", "安全带", "Please fasten your seat belt.", "请系好安全带。", "airport"},
	{"overhead compartment", "/ˌəʊvəˈhed kəmˈpɑːrtmənt/", "头顶行李舱", "Please close the overhead compartment.", "请关好头顶行李舱。", "airport"},
	{"baggage tag", "/ˈbægɪdʒ tæɡ/", "行李标签", "Please attach the baggage tag to your suitcase.", "请将行李标签贴到您的箱子上。", "airport"},
	{"flight attendant", "/flaɪt əˈtendənt/", "空乘人员", "Please call the flight attendant if you need help.", "如果需要帮助，请呼叫空乘人员。", "airport"},
	{"economy class", "/ɪˈkɒnəmi klɑːs/", "经济舱", "I booked a seat in economy class.", "我预订了经济舱的座位。", "airport"},
	{"business class", "/ˈbɪznɪs klɑːs/", "商务舱", "The business class seats are very comfortable.", "商务舱的座位非常舒适。", "airport"},
	{"in-flight meal", "/ɪn flaɪt miːl/", "飞机餐", "Would you like the chicken or beef in-flight meal?", "您想要鸡肉还是牛肉飞机餐？", "airport"},
	{"departure board", "/dɪˈpɑːtʃər bɔːrd/", "出发信息显示屏", "Check the departure board for gate information.", "查看出发信息显示屏了解登机门信息。", "airport"},
	{"arrivals hall", "/əˈraɪvlz hɔːl/", "到达大厅", "Please meet me at the arrivals hall.", "请在到达大厅等我。", "airport"},
	{"travel adapter", "/ˈtrævl əˈdæptər/", "旅行转接头", "I forgot to pack my travel adapter.", "我忘记带旅行转接头了。", "airport"},
	{"time zone", "/ˈtaɪm zəʊn/", "时区", "We crossed three time zones on this flight.", "我们这次飞行跨越了三个时区。", "airport"},
	{"stopover", "/ˈstɒpəʊvər/", "中途停留", "We have a stopover in Singapore.", "我们在新加坡有一次中途停留。", "airport"},
	{"e-ticket", "/ˈiː ˌtɪkɪt/", "电子机票", "Please print your e-ticket or show it on your phone.", "请打印您的电子机票或在手机上出示。", "airport"},

	// ===== 扩充：酒店更多短语 =====
	{"hotel lobby", "/ˈhəʊtl ˈlɒbi/", "酒店大堂", "Please wait for me in the hotel lobby.", "请在酒店大堂等我。", "hotel"},
	{"room number", "/ruːm ˈnʌmbər/", "房间号", "What is your room number?", "您的房间号是多少？", "hotel"},
	{"ocean view room", "/ˈəʊʃn vjuː ruːm/", "海景房", "I'd like an ocean view room if possible.", "如果可以，我想要一间海景房。", "hotel"},
	{"swimming pool", "/ˈswɪmɪŋ puːl/", "游泳池", "The swimming pool is open until 10 PM.", "游泳池营业到晚上10点。", "hotel"},
	{"breakfast buffet", "/ˈbrekfəst ˈbʊfeɪ/", "早餐自助餐", "The breakfast buffet starts at 7 AM.", "早餐自助餐从早上7点开始。", "hotel"},
	{"hotel amenities", "/həʊtel əˈmiːnɪtɪz/", "酒店设施", "The hotel amenities include a gym and spa.", "酒店设施包括健身房和水疗中心。", "hotel"},
	{"room rate", "/ruːm reɪt/", "房价", "What is the room rate per night?", "每晚的房价是多少？", "hotel"},
	{"booking confirmation", "/ˈbʊkɪŋ ˌkɒnfəˈmeɪʃn/", "预订确认", "Please show your booking confirmation at check-in.", "请在入住时出示您的预订确认。", "hotel"},
	{"extra bed", "/ˈekstrə bed/", "加床", "Can I request an extra bed?", "我可以申请加床吗？", "hotel"},
	{"pillow", "/ˈpɪləʊ/", "枕头", "Could I have an extra pillow, please?", "请问可以多给我一个枕头吗？", "hotel"},
	{"blanket", "/ˈblæŋkɪt/", "毯子", "I need an extra blanket.", "我需要一条额外的毯子。", "hotel"},
	{"room cleaning", "/ruːm ˈkliːnɪŋ/", "客房清洁", "Can you skip room cleaning today?", "今天可以不打扫房间吗？", "hotel"},
	{"spa", "/spɑː/", "水疗中心", "I'd like to book a treatment at the spa.", "我想预约水疗中心的护理项目。", "hotel"},
	{"24-hour front desk", "/ˌtwenti fɔːr ˈaʊər frʌnt desk/", "24小时前台", "The front desk is available 24 hours.", "前台24小时提供服务。", "hotel"},
	{"noise complaint", "/nɔɪz kəmˈpleɪnt/", "噪音投诉", "I'd like to make a noise complaint.", "我想投诉噪音问题。", "hotel"},
	{"lost and found", "/ˌlɒst ənd ˈfaʊnd/", "失物招领处", "I left my charger in the room. Is there a lost and found?", "我把充电器落在房间里了，有失物招领处吗？", "hotel"},
	{"hotel shuttle", "/ˈhəʊtl ˈʃʌtl/", "酒店班车", "Does the hotel have a shuttle to the airport?", "酒店有去机场的班车吗？", "hotel"},

	// ===== 扩充：餐厅更多短语 =====
	{"buffet restaurant", "/ˈbʊfeɪ ˈrestrɒnt/", "自助餐厅", "This is an all-you-can-eat buffet restaurant.", "这是一家可以无限吃的自助餐厅。", "restaurant"},
	{"fine dining", "/faɪn ˈdaɪnɪŋ/", "高档餐饮", "We'd like a fine dining experience tonight.", "今晚我们想体验一下高档餐饮。", "restaurant"},
	{"outdoor seating", "/ˈaʊtdɔːr ˈsiːtɪŋ/", "户外座位", "Do you have outdoor seating?", "你们有户外座位吗？", "restaurant"},
	{"half portion", "/hɑːf ˈpɔːrʃn/", "半份", "Can I have a half portion of pasta?", "我可以要半份意大利面吗？", "restaurant"},
	{"refill", "/ˌriːˈfɪl/", "续杯", "Free refills are included.", "包含免费续杯。", "restaurant"},
	{"local specialty", "/ˈləʊkl spəˈʃælɪti/", "当地特色菜", "I'd like to try a local specialty.", "我想尝试一下当地特色菜。", "restaurant"},
	{"three-course meal", "/ˌθriː kɔːrs ˈmiːl/", "三道式套餐", "A three-course meal includes starter, main, and dessert.", "三道式套餐包括前菜、主菜和甜点。", "restaurant"},
	{"sharing platter", "/ˈʃeərɪŋ ˈplætər/", "拼盘", "Let's order a sharing platter.", "我们点一个拼盘吧。", "restaurant"},
	{"sparkling wine", "/ˌspɑːklɪŋ ˈwaɪn/", "气泡葡萄酒", "Would you like some sparkling wine?", "您想来点气泡葡萄酒吗？", "restaurant"},
	{"draft beer", "/drɑːft bɪər/", "生啤", "I'll have a pint of draft beer.", "我要一品脱生啤。", "restaurant"},
	{"non-alcoholic drink", "/ˌnɒn ˌælkəˈhɒlɪk drɪŋk/", "无酒精饮料", "Do you have any non-alcoholic drinks?", "你们有无酒精饮料吗？", "restaurant"},
	{"kids menu", "/kɪdz ˈmenjuː/", "儿童菜单", "Do you have a kids menu?", "你们有儿童菜单吗？", "restaurant"},
	{"reservation required", "/ˌrezəˈveɪʃn rɪˈkwaɪərd/", "需要预约", "Reservation is required for dinner.", "晚餐需要提前预约。", "restaurant"},
	{"happy hour", "/ˈhæpi ˌaʊər/", "欢乐时光（折扣时段）", "Drinks are half price during happy hour.", "欢乐时光期间饮品半价。", "restaurant"},
	{"food court", "/fuːd kɔːrt/", "美食广场", "There is a food court on the third floor.", "三楼有一个美食广场。", "restaurant"},
	{"outdoor cafe", "/ˈaʊtdɔːr ˈkæfeɪ/", "户外咖啡馆", "Let's sit at the outdoor cafe.", "我们坐在户外咖啡馆吧。", "restaurant"},
	{"vending machine", "/ˈvendɪŋ məˌʃiːn/", "自动售货机", "There is a vending machine in the hallway.", "走廊里有一台自动售货机。", "restaurant"},
	{"chopsticks", "/ˈtʃɒpstɪks/", "筷子", "Can I have some chopsticks?", "我可以要一双筷子吗？", "restaurant"},
	{"straw", "/strɔː/", "吸管", "Can I have a straw for this drink?", "这杯饮料可以给我吸管吗？", "restaurant"},
	{"ice", "/aɪs/", "冰块", "I'd like my drink with ice, please.", "我的饮料请加冰。", "restaurant"},
	{"takeaway box", "/ˈteɪkəweɪ bɒks/", "外卖盒", "Can I have a takeaway box for this?", "这个可以给我打包盒吗？", "restaurant"},

	// ===== 扩充：问路更多短语 =====
	{"tourist information center", "/ˈtʊərɪst ˌɪnfəˈmeɪʃn ˈsentər/", "旅游信息中心", "The tourist information center is near the entrance.", "旅游信息中心在入口附近。", "direction"},
	{"on the left", "/ɒn ðə left/", "在左边", "The bank is on the left.", "银行在左边。", "direction"},
	{"on the right", "/ɒn ðə raɪt/", "在右边", "Turn right and the hotel is on the right.", "右转后酒店就在右边。", "direction"},
	{"at the end of", "/æt ðə end əv/", "在…的尽头", "The museum is at the end of this street.", "博物馆在这条街的尽头。", "direction"},
	{"next to", "/ˈnekst tuː/", "紧邻", "The pharmacy is next to the supermarket.", "药店紧邻超市。", "direction"},
	{"opposite", "/ˈɒpəzɪt/", "对面", "The restaurant is opposite the hotel.", "餐厅在酒店对面。", "direction"},
	{"public transport", "/ˈpʌblɪk ˈtrænspɔːrt/", "公共交通", "How do I get there by public transport?", "乘公共交通怎么去那里？", "direction"},
	{"local map", "/ˈləʊkl mæp/", "当地地图", "Can I get a local map from here?", "我可以在这里拿到当地地图吗？", "direction"},
	{"get lost", "/ɡet lɒst/", "迷路", "I think we got lost.", "我觉得我们迷路了。", "direction"},
	{"road sign", "/ˈrəʊd saɪn/", "路标", "Follow the road signs to the city center.", "跟着路标走到市中心。", "direction"},
	{"pedestrian zone", "/pəˈdestriən zəʊn/", "步行区", "This area is a pedestrian zone.", "这个区域是步行区。", "direction"},
	{"dead end", "/ˌded ˈend/", "死胡同", "This road is a dead end.", "这条路是死胡同。", "direction"},
	{"shortcut", "/ˈʃɔːtkʌt/", "捷径", "Is there a shortcut to the station?", "去车站有捷径吗？", "direction"},
	{"ten minutes away", "/ten ˈmɪnɪts əˈweɪ/", "十分钟路程", "The station is about ten minutes away.", "车站大约十分钟路程。", "direction"},
	{"take the first turning", "/teɪk ðə fɜːst ˈtɜːrnɪŋ/", "在第一个路口转", "Take the first turning on the left.", "在左边第一个路口转。", "direction"},

	// ===== 扩充：购物更多短语 =====
	{"local market", "/ˈləʊkl ˈmɑːrkɪt/", "当地市集", "Let's browse the local market for souvenirs.", "我们去当地市集找纪念品吧。", "shopping"},
	{"window shopping", "/ˈwɪndəʊ ˌʃɒpɪŋ/", "逛街不购物", "I'm just window shopping today.", "我今天只是逛逛不买东西。", "shopping"},
	{"limited edition", "/ˌlɪmɪtɪd ɪˈdɪʃn/", "限量版", "This is a limited edition souvenir.", "这是一款限量版纪念品。", "shopping"},
	{"pay in cash", "/peɪ ɪn kæʃ/", "用现金付款", "Can I pay in cash?", "我可以用现金付款吗？", "shopping"},
	{"debit card", "/ˈdebɪt kɑːrd/", "借记卡", "I'd like to pay with my debit card.", "我想用借记卡付款。", "shopping"},
	{"mobile payment", "/ˈməʊbaɪl ˈpeɪmənt/", "手机支付", "Do you accept mobile payment?", "你们接受手机支付吗？", "shopping"},
	{"bargain price", "/ˈbɑːɡɪn praɪs/", "特价", "This jacket is at a bargain price today.", "这件夹克今天是特价。", "shopping"},
	{"clearance sale", "/ˈklɪərəns seɪl/", "清仓甩卖", "Everything is 50% off in the clearance sale.", "清仓甩卖所有商品半价。", "shopping"},
	{"try on", "/traɪ ɒn/", "试穿", "Can I try this on?", "我可以试穿这件吗？", "shopping"},
	{"size chart", "/saɪz tʃɑːrt/", "尺码表", "Can I see the size chart?", "我可以看看尺码表吗？", "shopping"},
	{"gift wrapping", "/ˈɡɪft ˌræpɪŋ/", "礼品包装", "Can you do gift wrapping for this?", "这个可以做礼品包装吗？", "shopping"},
	{"membership card", "/ˈmembəʃɪp kɑːrd/", "会员卡", "Do you have a membership card?", "您有会员卡吗？", "shopping"},
	{"points card", "/pɔɪnts kɑːrd/", "积分卡", "Can I use my points card here?", "我可以在这里使用积分卡吗？", "shopping"},
	{"open 24 hours", "/ˈəʊpən ˌtwenti fɔːr ˈaʊərz/", "24小时营业", "The convenience store is open 24 hours.", "便利店24小时营业。", "shopping"},
	{"flea market", "/ˈfliː ˌmɑːrkɪt/", "跳蚤市场", "You can find antiques at the flea market.", "你可以在跳蚤市场找到古董。", "shopping"},
	{"supermarket", "/ˈsuːpəmɑːrkɪt/", "超级市场", "There is a supermarket near the hotel.", "酒店附近有一家超市。", "shopping"},
	{"convenience store", "/kənˈviːniəns stɔːr/", "便利店", "I'll stop at the convenience store.", "我要去便利店一趟。", "shopping"},

	// ===== 扩充：紧急情况更多短语 =====
	{"call an ambulance", "/kɔːl ən ˈæmbjələns/", "叫救护车", "Please call an ambulance immediately.", "请立即叫救护车。", "emergency"},
	{"first aid kit", "/ˌfɜːrst ˈeɪd kɪt/", "急救箱", "Do you have a first aid kit?", "你们有急救箱吗？", "emergency"},
	{"sunburn", "/ˈsʌnbɜːrn/", "晒伤", "I have a bad sunburn.", "我被晒伤了。", "emergency"},
	{"food poisoning", "/fuːd ˈpɔɪzənɪŋ/", "食物中毒", "I think I have food poisoning.", "我觉得我食物中毒了。", "emergency"},
	{"sprained ankle", "/spreɪnd ˈæŋkl/", "脚踝扭伤", "I've sprained my ankle.", "我的脚踝扭伤了。", "emergency"},
	{"dehydration", "/ˌdiːhaɪˈdreɪʃn/", "脱水", "I feel dizzy due to dehydration.", "我因脱水感到头晕。", "emergency"},
	{"motion sickness", "/ˈməʊʃn ˌsɪknɪs/", "晕动症", "I have motion sickness on long bus rides.", "坐长途汽车我会晕车。", "emergency"},
	{"antihistamine", "/ˌæntiˈhɪstəmiːn/", "抗组胺药", "Do you have any antihistamine tablets?", "你们有抗组胺药片吗？", "emergency"},
	{"sanitary napkin", "/ˈsænɪtri ˈnæpkɪn/", "卫生巾", "Where can I buy sanitary napkins?", "我在哪里可以买到卫生巾？", "emergency"},
	{"fever reducer", "/ˈfiːvər rɪˌdjuːsər/", "退烧药", "I need a fever reducer.", "我需要退烧药。", "emergency"},
	{"bandage", "/ˈbændɪdʒ/", "绷带", "Can I have a bandage for this wound?", "这个伤口可以给我绷带吗？", "emergency"},
	{"insect repellent", "/ˈɪnsekt rɪˈpelənt/", "驱虫剂", "I need some insect repellent.", "我需要驱虫剂。", "emergency"},
	{"sunscreen", "/ˈsʌnskriːn/", "防晒霜", "Don't forget to apply sunscreen.", "别忘了涂防晒霜。", "emergency"},
	{"consulate", "/ˈkɒnsjələt/", "领事馆", "I need to contact the consulate.", "我需要联系领事馆。", "emergency"},
	{"emergency hotline", "/ɪˈmɜːrdʒənsi ˈhɒtlaɪn/", "紧急热线", "Call the emergency hotline for help.", "拨打紧急热线寻求帮助。", "emergency"},

	// ===== 扩充：交通出行更多短语 =====
	{"bus route", "/bʌs ruːt/", "公交路线", "Which bus route goes to the old town?", "哪路公交车去老城区？", "transportation"},
	{"subway map", "/ˈsʌbweɪ mæp/", "地铁图", "Can I have a subway map?", "我可以要一张地铁图吗？", "transportation"},
	{"next stop", "/nekst stɒp/", "下一站", "What is the next stop?", "下一站是哪里？", "transportation"},
	{"last train", "/lɑːst treɪn/", "末班车", "What time is the last train?", "末班车是几点？", "transportation"},
	{"first train", "/fɜːrst treɪn/", "首班车", "What time is the first train?", "首班车是几点？", "transportation"},
	{"train station", "/treɪn ˈsteɪʃn/", "火车站", "How do I get to the train station?", "去火车站怎么走？", "transportation"},
	{"bus terminal", "/bʌs ˈtɜːrmɪnl/", "汽车总站", "The bus terminal is 10 minutes away.", "汽车总站10分钟路程。", "transportation"},
	{"rental car", "/ˈrentl kɑːr/", "租赁汽车", "I'd like to pick up my rental car.", "我想取我的租赁汽车。", "transportation"},
	{"driving license", "/ˈdraɪvɪŋ ˈlaɪsəns/", "驾照", "I need to show my driving license.", "我需要出示我的驾照。", "transportation"},
	{"fuel up", "/ˈfjuːəl ʌp/", "加油", "I need to fuel up before the long drive.", "长途驾驶前我需要加油。", "transportation"},
	{"speed limit", "/ˈspiːd ˌlɪmɪt/", "限速", "The speed limit on this road is 60 km/h.", "这条路的限速是每小时60公里。", "transportation"},
	{"traffic jam", "/ˈtræfɪk dʒæm/", "交通堵塞", "There is a traffic jam on the main road.", "主干道上发生了交通堵塞。", "transportation"},
	{"taxi meter", "/ˈtæksi ˌmiːtər/", "出租车计价表", "Make sure the taxi meter is running.", "确保出租车计价表在运行。", "transportation"},
	{"taxi stand", "/ˈtæksi stænd/", "出租车站", "There is a taxi stand outside the hotel.", "酒店外面有出租车站。", "transportation"},
	{"boat tour", "/bəʊt tʊər/", "游船观光", "Let's book a boat tour of the harbor.", "我们预订一个港口游船观光吧。", "transportation"},
	{"cable car", "/ˈkeɪbl kɑːr/", "缆车", "The cable car takes you to the top of the mountain.", "缆车带你到山顶。", "transportation"},
	{"airport express", "/ˈeəpɔːrt ɪkˈspres/", "机场快线", "Take the airport express to save time.", "乘机场快线节省时间。", "transportation"},
	{"travel pass", "/ˈtrævl pɑːs/", "旅游通票", "A travel pass gives you unlimited rides.", "旅游通票让你无限次乘车。", "transportation"},

	// ===== 扩充：景点娱乐更多短语 =====
	{"boat trip", "/bəʊt trɪp/", "乘船游览", "A boat trip along the coast is wonderful.", "沿海岸乘船游览非常美妙。", "entertainment"},
	{"city tour", "/ˈsɪti tʊər/", "城市游览", "Book a half-day city tour.", "预订半天城市游览。", "entertainment"},
	{"day trip", "/deɪ trɪp/", "一日游", "We went on a day trip to the countryside.", "我们去乡村进行了一日游。", "entertainment"},
	{"souvenir shop", "/suːˈvɪənɪr ʃɒp/", "纪念品店", "Don't forget to stop at the souvenir shop.", "别忘了去纪念品店看看。", "entertainment"},
	{"theme park", "/ˈθiːm pɑːrk/", "主题公园", "The theme park is great for families.", "主题公园非常适合家庭游玩。", "entertainment"},
	{"botanical garden", "/bəˈtænɪkl ˈɡɑːrdn/", "植物园", "The botanical garden has rare tropical plants.", "植物园里有珍稀热带植物。", "entertainment"},
	{"aquarium", "/əˈkweəriəm/", "水族馆", "The aquarium has thousands of fish species.", "水族馆里有数千种鱼类。", "entertainment"},
	{"zip line", "/ˈzɪp laɪn/", "高空滑索", "Let's try the zip line over the canyon.", "我们试试横跨峡谷的高空滑索吧。", "entertainment"},
	{"snorkeling", "/ˈsnɔːrkəlɪŋ/", "浮潜", "The coral reef is perfect for snorkeling.", "珊瑚礁非常适合浮潜。", "entertainment"},
	{"bungee jumping", "/ˈbʌndʒi ˌdʒʌmpɪŋ/", "蹦极", "Are you brave enough to try bungee jumping?", "你够勇敢去尝试蹦极吗？", "entertainment"},
	{"whale watching", "/weɪl ˈwɒtʃɪŋ/", "观鲸", "Join a whale watching tour this morning.", "今天早上参加一个观鲸之旅吧。", "entertainment"},
	{"cultural show", "/ˈkʌltʃərəl ʃəʊ/", "文化表演", "The cultural show starts at 8 PM.", "文化表演晚上8点开始。", "entertainment"},
	{"folk dance", "/fəʊk dɑːns/", "民间舞蹈", "The folk dance performance was amazing.", "民间舞蹈表演非常精彩。", "entertainment"},
	{"cooking class", "/ˈkʊkɪŋ klɑːs/", "烹饪课", "I signed up for a local cooking class.", "我报名参加了当地的烹饪课。", "entertainment"},
	{"wine tasting", "/waɪn ˈteɪstɪŋ/", "葡萄酒品鉴", "The vineyard offers free wine tasting.", "这家葡萄酒庄提供免费品酒。", "entertainment"},
	{"photography tour", "/fəˈtɒɡrəfi tʊər/", "摄影之旅", "Join a photography tour at sunrise.", "在日出时参加摄影之旅。", "entertainment"},
	{"spa treatment", "/spɑː ˈtriːtmənt/", "水疗护理", "I booked a relaxing spa treatment.", "我预约了一个放松的水疗护理。", "entertainment"},
	{"language exchange", "/ˈlæŋɡwɪdʒ ɪksˈtʃeɪndʒ/", "语言交流", "Language exchange helps you practice English.", "语言交流帮助你练习英语。", "entertainment"},
	{"local festival", "/ˈləʊkl ˈfestɪvl/", "当地节日", "We were lucky to see a local festival.", "我们很幸运地见到了当地节日。", "entertainment"},
	{"street performance", "/striːt pəˈfɔːrməns/", "街头表演", "The street performance drew a big crowd.", "街头表演吸引了大批观众。", "entertainment"},
}

func main() {
	oxfordCSV := flag.String("oxford", "/tmp/oxford3000.csv", "Oxford 3000 词表 CSV 路径")
	ecdictCSV := flag.String("ecdict", "/tmp/ecdict.csv", "ECDICT 词典 CSV 路径")
	output := flag.String("out", "data/libraries/travel.json", "输出文件路径")
	flag.Parse()

	log.Println("开始构建旅游词库...")
	log.Printf("Oxford 3000: %s", *oxfordCSV)
	log.Printf("ECDICT: %s", *ecdictCSV)

	// 1. 读取 Oxford 3000，筛选旅游相关词
	oxford3000TravelWords, err := loadOxfordTravelWords(*oxfordCSV)
	if err != nil {
		log.Fatalf("读取 Oxford 3000 失败: %v", err)
	}
	log.Printf("从 Oxford 3000 筛选出旅游相关词: %d 个", len(oxford3000TravelWords))

	// 2. 读取 ECDICT，建立词汇查找表
	ecdictMap, err := loadECDICT(*ecdictCSV, oxford3000TravelWords)
	if err != nil {
		log.Fatalf("读取 ECDICT 失败: %v", err)
	}
	log.Printf("在 ECDICT 中找到匹配词条: %d 个", len(ecdictMap))

	// 3. 构建词库
	words := make([]OutputWord, 0, 1000)
	counter := 1

	// 3.1 添加 Oxford 3000 旅游词
	for word, category := range oxford3000TravelWords {
		entry := ecdictMap[strings.ToLower(word)]
		if entry == nil {
			continue
		}
		chinese := cleanTranslation(entry.Translation)
		if chinese == "" {
			continue
		}
		w := OutputWord{
			ID:       fmt.Sprintf("travel_%04d", counter),
			English:  word,
			Phonetic: entry.Phonetic,
			Chinese:  chinese,
			Example:  "", // 例句在 step5 中通过模板生成
			Category: category,
			Source:   "oxford3000",
		}
		words = append(words, w)
		counter++
	}

	// 3.2 添加旅游专项短语
	for _, phrase := range travelPhrases {
		w := OutputWord{
			ID:        fmt.Sprintf("travel_%04d", counter),
			English:   phrase[0],
			Phonetic:  phrase[1],
			Chinese:   phrase[2],
			Example:   phrase[3],
			ExampleCN: phrase[4],
			Category:  phrase[5],
			Source:    "travel_phrase",
		}
		words = append(words, w)
		counter++
	}

	log.Printf("词库总词条: %d 个", len(words))

	// 4. 输出 JSON
	lib := OutputLibrary{
		ID:    "travel",
		Name:  "旅游场景",
		Desc:  fmt.Sprintf("基于 Oxford 3000 高频词表筛选 + 旅游专项短语，共 %d 词，覆盖机场、酒店、餐厅、问路、购物等场景", len(words)),
		Words: words,
	}

	// 确保输出目录存在
	dir := "data/libraries"
	if err := os.MkdirAll(dir, 0755); err != nil {
		log.Fatalf("创建目录失败: %v", err)
	}

	f, err := os.Create(*output)
	if err != nil {
		log.Fatalf("创建输出文件失败: %v", err)
	}
	defer f.Close()

	enc := json.NewEncoder(f)
	enc.SetIndent("", "  ")
	if err := enc.Encode(lib); err != nil {
		log.Fatalf("写入 JSON 失败: %v", err)
	}

	log.Printf("词库已生成: %s", *output)

	// 统计各场景词数
	stats := map[string]int{}
	for _, w := range words {
		stats[w.Category]++
	}
	for cat, cnt := range stats {
		log.Printf("  %s: %d 词", cat, cnt)
	}
}

// loadOxfordTravelWords 读取 Oxford 3000 CSV，仅保留手工标注的旅游词，返回 word->category 映射
// 不使用宽泛关键词规则，确保分类准确
func loadOxfordTravelWords(path string) (map[string]string, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer f.Close()

	result := make(map[string]string)
	reader := csv.NewReader(f)
	reader.LazyQuotes = true
	reader.FieldsPerRecord = -1

	// 跳过头部
	if _, err := reader.Read(); err != nil {
		return nil, err
	}

	for {
		record, err := reader.Read()
		if err == io.EOF {
			break
		}
		if err != nil {
			continue
		}
		if len(record) < 1 {
			continue
		}

		word := strings.TrimSpace(record[0])
		wordLower := strings.ToLower(word)

		// 仅使用手工精确标注，不使用宽泛关键词匹配
		if cat, ok := oxford3000TravelCategory[wordLower]; ok {
			result[word] = cat
		}
	}

	return result, nil
}

// loadECDICT 读取 ECDICT CSV，仅加载目标词汇的词条，返回 word->entry 映射
func loadECDICT(path string, targetWords map[string]string) (map[string]*WordEntry, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer f.Close()

	// 建立目标词小写集合
	targetSet := make(map[string]bool, len(targetWords))
	for w := range targetWords {
		targetSet[strings.ToLower(w)] = true
	}

	result := make(map[string]*WordEntry)
	reader := csv.NewReader(f)
	reader.LazyQuotes = true
	reader.FieldsPerRecord = -1 // 允许可变字段数

	// 跳过头部：word,phonetic,definition,translation,pos,collins,oxford,tag,bnc,frq,exchange,detail,audio
	if _, err := reader.Read(); err != nil {
		return nil, err
	}

	count := 0
	for {
		record, err := reader.Read()
		if err == io.EOF {
			break
		}
		if err != nil {
			continue
		}
		if len(record) < 4 {
			continue
		}

		word := strings.TrimSpace(record[0])
		wordLower := strings.ToLower(word)
		if !targetSet[wordLower] {
			continue
		}

		phonetic := ""
		if len(record) > 1 {
			phonetic = record[1]
		}
		translation := ""
		if len(record) > 3 {
			translation = record[3]
		}

		result[wordLower] = &WordEntry{
			Word:        word,
			Phonetic:    phonetic,
			Translation: translation,
		}
		count++
	}

	log.Printf("ECDICT 扫描完成，匹配 %d 词", count)
	return result, nil
}

// cleanTranslation 清理 ECDICT 中文翻译，提取第一个简洁的核心释义
func cleanTranslation(raw string) string {
	if raw == "" {
		return ""
	}
	// 按换行符分割，逐行处理
	lines := strings.Split(raw, "\n")
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}
		// 跳过各类注解行
		if strings.HasPrefix(line, "[网络]") ||
			strings.HasPrefix(line, "[医]") ||
			strings.HasPrefix(line, "[化]") ||
			strings.HasPrefix(line, "[计]") ||
			strings.HasPrefix(line, "[经]") ||
			strings.HasPrefix(line, "[法]") ||
			strings.HasPrefix(line, "[军]") ||
			strings.HasPrefix(line, "[体]") ||
			strings.HasPrefix(line, "[电]") ||
			strings.HasPrefix(line, "[物]") {
			continue
		}
		// 去掉词性前缀（vi. vt. n. v. adj. adv. prep. conj. num. a. 等）
		posPrefixes := []string{
			"vi. ", "vt. ", "n. ", "v. ", "adj. ", "adv. ",
			"prep. ", "conj. ", "num. ", "a. ", "pron. ", "art. ",
			"int. ", "abbr. ", "aux. ",
		}
		for _, prefix := range posPrefixes {
			if strings.HasPrefix(line, prefix) {
				line = strings.TrimPrefix(line, prefix)
				break
			}
		}
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}
		// 取第一个分隔符前的内容（中英文逗号、分号均处理）
		for _, sep := range []string{"；", "，", ";", ","} {
			if idx := strings.Index(line, sep); idx > 0 {
				candidate := strings.TrimSpace(line[:idx])
				// 至少保留2个字符的中文释义
				if len([]rune(candidate)) >= 2 {
					return candidate
				}
			}
		}
		// 没有分隔符则取整行（限制10个汉字）
		runes := []rune(line)
		if len(runes) > 10 {
			runes = runes[:10]
		}
		return strings.TrimSpace(string(runes))
	}
	return ""
}
