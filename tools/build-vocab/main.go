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
	// 景点扩充
	"beach":       "entertainment",
	"adventure":   "entertainment",
	"landscape":   "entertainment",
	"photography": "entertainment",
	"relaxing":    "entertainment",

	// ===== airport 机场扩充 =====
	"document": "airport",
	"stamp":    "airport",
	"form":     "airport",
	"window":   "airport",
	"weight":   "airport",
	"declare":  "airport",

	// ===== hotel 酒店扩充 =====
	"quiet":       "hotel",
	"noise":       "hotel",
	"comfortable": "hotel",
	"soap":        "hotel",
	"repair":      "hotel",

	// ===== restaurant 餐厅扩充 =====
	"pepper": "restaurant",
	"onion":  "restaurant",
	"nut":    "restaurant",
	"raw":    "restaurant",
	"plate":  "restaurant",
	"bowl":   "restaurant",
	"cup":    "restaurant",
	"fork":   "restaurant",
	"knife":  "restaurant",
	"spoon":  "restaurant",
	"glass":  "restaurant",
	"lemon":  "restaurant",
	"butter": "restaurant",
	"cream":  "restaurant",

	// ===== direction 问路扩充 =====
	"block":   "direction",
	"follow":  "direction",
	"towards": "direction",
	"beyond":  "direction",

	// ===== shopping 购物扩充 =====
	"brand":   "shopping",
	"quality": "shopping",
	"size":    "shopping",
	"colour":  "shopping",
	"cloth":   "shopping",
	"wear":    "shopping",
	"fashion": "shopping",
	"style":   "shopping",
	"carry":   "shopping",
	"wrap":    "shopping",
	"label":   "shopping",
	"item":    "shopping",

	// ===== emergency 紧急情况扩充 =====
	"throat":  "emergency",
	"stomach": "emergency",
	"breathe": "emergency",
	"chest":   "emergency",
	"head":    "emergency",
	"back":    "emergency",
	"leg":     "emergency",
	"arm":     "emergency",
	"eye":     "emergency",
	"ear":     "emergency",
	"tooth":   "emergency",
	"blood":   "emergency",
	"cut":     "emergency",
	"swallow": "emergency",
	"faint":   "emergency",

	// ===== transportation 交通扩充 =====
	"speed":   "transportation",
	"express": "transportation",
	"pass":    "transportation",

	// ===== general 旅行通用交际 =====
	"hello":       "general",
	"goodbye":     "general",
	"please":      "general",
	"sorry":       "general",
	"thank":       "general",
	"yes":         "general",
	"no":          "general",
	"understand":  "general",
	"speak":       "general",
	"name":        "general",
	"question":    "general",
	"answer":      "general",
	"repeat":      "general",
	"today":       "general",
	"tomorrow":    "general",
	"yesterday":   "general",
	"morning":     "general",
	"afternoon":   "general",
	"evening":     "general",
	"night":       "general",
	"hour":        "general",
	"minute":      "general",
	"week":        "general",
	"month":       "general",
	"time":        "general",
	"clock":       "general",
	"early":       "general",
	"late":        "general",
	"hot":         "general",
	"warm":        "general",
	"rain":        "general",
	"sun":         "general",
	"man":         "general",
	"woman":       "general",
	"child":       "general",
	"family":      "general",
	"friend":      "general",
	"people":      "general",
	"person":      "general",
	"number":      "general",
	"first":       "general",
	"second":      "general",
	"last":        "general",
	"next":        "general",
	"much":        "general",
	"many":        "general",
	"enough":      "general",
	"more":        "general",
	"good":        "general",
	"bad":         "general",
	"nice":        "general",
	"big":         "general",
	"small":       "general",
	"new":         "general",
	"old":         "general",
	"long":        "general",
	"short":       "general",
	"fast":        "general",
	"slow":        "general",
	"need":        "general",
	"know":        "general",
	"think":       "general",
	"use":         "general",
	"find":        "general",
	"try":         "general",
	"call":        "general",
	"talk":        "general",
	"bring":       "general",
	"take":        "general",
	"give":        "general",
	"get":         "general",
	"come":        "general",
	"see":         "general",
	"look":        "general",
	"show":        "general",
	"stop":        "general",
	"wait":        "general",
	"return":      "general",
	"leave":       "general",
	"write":       "general",
	"read":        "general",
	"alone":       "general",
	"together":    "general",
	"group":       "general",
	"problem":     "general",
	"important":   "general",
	"information": "general",
	"office":      "general",
	"area":        "general",
	"picture":     "general",
	"open":        "general",
	"near":        "general",
	"full":        "general",
	"empty":       "general",
	"light":       "general",
	"heavy":       "general",
	"around":      "general",
	"through":     "general",
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

	// ===== 通用交际（general） =====
	{"do you speak English", "/duː jʊ spiːk ˈɪŋɡlɪʃ/", "你会说英语吗", "Do you speak English?", "你会说英语吗？", "general"},
	{"I don't understand", "/aɪ dəʊnt ˌʌndəˈstænd/", "我听不懂", "I'm sorry, I don't understand.", "对不起，我听不懂。", "general"},
	{"could you repeat that", "/kʊd jʊ rɪˈpiːt ðæt/", "请再说一遍", "Could you repeat that more slowly?", "请再慢一点说一遍好吗？", "general"},
	{"please speak slowly", "/pliːz spiːk ˈsləʊli/", "请说慢一点", "Please speak slowly, I'm learning English.", "请说慢一点，我在学英语。", "general"},
	{"can you write it down", "/kæn jʊ raɪt ɪt daʊn/", "请写下来", "Can you write it down for me?", "能帮我写下来吗？", "general"},
	{"thank you very much", "/θæŋk jʊ ˈveri mʌtʃ/", "非常感谢", "Thank you very much for your help.", "非常感谢你的帮助。", "general"},
	{"you're welcome", "/jɔːr ˈwelkəm/", "不客气", "You're welcome, have a nice trip!", "不客气，祝旅途愉快！", "general"},
	{"excuse me please", "/ɪkˈskjuːz miː pliːz/", "劳驾打扰一下", "Excuse me please, where is the exit?", "劳驾，出口在哪里？", "general"},
	{"no problem", "/nəʊ ˈprɒbləm/", "没问题", "No problem, I can help you.", "没问题，我可以帮你。", "general"},
	{"I'm sorry", "/aɪm ˈsɒri/", "对不起", "I'm sorry for the inconvenience.", "对不起给您带来不便。", "general"},
	{"what time is it", "/wɒt taɪm ɪz ɪt/", "现在几点", "Excuse me, what time is it now?", "打扰一下，现在几点了？", "general"},
	{"how long does it take", "/haʊ lɒŋ dʌz ɪt teɪk/", "需要多长时间", "How long does it take to get there?", "去那里需要多长时间？", "general"},
	{"where is the bathroom", "/weər ɪz ðə ˈbɑːθruːm/", "洗手间在哪里", "Excuse me, where is the bathroom?", "打扰一下，洗手间在哪里？", "general"},
	{"is there wifi here", "/ɪz ðeər ˈwaɪfaɪ hɪər/", "这里有WiFi吗", "Is there free wifi here?", "这里有免费WiFi吗？", "general"},
	{"what is the wifi password", "/wɒt ɪz ðə ˈwaɪfaɪ ˈpɑːswɜːrd/", "WiFi密码是多少", "What is the wifi password?", "WiFi密码是多少？", "general"},
	{"can I charge my phone", "/kæn aɪ tʃɑːrdʒ maɪ fəʊn/", "我可以给手机充电吗", "Can I charge my phone here?", "我可以在这里给手机充电吗？", "general"},
	{"I need help", "/aɪ niːd help/", "我需要帮助", "I need help, please!", "我需要帮助，请！", "general"},
	{"can you help me", "/kæn jʊ help miː/", "你能帮助我吗", "Can you help me find my hotel?", "你能帮我找到我的酒店吗？", "general"},
	{"do you accept", "/duː jʊ əkˈsept/", "你们接受吗", "Do you accept credit cards?", "你们接受信用卡吗？", "general"},
	{"how much is it", "/haʊ mʌtʃ ɪz ɪt/", "这个多少钱", "How much is it in total?", "总共多少钱？", "general"},
	{"is it included", "/ɪz ɪt ɪnˈkluːdɪd/", "包含在内吗", "Is breakfast included in the price?", "价格中包含早餐吗？", "general"},
	{"I'm looking for", "/aɪm ˈlʊkɪŋ fɔːr/", "我在找", "I'm looking for a pharmacy nearby.", "我在找附近的药店。", "general"},
	{"just a moment", "/dʒʌst ə ˈməʊmənt/", "请稍等", "Just a moment, I'll check for you.", "请稍等，我帮您查一下。", "general"},
	{"I don't know", "/aɪ dəʊnt nəʊ/", "我不知道", "I'm sorry, I don't know the way.", "对不起，我不知道路。", "general"},
	{"what does this mean", "/wɒt dʌz ðɪs miːn/", "这是什么意思", "What does this word mean?", "这个词是什么意思？", "general"},
	{"is this right", "/ɪz ðɪs raɪt/", "这样对吗", "Is this the right bus stop?", "这是正确的公交站吗？", "general"},
	{"good morning", "/ɡʊd ˈmɔːrnɪŋ/", "早上好", "Good morning! How can I help you?", "早上好！请问有什么可以帮您？", "general"},
	{"good evening", "/ɡʊd ˈiːvnɪŋ/", "晚上好", "Good evening, welcome to our restaurant.", "晚上好，欢迎光临我们的餐厅。", "general"},
	{"nice to meet you", "/naɪs tʊ miːt jʊ/", "很高兴见到你", "Nice to meet you, I'm visiting from China.", "很高兴见到你，我从中国来旅游。", "general"},
	{"have a nice day", "/hæv ə naɪs deɪ/", "祝你今天愉快", "Thank you, have a nice day!", "谢谢你，祝你今天愉快！", "general"},
	{"enjoy your trip", "/ɪnˈdʒɔɪ jɔːr trɪp/", "旅途愉快", "Thank you, enjoy your trip!", "谢谢，旅途愉快！", "general"},

	// ===== 机场扩充（airport） =====
	{"my flight is delayed", "/maɪ flaɪt ɪz dɪˈleɪd/", "我的航班延误了", "My flight is delayed by two hours.", "我的航班延误了两小时。", "airport"},
	{"I missed my flight", "/aɪ mɪst maɪ flaɪt/", "我错过了航班", "I missed my flight, what should I do?", "我错过了航班，我应该怎么办？", "airport"},
	{"I need to rebook", "/aɪ niːd tʊ ˌriːˈbʊk/", "我需要改签", "I need to rebook my flight to tomorrow.", "我需要把航班改签到明天。", "airport"},
	{"where is the information desk", "/weər ɪz ðə ˌɪnfəˈmeɪʃn desk/", "问询处在哪里", "Where is the information desk?", "问询处在哪里？", "airport"},
	{"is this the right gate", "/ɪz ðɪs ðə raɪt ɡeɪt/", "这是正确的登机口吗", "Is this the right gate for flight CA123?", "这是CA123航班的登机口吗？", "airport"},
	{"how long is the delay", "/haʊ lɒŋ ɪz ðə dɪˈleɪ/", "延误多长时间", "How long is the delay?", "延误多长时间？", "airport"},
	{"can I upgrade my seat", "/kæn aɪ ˈʌpɡreɪd maɪ siːt/", "可以升级座位吗", "Can I upgrade my seat to business class?", "可以把我的座位升级到商务舱吗？", "airport"},
	{"I need a wheelchair", "/aɪ niːd ə ˈwiːltʃeər/", "我需要轮椅", "I need a wheelchair for my elderly mother.", "我的年迈母亲需要轮椅。", "airport"},
	{"is the flight on time", "/ɪz ðə flaɪt ɒn taɪm/", "航班准时吗", "Is the flight to London on time?", "飞伦敦的航班准时吗？", "airport"},
	{"where do I check in", "/weər duː aɪ tʃek ɪn/", "在哪里办理登机手续", "Where do I check in for this flight?", "这次航班在哪里办理登机手续？", "airport"},
	{"can I take this on board", "/kæn aɪ teɪk ðɪs ɒn bɔːrd/", "这个可以带上飞机吗", "Can I take this liquid on board?", "这瓶液体可以带上飞机吗？", "airport"},
	{"where is the transit lounge", "/weər ɪz ðə ˈtrænsɪt laʊndʒ/", "中转候机室在哪里", "Where is the transit lounge for my connection?", "我的中转候机室在哪里？", "airport"},
	{"my baggage is missing", "/maɪ ˈbægɪdʒ ɪz ˈmɪsɪŋ/", "我的行李丢失了", "My baggage is missing, who can I talk to?", "我的行李丢失了，我应该找谁？", "airport"},
	{"fill in the arrival form", "/fɪl ɪn ðə əˈraɪvl fɔːrm/", "填写入境表格", "Please fill in the arrival form before landing.", "请在降落前填写入境表格。", "airport"},
	{"nothing to declare", "/ˈnʌθɪŋ tʊ dɪˈkleər/", "没有需要申报的物品", "I have nothing to declare at customs.", "我在海关没有需要申报的物品。", "airport"},

	// ===== 酒店扩充（hotel） =====
	{"the air conditioning is broken", "/ðə eər kənˈdɪʃənɪŋ ɪz ˈbrəʊkən/", "空调坏了", "The air conditioning in my room is broken.", "我房间里的空调坏了。", "hotel"},
	{"there is no hot water", "/ðeər ɪz nəʊ hɒt ˈwɔːtər/", "没有热水", "There is no hot water in the shower.", "淋浴没有热水。", "hotel"},
	{"can I have more towels", "/kæn aɪ hæv mɔːr ˈtaʊəlz/", "我可以多要几条毛巾吗", "Can I have more towels for my room?", "我的房间可以多要几条毛巾吗？", "hotel"},
	{"the wifi is not working", "/ðə ˈwaɪfaɪ ɪz nɒt ˈwɜːrkɪŋ/", "WiFi不能用", "The wifi in my room is not working.", "我房间的WiFi不能用。", "hotel"},
	{"can I have a receipt", "/kæn aɪ hæv ə rɪˈsiːt/", "我可以要收据吗", "Can I have a receipt for my stay?", "我可以要一张住宿收据吗？", "hotel"},
	{"can I store my luggage", "/kæn aɪ stɔːr maɪ ˈlʌɡɪdʒ/", "我可以存放行李吗", "Can I store my luggage here after checkout?", "退房后我可以把行李存放在这里吗？", "hotel"},
	{"is breakfast included", "/ɪz ˈbrekfəst ɪnˈkluːdɪd/", "包含早餐吗", "Is breakfast included in the room rate?", "房价包含早餐吗？", "hotel"},
	{"what time is breakfast served", "/wɒt taɪm ɪz ˈbrekfəst sɜːrvd/", "早餐什么时候供应", "What time is breakfast served?", "早餐什么时候供应？", "hotel"},
	{"I'd like to extend my stay", "/aɪd laɪk tʊ ɪkˈstend maɪ steɪ/", "我想延长住宿时间", "I'd like to extend my stay by one night.", "我想延长住宿一晚。", "hotel"},
	{"the toilet is blocked", "/ðə ˈtɔɪlɪt ɪz blɒkt/", "马桶堵了", "The toilet in my bathroom is blocked.", "我浴室的马桶堵了。", "hotel"},
	{"can I have extra hangers", "/kæn aɪ hæv ˈekstrə ˈhæŋərz/", "我可以多要几个衣架吗", "Can I have some extra hangers please?", "请问可以多给我几个衣架吗？", "hotel"},
	{"I need a hairdryer", "/aɪ niːd ə ˈheərdraɪər/", "我需要吹风机", "Is there a hairdryer in the room?", "房间里有吹风机吗？", "hotel"},
	{"can I get a room on a higher floor", "/kæn aɪ ɡet ə ruːm ɒn ə ˈhaɪər flɔːr/", "我可以要高楼层的房间吗", "Can I get a room on a higher floor?", "我可以要一间高楼层的房间吗？", "hotel"},
	{"the room is too noisy", "/ðə ruːm ɪz tuː ˈnɔɪzi/", "房间太吵了", "The room is too noisy, can I change?", "房间太吵了，我可以换房间吗？", "hotel"},
	{"I locked myself out", "/aɪ lɒkt maɪˈself aʊt/", "我把自己锁在门外了", "I locked myself out of my room.", "我把自己锁在房间外了。", "hotel"},
	{"the light is not working", "/ðə laɪt ɪz nɒt ˈwɜːrkɪŋ/", "灯不亮", "The light in the bathroom is not working.", "浴室的灯不亮。", "hotel"},

	// ===== 餐厅扩充（restaurant） =====
	{"I'm vegetarian", "/aɪm ˌvedʒɪˈteəriən/", "我是素食者", "I'm vegetarian, do you have meat-free dishes?", "我是素食者，你们有不含肉的菜吗？", "restaurant"},
	{"I'm vegan", "/aɪm ˈviːɡən/", "我是纯素食者", "I'm vegan, no dairy or eggs please.", "我是纯素食者，请不要加奶制品或鸡蛋。", "restaurant"},
	{"I'm allergic to nuts", "/aɪm əˈlɜːrdʒɪk tʊ nʌts/", "我对坚果过敏", "I'm allergic to nuts, please be careful.", "我对坚果过敏，请注意。", "restaurant"},
	{"no MSG please", "/nəʊ ˌem es ˈdʒiː pliːz/", "请不要加味精", "No MSG please, I'm sensitive to it.", "请不要加味精，我对它敏感。", "restaurant"},
	{"can you make it less spicy", "/kæn jʊ meɪk ɪt les ˈspaɪsi/", "可以少放辣吗", "Can you make it less spicy for me?", "可以帮我做得少辣一点吗？", "restaurant"},
	{"I'd like the bill please", "/aɪd laɪk ðə bɪl pliːz/", "请给我结账", "I'd like the bill please, we're done.", "请给我结账，我们吃好了。", "restaurant"},
	{"is service charge included", "/ɪz ˈsɜːrvɪs tʃɑːrdʒ ɪnˈkluːdɪd/", "包含服务费吗", "Is service charge included in the bill?", "账单里包含服务费吗？", "restaurant"},
	{"what's the soup of the day", "/wɒts ðə suːp əv ðə deɪ/", "今日汤品是什么", "What's the soup of the day?", "今天的汤品是什么？", "restaurant"},
	{"can I see the dessert menu", "/kæn aɪ siː ðə dɪˈzɜːrt ˈmenjuː/", "我可以看甜点菜单吗", "Can I see the dessert menu please?", "请问可以看一下甜点菜单吗？", "restaurant"},
	{"do you take reservations", "/duː jʊ teɪk ˌrezəˈveɪʃnz/", "你们接受预约吗", "Do you take reservations for dinner?", "你们晚餐接受预约吗？", "restaurant"},
	{"we have a reservation", "/wiː hæv ə ˌrezəˈveɪʃn/", "我们有预约", "We have a reservation for two at 7 PM.", "我们预约了晚上7点两人位。", "restaurant"},
	{"how long is the wait", "/haʊ lɒŋ ɪz ðə weɪt/", "需要等多久", "How long is the wait for a table?", "等座位需要多久？", "restaurant"},
	{"can we sit outside", "/kæn wiː sɪt ˈaʊtsaɪd/", "我们可以坐外面吗", "Can we sit outside in the garden?", "我们可以坐在花园外面吗？", "restaurant"},
	{"this is not what I ordered", "/ðɪs ɪz nɒt wɒt aɪ ˈɔːrdərd/", "这不是我点的", "Excuse me, this is not what I ordered.", "不好意思，这不是我点的菜。", "restaurant"},
	{"can I have some more water", "/kæn aɪ hæv sʌm mɔːr ˈwɔːtər/", "可以再给我一些水吗", "Can I have some more water please?", "请问可以再给我一些水吗？", "restaurant"},
	{"what do you recommend", "/wɒt duː jʊ ˌrekəˈmend/", "你有什么推荐", "What do you recommend from the menu?", "菜单上你有什么推荐？", "restaurant"},
	{"I'd like mine without", "/aɪd laɪk maɪn wɪˈðaʊt/", "我的不要加…", "I'd like mine without onions please.", "我的请不要加洋葱。", "restaurant"},
	{"is this dish gluten-free", "/ɪz ðɪs dɪʃ ˈɡluːtnfriː/", "这道菜不含麸质吗", "Is this dish gluten-free?", "这道菜不含麸质吗？", "restaurant"},
	{"can I have a doggy bag", "/kæn aɪ hæv ə ˈdɒɡi bæɡ/", "可以打包带走吗", "Can I have a doggy bag for the leftovers?", "剩下的食物可以帮我打包带走吗？", "restaurant"},
	{"medium steak please", "/ˈmiːdiəm steɪk pliːz/", "牛排要五分熟", "I'll have the ribeye steak, medium please.", "我要一份肋眼牛排，五分熟。", "restaurant"},

	// ===== 问路扩充（direction） =====
	{"I'm trying to get to", "/aɪm ˈtraɪɪŋ tʊ ɡet tʊ/", "我想去…", "I'm trying to get to the central station.", "我想去中央车站。", "direction"},
	{"is this the right way to", "/ɪz ðɪs ðə raɪt weɪ tʊ/", "这是去…的方向吗", "Is this the right way to the museum?", "这是去博物馆的方向吗？", "direction"},
	{"can you show me on the map", "/kæn jʊ ʃəʊ miː ɒn ðə mæp/", "你能在地图上指给我看吗", "Can you show me on the map where we are?", "你能在地图上指出我们在哪里吗？", "direction"},
	{"how many stops is it", "/haʊ ˈmeni stɒps ɪz ɪt/", "需要几站", "How many stops is it to the airport?", "去机场需要几站？", "direction"},
	{"do I need to transfer", "/duː aɪ niːd tʊ trænsˈfɜːr/", "我需要换乘吗", "Do I need to transfer to get there?", "去那里需要换乘吗？", "direction"},
	{"can I walk there", "/kæn aɪ wɔːk ðeər/", "我可以步行去吗", "Can I walk there from the hotel?", "从酒店可以步行去吗？", "direction"},
	{"where exactly are we", "/weər ɪɡˈzæktli ɑːr wiː/", "我们确切在哪里", "Can you show me where exactly we are?", "你能告诉我们确切在哪里吗？", "direction"},
	{"which direction is north", "/wɪtʃ dɪˈrekʃn ɪz nɔːrθ/", "哪个方向是北边", "Which direction is north from here?", "从这里哪个方向是北边？", "direction"},
	{"is there a shortcut", "/ɪz ðeər ə ˈʃɔːtkʌt/", "有捷径吗", "Is there a shortcut to the station?", "去车站有捷径吗？", "direction"},
	{"follow the signs", "/ˈfɒləʊ ðə saɪnz/", "跟着路标走", "Just follow the signs to the terminal.", "只需跟着路标走到航站楼。", "direction"},
	{"it's about a ten minute walk", "/ɪts əˈbaʊt ə ten ˈmɪnɪt wɔːk/", "大约步行十分钟", "It's about a ten minute walk from here.", "从这里大约步行十分钟。", "direction"},
	{"take the second exit", "/teɪk ðə ˈsekənd ˈeɡzɪt/", "走第二个出口", "At the roundabout, take the second exit.", "在环形交叉路口走第二个出口。", "direction"},
	{"go past the church", "/ɡəʊ pɑːst ðə tʃɜːrtʃ/", "路过教堂继续走", "Go past the church and turn left.", "路过教堂后左转。", "direction"},
	{"you can't miss it", "/jʊ kɑːnt mɪs ɪt/", "你不会错过的", "It's a big red building, you can't miss it.", "那是一栋大红色建筑，你不会错过的。", "direction"},
	{"I think I'm lost", "/aɪ θɪŋk aɪm lɒst/", "我觉得我迷路了", "I think I'm lost, can you help me?", "我觉得我迷路了，你能帮我吗？", "direction"},

	// ===== 购物扩充（shopping） =====
	{"I'm just looking", "/aɪm dʒʌst ˈlʊkɪŋ/", "我只是随便看看", "I'm just looking, thank you.", "我只是随便看看，谢谢。", "shopping"},
	{"do you have this in another colour", "/duː jʊ hæv ðɪs ɪn əˈnʌðər ˈkʌlər/", "这个有其他颜色吗", "Do you have this in another colour?", "这个有其他颜色吗？", "shopping"},
	{"do you have a larger size", "/duː jʊ hæv ə ˈlɑːrdʒər saɪz/", "有大一号的吗", "Do you have this in a larger size?", "这个有大一号的吗？", "shopping"},
	{"do you have a smaller size", "/duː jʊ hæv ə ˈsmɔːlər saɪz/", "有小一号的吗", "Do you have this in a smaller size?", "这个有小一号的吗？", "shopping"},
	{"can I try this on", "/kæn aɪ traɪ ðɪs ɒn/", "我可以试穿吗", "Can I try this on, please?", "我可以试穿这件吗？", "shopping"},
	{"is this handmade", "/ɪz ðɪs ˈhændmeɪd/", "这是手工制作的吗", "Is this handmade by local artisans?", "这是当地工匠手工制作的吗？", "shopping"},
	{"is this authentic", "/ɪz ðɪs ɔːˈθentɪk/", "这是正品吗", "Is this authentic or a copy?", "这是正品还是复制品？", "shopping"},
	{"I'd like a refund", "/aɪd laɪk ə ˈriːfʌnd/", "我想退款", "I'd like a refund for this item.", "我想退掉这件商品。", "shopping"},
	{"this is damaged", "/ðɪs ɪz ˈdæmɪdʒd/", "这个坏了", "This item is damaged, can I exchange it?", "这件商品坏了，我可以换货吗？", "shopping"},
	{"do you have a warranty", "/duː jʊ hæv ə ˈwɒrənti/", "有保修吗", "Does this product come with a warranty?", "这个产品有保修吗？", "shopping"},
	{"can you give me a discount", "/kæn jʊ ɡɪv miː ə ˈdɪskaʊnt/", "可以给我折扣吗", "Can you give me a discount if I buy two?", "如果我买两件，可以给我折扣吗？", "shopping"},
	{"what's your best price", "/wɒts jɔːr best praɪs/", "你们最低价格是多少", "What's your best price for this bag?", "这个包你们最低价格是多少？", "shopping"},
	{"I'll take it", "/aɪl teɪk ɪt/", "我要了", "The price is good, I'll take it.", "价格不错，我要了。", "shopping"},
	{"do you ship internationally", "/duː jʊ ʃɪp ˌɪntəˈnæʃnəli/", "你们国际配送吗", "Do you ship internationally to China?", "你们配送到中国吗？", "shopping"},
	{"this doesn't fit", "/ðɪs ˈdʌznt fɪt/", "这个不合适", "This jacket doesn't fit me, it's too tight.", "这件夹克不合适，太紧了。", "shopping"},

	// ===== 紧急情况扩充（emergency） =====
	{"I need a doctor urgently", "/aɪ niːd ə ˈdɒktər ˈɜːrdʒəntli/", "我急需看医生", "I need a doctor urgently, I feel very ill.", "我急需看医生，我感觉很不舒服。", "emergency"},
	{"I've been robbed", "/aɪv biːn rɒbd/", "我被抢劫了", "I've been robbed, please call the police.", "我被抢劫了，请报警。", "emergency"},
	{"my bag was stolen", "/maɪ bæɡ wɒz ˈstəʊlən/", "我的包被偷了", "My bag was stolen at the market.", "我的包在市场被偷了。", "emergency"},
	{"I'm having chest pain", "/aɪm ˈhævɪŋ tʃest peɪn/", "我胸口痛", "I'm having chest pain, please help me.", "我胸口痛，请帮帮我。", "emergency"},
	{"I can't breathe", "/aɪ kɑːnt briːð/", "我喘不过气", "I can't breathe, I need help immediately.", "我喘不过气，我立刻需要帮助。", "emergency"},
	{"I feel faint", "/aɪ fiːl feɪnt/", "我感觉要晕倒了", "I feel faint, I need to sit down.", "我感觉要晕倒了，我需要坐下来。", "emergency"},
	{"I'm diabetic", "/aɪm ˌdaɪəˈbetɪk/", "我是糖尿病患者", "I'm diabetic and need sugar now.", "我是糖尿病患者，现在需要糖分。", "emergency"},
	{"I have high blood pressure", "/aɪ hæv haɪ blʌd ˈpreʃər/", "我有高血压", "I have high blood pressure and need my medication.", "我有高血压，需要我的药。", "emergency"},
	{"I need an interpreter", "/aɪ niːd ən ɪnˈtɜːrprɪtər/", "我需要翻译", "I need an interpreter, I don't speak English well.", "我需要翻译，我英语说得不好。", "emergency"},
	{"I've lost my phone", "/aɪv lɒst maɪ fəʊn/", "我的手机丢了", "I've lost my phone, can you help me?", "我的手机丢了，你能帮我吗？", "emergency"},
	{"I need to go to hospital", "/aɪ niːd tʊ ɡəʊ tʊ ˈhɒspɪtl/", "我需要去医院", "I need to go to hospital immediately.", "我需要立刻去医院。", "emergency"},
	{"do you have pain relief", "/duː jʊ hæv peɪn rɪˈliːf/", "你有止痛药吗", "Do you have any pain relief medication?", "你有止痛药吗？", "emergency"},
	{"I have a fever", "/aɪ hæv ə ˈfiːvər/", "我发烧了", "I have a fever and a headache.", "我发烧了，而且头疼。", "emergency"},
	{"I was in an accident", "/aɪ wɒz ɪn ən ˈæksɪdənt/", "我出了事故", "I was in an accident and need help.", "我出了事故，需要帮助。", "emergency"},
	{"where is the nearest clinic", "/weər ɪz ðə ˈnɪərɪst ˈklɪnɪk/", "最近的诊所在哪里", "Where is the nearest clinic or hospital?", "最近的诊所或医院在哪里？", "emergency"},

	// ===== 交通扩充（transportation） =====
	{"which platform for", "/wɪtʃ ˈplætfɔːrm fɔːr/", "几号站台去…", "Which platform for the train to Paris?", "去巴黎的火车在几号站台？", "transportation"},
	{"is this seat taken", "/ɪz ðɪs siːt ˈteɪkən/", "这个座位有人吗", "Excuse me, is this seat taken?", "打扰一下，这个座位有人吗？", "transportation"},
	{"does this bus go to", "/dʌz ðɪs bʌs ɡəʊ tʊ/", "这路公交去…吗", "Does this bus go to the city center?", "这路公交去市中心吗？", "transportation"},
	{"where should I get off", "/weər ʃʊd aɪ ɡet ɒf/", "我应该在哪里下车", "Where should I get off for the museum?", "去博物馆我应该在哪里下车？", "transportation"},
	{"can I buy a ticket on board", "/kæn aɪ baɪ ə ˈtɪkɪt ɒn bɔːrd/", "可以在车上买票吗", "Can I buy a ticket on board?", "可以在车上买票吗？", "transportation"},
	{"is there a day pass", "/ɪz ðeər ə deɪ pɑːs/", "有日票吗", "Is there a day pass for all buses?", "有适用于所有公交的日票吗？", "transportation"},
	{"how much is the fare", "/haʊ mʌtʃ ɪz ðə feər/", "票价是多少", "How much is the fare to the airport?", "去机场的票价是多少？", "transportation"},
	{"the train is late", "/ðə treɪn ɪz leɪt/", "火车晚点了", "The train is late by thirty minutes.", "火车晚点了三十分钟。", "transportation"},
	{"is there a faster option", "/ɪz ðeər ə ˈfɑːstər ˈɒpʃn/", "有更快的选择吗", "Is there a faster option than the bus?", "有比公交更快的选择吗？", "transportation"},
	{"book a taxi in advance", "/bʊk ə ˈtæksi ɪn ədˈvɑːns/", "提前预约出租车", "I'd like to book a taxi in advance.", "我想提前预约一辆出租车。", "transportation"},
	{"the taxi driver took a wrong turn", "/ðə ˈtæksi ˈdraɪvər tʊk ə rɒŋ tɜːrn/", "出租车司机走错路了", "The taxi driver took a wrong turn.", "出租车司机走错路了。", "transportation"},
	{"how often does it run", "/haʊ ˈɒfən dʌz ɪt rʌn/", "多久一班", "How often does the subway run here?", "这里地铁多久一班？", "transportation"},
	{"is there a night service", "/ɪz ðeər ə naɪt ˈsɜːrvɪs/", "有夜班车吗", "Is there a night service after midnight?", "午夜后有夜班车吗？", "transportation"},
	{"can I rent a bicycle", "/kæn aɪ rent ə ˈbaɪsɪkl/", "我可以租一辆自行车吗", "Can I rent a bicycle for the day?", "我可以租一辆自行车用一天吗？", "transportation"},
	{"where is the car park", "/weər ɪz ðə kɑːr pɑːrk/", "停车场在哪里", "Where is the nearest car park?", "最近的停车场在哪里？", "transportation"},

	// ===== 景点娱乐扩充（entertainment） =====
	{"what's on tonight", "/wɒts ɒn təˈnaɪt/", "今晚有什么节目", "What's on tonight at the theater?", "今晚剧院有什么演出？", "entertainment"},
	{"can I buy tickets here", "/kæn aɪ baɪ ˈtɪkɪts hɪər/", "我可以在这里买票吗", "Can I buy tickets here for the tour?", "我可以在这里购买游览票吗？", "entertainment"},
	{"is photography allowed", "/ɪz fəˈtɒɡrəfi əˈlaʊd/", "允许拍照吗", "Is photography allowed inside the museum?", "博物馆内允许拍照吗？", "entertainment"},
	{"are there student discounts", "/ɑːr ðeər ˈstjuːdənt ˈdɪskaʊnts/", "有学生折扣吗", "Are there student discounts for the entrance?", "门票有学生折扣吗？", "entertainment"},
	{"how long does the tour last", "/haʊ lɒŋ dʌz ðə tʊər lɑːst/", "游览要多长时间", "How long does the guided tour last?", "导游带队游览要多长时间？", "entertainment"},
	{"is there an English audio guide", "/ɪz ðeər ən ˈɪŋɡlɪʃ ˈɔːdiəʊ ɡaɪd/", "有英文语音导览吗", "Is there an English audio guide available?", "有英文语音导览可用吗？", "entertainment"},
	{"where is the best viewpoint", "/weər ɪz ðə best ˈvjuːpɔɪnt/", "最佳观景点在哪里", "Where is the best viewpoint for photos?", "拍照的最佳观景点在哪里？", "entertainment"},
	{"is it crowded at this time", "/ɪz ɪt ˈkraʊdɪd æt ðɪs taɪm/", "这个时候人多吗", "Is it very crowded at this time of day?", "一天中这个时候人多吗？", "entertainment"},
	{"I'd like to join a tour group", "/aɪd laɪk tʊ dʒɔɪn ə tʊər ɡruːp/", "我想加入旅游团", "I'd like to join a tour group for this site.", "我想加入这个景点的旅游团。", "entertainment"},
	{"is there a sunset cruise", "/ɪz ðeər ə ˈsʌnset kruːz/", "有日落游船吗", "Is there a sunset cruise on the harbor?", "港口有日落游船吗？", "entertainment"},
	{"where can I rent snorkel gear", "/weər kæn aɪ rent ˈsnɔːrkl ɡɪər/", "哪里可以租浮潜装备", "Where can I rent snorkel gear?", "哪里可以租浮潜装备？", "entertainment"},
	{"this view is breathtaking", "/ðɪs vjuː ɪz ˈbreθteɪkɪŋ/", "这景色令人叹为观止", "This view is absolutely breathtaking!", "这景色真是令人叹为观止！", "entertainment"},
	{"can I book in advance", "/kæn aɪ bʊk ɪn ədˈvɑːns/", "可以提前预订吗", "Can I book the tour in advance online?", "可以提前在网上预订游览吗？", "entertainment"},
	{"what is the dress code", "/wɒt ɪz ðə dres kəʊd/", "着装要求是什么", "What is the dress code for this temple?", "参观这座寺庙的着装要求是什么？", "entertainment"},
	{"is the museum free", "/ɪz ðə mjuːˈziːəm friː/", "博物馆免费吗", "Is the museum free on weekends?", "博物馆周末免费吗？", "entertainment"},
	{"I'd like a souvenir", "/aɪd laɪk ə suːˈvɪənɪr/", "我想买纪念品", "I'd like to buy a souvenir for my family.", "我想给家人买一件纪念品。", "entertainment"},
	{"where is the gift shop", "/weər ɪz ðə ɡɪft ʃɒp/", "礼品店在哪里", "Where is the gift shop in this museum?", "这家博物馆的礼品店在哪里？", "entertainment"},
	{"can I take a selfie here", "/kæn aɪ teɪk ə ˈselfi hɪər/", "我可以在这里自拍吗", "Can I take a selfie here?", "我可以在这里自拍吗？", "entertainment"},
	{"what time does it close", "/wɒt taɪm dʌz ɪt kləʊz/", "几点关门", "What time does the park close today?", "公园今天几点关门？", "entertainment"},
	{"is there a free walking tour", "/ɪz ðeər ə friː ˈwɔːkɪŋ tʊər/", "有免费步行游览吗", "Is there a free walking tour of the old town?", "有老城区的免费步行游览吗？", "entertainment"},

	// ===== 通用交际补充（general extra） =====
	{"I'm a tourist", "/aɪm ə ˈtʊərɪst/", "我是游客", "I'm a tourist visiting for the first time.", "我是第一次来的游客。", "general"},
	{"I'm from China", "/aɪm frɒm ˈtʃaɪnə/", "我来自中国", "I'm from China, this is my first trip.", "我来自中国，这是我的第一次旅行。", "general"},
	{"do you have a map", "/duː jʊ hæv ə mæp/", "你有地图吗", "Do you have a free map of the city?", "你有免费的城市地图吗？", "general"},
	{"can you recommend", "/kæn jʊ ˌrekəˈmend/", "你能推荐吗", "Can you recommend a good restaurant nearby?", "你能推荐一家附近的好餐厅吗？", "general"},
	{"what is this place", "/wɒt ɪz ðɪs pleɪs/", "这个地方是什么", "What is this place called?", "这个地方叫什么名字？", "general"},
	{"how do I get there", "/haʊ duː aɪ ɡet ðeər/", "我怎么去那里", "How do I get there by public transport?", "乘公共交通怎么去那里？", "general"},
	{"is it far from here", "/ɪz ɪt fɑːr frɒm hɪər/", "离这里远吗", "Is it far from here?", "离这里远吗？", "general"},
	{"can I pay by card", "/kæn aɪ peɪ baɪ kɑːrd/", "我可以刷卡吗", "Can I pay by card or cash only?", "可以刷卡还是只收现金？", "general"},
	{"please call a taxi", "/pliːz kɔːl ə ˈtæksi/", "请帮我叫出租车", "Please call a taxi for me.", "请帮我叫一辆出租车。", "general"},
	{"I'm checking in", "/aɪm ˈtʃekɪŋ ɪn/", "我要办理入住", "I'm checking in. My name is Li Ming.", "我要办理入住，我叫李明。", "general"},
	{"I need a receipt", "/aɪ niːd ə rɪˈsiːt/", "我需要收据", "I need a receipt for this purchase.", "我需要这次购物的收据。", "general"},
	{"is this seat free", "/ɪz ðɪs siːt friː/", "这个座位有人吗", "Excuse me, is this seat free?", "打扰一下，这个座位有人吗？", "general"},
	{"could you take a photo", "/kʊd jʊ teɪk ə ˈfəʊtəʊ/", "能帮我拍照吗", "Could you take a photo of me?", "能帮我拍张照吗？", "general"},
	{"what's the exchange rate", "/wɒts ðə ɪksˈtʃeɪndʒ reɪt/", "汇率是多少", "What's the exchange rate today?", "今天的汇率是多少？", "general"},
	{"where can I change money", "/weər kæn aɪ tʃeɪndʒ ˈmʌni/", "我在哪里可以换钱", "Where can I change money nearby?", "附近哪里可以换钱？", "general"},

	// ===== 机场补充（airport extra） =====
	{"is duty free open", "/ɪz ˈdjuːti friː ˈəʊpən/", "免税店开着吗", "Is the duty free shop open now?", "免税店现在开着吗？", "airport"},
	{"I need a boarding pass", "/aɪ niːd ə ˈbɔːrdɪŋ pɑːs/", "我需要登机牌", "I need to print my boarding pass.", "我需要打印登机牌。", "airport"},
	{"which terminal", "/wɪtʃ ˈtɜːrmɪnəl/", "哪个航站楼", "Which terminal does my flight depart from?", "我的航班从哪个航站楼出发？", "airport"},
	{"where is baggage claim", "/weər ɪz ˈbæɡɪdʒ kleɪm/", "行李提取处在哪里", "Where is the baggage claim area?", "行李提取处在哪里？", "airport"},
	{"my flight number is", "/maɪ flaɪt ˈnʌmbər ɪz/", "我的航班号是", "My flight number is CA123.", "我的航班号是CA123。", "airport"},
	{"I have a stopover", "/aɪ hæv ə ˈstɒpəʊvər/", "我有中转", "I have a three-hour stopover in Dubai.", "我在迪拜有三小时中转。", "airport"},
	{"is there a lounge", "/ɪz ðeər ə laʊndʒ/", "有休息室吗", "Is there a business lounge I can use?", "有可以使用的商务休息室吗？", "airport"},
	{"I need to pick up luggage", "/aɪ niːd tə pɪk ʌp ˈlʌɡɪdʒ/", "我需要取行李", "I need to pick up my luggage at carousel 3.", "我需要在3号传送带取行李。", "airport"},
	{"when does boarding start", "/wen dʌz ˈbɔːrdɪŋ stɑːrt/", "什么时候开始登机", "When does boarding start for flight CA100?", "CA100航班什么时候开始登机？", "airport"},
	{"aisle or window seat", "/aɪl ɔːr ˈwɪndəʊ siːt/", "靠走道还是靠窗", "Do you prefer an aisle or window seat?", "你喜欢靠走道还是靠窗的座位？", "airport"},

	// ===== 酒店补充（hotel extra） =====
	{"what's the check-out time", "/wɒts ðə tʃek aʊt taɪm/", "退房时间是几点", "What's the check-out time?", "退房时间是几点？", "hotel"},
	{"can I have a late check-out", "/kæn aɪ hæv ə leɪt tʃek aʊt/", "可以延迟退房吗", "Can I have a late check-out until 2pm?", "可以延迟到下午2点退房吗？", "hotel"},
	{"is there a gym", "/ɪz ðeər ə dʒɪm/", "有健身房吗", "Is there a gym in this hotel?", "这家酒店有健身房吗？", "hotel"},
	{"what floor is my room", "/wɒt flɔːr ɪz maɪ ruːm/", "我的房间在几楼", "What floor is my room on?", "我的房间在几楼？", "hotel"},
	{"can I get a room upgrade", "/kæn aɪ ɡet ə ruːm ˈʌpɡreɪd/", "可以升级房间吗", "Can I get a room upgrade?", "可以给我升级房间吗？", "hotel"},
	{"where is the swimming pool", "/weər ɪz ðə ˈswɪmɪŋ puːl/", "游泳池在哪里", "Where is the hotel swimming pool?", "酒店游泳池在哪里？", "hotel"},
	{"can you arrange an airport transfer", "/kæn jʊ əˈreɪndʒ æn ˈeərpɔːrt ˈtrænsfɜːr/", "能安排机场接送吗", "Can you arrange an airport transfer for me?", "能帮我安排机场接送吗？", "hotel"},
	{"I left something in my room", "/aɪ left ˈsʌmθɪŋ ɪn maɪ ruːm/", "我把东西落在房间里了", "I left my passport in my room.", "我把护照落在房间里了。", "hotel"},
	{"can I have a wake-up call", "/kæn aɪ hæv ə weɪk ʌp kɔːl/", "可以叫醒服务吗", "Can I have a wake-up call at 7am?", "能在早上7点帮我叫醒吗？", "hotel"},
	{"is there parking available", "/ɪz ðeər ˈpɑːrkɪŋ əˈveɪləbl/", "有停车位吗", "Is there parking available at the hotel?", "酒店有停车位吗？", "hotel"},

	// ===== 餐厅补充（restaurant extra） =====
	{"I'd like to order", "/aɪd laɪk tə ˈɔːrdər/", "我想点餐", "I'd like to order now, please.", "我现在想点餐，谢谢。", "restaurant"},
	{"what are today's specials", "/wɒt ɑːr təˈdeɪz ˈspeʃəlz/", "今日特餐有哪些", "What are today's specials?", "今日特餐有哪些？", "restaurant"},
	{"can I split the bill", "/kæn aɪ splɪt ðə bɪl/", "可以分开结账吗", "Can we split the bill please?", "我们可以分开结账吗？", "restaurant"},
	{"do you have a kids menu", "/duː jʊ hæv ə kɪdz ˈmenjuː/", "有儿童菜单吗", "Do you have a kids menu?", "你们有儿童菜单吗？", "restaurant"},
	{"the food is excellent", "/ðə fuːd ɪz ˈeksələnt/", "食物非常好", "The food here is excellent!", "这里的食物非常好！", "restaurant"},
	{"I'd like some more bread", "/aɪd laɪk sʌm mɔːr bred/", "我想再要一些面包", "Could I have some more bread please?", "请再给我一些面包好吗？", "restaurant"},
	{"is this dish popular", "/ɪz ðɪs dɪʃ ˈpɒpjʊlər/", "这道菜受欢迎吗", "Is this dish popular with tourists?", "这道菜受游客欢迎吗？", "restaurant"},
	{"can I have chopsticks", "/kæn aɪ hæv ˈtʃɒpstɪks/", "可以给我筷子吗", "Can I have chopsticks please?", "能给我筷子吗？", "restaurant"},
	{"no ice please", "/nəʊ aɪs pliːz/", "不要加冰", "No ice please, just room temperature.", "不要加冰，常温就好。", "restaurant"},
	{"this tastes amazing", "/ðɪs teɪsts əˈmeɪzɪŋ/", "这个味道棒极了", "This tastes amazing, what's in it?", "这个味道棒极了，里面有什么？", "restaurant"},

	// ===== 购物补充（shopping extra） =====
	{"do you have this in stock", "/duː jʊ hæv ðɪs ɪn stɒk/", "这个有货吗", "Do you have this model in stock?", "这款有货吗？", "shopping"},
	{"can I get a tax refund", "/kæn aɪ ɡet ə tæks ˈriːfʌnd/", "可以退税吗", "Can I get a tax refund as a tourist?", "作为游客可以退税吗？", "shopping"},
	{"where is the fitting room", "/weər ɪz ðə ˈfɪtɪŋ ruːm/", "试衣间在哪里", "Where is the fitting room?", "试衣间在哪里？", "shopping"},
	{"I'd like to exchange this", "/aɪd laɪk tə ɪksˈtʃeɪndʒ ðɪs/", "我想换货", "I'd like to exchange this for a different size.", "我想换一个不同尺码。", "shopping"},
	{"what are your opening hours", "/wɒt ɑːr jɔːr ˈəʊpənɪŋ aʊərz/", "营业时间是几点", "What are your opening hours?", "你们的营业时间是几点？", "shopping"},
	{"do you deliver", "/duː jʊ dɪˈlɪvər/", "你们送货吗", "Do you deliver to hotels?", "你们送货到酒店吗？", "shopping"},
	{"is there a sale on", "/ɪz ðeər ə seɪl ɒn/", "现在有打折吗", "Is there a sale on at the moment?", "现在有打折活动吗？", "shopping"},
	{"this is a gift", "/ðɪs ɪz ə ɡɪft/", "这是礼物", "This is a gift, can you wrap it nicely?", "这是礼物，能包装好看一点吗？", "shopping"},
	{"I'll think about it", "/aɪl θɪŋk əˈbaʊt ɪt/", "我考虑一下", "I'll think about it and come back.", "我考虑一下再回来。", "shopping"},
	{"do you have a loyalty card", "/duː jʊ hæv ə ˈlɔɪəlti kɑːrd/", "有会员卡吗", "Do you have a loyalty card?", "你们有会员卡吗？", "shopping"},

	// ===== 交通补充（transportation extra） =====
	{"is there a direct train", "/ɪz ðeər ə dɪˈrekt treɪn/", "有直达列车吗", "Is there a direct train to the city centre?", "有去市中心的直达列车吗？", "transportation"},
	{"I need to get to the airport", "/aɪ niːd tə ɡet tə ðə ˈeərpɔːrt/", "我需要去机场", "I need to get to the airport by 9am.", "我需要在上午9点前到机场。", "transportation"},
	{"can I bring my luggage", "/kæn aɪ brɪŋ maɪ ˈlʌɡɪdʒ/", "我可以带行李吗", "Can I bring my suitcase on the bus?", "我可以把行李箱带上公共汽车吗？", "transportation"},
	{"is it the next stop", "/ɪz ɪt ðə nekst stɒp/", "是下一站吗", "Is the museum the next stop?", "博物馆是下一站吗？", "transportation"},
	{"how do I validate my ticket", "/haʊ duː aɪ ˈvælɪdeɪt maɪ ˈtɪkɪt/", "怎么验票", "How do I validate my ticket on this bus?", "这辆公共汽车怎么验票？", "transportation"},
	{"is there a luggage storage", "/ɪz ðeər ə ˈlʌɡɪdʒ ˈstɔːrɪdʒ/", "有寄存行李的地方吗", "Is there luggage storage at the station?", "车站有寄存行李的地方吗？", "transportation"},
	{"I'd like a single ticket", "/aɪd laɪk ə ˈsɪŋɡəl ˈtɪkɪt/", "我想买单程票", "I'd like a single ticket to downtown.", "我想买一张去市中心的单程票。", "transportation"},
	{"I'd like a return ticket", "/aɪd laɪk ə rɪˈtɜːrn ˈtɪkɪt/", "我想买往返票", "I'd like a return ticket to the airport.", "我想买一张去机场的往返票。", "transportation"},
	{"what time is the last train", "/wɒt taɪm ɪz ðə lɑːst treɪn/", "最后一班列车几点", "What time is the last train tonight?", "今晚最后一班列车几点？", "transportation"},
	{"is there a bus to the beach", "/ɪz ðeər ə bʌs tə ðə biːtʃ/", "有去海滩的公交吗", "Is there a bus to the beach from here?", "从这里有去海滩的公交吗？", "transportation"},

	// ===== 紧急补充（emergency extra） =====
	{"call the police", "/kɔːl ðə pəˈliːs/", "叫警察", "Please call the police immediately!", "请立刻叫警察！", "emergency"},
	{"I need an ambulance", "/aɪ niːd æn ˈæmbjʊləns/", "我需要救护车", "I need an ambulance, it's urgent!", "我需要救护车，很紧急！", "emergency"},
	{"I've lost my passport", "/aɪv lɒst maɪ ˈpɑːspɔːrt/", "我的护照丢了", "I've lost my passport, what should I do?", "我的护照丢了，我该怎么办？", "emergency"},
	{"where is the embassy", "/weər ɪz ðə ˈembəsi/", "大使馆在哪里", "Where is the Chinese embassy?", "中国大使馆在哪里？", "emergency"},
	{"I've been overcharged", "/aɪv biːn ˌəʊvərˈtʃɑːrdʒd/", "我被多收费了", "I think I've been overcharged.", "我觉得我被多收费了。", "emergency"},
	{"I need to make a call", "/aɪ niːd tə meɪk ə kɔːl/", "我需要打电话", "I need to make an emergency call.", "我需要打紧急电话。", "emergency"},
	{"I have travel insurance", "/aɪ hæv ˈtrævəl ɪnˈʃʊərəns/", "我有旅游保险", "I have travel insurance, who do I call?", "我有旅游保险，我该打给谁？", "emergency"},
	{"where is the police station", "/weər ɪz ðə pəˈliːs ˈsteɪʃən/", "警察局在哪里", "Where is the nearest police station?", "最近的警察局在哪里？", "emergency"},
	{"I need a translator", "/aɪ niːd ə trænsˈleɪtər/", "我需要翻译", "I need a Chinese translator please.", "我需要一名中文翻译。", "emergency"},
	{"I have an allergy", "/aɪ hæv æn ˈælərdʒi/", "我有过敏症", "I have a severe nut allergy.", "我对坚果严重过敏。", "emergency"},

	// ===== 景点与文化（entertainment extra2） =====
	{"what's the admission fee", "/wɒts ðə ədˈmɪʃən fiː/", "门票多少钱", "What's the admission fee for adults?", "成人门票多少钱？", "entertainment"},
	{"are there guided tours", "/ɑːr ðeər ˈɡaɪdɪd tʊərz/", "有导游游览吗", "Are there guided tours in English?", "有英文导游游览吗？", "entertainment"},
	{"when is the next show", "/wen ɪz ðə nekst ʃəʊ/", "下一场表演是什么时候", "When is the next show?", "下一场表演是什么时候？", "entertainment"},
	{"where can I rent a bike", "/weər kæn aɪ rent ə baɪk/", "哪里可以租自行车", "Where can I rent a bike around here?", "附近哪里可以租自行车？", "entertainment"},
	{"is this a UNESCO site", "/ɪz ðɪs ə ˌjuːnesˈkəʊ saɪt/", "这是联合国教科文组织遗址吗", "Is this a UNESCO World Heritage site?", "这是联合国教科文组织世界遗产吗？", "entertainment"},
	{"the scenery is beautiful", "/ðə ˈsiːnəri ɪz ˈbjuːtɪfəl/", "风景很美", "The scenery here is absolutely beautiful.", "这里的风景真的很美。", "entertainment"},
	{"is there a night market", "/ɪz ðeər ə naɪt ˈmɑːrkɪt/", "有夜市吗", "Is there a night market nearby?", "附近有夜市吗？", "entertainment"},
	{"what time does it open", "/wɒt taɪm dʌz ɪt ˈəʊpən/", "几点开门", "What time does the attraction open?", "景点几点开门？", "entertainment"},
	{"is there an app for this", "/ɪz ðeər æn æp fɔːr ðɪs/", "这个有应用程序吗", "Is there an official app for this museum?", "这家博物馆有官方应用程序吗？", "entertainment"},
	{"I'd like a group photo", "/aɪd laɪk ə ɡruːp ˈfəʊtəʊ/", "我想拍合影", "Could you take a group photo for us?", "能帮我们拍张合影吗？", "entertainment"},

	// ===== 问路补充（direction extra2） =====
	{"is there a shortcut", "/ɪz ðeər ə ˈʃɔːtkʌt/", "有捷径吗", "Is there a shortcut to the train station?", "有去火车站的捷径吗？", "direction"},
	{"can you point me in the right direction", "/kæn jʊ pɔɪnt miː ɪn ðə raɪt dɪˈrekʃən/", "能给我指个方向吗", "Can you point me in the right direction?", "能给我指个方向吗？", "direction"},
	{"I've taken the wrong bus", "/aɪv ˈteɪkən ðə rɒŋ bʌs/", "我坐错公交了", "I think I've taken the wrong bus.", "我觉得我坐错公共汽车了。", "direction"},
	{"how do I get back to the hotel", "/haʊ duː aɪ ɡet bæk tə ðə həʊˈtel/", "怎么回酒店", "How do I get back to my hotel from here?", "从这里怎么回我的酒店？", "direction"},
	{"is it within walking distance", "/ɪz ɪt wɪðɪn ˈwɔːkɪŋ ˈdɪstəns/", "走路能到吗", "Is the beach within walking distance?", "海滩走路能到吗？", "direction"},
	{"which exit should I take", "/wɪtʃ ˈeksɪt ʃʊd aɪ teɪk/", "我应该走哪个出口", "Which exit should I take for the museum?", "去博物馆应该走哪个出口？", "direction"},
	{"is this on the tourist map", "/ɪz ðɪs ɒn ðə ˈtʊərɪst mæp/", "这个在旅游地图上吗", "Is this place on the tourist map?", "这个地方在旅游地图上吗？", "direction"},
	{"north south east west", "/nɔːrθ saʊθ iːst west/", "东南西北", "Which direction is north from here?", "从这里往哪个方向是北？", "direction"},

	// ===== 通用补充（general extra2） =====
	{"do you understand me", "/duː jʊ ˌʌndəˈstænd miː/", "你能听懂我说的吗", "Do you understand what I'm saying?", "你能听懂我说的吗？", "general"},
	{"let me show you", "/let miː ʃəʊ jʊ/", "让我给你看", "Let me show you on my phone.", "让我在手机上给你看。", "general"},
	{"I'll use a translator", "/aɪl juːz ə trænsˈleɪtər/", "我用翻译软件", "I'll use a translator app on my phone.", "我用手机上的翻译软件。", "general"},
	{"it's my first time here", "/ɪts maɪ fɜːrst taɪm hɪər/", "这是我第一次来这里", "It's my first time visiting this country.", "这是我第一次来这个国家。", "general"},
	{"I love this place", "/aɪ lʌv ðɪs pleɪs/", "我喜欢这个地方", "I love this city, it's so vibrant!", "我喜欢这座城市，真是生机勃勃！", "general"},
	{"the weather is nice today", "/ðə ˈweðər ɪz naɪs təˈdeɪ/", "今天天气真好", "The weather is nice today for sightseeing.", "今天天气很好，适合观光。", "general"},
	{"can I take a picture here", "/kæn aɪ teɪk ə ˈpɪktʃər hɪər/", "我可以在这里拍照吗", "Can I take a picture here?", "我可以在这里拍照吗？", "general"},
	{"I'm traveling alone", "/aɪm ˈtrævəlɪŋ əˈləʊn/", "我独自旅行", "I'm traveling alone for the first time.", "我第一次独自旅行。", "general"},
	{"my Chinese is limited", "/maɪ ˌtʃaɪˈniːz ɪz ˈlɪmɪtɪd/", "我的中文有限", "Sorry, my Chinese is very limited.", "对不起，我的中文很有限。", "general"},
	{"have a safe trip", "/hæv ə seɪf trɪp/", "旅途平安", "Have a safe trip back home!", "旅途平安回家！", "general"},

	// ===== 最终补充（各分类补齐） =====
	{"are you open on Sunday", "/ɑːr jʊ ˈəʊpən ɒn ˈsʌndeɪ/", "周日营业吗", "Are you open on Sundays?", "你们周日营业吗？", "shopping"},
	{"do you offer student discount", "/duː jʊ ˈɒfər ˈstjuːdənt ˈdɪskaʊnt/", "有学生折扣吗", "Do you offer a student discount?", "你们有学生折扣吗？", "shopping"},
	{"is this waterproof", "/ɪz ðɪs ˈwɔːtərpruːf/", "这个防水吗", "Is this jacket waterproof?", "这件夹克防水吗？", "shopping"},
	{"can I pay in installments", "/kæn aɪ peɪ ɪn ɪnˈstɔːlmənts/", "可以分期付款吗", "Can I pay in monthly installments?", "可以按月分期付款吗？", "shopping"},
	{"where is the nearest ATM", "/weər ɪz ðə ˈnɪərɪst eɪ tiː em/", "最近的ATM在哪里", "Where is the nearest ATM machine?", "最近的ATM取款机在哪里？", "general"},
	{"I need to top up my card", "/aɪ niːd tə tɒp ʌp maɪ kɑːrd/", "我需要给卡充值", "I need to top up my travel card.", "我需要给旅行卡充值。", "transportation"},
	{"does this include tour guide", "/dʌz ðɪs ɪnˈkluːd tʊər ɡaɪd/", "包含导游服务吗", "Does this package include a tour guide?", "这个套餐包含导游服务吗？", "entertainment"},
	{"is there a cloakroom", "/ɪz ðeər ə ˈkləʊkruːm/", "有存衣处吗", "Is there a cloakroom at the entrance?", "入口处有存衣处吗？", "entertainment"},
	{"the flight has been cancelled", "/ðə flaɪt hæz biːn ˈkænsəld/", "航班已被取消", "I was told my flight has been cancelled.", "我被告知我的航班已被取消。", "airport"},
	{"can I check my bag in early", "/kæn aɪ tʃek maɪ bæɡ ɪn ˈɜːrli/", "可以提前托运行李吗", "Can I check my bag in early?", "可以提前托运行李吗？", "airport"},
	{"I'd like a non-smoking room", "/aɪd laɪk ə nɒn ˈsməʊkɪŋ ruːm/", "我想要一间禁烟房", "I'd like a non-smoking room please.", "我想要一间禁烟房。", "hotel"},
	{"is room service available", "/ɪz ruːm ˈsɜːrvɪs əˈveɪləbl/", "有客房服务吗", "Is room service available after midnight?", "午夜之后有客房服务吗？", "hotel"},
	{"could I have extra pillows", "/kʊd aɪ hæv ˈekstrə ˈpɪləʊz/", "可以再要几个枕头吗", "Could I have two extra pillows please?", "可以再给我两个枕头吗？", "hotel"},
	{"I have a food allergy", "/aɪ hæv ə fuːd ˈælərdʒi/", "我有食物过敏", "I have a food allergy to shellfish.", "我对贝类食物过敏。", "restaurant"},
	{"the service was wonderful", "/ðə ˈsɜːrvɪs wɒz ˈwʌndərfəl/", "服务非常好", "The service was wonderful, thank you!", "服务非常好，谢谢！", "restaurant"},
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
