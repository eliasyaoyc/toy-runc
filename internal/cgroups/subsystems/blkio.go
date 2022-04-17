package subsystems

type BlkioSubsystem struct {
}

func (b *BlkioSubsystem) Name() string {
	//TODO implement me
	panic("implement me")
}

func (b *BlkioSubsystem) Set(cgroupPath string, res *ResourceConfig) error {
	//TODO implement me
	panic("implement me")
}

func (b *BlkioSubsystem) Apply(cgroupPath string, pid int) error {
	//TODO implement me
	panic("implement me")
}

func (b *BlkioSubsystem) Remove(cgroupPath string) error {
	//TODO implement me
	panic("implement me")
}
