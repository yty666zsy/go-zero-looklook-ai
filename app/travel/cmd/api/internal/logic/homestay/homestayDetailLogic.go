package homestay

import (
	"context"
	"math"

	"looklook/app/travel/cmd/api/internal/conv"
	"looklook/app/travel/cmd/api/internal/svc"
	"looklook/app/travel/cmd/api/internal/types"
	"looklook/app/travel/model"
	"looklook/pkg/xerr"

	"github.com/pkg/errors"
	"github.com/zeromicro/go-zero/core/logx"
)

type HomestayDetailLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewHomestayDetailLogic(ctx context.Context, svcCtx *svc.ServiceContext) HomestayDetailLogic {
	return HomestayDetailLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

type scoreStatRow struct {
	Clean    float64 `db:"clean"`
	Location float64 `db:"location"`
	Service  float64 `db:"service"`
	Cnt      int64   `db:"cnt"`
}

// HomestayDetail 民宿详情聚合：民宿本体 + 轮播图 + 店铺信息 + 评分维度
// 这里直接读本地 model：详情属于"只被自己用"的聚合查询，走 rpc 反而多一次序列化
func (l *HomestayDetailLogic) HomestayDetail(req types.HomestayDetailReq) (*types.HomestayDetailResp, error) {

	homestay, err := l.svcCtx.HomestayModel.FindOne(l.ctx, req.Id)
	if err != nil && err != model.ErrNotFound {
		return nil, errors.Wrapf(xerr.NewErrCode(xerr.DB_ERROR), "get homestay detail db err , id : %d , err : %v", req.Id, err)
	}
	if homestay == nil {
		return nil, errors.Wrapf(xerr.NewErrMsg("民宿不存在"), "homestay not found id : %d", req.Id)
	}

	typeHomestay := conv.HomestayToType(homestay)

	// 店铺信息
	var business types.HomestayBusiness
	if biz, err := l.svcCtx.HomestayBusinessModel.FindOne(l.ctx, homestay.HomestayBusinessId); err != nil {
		if err != model.ErrNotFound {
			logx.WithContext(l.ctx).Errorf("get business fail id:%d err:%v", homestay.HomestayBusinessId, err)
		}
	} else if biz != nil {
		business = types.HomestayBusiness{
			Id:        biz.Id,
			Title:     biz.Title,
			Info:      biz.Info,
			Tags:      biz.Tags,
			Cover:     biz.Cover,
			Star:      biz.Star,
			HeaderImg: biz.HeaderImg,
		}
	}

	// 评分维度：star 字段是 json 数组 [卫生,位置,服务]
	var stat scoreStatRow
	statQuery := "select ifnull(avg(cast(json_extract(`star`,'$[0]') as decimal(3,1))),0) as clean," +
		" ifnull(avg(cast(json_extract(`star`,'$[1]') as decimal(3,1))),0) as location," +
		" ifnull(avg(cast(json_extract(`star`,'$[2]') as decimal(3,1))),0) as service," +
		" count(*) as cnt from `homestay_comment` where del_state = 0 and homestay_id = ?"
	if err := l.svcCtx.SqlConn.QueryRowCtx(l.ctx, &stat, statQuery, req.Id); err != nil {
		logx.WithContext(l.ctx).Errorf("get score stat fail id:%d err:%v", req.Id, err)
	}

	scoreStat := types.ScoreStat{
		Clean:    round1(stat.Clean),
		Location: round1(stat.Location),
		Service:  round1(stat.Service),
		Count:    stat.Cnt,
	}
	if scoreStat.Count > 0 {
		scoreStat.Score = round1((stat.Clean + stat.Location + stat.Service) / 3)
	} else {
		// 没有评价时退化为民宿自带评分，避免前端出现 0 分
		scoreStat.Score = homestay.Score
	}

	return &types.HomestayDetailResp{
		Homestay:  typeHomestay,
		Images:    typeHomestay.Images,
		Business:  business,
		ScoreStat: scoreStat,
	}, nil
}

func round1(v float64) float64 {
	return math.Round(v*10) / 10
}
