package logic

import (
	"context"
	"encoding/json"
	"github.com/hibiken/asynq"
	"looklook/app/mqueue/cmd/job/jobtype"
	"strings"
	"time"

	"looklook/app/order/cmd/rpc/internal/svc"
	"looklook/app/order/cmd/rpc/pb"
	"looklook/app/order/model"
	"looklook/app/travel/cmd/rpc/travel"
	"looklook/pkg/booking"
	"looklook/pkg/tool"
	"looklook/pkg/uniqueid"
	"looklook/pkg/xerr"

	"github.com/pkg/errors"
	"github.com/zeromicro/go-zero/core/logx"
	"github.com/zeromicro/go-zero/core/stores/sqlx"
)

const (
	CloseOrderTimeMinutes = 30 //defer close order time

	maxBookingNights = 365 //单次预订最多占用的晚数，防止刷日历
	lockTTLSeconds   = 8   //分布式锁自动过期时间（进程崩了也不会锁死房源）
	lockWaitMillis   = 1500 //抢锁自旋等待，超过就提示「预订火爆」
)

type CreateHomestayOrderLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewCreateHomestayOrderLogic(ctx context.Context, svcCtx *svc.ServiceContext) *CreateHomestayOrderLogic {
	return &CreateHomestayOrderLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

// CreateHomestayOrder.
func (l *CreateHomestayOrderLogic) CreateHomestayOrder(in *pb.CreateHomestayOrderReq) (*pb.CreateHomestayOrderResp, error) {

	//1、Create Order
	if in.LiveEndTime <= in.LiveStartTime {
		return nil, errors.Wrapf(xerr.NewErrMsg("Stay at least one night"), "Place an order at a B&B. The end time of your stay must be greater than the start time. in : %+v", in)
	}

	resp, err := l.svcCtx.TravelRpc.HomestayDetail(l.ctx, &travel.HomestayDetailReq{
		Id: in.HomestayId,
	})
	if err != nil {
		return nil, errors.Wrapf(xerr.NewErrMsg("Failed to query the record"), "Failed to query the record  rpc HomestayDetail fail , homestayId : %d , err : %v", in.HomestayId, err)
	}
	if resp.Homestay == nil {
		return nil, errors.Wrapf(xerr.NewErrMsg("This record does not exist"), "This record does not exist , homestayId : %d ", in.HomestayId)
	}

	var cover string //Get the cover...
	if len(resp.Homestay.Banner) > 0 {
		cover = strings.Split(resp.Homestay.Banner, ",")[0]
	}

	order := new(model.HomestayOrder)
	order.Sn = uniqueid.GenSn(uniqueid.SN_PREFIX_HOMESTAY_ORDER)
	order.UserId = in.UserId
	order.HomestayId = in.HomestayId
	order.Title = resp.Homestay.Title
	order.SubTitle = resp.Homestay.SubTitle
	order.Cover = cover
	order.Info = resp.Homestay.Info
	order.PeopleNum = resp.Homestay.PeopleNum
	order.RowType = resp.Homestay.RowType
	order.HomestayPrice = resp.Homestay.HomestayPrice
	order.MarketHomestayPrice = resp.Homestay.MarketHomestayPrice
	order.HomestayBusinessId = resp.Homestay.HomestayBusinessId
	order.HomestayUserId = resp.Homestay.UserId
	order.LivePeopleNum = in.LivePeopleNum
	order.TradeState = model.HomestayOrderTradeStateWaitPay
	order.TradeCode = tool.Krand(8, tool.KC_RAND_KIND_ALL)
	order.Remark = in.Remark
	order.FoodInfo = resp.Homestay.FoodInfo
	order.FoodPrice = resp.Homestay.FoodPrice
	order.LiveStartDate = time.Unix(in.LiveStartTime, 0)
	order.LiveEndDate = time.Unix(in.LiveEndTime, 0)

	liveDays := int64(order.LiveEndDate.Sub(order.LiveStartDate).Seconds() / 86400) //Stayed a few days in total

	order.HomestayTotalPrice = int64(resp.Homestay.HomestayPrice * liveDays) //Calculate the total price of the B&B
	if in.IsFood {
		order.NeedFood = model.HomestayOrderNeedFoodYes
		//Calculate the total price of the meal.
		order.FoodTotalPrice = int64(resp.Homestay.FoodPrice * in.LivePeopleNum * liveDays)
	}

	order.OrderTotalPrice = order.HomestayTotalPrice + order.FoodTotalPrice //Calculate total order price.

	//1.1、高并发防超卖：房源粒度分布式锁 + 逐日占用日历（唯一键兜底）。
	//     锁把「查占用 → 占位 → 落订单」串起来，唯一键保证即使锁失效也不会卖重。
	dates := booking.Dates(order.LiveStartDate, order.LiveEndDate)
	if len(dates) == 0 {
		return nil, errors.Wrapf(xerr.NewErrMsg("Stay at least one night"), "no occupiable date, in : %+v", in)
	}
	if len(dates) > maxBookingNights {
		return nil, errors.Wrapf(xerr.NewErrMsg("One booking supports at most 365 nights"), "too many nights : %d", len(dates))
	}

	lockKey := booking.LockKey(in.HomestayId)
	token, err := l.svcCtx.Lock.Acquire(l.ctx, lockKey, lockTTLSeconds, lockWaitMillis)
	if err != nil {
		if errors.Is(err, booking.ErrLockBusy) {
			l.Errorf("create order reject, booking lock busy, homestayId : %d", in.HomestayId)
			return nil, xerr.NewErrBiz("当前预订火爆，请稍后重试")
		}
		return nil, errors.Wrapf(xerr.NewErrCode(xerr.DB_ERROR), "acquire booking lock fail homestayId : %d , err : %v", in.HomestayId, err)
	}
	defer func() {
		// 用独立 ctx 释放，避免请求被取消后锁要等 TTL 才能放开
		releaseCtx, cancel := context.WithTimeout(context.Background(), time.Second)
		defer cancel()
		if err := l.svcCtx.Lock.Release(releaseCtx, lockKey, token); err != nil {
			logx.WithContext(l.ctx).Errorf("release booking lock fail, key : %s , err : %v", lockKey, err)
		}
	}()

	// 锁内先查一遍冲突日期，给出「哪几晚被订走了」的友好提示（不是最终保证，最终靠唯一键）
	if taken, err := l.svcCtx.Calendar.TakenDates(l.ctx, nil, in.HomestayId, dates); err == nil && len(taken) > 0 {
		nights := make([]string, 0, len(taken))
		for _, d := range taken {
			nights = append(nights, d.Format("01-02"))
		}
		l.Errorf("create order reject, dates taken, homestayId : %d , nights : %v", in.HomestayId, nights)
		return nil, xerr.NewErrBiz("所选日期已被预订（" + strings.Join(nights, "、") + "），请更换日期")
	}

	// 订单和日历占用写在同一个事务里：要么都成功，要么都回滚（不会出现「占了日历没订单」）
	txErr := l.svcCtx.DB.TransactCtx(l.ctx, func(ctx context.Context, session sqlx.Session) error {
		if err := l.svcCtx.Calendar.Occupy(ctx, session, in.HomestayId, dates, order.Sn); err != nil {
			return err
		}
		_, err := l.svcCtx.HomestayOrderModel.Insert(ctx, session, order)
		return err
	})
	if txErr != nil {
		if errors.Is(txErr, booking.ErrDateTaken) {
			l.Errorf("occupy calendar conflict, homestayId : %d , dates : %v , err : %v", in.HomestayId, dates, txErr)
			return nil, xerr.NewErrBiz("所选日期已被预订，请更换日期")
		}
		return nil, errors.Wrapf(xerr.NewErrCode(xerr.DB_ERROR), "Order Database Exception order : %+v , err: %v", order, txErr)
	}

	//2、Delayed closing of order tasks.
	payload, err := json.Marshal(jobtype.DeferCloseHomestayOrderPayload{Sn: order.Sn})
	if err != nil {
		logx.WithContext(l.ctx).Errorf("create defer close order task json Marshal fail err :%+v , sn : %s", err, order.Sn)
	} else {
		_, err = l.svcCtx.AsynqClient.Enqueue(asynq.NewTask(jobtype.DeferCloseHomestayOrder, payload), asynq.ProcessIn(CloseOrderTimeMinutes*time.Minute))
		if err != nil {
			logx.WithContext(l.ctx).Errorf("create defer close order task insert queue fail err :%+v , sn : %s", err, order.Sn)
		}
	}

	return &pb.CreateHomestayOrderResp{
		Sn: order.Sn,
	}, nil
}
