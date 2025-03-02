package iface

type IServer interface {
	Start()
	SetRegistry(registry IRegistry)
}
