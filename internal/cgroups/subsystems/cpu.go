package subsystems

import (
	"fmt"
	"io/ioutil"
	"path"
	"strconv"
	"syscall"
)

type CpuSubsystem struct {
}

func (c *CpuSubsystem) Name() string {
	return "cpu"
}

func (c *CpuSubsystem) Set(cgroupPath string, res *ResourceConfig) error {
	if subsysCgroupPath, err := GetCgroupPath(c.Name(), cgroupPath, true); err == nil {
		if res.CpuShare != "" {
			if err := ioutil.WriteFile(path.Join(subsysCgroupPath, "cpu.shares"), []byte(res.CpuShare), 0644); err != nil {
				return fmt.Errorf("set cgroup cpu share fail %v", err)
			}
		}
		return nil
	} else {
		return err
	}
}

func (c *CpuSubsystem) Apply(cgroupPath string, pid int) error {
	if subsysCgroupPath, err := GetCgroupPath(c.Name(), cgroupPath, false); err == nil {
		if err := ioutil.WriteFile(path.Join(subsysCgroupPath, "tasks"), []byte(strconv.Itoa(pid)), 0644); err != nil {
			return fmt.Errorf("set cgroup proc fail %v", err)
		}
		return nil
	} else {
		return fmt.Errorf("get cgroup %s error: %v", cgroupPath, err)
	}
}

func (c *CpuSubsystem) Remove(cgroupPath string) error {
	if subsysCgroupPath, err := GetCgroupPath(c.Name(), cgroupPath, false); err == nil {
		return syscall.Rmdir(subsysCgroupPath)
	} else {
		return err
	}
}
