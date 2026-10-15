package config

import (
	"github.com/zeromicro/go-zero/core/stores/redis"
	"github.com/zeromicro/go-zero/rest"
)

type Config struct {
	rest.RestConf

	// C 端用户 JWT（与 usercenter-api 共用密钥）
	JwtAuth struct {
		AccessSecret string
		AccessExpire int64
	}

	// 管理端 JWT（与 admin-api 共用密钥）
	AdminJwtAuth struct {
		AccessSecret string
		AccessExpire int64
	}

	Redis redis.RedisConf

	// Agent 自己的库
	AiDB struct {
		DataSource string
	}

	// 业务服务地址：工具调用复用现有 HTTP 接口
	TravelAPI struct {
		BaseUrl string
		Timeout int64
	}
	OrderAPI struct {
		BaseUrl string
		Timeout int64
	}
	AdminAPI struct {
		BaseUrl string
		Timeout int64
	}

	LLM   LLMConf
	Agent AgentConf
	Rag   RagConf
	Cron  CronConf
}

// LLMConf 大模型配置。Provider 支持 mock / openai（OpenAI 兼容协议）。
type LLMConf struct {
	Provider    string
	BaseUrl     string
	ApiKey      string
	Model       string
	Temperature float64
	TimeoutMs   int64
	MaxTokens   int
}

// AgentConf Agent 运行时约束，防止模型陷入无限工具调用。
type AgentConf struct {
	MaxSteps      int   // 单轮最多几次「思考-行动」
	MaxHistory    int   // 带多少条历史消息进上下文
	ToolTimeoutMs int64 // 单个工具执行超时
	ActionTTLSec  int64 // 待确认写操作令牌的有效期
}

// RagConf 知识库检索参数。
type RagConf struct {
	TopK        int
	MinScore    float64
	MaxDocChars int
}

// CronConf 无人值守任务：经营日报 + 异常巡检。
//
// 这两个是「主动型 Agent」——不等用户提问，由服务自己按节奏跑，
// 把结果沉淀成日报与告警，运营只需要看结论、点处置。
type CronConf struct {
	Enabled           bool    // 总开关
	RunOnStart        bool    // 启动后先跑一次（便于演示，不用等到整点）
	PatrolIntervalMin int     // 巡检间隔（分钟）
	ReportHour        int     // 每天几点生成日报（0-23）
	ServiceUid        int64   // 内部服务账号的 adminId（ai-agent），无人值守调用以此身份发起
	MaxAlertsPerRule  int     // 单条规则单次最多产出多少条告警，避免刷屏
	PendingPayMin     int     // 待支付超过多少分钟判为异常
	IdleHomestayDays  int     // 在售房源连续多少天没有有效订单判为异常
	LowScore          float64 // 房源评分低于该值判为异常
	MinCommentCount   int64   // 触发评分异常所需的最少评价数（样本太少不算）
	HighCancelRate    float64 // 用户取消率高于该值判为异常（0-1）
	MinOrderForCancel int64   // 计算取消率所需的最少订单数
}
