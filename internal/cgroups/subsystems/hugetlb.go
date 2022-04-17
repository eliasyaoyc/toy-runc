package subsystems

type HugetlbSubsystem struct {
}

func (h *HugetlbSubsystem) Name() string {
	//TODO implement me
	panic("implement me")
}

func (h *HugetlbSubsystem) Set(cgroupPath string, res *ResourceConfig) error {
	//TODO implement me
	panic("implement me")
}

func (h *HugetlbSubsystem) Apply(cgroupPath string, pid int) error {
	//TODO implement me
	panic("implement me")
}

func (h *HugetlbSubsystem) Remove(cgroupPath string) error {
	//TODO implement me
	panic("implement me")
}
