package ziface

type IRouter interface {
	//处理业务之前的方法
	PreHandler(IRequest)
	//处理业务的方法
	Handler(IRequest)

	//处理业务之后的方法
	PostHandler(IRequest)
}
