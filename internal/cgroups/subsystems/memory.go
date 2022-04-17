package subsystems

import (
	"errors"
	"fmt"
	"io/ioutil"
	"path"
	"strconv"
	"syscall"
)

type MemorySubsystem struct {
}

func (m *MemorySubsystem) Name() string {
	return "memory"
}

func (m *MemorySubsystem) Set(cgroupPath string, res *ResourceConfig) error {
	if subsysCgroupPath, err := GetCgroupPath(m.Name(), cgroupPath, true); err == nil {
		if res.MemoryLimit != "" {
			if err := ioutil.WriteFile(path.Join(subsysCgroupPath, "memory.limit_in_bytes"), []byte(res.MemoryLimit), 0644); err != nil {
				return errors.New(fmt.Sprintf("set cgroup memory fail; %v", err))
			}
		}
		return nil
	} else {
		return err
	}
}

func (m *MemorySubsystem) Apply(cgroupPath string, pid int) error {
	if subsysCgroupPath, err := GetCgroupPath(m.Name(), cgroupPath, false); err == nil {
		if err := ioutil.WriteFile(path.Join(subsysCgroupPath, "tasks"), []byte(strconv.Itoa(pid)), 0644); err != nil {
			return errors.New(fmt.Sprintf("set cgroup proc fail; %v", err))
		}
		return nil
	} else {
		return errors.New(fmt.Sprintf("get cgroup %s error: %v", cgroupPath, err))
	}
}

func (m *MemorySubsystem) Remove(cgroupPath string) error {
	if subsysCgroupPath, err := GetCgroupPath(m.Name(), cgroupPath, false); err == nil {
		return syscall.Rmdir(subsysCgroupPath)
	} else {
		return err
	}
}
