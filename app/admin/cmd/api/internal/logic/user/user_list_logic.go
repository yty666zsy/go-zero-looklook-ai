package user

import (
	"context"

	"looklook/app/admin/cmd/api/internal/conv"
	"looklook/app/admin/cmd/api/internal/svc"
	"looklook/app/admin/cmd/api/internal/types"

	"github.com/zeromicro/go-zero/core/logx"
)

type UserListLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

// 用户列表
func NewUserListLogic(ctx context.Context, svcCtx *svc.ServiceContext) *UserListLogic {
	return &UserListLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *UserListLogic) UserList(req *types.UserListReq) (resp *types.UserListResp, err error) {
	page, pageSize := conv.NormalizePage(req.Page, req.PageSize)

	rows, total, err := l.svcCtx.Dao.ListUsers(l.ctx, page, pageSize, req.Keyword)
	if err != nil {
		return nil, err
	}

	list := make([]types.UserItem, 0, len(rows))
	for _, r := range rows {
		list = append(list, conv.UserItem(r))
	}
	return &types.UserListResp{
		List:     list,
		Page:     page,
		PageSize: pageSize,
		Total:    total,
	}, nil
}
