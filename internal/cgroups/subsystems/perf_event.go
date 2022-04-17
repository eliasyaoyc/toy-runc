package subsystems

type PerfEventSubsystem struct {
}

func (p *PerfEventSubsystem) Name() string {
	//TODO implement me
	panic("implement me")
}

func (p *PerfEventSubsystem) Set(cgroupPath string, res *ResourceConfig) error {
	//TODO implement me
	panic("implement me")
}

func (p *PerfEventSubsystem) Apply(cgroupPath string, pid int) error {
	//TODO implement me
	panic("implement me")
}

func (p *PerfEventSubsystem) Remove(cgroupPath string) error {
	//TODO implement me
	panic("implement me")
}
