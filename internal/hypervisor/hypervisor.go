package hypervisor

import "errors"

type VMConfig struct {
	Name     string
	CPU      int
	MemoryMB int
	DiskGB   int
}

type VMInfo struct {
	ID       string
	Name     string
	Status   string
	CPU      int
	MemoryMB int
	DiskGB   int
}

type Hypervisor interface {
	CreateVM(cfg VMConfig) (*VMInfo, error)
	StartVM(id string) error
	StopVM(id string) error
	DeleteVM(id string) error
	ResizeVM(id string, cfg VMConfig) error
}

var ErrVMNotFound = errors.New("VM not found")
