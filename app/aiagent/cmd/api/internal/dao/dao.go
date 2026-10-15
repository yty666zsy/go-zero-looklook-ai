// Package dao 是 AI Agent 服务的数据访问层。
//
// 这里只负责 Agent 自己的库（looklook_aiagent）：会话、消息、知识库、
// 待确认动作令牌、工具审计；业务数据一律通过 internal/biz 调现有接口获取。
package dao

import (
	"context"
	"database/sql"
	"encoding/json"
	"time"

	"github.com/zeromicro/go-zero/core/stores/sqlx"
)

// 场景
const (
	SceneCustomer = "c"
	SceneAdmin    = "admin"
)

// 会话
type Conversation struct {
	Id         int64     `db:"id"`
	Scene      string    `db:"scene"`
	OwnerId    int64     `db:"owner_id"`
	Title      string    `db:"title"`
	MsgCount   int64     `db:"msg_count"`
	LastMsg    string    `db:"last_msg"`
	CreateTime time.Time `db:"create_time"`
	UpdateTime time.Time `db:"update_time"`
}

// 消息
type Message struct {
	Id               int64          `db:"id"`
	ConversationId   int64          `db:"conversation_id"`
	Scene            string         `db:"scene"`
	OwnerId          int64          `db:"owner_id"`
	Role             string         `db:"role"`
	Content          string         `db:"content"`
	Steps            sql.NullString `db:"steps"`
	Cards            sql.NullString `db:"cards"`
	Model            string         `db:"model"`
	PromptTokens     int64          `db:"prompt_tokens"`
	CompletionTokens int64          `db:"completion_tokens"`
	LatencyMs        int64          `db:"latency_ms"`
	CreateTime       time.Time      `db:"create_time"`
}

// 知识库条目
type Knowledge struct {
	Id         int64     `db:"id"`
	Scene      string    `db:"scene"`
	Category   string    `db:"category"`
	Title      string    `db:"title"`
	Content    string    `db:"content"`
	Keywords   string    `db:"keywords"`
	Enabled    int64     `db:"enabled"`
	UpdateTime time.Time `db:"update_time"`
}

// 待确认动作令牌
type ActionToken struct {
	Token          string         `db:"token"`
	Scene          string         `db:"scene"`
	OwnerId        int64          `db:"owner_id"`
	ConversationId int64          `db:"conversation_id"`
	Tool           string         `db:"tool"`
	Args           sql.NullString `db:"args"`
	Summary        string         `db:"summary"`
	Status         int64          `db:"status"`
	ExpireTime     time.Time      `db:"expire_time"`
	CreateTime     time.Time      `db:"create_time"`
}

const (
	ActionStatusPending   int64 = 0
	ActionStatusDone      int64 = 1
	ActionStatusRejected  int64 = 2
	ActionStatusExpired   int64 = 3
)

type Dao struct {
	conn sqlx.SqlConn
}

func New(conn sqlx.SqlConn) *Dao {
	return &Dao{conn: conn}
}

// ---------- 会话 ----------

func (d *Dao) CreateConversation(ctx context.Context, scene string, ownerId int64, title string) (int64, error) {
	res, err := d.conn.ExecCtx(ctx,
		"insert into ai_conversation (scene, owner_id, title, last_msg) values (?, ?, ?, '')",
		scene, ownerId, title)
	if err != nil {
		return 0, err
	}
	return res.LastInsertId()
}

