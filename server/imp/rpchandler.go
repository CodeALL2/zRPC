package imp

import (
	"context"
	"fmt"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/types/known/anypb"
	"reflect"
	"zRPC/gzinx/ziface"
	"zRPC/gzinx/znet"
	"zRPC/protobufdemo/pp/pb"
	"zRPC/server/iface"
	"zRPC/server/util"
)

type TwoHandler struct {
	znet.BaseRouter
	Server iface.IServer
}

func (this *TwoHandler) PreHandler(request ziface.IRequest) {
	fmt.Println("调用之前的函数", request.GetConnection().GetRemoteAddr().String())
}

func (this *TwoHandler) Handler(request ziface.IRequest) {
	fmt.Println("方法2的handler执行", request.GetConnection().GetRemoteAddr().String())
	fmt.Println("数据体:", request.GetData())
	//err := request.GetConnection().SendMessage(request.GetMsgID(), []byte("hello this is zinx"))
	//if err != nil {
	//	fmt.Println(err)
	//}
	//开始拆
	//解析zRPC请求
	rpcRequest := &pb.RpcRequest{}
	if err := proto.Unmarshal(request.GetData(), rpcRequest); err != nil {
		fmt.Println("rpc请求解析失败", err)
		return
	}

	serviceInfo, err := this.Server.GetRegistry().GetService(rpcRequest.ServiceName)
	fmt.Println("请求方法的服务名", rpcRequest.ServiceName)
	if err != nil {
		fmt.Println(err)
		//给客户端写回数据
		return
	}
	fmt.Println("请求方法的方法名", rpcRequest.MethodName)

	method := serviceInfo.ServiceValue.MethodByName(rpcRequest.MethodName)

	if !method.IsValid() {
		fmt.Errorf("方法 %s 不存在", rpcRequest.MethodName)
		//给客户端返回错误信息
		return
	}
	//

	//获取方法的参数类型
	methodType := method.Type()
	numParams := methodType.NumIn()

	//if len(rpcRequest.GetParameters()) != numParams {
	//	fmt.Println("参数数量不匹配: 期望 %d, 实际 %d", numParams, len(rpcRequest.Parameters))
	//	//给客户端返回错误信息
	//	return
	//}
	//
	fmt.Println("方法的参数个数:", numParams)

	//准备方法的参数
	args := make([]reflect.Value, numParams)
	ctx := context.Background()
	args[0] = reflect.ValueOf(ctx)

	fmt.Println("请求的参数类型", rpcRequest.ParamName)
	paramValue, err := util.NewTypeRegistry().GetType(rpcRequest.ParamName)
	if err != nil {
		fmt.Println(err)
		return
	}

	// 检查 `paramValue` 是否为 nil
	if paramValue == nil {
		fmt.Println("GetType() 返回了 nil")
		return
	}

	// 检查是否为空
	if rpcRequest.Value == nil {
		fmt.Println("ValueList[i] 为空，无法解析参数")
		return
	}

	// ✅ 正确的 `UnmarshalTo()`
	if err := rpcRequest.Value.UnmarshalTo(paramValue); err != nil {
		fmt.Println("参数解析失败:", err)
		return
	}
	fmt.Println("参数解析成功", paramValue)
	args[1] = reflect.ValueOf(paramValue)
	//response := &pb.GetUserResponse{
	//	User:    nil,
	//	Success: false}
	//marshal, _ := proto.Marshal(response)
	//if err != nil {
	//	fmt.Println(err)
	//	request.GetConnection().SendMessage(1,[]byte(""))

	result := method.Call(args)

	if err, _ := result[1].Interface().(error); err != nil {
		fmt.Println(err)
		return
	}

	var resultTypeName string

	if result[0].Type().Kind() == reflect.Ptr {
		fmt.Printf("类型 %s (指针类型)\n", result[0].Type().Elem().Name())
		resultTypeName = result[0].Type().Elem().Name()
	} else {
		fmt.Printf("类型 %s (值类型)\n", result[0].Type().Name())
		resultTypeName = result[0].Type().Name()
	}

	fmt.Println("返回参数的类型", resultTypeName)
	anyValue, err := anypb.New(result[0].Interface().(proto.Message))
	if err != nil {
		fmt.Println("返回类型转成protobuf Message失败")
		return
	}

	response := &pb.RpcResponse{
		MsgId:           rpcRequest.MsgId,
		TypeName:        resultTypeName,
		ResponseValue:   anyValue,
		ResponseMessage: "成功调用",
		ResponseStatue:  200,
	}

	marshal, err := proto.Marshal(response)
	request.GetConnection().SendMessage(request.GetMsgID(), marshal)
}

func (this *TwoHandler) PostHandler(request ziface.IRequest) {
	fmt.Println("handler结束后的函数", request.GetConnection().GetRemoteAddr().String())
}
