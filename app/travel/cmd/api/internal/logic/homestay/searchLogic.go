package homestay

import (
	"context"
	"strings"

	"looklook/app/travel/cmd/api/internal/conv"
	"looklook/app/travel/cmd/api/internal/svc"
	"looklook/app/travel/cmd/api/internal/types"
	"looklook/pkg/tool"
	"looklook/pkg/xerr"

	"github.com/Masterminds/squirrel"
	"github.com/pkg/errors"
	"github.com/zeromicro/go-zero/core/logx"
)

type SearchLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewSearchLogic(ctx context.Context, svcCtx *svc.ServiceContext) SearchLogic {
	return SearchLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

// Search 民宿搜索/筛选，支持关键词、城市、价格区间、人数与排序
func (l *SearchLogic) Search(req types.SearchReq) (*types.SearchResp, error) {

	page := req.Page
	if page < 1 {
		page = 1
	}
	pageSize := req.PageSize
	if pageSize <= 0 {
		pageSize = 10
	}
	if pageSize > 50 {
		pageSize = 50
	}

	builder := l.svcCtx.HomestayModel.SelectBuilder().Where(squirrel.Eq{"row_state": 1})

	if city := strings.TrimSpace(req.City); city != "" {
		builder = builder.Where(squirrel.Eq{"city": city})
	}

	if keyword := strings.TrimSpace(req.Keyword); keyword != "" {
		like := "%" + keyword + "%"
		builder = builder.Where(squirrel.Or{
			squirrel.Like{"title": like},
			squirrel.Like{"sub_title": like},
			squirrel.Like{"address": like},
			squirrel.Like{"tags": like},
		})
	}

	if req.MinPrice > 0 {
		builder = builder.Where(squirrel.GtOrEq{"homestay_price": tool.Yuan2Fen(req.MinPrice)})
	}
	if req.MaxPrice > 0 {
		builder = builder.Where(squirrel.LtOrEq{"homestay_price": tool.Yuan2Fen(req.MaxPrice)})
	}
	if req.PeopleNum > 0 {
		builder = builder.Where(squirrel.GtOrEq{"people_num": req.PeopleNum})
	}

	orderBy := "id desc"
	switch req.SortBy {
	case "priceAsc":
		orderBy = "homestay_price asc"
	case "priceDesc":
		orderBy = "homestay_price desc"
	case "score":
		orderBy = "score desc, comment_count desc"
	case "new":
		orderBy = "id desc"
	case "default", "":
		orderBy = "id desc"
	}

	list, total, err := l.svcCtx.HomestayModel.FindPageListByPageWithTotal(l.ctx, builder, page, pageSize, orderBy)
	if err != nil {
		return nil, errors.Wrapf(xerr.NewErrCode(xerr.DB_ERROR), "search homestay db err req : %+v , err : %v", req, err)
	}

	return &types.SearchResp{
		List:  conv.HomestayListToType(list),
		Total: total,
	}, nil
}
