package ai

import (
	"context"

	"looklook/app/aiagent/cmd/api/internal/conv"
	"looklook/app/aiagent/cmd/api/internal/dao"
	"looklook/app/aiagent/cmd/api/internal/svc"
	"looklook/app/aiagent/cmd/api/internal/types"
	"looklook/pkg/ctxdata"
	"looklook/pkg/xerr"

	"github.com/zeromicro/go-zero/core/logx"
)

type ReportListLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewReportListLogic(ctx context.Context, svcCtx *svc.ServiceContext) *ReportListLogic {
	return &ReportListLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *ReportListLogic) ReportList(req *types.ReportListReq) (resp *types.ReportListResp, err error) {
	rows, err := l.svcCtx.Dao.ListReports(l.ctx, req.Limit)
	if err != nil {
		return nil, err
	}
	return &types.ReportListResp{List: conv.Reports(rows)}, nil
}

// ---------- 详情 ----------

type ReportDetailLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewReportDetailLogic(ctx context.Context, svcCtx *svc.ServiceContext) *ReportDetailLogic {
	return &ReportDetailLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *ReportDetailLogic) ReportDetail(req *types.ReportDetailReq) (resp *types.ReportDetailResp, err error) {
	if req.Id <= 0 {
		return nil, xerr.NewErrMsg("缺少日报 id")
	}
	row, err := l.svcCtx.Dao.GetReport(l.ctx, req.Id)
	if err != nil {
		return nil, err
	}
	if row == nil {
		return nil, xerr.NewErrMsg("日报不存在")
	}
	item := conv.Report(*row, true)
	return &types.ReportDetailResp{Report: &item}, nil
}

// ---------- 生成 ----------

type ReportGenerateLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewReportGenerateLogic(ctx context.Context, svcCtx *svc.ServiceContext) *ReportGenerateLogic {
	return &ReportGenerateLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

// ReportGenerate 手动触发一次日报生成（运营在页面上点「立即生成」）。
func (l *ReportGenerateLogic) ReportGenerate(req *types.ReportGenerateReq) (resp *types.ReportGenerateResp, err error) {
	row, created, err := l.svcCtx.Insight.GenerateReport(l.ctx, req.Type, req.Date, "manual", req.Force)
	if err != nil {
		return nil, xerr.NewErrMsg(err.Error())
	}
	item := conv.Report(*row, true)
	return &types.ReportGenerateResp{Report: &item, Created: created}, nil
}

// ---------- 巡检 ----------

type PatrolRunLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewPatrolRunLogic(ctx context.Context, svcCtx *svc.ServiceContext) *PatrolRunLogic {
	return &PatrolRunLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *PatrolRunLogic) PatrolRun(_ *types.PatrolRunReq) (resp *types.PatrolRunResp, err error) {
	res, err := l.svcCtx.Insight.RunPatrol(l.ctx, "manual")
	if err != nil {
		return nil, xerr.NewErrMsg(err.Error())
	}
	return conv.Patrol(res), nil
}

// ---------- 告警 ----------

type AlertListLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewAlertListLogic(ctx context.Context, svcCtx *svc.ServiceContext) *AlertListLogic {
	return &AlertListLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *AlertListLogic) AlertList(req *types.AlertListReq) (resp *types.AlertListResp, err error) {
	status := req.Status
	if status < -1 {
		status = 0
	}
	list, err := l.svcCtx.Dao.ListAlerts(l.ctx, status, req.Level, req.Limit)
	if err != nil {
		return nil, err
	}
	pending, err := l.svcCtx.Dao.CountAlerts(l.ctx, dao.AlertPending)
	if err != nil {
		return nil, err
	}
	total, err := l.svcCtx.Dao.CountAlerts(l.ctx, -1)
	if err != nil {
		return nil, err
	}
	return &types.AlertListResp{
		List:         conv.Alerts(list),
		PendingCount: pending,
		TotalCount:   total,
	}, nil
}

type AlertActionLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewAlertActionLogic(ctx context.Context, svcCtx *svc.ServiceContext) *AlertActionLogic {
	return &AlertActionLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

// AlertAction 把告警里预置的处置动作变成一次待确认动作。
// 这一步不会改动业务数据，必须再调 /action/confirm 才会真正执行。
func (l *AlertActionLogic) AlertAction(req *types.AlertActionReq) (resp *types.AlertActionResp, err error) {
	if req.Id <= 0 {
		return nil, xerr.NewErrMsg("缺少告警 id")
	}
	uid := ctxdata.GetUidFromCtx(l.ctx)
	token, text, ttl, err := l.svcCtx.Insight.AlertAction(l.ctx, req.Id, uid)
	if err != nil {
		return nil, xerr.NewErrMsg(err.Error())
	}
	return &types.AlertActionResp{
		AlertId:     req.Id,
		Token:       token,
		ConfirmText: text,
		ExpireIn:    ttl,
	}, nil
}

type AlertStatusLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewAlertStatusLogic(ctx context.Context, svcCtx *svc.ServiceContext) *AlertStatusLogic {
	return &AlertStatusLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *AlertStatusLogic) AlertStatus(req *types.AlertStatusReq) (resp *types.AlertStatusResp, err error) {
	if req.Id <= 0 {
		return nil, xerr.NewErrMsg("缺少告警 id")
	}
	if err := l.svcCtx.Insight.SetAlertStatus(l.ctx, req.Id, req.Status); err != nil {
		return nil, xerr.NewErrMsg(err.Error())
	}
	return &types.AlertStatusResp{Ok: true, Status: req.Status}, nil
}
