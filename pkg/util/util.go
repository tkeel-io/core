package util

import v1 "github.com/tkeel-io/tkeel-interface/openapi/v1"

func GetV1ResultOK() *v1.Result {
	return &v1.Result{
		Ret: v1.Retcode_ok,
		Msg: "ok",
	}
}

func GetV1ResultBadRequest(msg string) *v1.Result {
	return &v1.Result{
		Ret: v1.Retcode_badRequest,
		Msg: msg,
	}
}

func GetV1ResultInternalError(msg string) *v1.Result {
	return &v1.Result{
		Ret: v1.Retcode_internalError,
		Msg: msg,
	}
}
