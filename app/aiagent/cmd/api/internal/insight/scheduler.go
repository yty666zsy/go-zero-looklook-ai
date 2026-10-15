package insight

import (
	"context"
	"time"

	"looklook/app/aiagent/cmd/api/internal/dao"

	"github.com/zeromicro/go-zero/core/logx"
)

// Start 启动无人值守任务。
//
// 两个循环互相独立：
//   - 巡检：固定间隔跑一次（默认 30 分钟），发现异常即落库
//   - 日报：每 10 分钟检查一次，到了设定时间且当天还没有日报就生成
//
// 都是「先检查再执行」的幂等设计，服务重启不会产生重复日报/重复告警。
func (s *Service) Start(ctx context.Context) {
	if !s.cfg.Cron.Enabled {
		logx.Info("insight cron disabled (Cron.Enabled=false)")
		return
	}
	go s.patrolLoop(ctx)
	go s.reportLoop(ctx)
	logx.Infof("insight cron started: patrol every %dmin, report at %02d:00, serviceUid=%d",
		s.patrolInterval(), s.reportHour(), s.uid())
}

func (s *Service) patrolInterval() int {
	if s.cfg.Cron.PatrolIntervalMin <= 0 {
		return 30
	}
	return s.cfg.Cron.PatrolIntervalMin
}

func (s *Service) reportHour() int {
	h := s.cfg.Cron.ReportHour
	if h < 0 || h > 23 {
		return 9
	}
	return h
}

func (s *Service) patrolLoop(ctx context.Context) {
	if s.cfg.Cron.RunOnStart {
		// 稍等一下再跑：等服务注册完成、业务接口可用
		if !sleepCtx(ctx, 20*time.Second) {
			return
		}
		s.safePatrol(ctx, "cron:startup")
	}

	tick := time.NewTicker(time.Duration(s.patrolInterval()) * time.Minute)
	defer tick.Stop()
	for {
		select {
		case <-ctx.Done():
			return
		case <-tick.C:
			s.safePatrol(ctx, "cron")
		}
	}
}

func (s *Service) reportLoop(ctx context.Context) {
	tick := time.NewTicker(10 * time.Minute)
	defer tick.Stop()

	check := func() {
		now := time.Now()
		date := now.Format("2006-01-02")
		if s.LatestReportDate() == date {
			return // 今天已经生成过
		}
		if !s.cfg.Cron.RunOnStart && now.Hour() < s.reportHour() {
			return // 还没到日报时间
		}
		old, err := s.dao.GetReportByDate(ctx, dao.ReportDaily, date)
		if err != nil {
			logx.WithContext(ctx).Errorf("check report err: %v", err)
			return
		}
		if old != nil {
			s.mu.Lock()
			s.lastReportDate = date
			s.mu.Unlock()
			return
		}
		s.safeReport(ctx, "cron")
	}

	check()
	for {
		select {
		case <-ctx.Done():
			return
		case <-tick.C:
			check()
		}
	}
}

func (s *Service) safePatrol(ctx context.Context, trigger string) {
	defer func() {
		if rec := recover(); rec != nil {
			logx.WithContext(ctx).Errorf("patrol panic: %v", rec)
		}
	}()
	if _, err := s.RunPatrol(ctx, trigger); err != nil {
		logx.WithContext(ctx).Errorf("patrol failed: %v", err)
	}
}

func (s *Service) safeReport(ctx context.Context, trigger string) {
	defer func() {
		if rec := recover(); rec != nil {
			logx.WithContext(ctx).Errorf("report panic: %v", rec)
		}
	}()
	if _, _, err := s.GenerateReport(ctx, dao.ReportDaily, time.Now().Format("2006-01-02"), trigger, false); err != nil {
		logx.WithContext(ctx).Errorf("report failed: %v", err)
	}
}

func sleepCtx(ctx context.Context, d time.Duration) bool {
	t := time.NewTimer(d)
	defer t.Stop()
	select {
	case <-ctx.Done():
		return false
	case <-t.C:
		return true
	}
}
