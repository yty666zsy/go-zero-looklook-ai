package dao

import (
	"context"
	"database/sql"
	"time"

	"github.com/zeromicro/go-zero/core/stores/sqlx"
)

// 告警状态
const (
	AlertPending  = 0 // 待处理
	AlertResolved = 1 // 已处置（人工点了一键处置）
	AlertIgnored  = 2 // 已忽略（人工）
	AlertCleared  = 3 // 已消除（下次巡检发现异常自己消失了）
)

// 日报类型
const (
	ReportDaily = "daily"
)

// ---------- 经营日报 ----------

type Report struct {
	Id          int64          `db:"id"`
	ReportType  string         `db:"report_type"`
	ReportDate  string         `db:"report_date"`
	Title       string         `db:"title"`
	Summary     string         `db:"summary"`
	Content     string         `db:"content"`
	Metrics     sql.NullString `db:"metrics"`
	Cards       sql.NullString `db:"cards"`
	Steps       sql.NullString `db:"steps"`
	Highlights  sql.NullString `db:"highlights"`
	Model       string         `db:"model"`
	GeneratedBy string         `db:"generated_by"`
	LatencyMs   int64          `db:"latency_ms"`
	CreateTime  time.Time      `db:"create_time"`
	UpdateTime  time.Time      `db:"update_time"`
}

type ReportWrite struct {
	ReportType  string
	ReportDate  string
	Title       string
	Summary     string
	Content     string
	Metrics     any
	Cards       any
	Steps       any
	Highlights  any
	Model       string
	GeneratedBy string
	LatencyMs   int64
}

const reportCols = "id, report_type, report_date, title, summary, content, metrics, cards, steps, highlights, model, generated_by, latency_ms, create_time, update_time"

// UpsertReport 同一天同一类型只保留一份（定时任务可能重跑，重跑即覆盖）。
func (d *Dao) UpsertReport(ctx context.Context, w ReportWrite) (int64, error) {
	metrics, err := toJSON(w.Metrics)
	if err != nil {
		return 0, err
	}
	cards, err := toJSON(w.Cards)
	if err != nil {
		return 0, err
	}
	steps, err := toJSON(w.Steps)
	if err != nil {
		return 0, err
	}
	highlights, err := toJSON(w.Highlights)
	if err != nil {
		return 0, err
	}

	old, err := d.GetReportByDate(ctx, w.ReportType, w.ReportDate)
	if err != nil {
		return 0, err
	}
	if old != nil {
		_, err = d.conn.ExecCtx(ctx,
			"update ai_report set title = ?, summary = ?, content = ?, metrics = ?, cards = ?, steps = ?, highlights = ?, "+
				"model = ?, generated_by = ?, latency_ms = ? where id = ?",
			w.Title, w.Summary, w.Content, metrics, cards, steps, highlights, w.Model, w.GeneratedBy, w.LatencyMs, old.Id)
		if err != nil {
			return 0, err
		}
		return old.Id, nil
	}

	res, err := d.conn.ExecCtx(ctx,
		"insert into ai_report (report_type, report_date, title, summary, content, metrics, cards, steps, highlights, model, generated_by, latency_ms) "+
			"values (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)",
		w.ReportType, w.ReportDate, w.Title, w.Summary, w.Content, metrics, cards, steps, highlights, w.Model, w.GeneratedBy, w.LatencyMs)
	if err != nil {
		return 0, err
	}
	return res.LastInsertId()
}

