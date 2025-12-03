package hypervisor

type QEMUHypervisor struct{}

func NewQEMUHypervisor() *QEMUHypervisor {
	return &QEMUHypervisor{}
}

func (q *QEMUHypervisor) CreateVM(cfg VMConfig) (*VMInfo, error) {
	// TODO: 调用 QEMU CLI 或 libvirt API
	return &VMInfo{
		ID:       "vm-123",
		Name:     cfg.Name,
		Status:   "stopped",
		CPU:      cfg.CPU,
		MemoryMB: cfg.MemoryMB,
		DiskGB:   cfg.DiskGB,
	}, nil
}

func (q *QEMUHypervisor) StartVM(id string) error                { return nil }
func (q *QEMUHypervisor) StopVM(id string) error                 { return nil }
func (q *QEMUHypervisor) DeleteVM(id string) error               { return nil }
func (q *QEMUHypervisor) ResizeVM(id string, cfg VMConfig) error { return nil }
