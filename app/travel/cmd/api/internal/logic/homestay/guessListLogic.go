package homestay

import (
	"context"

	"looklook/app/travel/cmd/api/internal/conv"
	"looklook/app/travel/cmd/api/internal/svc"
	"looklook/app/travel/cmd/api/internal/types"
	"looklook/pkg/xerr"

	"github.com/pkg/errors"
	"github.com/zeromicro/go-zero/core/logx"
)

type GuessListLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewGuessListLogic(ctx context.Context, svcCtx *svc.ServiceContext) GuessListLogic {
	return GuessListLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *GuessListLogic) GuessList(req types.GuessListReq) (*types.GuessListResp, error) {
	var resp []types.Homestay

	list, err := l.svcCtx.HomestayModel.FindPageListByIdDESC(l.ctx, l.svcCtx.HomestayModel.SelectBuilder(), 0, 6)
	if err != nil {
		return nil, errors.Wrapf(xerr.NewErrCode(xerr.DB_ERROR), "GuessList db err req : %+v , err : %v", req, err)
	}

	resp = conv.HomestayListToType(list)

	return &types.GuessListResp{
		List: resp,
	}, nil
}
