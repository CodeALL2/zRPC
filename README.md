# zRPC
基于go-zinx轻量级tcp服务器开发的RPC

golang底层的netpoll的优化封装，使golang天生支持i/o多路复用，用户不用过多的关心连接池、多路复用管理
因此zRPC选择zinx轻量级tcp服务器进行开发，极大的简化了开发成本，在zinx服务器上我们采用reactor的网络io模型
能支持高并发连接，搭配handler处理池，业务响应速度更改

待开发...
