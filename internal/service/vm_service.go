package service

import (
	"errors"
	"time"

	"StarstreamAstra/internal/hypervisor"
	"StarstreamAstra/internal/model"

	"gorm.io/gorm"
)

type VMService struct {
	db         *gorm.DB
	hypervisor hypervisor.Hypervisor
}

func NewVMService(db *gorm.DB, hv hypervisor.Hypervisor) *VMService {
	return &VMService{db: db, hypervisor: hv}
}

type VMCreateRequest struct {
	Name        string `json:"name" binding:"required"`
	CPU         int    `json:"cpu" binding:"required"`
	MemoryMB    int    `json:"memory_mb" binding:"required"`
	DiskGB      int    `json:"disk_gb" binding:"required"`
	Description string `json:"description"`
}

func (s *VMService) CreateVM(req VMCreateRequest) (*model.VM, error) {
	cfg := hypervisor.VMConfig{
		Name: req.Name, CPU: req.CPU, MemoryMB: req.MemoryMB, DiskGB: req.DiskGB,
	}
	info, err := s.hypervisor.CreateVM(cfg)
	if err != nil {
		return nil, err
	}

	vm := &model.VM{
		Name:         info.Name,
		CPU:          info.CPU,
		MemoryMB:     info.MemoryMB,
		DiskGB:       info.DiskGB,
		Status:       info.Status,
		Description:  req.Description,
		CreatedAt:    time.Now(),
		UpdatedAt:    time.Now(),
		HypervisorID: info.ID,
	}
	if err := s.db.Create(vm).Error; err != nil {
		return nil, err
	}
	return vm, nil
}

func (s *VMService) ListVMs() ([]*model.VM, error) {
	var vms []*model.VM
	if err := s.db.Find(&vms).Error; err != nil {
		return nil, err
	}
	return vms, nil
}

func (s *VMService) GetVMByID(id uint) (*model.VM, error) {
	var vm model.VM
	if err := s.db.First(&vm, id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return &vm, nil
}

func (s *VMService) UpdateVMStatus(id uint, status string) error {
	return s.db.Model(&model.VM{}).Where("id = ?", id).Update("status", status).Error
}

func (s *VMService) StartVM(id uint) error {
	vm, err := s.GetVMByID(id)
	if err != nil || vm == nil {
		return err
	}
	if err := s.hypervisor.StartVM(vm.HypervisorID); err != nil {
		return err
	}
	return s.UpdateVMStatus(id, "running")
}

func (s *VMService) StopVM(id uint) error {
	vm, err := s.GetVMByID(id)
	if err != nil || vm == nil {
		return err
	}
	if err := s.hypervisor.StopVM(vm.HypervisorID); err != nil {
		return err
	}
	return s.UpdateVMStatus(id, "stopped")
}

func (s *VMService) DeleteVM(id uint) error {
	vm, err := s.GetVMByID(id)
	if err != nil || vm == nil {
		return err
	}
	if err := s.hypervisor.DeleteVM(vm.HypervisorID); err != nil {
		return err
	}
	return s.db.Delete(&model.VM{}, id).Error
}

func (s *VMService) ResizeVM(id uint, cfg hypervisor.VMConfig) error {
	vm, err := s.GetVMByID(id)
	if err != nil || vm == nil {
		return err
	}
	if err := s.hypervisor.ResizeVM(vm.HypervisorID, cfg); err != nil {
		return err
	}
	vm.CPU = cfg.CPU
	vm.MemoryMB = cfg.MemoryMB
	vm.DiskGB = cfg.DiskGB
	vm.UpdatedAt = time.Now()
	return s.db.Save(vm).Error
}
