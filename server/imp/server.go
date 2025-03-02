package imp

import (
	"context"
	"fmt"
	"github.com/golang/protobuf/proto"
	"reflect"
	"zRPC/gzinx/ziface"
	"zRPC/gzinx/znet"
	"zRPC/protobufdemo/pp/pb"
	"zRPC/server/iface"
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
	userRequest := &pb.GetUserRequest{}
	err := proto.Unmarshal(request.GetData(), userRequest)
	if err != nil {
		fmt.Println(err)
		return
	}

	fmt.Println("zRPC已经收到消息了,获取用户id为:", userRequest.UserId)
	service, _ := this.Server.registry.GetService("IUserService")
	method := service.ServiceValue.MethodByName("GetUser")

	if !method.IsValid() {
		fmt.Println("方法不存在")
		return
	}
	ctx := context.Background()

	args := []reflect.Value{
		reflect.ValueOf(ctx),
		reflect.ValueOf(userRequest),
	}
	result := method.Call(args)

	fmt.Println("调用返回的结果为:", result)

	//response := &pb.GetUserResponse{
	//	User:    nil,
	//	Success: false}
	//marshal, _ := proto.Marshal(response)
	//if err != nil {
	//	fmt.Println(err)
	//	request.GetConnection().SendMessage(1,[]byte(""))
}

func (this *TwoHandler) PostHandler(request ziface.IRequest) {
	fmt.Println("handler结束后的函数", request.GetConnection().GetRemoteAddr().String())
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
