package imp

import (
	"context"
	"encoding/json"
	"fmt"
	"google.golang.org/protobuf/proto"
	"reflect"
	"zRPC/gzinx/ziface"
	"zRPC/gzinx/znet"
	"zRPC/protobufdemo/pp/pb"

	"zRPC/server/iface"
	"zRPC/server/util"
)

type Server struct {
	registry iface.IRegistry //registry注册器
}

func NewServer() iface.IServer {
	return &Server{
		registry: nil,
	}
}

type TwoHandler struct {
	znet.BaseRouter
	Server *Server
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

	serviceInfo, err := this.Server.registry.GetService(rpcRequest.ServiceName)
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

	for i := 0; i < len(rpcRequest.ParamList); i++ {
		fmt.Println("进入", i)
		paramValue, err := util.NewTypeRegistry().GetType(rpcRequest.ParamList[i])
		if err != nil {
			fmt.Println(err)
			return
		}

		// 检查 `paramValue` 是否为 nil
		if paramValue == nil {
			fmt.Println("GetType() 返回了 nil")
			return
		}

		// 检查 `ValueList[i]` 是否为空
		if rpcRequest.ValueList[i] == nil {
			fmt.Println("ValueList[i] 为空，无法解析参数")
			return
		}

		// ✅ 正确的 `UnmarshalTo()`
		if err := rpcRequest.ValueList[i].UnmarshalTo(paramValue); err != nil {
			fmt.Println("参数解析失败:", err)
			return
		}

		fmt.Println("参数解析成功", paramValue)

		args[i+1] = reflect.ValueOf(paramValue)
		//response := &pb.GetUserResponse{
		//	User:    nil,
		//	Success: false}
		//marshal, _ := proto.Marshal(response)
		//if err != nil {
		//	fmt.Println(err)
		//	request.GetConnection().SendMessage(1,[]byte(""))
	}
	result := method.Call(args)

	if err, _ := result[1].Interface().(error); err != nil {
		fmt.Println(err)
		return
	}
	resPonse, ok := result[0].Interface().(*pb.RpcResponse)

	if !ok {
		fmt.Println("返回值类型出错")
		return
	}
	fmt.Println("返回状态", resPonse.ResponseStatue, "返回的类型", resPonse.TypeName)

	marshal, err := proto.Marshal(resPonse)
	request.GetConnection().SendMessage(request.GetMsgID(), marshal)
}

func (this *TwoHandler) PostHandler(request ziface.IRequest) {
	fmt.Println("handler结束后的函数", request.GetConnection().GetRemoteAddr().String())
}

// 反序列化参数
func (this *TwoHandler) deserializeParameter(data []byte, dest interface{}) error {
	// 这里使用 JSON 作为中间格式，您也可以使用其他序列化方式
	return json.Unmarshal(data, dest)
}

func (s *Server) Start() {

	zRPCServer := znet.NewServer("zRPC")
	//需要初始话handler
	handle := znet.NewMsgHandle()
	handle.AddMsgHandler(1, &TwoHandler{Server: s})
	zRPCServer.AddMsgHandler(handle)
	zRPCServer.Serve()
}

func (s *Server) SetRegistry(registry iface.IRegistry) {
	s.registry = registry
}
