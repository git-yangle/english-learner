# 伪代码文档 — English Learner

## 一、启动流程

```
main():
  加载配置（端口、数据目录）
  初始化 Repository（检查 data/ 目录，不存在则创建）
  加载所有词库到内存（from data/libraries/*.json）
  注册 HTTP 路由
  embed 前端静态文件
  启动 HTTP Server，打开浏览器 http://localhost:8080
```

---

## 二、初始化14天学习计划

```
PlanService.InitPlan(libraryID string):
  1. 检查是否已有计划 → 若存在提示覆盖确认
  2. 从词库加载所有单词列表（shuffled）
  3. 将单词平均分配到14天，每天约 15 新词
  4. 生成 DailyPlan[1..14]，每条包含：
       Day、Date（今天 + offset）、WordIDs、Status=pending
  5. 写入 data/user/plan.json
```

---

## 三、获取今日计划

```
PlanService.GetTodayPlan():
  1. 读取 plan.json，找到 Day 对应今天日期的 DailyPlan
  2. 调用 ReviewService.GetReviewWords(today) → 获取需复习单词
  3. 合并今日新词 + 复习词（去重），返回完整单词列表
  4. 按 Category 分组，便于前端展示

ReviewService.GetReviewWords(date):
  1. 读取 records.json，统计每个 WordID 的错误率
  2. 错误率 > 60%  → 加入今日复习
  3. 错误率 30-60% → 距上次学习 >= 2天 则加入
  4. 错误率 < 30%  → 艾宾浩斯间隔（1,2,4,7,15天）加入
  5. 返回需复习的 WordID 列表
```

---

## 四、学习模式

### 4.1 浏览模式

```
StudyService.Browse(day int):
  1. GetTodayPlan() 获取单词列表
  2. 按 Category 分组返回
  3. 每个单词包含：English、Phonetic、Chinese、Example、ExampleCN
  4. 前端翻卡片展示，用户可标记"认识/不认识"
  5. 标记结果作为 browse 类型 LearningRecord 写入
```

### 4.2 测试模式（Quiz）

```
StudyService.GenerateQuiz(day int, count int):
  1. 获取今日单词列表
  2. 随机选 count 个单词出题（优先选错误率高的）
  3. 每题生成4个选项：
       - 1个正确答案（中文释义）
       - 3个干扰项（从同词库随机选）
  4. 打乱选项顺序，返回题目列表

StudyService.SubmitQuizAnswer(wordID, selectedAnswer):
  1. 对比正确答案，判断 correct
  2. 写入 LearningRecord{Mode: quiz, Correct, WordID, Date}
  3. 返回是否正确 + 正确答案 + 例句
```

### 4.3 默写模式（Dictation）

```
StudyService.GetDictationWord(day int):
  1. 获取今日单词，优先返回错误率高的
  2. 只展示中文释义 + 例句（隐藏英文）
  3. 用户输入英文单词

StudyService.SubmitDictation(wordID, input string):
  1. 标准化处理：toLower、trim空格
  2. 与标准答案对比（exact match）
  3. 写入 LearningRecord{Mode: dictation, Correct, WordID}
  4. 返回是否正确 + 标准拼写 + 音标
```

---

## 五、打卡逻辑

```
CheckInService.CheckIn(day int):
  1. 统计今日所有 LearningRecord
  2. 计算各模式正确率：
       quizScore      = quiz正确数 / quiz总数
       dictationScore = dictation正确数 / dictation总数
       totalScore     = 综合加权（quiz*0.4 + dictation*0.6）
  3. 判断是否达标（totalScore >= 0.6 视为完成）
  4. 写入 CheckIn{Day, Date, Completed, Score}
  5. 更新 DailyPlan.Status = completed
  6. 触发次日计划调整（调用 ReviewService.GetReviewWords(tomorrow)）
```

---

## 六、统计数据

```
StatsService.GetOverview():
  返回：
    - 14天打卡日历（每天：日期、是否完成、得分）
    - 连续打卡天数
    - 总体正确率趋势（按天）
    - 词库掌握情况（按 Category 分组：已掌握/学习中/未学）

判断"已掌握"标准：
  该词最近3次作答正确率 = 100%
```

---

## 七、词库扩展

```
LibraryService.LoadLibraries():
  1. 扫描 data/libraries/*.json
  2. 每个文件解析为 WordLibrary
  3. 存入内存 map[libraryID]WordLibrary

用户添加新词库：
  1. 将新 JSON 文件放入 data/libraries/
  2. 调用 /api/libraries/reload 重新加载（无需重启）
```

---

## 八、数据文件结构

### data/libraries/travel.json
```json
{
  "id": "travel",
  "name": "旅游场景",
  "desc": "机场、酒店、餐厅、问路等旅游常用词",
  "words": [
    {
      "id": "w001",
      "english": "boarding pass",
      "phonetic": "/ˈbɔːrdɪŋ pæs/",
      "chinese": "登机牌",
      "example": "Please show your boarding pass at the gate.",
      "example_cn": "请在登机口出示您的登机牌。",
      "category": "airport"
    }
  ]
}
```

### data/user/plan.json
```json
{
  "library_id": "travel",
  "start_date": "2026-03-17",
  "days": [
    {
      "day": 1,
      "date": "2026-03-17",
      "word_ids": ["w001", "w002"],
      "review_word_ids": [],
      "status": "pending"
    }
  ]
}
```

### data/user/records.json
```json
[
  {
    "word_id": "w001",
    "mode": "quiz",
    "correct": true,
    "attempts": 1,
    "date": "2026-03-17"
  }
]
```

### data/user/checkins.json
```json
[
  {
    "day": 1,
    "date": "2026-03-17",
    "completed": true,
    "score": 0.85,
    "study_duration_sec": 1200
  }
]
```
