package util

import (
	"fmt"
	"google.golang.org/protobuf/proto"
	"reflect"
	"sync"
	"zRPC/protobufdemo/pp/pb"
)

// TypeRegistry 类型注册表
type TypeRegistry struct {
	types sync.Map // 存储类型名称到类型的映射
}

// NewTypeRegistry 创建新的类型注册表
func NewTypeRegistry() *TypeRegistry {
	registry := &TypeRegistry{}

	// 注册基础类型
	registry.types.Store("int32", &pb.Int32Value{})
	registry.types.Store("int64", &pb.Int64Value{})
	registry.types.Store("string", &pb.StringValue{})
	registry.types.Store("bool", &pb.BoolValue{})
	registry.types.Store("float", &pb.FloatValue{})
	registry.types.Store("double", &pb.DoubleValue{})
	registry.types.Store("bytes", &pb.BytesValue{})
	registry.types.Store("void", &pb.VoidValue{})

	// 注册业务类型
	registry.types.Store("User", &pb.User{})
	registry.types.Store("Order", &pb.Order{})
	registry.types.Store("RpcUser", &pb.RpcUser{})

	return registry
}

// GetType 获取类型信息，并返回新的实例
func (r *TypeRegistry) GetType(typeName string) (proto.Message, error) {
	value, ok := r.types.Load(typeName)
	if !ok {
		return nil, fmt.Errorf("类型未注册: %s", typeName)
	}

	// 断言类型
	protoType, ok := value.(proto.Message)
	if !ok {
		return nil, fmt.Errorf("类型 %s 不是 proto.Message", typeName)
	}

	// 通过反射创建新的实例
	newInstance := reflect.New(reflect.TypeOf(protoType).Elem()).Interface()

	// 断言回 proto.Message
	message, ok := newInstance.(proto.Message)
	if !ok {
		return nil, fmt.Errorf("无法转换 %s 为 proto.Message", typeName)
	}

	return message, nil
}
