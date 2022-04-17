package subsystems

type NetPrioritySubsystem struct {
}

func (n *NetPrioritySubsystem) Name() string {
	//TODO implement me
	panic("implement me")
}

func (n *NetPrioritySubsystem) Set(cgroupPath string, res *ResourceConfig) error {
	//TODO implement me
	panic("implement me")
}

func (n *NetPrioritySubsystem) Apply(cgroupPath string, pid int) error {
	//TODO implement me
	panic("implement me")
}

func (n *NetPrioritySubsystem) Remove(cgroupPath string) error {
	//TODO implement me
	panic("implement me")
}
