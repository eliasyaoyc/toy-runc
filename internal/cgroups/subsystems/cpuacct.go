package subsystems

type CpuacctSubsystem struct {
}

func (c *CpuacctSubsystem) Name() string {
	//TODO implement me
	panic("implement me")
}

func (c *CpuacctSubsystem) Set(cgroupPath string, res *ResourceConfig) error {
	//TODO implement me
	panic("implement me")
}

func (c *CpuacctSubsystem) Apply(cgroupPath string, pid int) error {
	//TODO implement me
	panic("implement me")
}

func (c *CpuacctSubsystem) Remove(cgroupPath string) error {
	//TODO implement me
	panic("implement me")
}
