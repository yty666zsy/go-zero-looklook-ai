package xerr

import (
	"fmt"
	"strings"

	"github.com/pkg/errors"
	"google.golang.org/grpc/status"
)

/**
常用通用固定错误
*/

type CodeError struct {
	errCode uint32
	errMsg  string
}

//返回给前端的错误码
func (e *CodeError) GetErrCode() uint32 {
	return e.errCode
}

//返回给前端显示端错误信息
func (e *CodeError) GetErrMsg() string {
	return e.errMsg
}

func (e *CodeError) Error() string {
	return fmt.Sprintf("ErrCode:%d，ErrMsg:%s", e.errCode, e.errMsg)
}

func NewErrCodeMsg(errCode uint32, errMsg string) *CodeError {
	return &CodeError{errCode: errCode, errMsg: errMsg}
}
func NewErrCode(errCode uint32) *CodeError {
	return &CodeError{errCode: errCode, errMsg: MapErrMsg(errCode)}
}

func NewErrMsg(errMsg string) *CodeError {
	return &CodeError{errCode: SERVER_COMMON_ERROR, errMsg: errMsg}
}

/**
业务错误（可以原样展示给用户的提示）
rpc 侧想给前端一句人话时用 NewErrBiz 构造，api 层用 BizMessage 取出来再返回。
约定：业务错误不要再 errors.Wrap 追加技术细节（细节写日志），否则文案会带上一串内部信息。
没带标记的错误（db / panic / 超时等）不会被透出，避免暴露底层实现。
*/

const BizMsgPrefix = "[biz]"

// NewErrBiz 业务错误：文案可以直接给用户看。
func NewErrBiz(errMsg string) *CodeError {
	return &CodeError{errCode: SERVER_COMMON_ERROR, errMsg: BizMsgPrefix + errMsg}
}

// BizMessage 从（可能是 grpc 包装过的）错误里取出业务提示。
func BizMessage(err error) (string, bool) {
	if err == nil {
		return "", false
	}
	if st, _ := status.FromError(errors.Cause(err)); st != nil {
		if msg, ok := trimBizPrefix(st.Message()); ok {
			return msg, true
		}
	}
	return trimBizPrefix(err.Error())
}

func trimBizPrefix(s string) (string, bool) {
	i := strings.Index(s, BizMsgPrefix)
	if i < 0 {
		return "", false
	}
	return strings.TrimSpace(s[i+len(BizMsgPrefix):]), true
}