func (d *Dao) GetConversation(ctx context.Context, id int64, scene string, ownerId int64) (*Conversation, error) {
	var c Conversation
	err := d.conn.QueryRowPartialCtx(ctx, &c,
		"select id, scene, owner_id, title, msg_count, last_msg, create_time, update_time "+
			"from ai_conversation where id = ? and scene = ? and owner_id = ? and del_state = 0",
		id, scene, ownerId)
	if err == sqlx.ErrNotFound {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return &c, nil
}

func (d *Dao) ListConversations(ctx context.Context, scene string, ownerId int64, limit int64) ([]Conversation, error) {
	if limit <= 0 || limit > 100 {
		limit = 30
	}
	var list []Conversation
	err := d.conn.QueryRowsPartialCtx(ctx, &list,
		"select id, scene, owner_id, title, msg_count, last_msg, create_time, update_time "+
			"from ai_conversation where scene = ? and owner_id = ? and del_state = 0 "+
			"order by update_time desc, id desc limit ?",
		scene, ownerId, limit)
	if err != nil {
		return nil, err
	}
	return list, nil
}

func (d *Dao) TouchConversation(ctx context.Context, id int64, lastMsg string) error {
	_, err := d.conn.ExecCtx(ctx,
		"update ai_conversation set msg_count = msg_count + 2, last_msg = ? where id = ?",
		cut(lastMsg, 200), id)
	return err
}

func (d *Dao) DeleteConversation(ctx context.Context, id int64, scene string, ownerId int64) (int64, error) {
	res, err := d.conn.ExecCtx(ctx,
		"update ai_conversation set del_state = 1 where id = ? and scene = ? and owner_id = ? and del_state = 0",
		id, scene, ownerId)
	if err != nil {
		return 0, err
	}
	return res.RowsAffected()
}

// ---------- 消息 ----------

type MessageWrite struct {
	ConversationId   int64
	Scene            string
	OwnerId          int64
	Role             string
	Content          string
	Steps            any
	Cards            any
	Model            string
	PromptTokens     int64
	CompletionTokens int64
	LatencyMs        int64
}

func (d *Dao) AppendMessage(ctx context.Context, m MessageWrite) (int64, error) {
	steps, err := toJSON(m.Steps)
	if err != nil {
		return 0, err
	}
	cards, err := toJSON(m.Cards)
	if err != nil {
		return 0, err
	}
	res, err := d.conn.ExecCtx(ctx,
		"insert into ai_message (conversation_id, scene, owner_id, role, content, steps, cards, model, prompt_tokens, completion_tokens, latency_ms) "+
			"values (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)",
		m.ConversationId, m.Scene, m.OwnerId, m.Role, m.Content, steps, cards, m.Model,
		m.PromptTokens, m.CompletionTokens, m.LatencyMs)
	if err != nil {
		return 0, err
	}
	return res.LastInsertId()
}

// ListMessages 取最近 limit 条（按时间正序返回），用于组装上下文。
func (d *Dao) ListMessages(ctx context.Context, conversationId int64, limit int64) ([]Message, error) {
	if limit <= 0 || limit > 100 {
		limit = 20
	}
	var list []Message
	err := d.conn.QueryRowsPartialCtx(ctx, &list,
		"select id, conversation_id, role, content, create_time from ai_message "+
			"where conversation_id = ? order by id desc limit ?",
		conversationId, limit)
	if err != nil {
		return nil, err
	}
	// 反转为时间正序
	for i, j := 0, len(list)-1; i < j; i, j = i+1, j-1 {
		list[i], list[j] = list[j], list[i]
	}
	return list, nil
}

func (d *Dao) ListMessagesPage(ctx context.Context, conversationId int64, scene string, ownerId int64, limit int64) ([]Message, error) {
	if limit <= 0 || limit > 200 {
		limit = 100
	}
	var list []Message
	err := d.conn.QueryRowsPartialCtx(ctx, &list,
		"select id, conversation_id, scene, owner_id, role, content, steps, cards, model, latency_ms, create_time "+
			"from ai_message where conversation_id = ? and scene = ? and owner_id = ? order by id asc limit ?",
		conversationId, scene, ownerId, limit)
	if err != nil {
		return nil, err
	}
	return list, nil
}

// ---------- 知识库 ----------

func (d *Dao) ListKnowledge(ctx context.Context, scenes ...string) ([]Knowledge, error) {
	q := "select id, scene, category, title, content, keywords, enabled, update_time from ai_knowledge where enabled = 1"
	args := make([]any, 0, len(scenes))
	if len(scenes) > 0 {
		q += " and scene in (" + placeholders(len(scenes)) + ")"
		for _, s := range scenes {
			args = append(args, s)
		}
	}
	var list []Knowledge
	if err := d.conn.QueryRowsPartialCtx(ctx, &list, q, args...); err != nil {
		return nil, err
	}
	return list, nil
}

// AdminListKnowledge 管理端列表：不受 enabled 限制，支持关键字与场景过滤。
func (d *Dao) AdminListKnowledge(ctx context.Context, keyword, scene string) ([]Knowledge, error) {
	q := "select id, scene, category, title, content, keywords, enabled, update_time from ai_knowledge where 1 = 1"
	args := []any{}
	if keyword != "" {
		q += " and (title like ? or content like ? or keywords like ?)"
		kw := "%" + keyword + "%"
		args = append(args, kw, kw, kw)
	}
	if scene != "" {
		q += " and scene = ?"
		args = append(args, scene)
	}
	q += " order by id desc limit 500"
	var list []Knowledge
	if err := d.conn.QueryRowsPartialCtx(ctx, &list, q, args...); err != nil {
		return nil, err
	}
	return list, nil
}

func (d *Dao) UpsertKnowledge(ctx context.Context, k Knowledge) (int64, error) {
	if k.Id > 0 {
		_, err := d.conn.ExecCtx(ctx,
			"update ai_knowledge set scene = ?, category = ?, title = ?, content = ?, keywords = ?, enabled = ? where id = ?",
			k.Scene, k.Category, k.Title, k.Content, k.Keywords, k.Enabled, k.Id)
		return k.Id, err
	}
	res, err := d.conn.ExecCtx(ctx,
		"insert into ai_knowledge (scene, category, title, content, keywords, enabled) values (?, ?, ?, ?, ?, ?)",
		k.Scene, k.Category, k.Title, k.Content, k.Keywords, k.Enabled)
	if err != nil {
		return 0, err
	}
	return res.LastInsertId()
}

func (d *Dao) DeleteKnowledge(ctx context.Context, id int64) (int64, error) {
	res, err := d.conn.ExecCtx(ctx, "delete from ai_knowledge where id = ?", id)
	if err != nil {
		return 0, err
	}
	return res.RowsAffected()
}

// ---------- 待确认动作 ----------

func (d *Dao) CreateActionToken(ctx context.Context, a ActionToken) error {
	_, err := d.conn.ExecCtx(ctx,
		"insert into ai_action_token (token, scene, owner_id, conversation_id, tool, args, summary, status, expire_time) "+
			"values (?, ?, ?, ?, ?, ?, ?, 0, ?)",
		a.Token, a.Scene, a.OwnerId, a.ConversationId, a.Tool, a.Args, a.Summary, a.ExpireTime)
	return err
}

func (d *Dao) GetActionToken(ctx context.Context, token, scene string, ownerId int64) (*ActionToken, error) {
	var a ActionToken
	err := d.conn.QueryRowPartialCtx(ctx, &a,
		"select token, scene, owner_id, conversation_id, tool, args, summary, status, expire_time, create_time "+
			"from ai_action_token where token = ? and scene = ? and owner_id = ?",
		token, scene, ownerId)
	if err == sqlx.ErrNotFound {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return &a, nil
}

func (d *Dao) MarkActionToken(ctx context.Context, token string, status int64) (int64, error) {
	res, err := d.conn.ExecCtx(ctx,
		"update ai_action_token set status = ? where token = ? and status = 0", status, token)
	if err != nil {
		return 0, err
	}
	return res.RowsAffected()
}

// ---------- 工具审计 ----------

func (d *Dao) AppendAudit(ctx context.Context, conversationId int64, scene string, ownerId int64, tool string, args any, result string, status string, costMs int64) error {
	a, err := toJSON(args)
	if err != nil {
		return err
	}
	_, err = d.conn.ExecCtx(ctx,
		"insert into ai_tool_audit (conversation_id, scene, owner_id, tool, args, result, status, cost_ms) "+
			"values (?, ?, ?, ?, ?, ?, ?, ?)",
		conversationId, scene, ownerId, tool, a, cut(result, 4000), status, costMs)
	return err
}

// ---------- 小工具 ----------

func toJSON(v any) (sql.NullString, error) {
	if v == nil {
		return sql.NullString{}, nil
	}
	// sql.NullString 直接透传：它自己就是「可能为空的字符串」，
	// 再 marshal 一次会变成 {"String":"...","Valid":true} 这种错误结构。
	switch t := v.(type) {
	case sql.NullString:
		if !t.Valid || t.String == "" {
			return sql.NullString{}, nil
		}
		return t, nil
	case *sql.NullString:
		if t == nil || !t.Valid || t.String == "" {
			return sql.NullString{}, nil
		}
		return *t, nil
	case string:
		if t == "" {
			return sql.NullString{}, nil
		}
		return sql.NullString{String: t, Valid: true}, nil
	}
	b, err := json.Marshal(v)
	if err != nil {
		return sql.NullString{}, err
	}
	if string(b) == "null" || string(b) == "[]" || string(b) == "{}" {
		return sql.NullString{}, nil
	}
	return sql.NullString{String: string(b), Valid: true}, nil
}

func placeholders(n int) string {
	if n <= 0 {
		return ""
	}
	b := make([]byte, 0, n*2-1)
	for i := 0; i < n; i++ {
		if i > 0 {
			b = append(b, ',')
		}
		b = append(b, '?')
	}
	return string(b)
}

func cut(s string, n int) string {
	r := []rune(s)
	if len(r) <= n {
		return s
	}
	return string(r[:n]) + "..."
}
