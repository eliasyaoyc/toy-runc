package subsystems

type FreezerSubsystem struct {
}

func (f *FreezerSubsystem) Name() string {
	//TODO implement me
	panic("implement me")
}

func (f *FreezerSubsystem) Set(cgroupPath string, res *ResourceConfig) error {
	//TODO implement me
	panic("implement me")
}

func (f *FreezerSubsystem) Apply(cgroupPath string, pid int) error {
	//TODO implement me
	panic("implement me")
}

func (f *FreezerSubsystem) Remove(cgroupPath string) error {
	//TODO implement me
	panic("implement me")
}
