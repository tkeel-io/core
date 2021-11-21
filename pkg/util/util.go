package util

import v1 "github.com/tkeel-io/tkeel-interface/openapi/v1"

func GetV1ResultOK() *v1.Result {
	return &v1.Result{
		Ret: v1.Retcode_OK,
		Msg: "ok",
	}
}

func GetV1ResultBadRequest(msg string) *v1.Result {
	return &v1.Result{
		Ret: v1.Retcode_BAD_REQEUST,
		Msg: msg,
	}
}

func GetV1ResultInternalError(msg string) *v1.Result {
	return &v1.Result{
		Ret: v1.Retcode_INTERNAL_ERROR,
		Msg: msg,
	}
}
