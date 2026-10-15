package homestay

import (
	"context"

	"looklook/app/admin/cmd/api/internal/svc"
	"looklook/app/admin/cmd/api/internal/types"

	"github.com/zeromicro/go-zero/core/logx"
)

type HomestayCityOptionsLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

// 城市下拉选项
func NewHomestayCityOptionsLogic(ctx context.Context, svcCtx *svc.ServiceContext) *HomestayCityOptionsLogic {
	return &HomestayCityOptionsLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *HomestayCityOptionsLogic) HomestayCityOptions(req *types.HomestayCityReq) (resp *types.HomestayCityResp, err error) {
	rows, err := l.svcCtx.Dao.CityOptions(l.ctx)
	if err != nil {
		return nil, err
	}
	list := make([]types.CityOption, 0, len(rows))
	for _, r := range rows {
		list = append(list, types.CityOption{City: r.City, Count: r.Count})
	}
	return &types.HomestayCityResp{List: list}, nil
}
