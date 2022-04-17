package subsystems

type PidsSubsystem struct {
}

func (p *PidsSubsystem) Name() string {
	//TODO implement me
	panic("implement me")
}

func (p *PidsSubsystem) Set(cgroupPath string, res *ResourceConfig) error {
	//TODO implement me
	panic("implement me")
}

func (p *PidsSubsystem) Apply(cgroupPath string, pid int) error {
	//TODO implement me
	panic("implement me")
}

func (p *PidsSubsystem) Remove(cgroupPath string) error {
	//TODO implement me
	panic("implement me")
}