func (d *Dao) GetReport(ctx context.Context, id int64) (*Report, error) {
	var r Report
	err := d.conn.QueryRowPartialCtx(ctx, &r,
		"select "+reportCols+" from ai_report where id = ? and del_state = 0", id)
	if err == sqlx.ErrNotFound {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return &r, nil
}

func (d *Dao) GetReportByDate(ctx context.Context, reportType, date string) (*Report, error) {
	var r Report
	err := d.conn.QueryRowPartialCtx(ctx, &r,
		"select "+reportCols+" from ai_report where report_type = ? and report_date = ? and del_state = 0", reportType, date)
	if err == sqlx.ErrNotFound {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return &r, nil
}

// ListReports 列表只取摘要字段，正文按需再拉。
func (d *Dao) ListReports(ctx context.Context, limit int64) ([]Report, error) {
	if limit <= 0 || limit > 100 {
		limit = 30
	}
	var list []Report
	err := d.conn.QueryRowsPartialCtx(ctx, &list,
		"select id, report_type, report_date, title, summary, content, metrics, cards, highlights, model, generated_by, latency_ms, create_time, update_time "+
			"from ai_report where del_state = 0 order by report_date desc, id desc limit ?", limit)
	if err != nil {
		return nil, err
	}
	return list, nil
}

// ---------- 异常巡检告警 ----------

type Alert struct {
	Id          int64          `db:"id"`
	AlertKey    string         `db:"alert_key"`
	Rule        string         `db:"rule"`
	Level       string         `db:"level"`
	Category    string         `db:"category"`
	Title       string         `db:"title"`
	Detail      string         `db:"detail"`
	Suggestion  string         `db:"suggestion"`
	ActionTool  string         `db:"action_tool"`
	ActionArgs  sql.NullString `db:"action_args"`
	ActionText  string         `db:"action_text"`
	RelatedType string         `db:"related_type"`
	RelatedId   string         `db:"related_id"`
	Status      int64          `db:"status"`
	HitCount    int64          `db:"hit_count"`
	Batch       string         `db:"batch"`
	LastHitTime time.Time      `db:"last_hit_time"`
	CreateTime  time.Time      `db:"create_time"`
	UpdateTime  time.Time      `db:"update_time"`
}

// UpsertAlert 幂等写入：
//   - 首次出现 → 新增（status=待处理）
//   - 再次命中 → 更新明细并累加 hit_count，待处理的保持待处理
//   - 已被系统「自动消除」的异常再次出现 → 重新打开（说明是新发生的问题）
//   - 已被人工「处置 / 忽略」的 → 保持关闭，只累加命中次数（人工可手动重开）
//
// 返回是否本次新建。
func (d *Dao) UpsertAlert(ctx context.Context, a Alert) (bool, error) {
	args, err := toJSON(a.ActionArgs)
	if err != nil {
		return false, err
	}

	old, err := d.GetAlertByKey(ctx, a.AlertKey)
	if err != nil {
		return false, err
	}
	if old != nil {
		reopen := old.Status == AlertCleared
		if reopen {
			_, err = d.conn.ExecCtx(ctx,
				"update ai_alert set level = ?, category = ?, title = ?, detail = ?, suggestion = ?, action_tool = ?, action_args = ?, action_text = ?, "+
					"related_type = ?, related_id = ?, status = ?, hit_count = hit_count + 1, batch = ?, last_hit_time = now() where id = ?",
				a.Level, a.Category, a.Title, a.Detail, a.Suggestion, a.ActionTool, args, a.ActionText,
				a.RelatedType, a.RelatedId, AlertPending, a.Batch, old.Id)
			return false, err
		}
		_, err = d.conn.ExecCtx(ctx,
			"update ai_alert set level = ?, category = ?, title = ?, detail = ?, suggestion = ?, action_tool = ?, action_args = ?, action_text = ?, "+
				"related_type = ?, related_id = ?, hit_count = hit_count + 1, batch = ?, last_hit_time = now() where id = ?",
			a.Level, a.Category, a.Title, a.Detail, a.Suggestion, a.ActionTool, args, a.ActionText,
			a.RelatedType, a.RelatedId, a.Batch, old.Id)
		return false, err
	}

	_, err = d.conn.ExecCtx(ctx,
		"insert into ai_alert (alert_key, rule, level, category, title, detail, suggestion, action_tool, action_args, action_text, "+
			"related_type, related_id, status, hit_count, batch, last_hit_time) "+
			"values (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, 0, 1, ?, now())",
		a.AlertKey, a.Rule, a.Level, a.Category, a.Title, a.Detail, a.Suggestion, a.ActionTool, args, a.ActionText,
		a.RelatedType, a.RelatedId, a.Batch)
	return true, err
}

const alertCols = "id, alert_key, rule, level, category, title, detail, suggestion, action_tool, action_args, action_text, " +
	"related_type, related_id, status, hit_count, batch, last_hit_time, create_time, update_time"

func (d *Dao) GetAlert(ctx context.Context, id int64) (*Alert, error) {
	var a Alert
	err := d.conn.QueryRowPartialCtx(ctx, &a,
		"select "+alertCols+" from ai_alert where id = ?", id)
	if err == sqlx.ErrNotFound {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return &a, nil
}

func (d *Dao) GetAlertByKey(ctx context.Context, key string) (*Alert, error) {
	var a Alert
	err := d.conn.QueryRowPartialCtx(ctx, &a,
		"select "+alertCols+" from ai_alert where alert_key = ?", key)
	if err == sqlx.ErrNotFound {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return &a, nil
}

// ListAlerts status < 0 表示不限状态。
func (d *Dao) ListAlerts(ctx context.Context, status int64, level string, limit int64) ([]Alert, error) {
	if limit <= 0 || limit > 200 {
		limit = 50
	}
	q := "select " + alertCols + " from ai_alert where 1 = 1"
	args := []any{}
	if status >= 0 {
		q += " and status = ?"
		args = append(args, status)
	}
	if level != "" {
		q += " and level = ?"
		args = append(args, level)
	}
	// 待处理优先、等级高的优先，再看时间
	q += " order by (status = 0) desc, field(level, 'error', 'warn', 'info'), last_hit_time desc, id desc limit ?"
	args = append(args, limit)

	var list []Alert
	if err := d.conn.QueryRowsPartialCtx(ctx, &list, q, args...); err != nil {
		return nil, err
	}
	return list, nil
}

func (d *Dao) CountAlerts(ctx context.Context, status int64) (int64, error) {
	var n int64
	q := "select count(*) from ai_alert where 1 = 1"
	args := []any{}
	if status >= 0 {
		q += " and status = ?"
		args = append(args, status)
	}
	if err := d.conn.QueryRowPartialCtx(ctx, &n, q, args...); err != nil {
		return 0, err
	}
	return n, nil
}

func (d *Dao) SetAlertStatus(ctx context.Context, id, status int64) (int64, error) {
	res, err := d.conn.ExecCtx(ctx, "update ai_alert set status = ? where id = ? and status <> ?", status, id, status)
	if err != nil {
		return 0, err
	}
	return res.RowsAffected()
}

// ClearMissingAlerts 把「上次是待处理、这次巡检没再命中」的告警自动标记为已消除。
// keepKeys 为空表示本次巡检没有任何异常。
func (d *Dao) ClearMissingAlerts(ctx context.Context, keepKeys []string) (int64, error) {
	q := "update ai_alert set status = ? where status = ?"
	args := []any{AlertCleared, AlertPending}
	if len(keepKeys) > 0 {
		q += " and alert_key not in (" + placeholders(len(keepKeys)) + ")"
		for _, k := range keepKeys {
			args = append(args, k)
		}
	}
	res, err := d.conn.ExecCtx(ctx, q, args...)
	if err != nil {
		return 0, err
	}
	return res.RowsAffected()
}
