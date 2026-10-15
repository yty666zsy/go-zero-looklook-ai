package homestay

import (
	"context"
	"strings"

	"looklook/app/travel/cmd/api/internal/conv"
	"looklook/app/travel/cmd/api/internal/svc"
	"looklook/app/travel/cmd/api/internal/types"
	"looklook/app/travel/model"
	"looklook/pkg/xerr"

	"github.com/Masterminds/squirrel"
	"github.com/pkg/errors"
	"github.com/zeromicro/go-zero/core/logx"
)

type HomeLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewHomeLogic(ctx context.Context, svcCtx *svc.ServiceContext) HomeLogic {
	return HomeLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

// cityItem 城市元数据，城市表不存在，这里用静态配置 + 实时房源数拼装
type cityItem struct {
	Code  string
	Name  string
	Cover string
}

var cityMaster = []cityItem{
	{"beijing", "北京", "https://picsum.photos/seed/city-beijing/600/400"},
	{"shanghai", "上海", "https://picsum.photos/seed/city-shanghai/600/400"},
	{"hangzhou", "杭州", "https://picsum.photos/seed/city-hangzhou/600/400"},
	{"chengdu", "成都", "https://picsum.photos/seed/city-chengdu/600/400"},
	{"xiamen", "厦门", "https://picsum.photos/seed/city-xiamen/600/400"},
	{"lijiang", "丽江", "https://picsum.photos/seed/city-lijiang/600/400"},
	{"sanya", "三亚", "https://picsum.photos/seed/city-sanya/600/400"},
	{"dali", "大理", "https://picsum.photos/seed/city-dali/600/400"},
}

// hotTagMaster 首页金刚区标签
var hotTagMaster = []string{"海景房", "独栋小院", "私汤温泉", "亲子友好", "设计感", "近地铁"}

type cityCountRow struct {
	City string `db:"city"`
	Cnt  int64  `db:"cnt"`
}

// Home 首页聚合接口：一次返回城市、轮播、优选、猜你喜欢，减少首屏请求
func (l *HomeLogic) Home(req types.HomeReq) (*types.HomeResp, error) {

	cityCountMap := make(map[string]int64, len(cityMaster))
	var rows []cityCountRow
	query := "select city, count(*) as cnt from `homestay` where del_state = 0 and row_state = 1 group by city"
	if err := l.svcCtx.SqlConn.QueryRowsCtx(l.ctx, &rows, query); err != nil {
		return nil, errors.Wrapf(xerr.NewErrCode(xerr.DB_ERROR), "home city count db err : %v", err)
	}
	for _, row := range rows {
		cityCountMap[row.City] = row.Cnt
	}

	cities := make([]types.City, 0, len(cityMaster))
	for _, c := range cityMaster {
		cities = append(cities, types.City{
			Code:  c.Code,
			Name:  c.Name,
			Cover: c.Cover,
			Count: cityCountMap[c.Name],
		})
	}

	// 轮播：取最新 5 间房的封面
	latest, err := l.svcCtx.HomestayModel.FindPageListByIdDESC(l.ctx,
		l.svcCtx.HomestayModel.SelectBuilder().Where(squirrel.Eq{"row_state": 1}), 0, 5)
	if err != nil {
		return nil, errors.Wrapf(xerr.NewErrCode(xerr.DB_ERROR), "home banner db err : %v", err)
	}
	banners := make([]types.Banner, 0, len(latest))
	for i, item := range latest {
		images := conv.SplitList(item.Banner)
		if len(images) == 0 {
			continue
		}
		banners = append(banners, types.Banner{
			Id:    item.Id,
			Cover: images[0],
			Title: item.Title,
			Link:  "",
		})
		_ = i
	}

	// 优选民宿：走活动表
	activityBuilder := l.svcCtx.HomestayActivityModel.SelectBuilder().Where(squirrel.Eq{
		"row_type":   model.HomestayActivityPreferredType,
		"row_status": model.HomestayActivityUpStatus,
	})
	activities, err := l.svcCtx.HomestayActivityModel.FindPageListByPage(l.ctx, activityBuilder, 1, 8, "id desc")
	if err != nil {
		return nil, errors.Wrapf(xerr.NewErrCode(xerr.DB_ERROR), "home activity db err : %v", err)
	}

	recommend := make([]types.Homestay, 0, len(activities))
	for _, a := range activities {
		h, err := l.svcCtx.HomestayModel.FindOne(l.ctx, a.DataId)
		if err != nil && err != model.ErrNotFound {
			logx.WithContext(l.ctx).Errorf("home recommend fail id:%d err:%v", a.DataId, err)
			continue
		}
		if h == nil || h.RowState != 1 {
			continue
		}
		recommend = append(recommend, conv.HomestayToType(h))
	}

	// 猜你喜欢：按城市筛选优先，否则取最新
	guessBuilder := l.svcCtx.HomestayModel.SelectBuilder().Where(squirrel.Eq{"row_state": 1})
	if strings.TrimSpace(req.City) != "" {
		guessBuilder = guessBuilder.Where(squirrel.Eq{"city": strings.TrimSpace(req.City)})
	}
	guessList, err := l.svcCtx.HomestayModel.FindPageListByIdDESC(l.ctx, guessBuilder, 0, 6)
	if err != nil {
		return nil, errors.Wrapf(xerr.NewErrCode(xerr.DB_ERROR), "home guess db err : %v", err)
	}

	recommendCity := strings.TrimSpace(req.City)
	if recommendCity == "" {
		recommendCity = "上海"
	}

	return &types.HomeResp{
		Cities:        cities,
		Banners:       banners,
		HotTags:       hotTagMaster,
		Recommend:     recommend,
		Guess:         conv.HomestayListToType(guessList),
		RecommendCity: recommendCity,
	}, nil
}
