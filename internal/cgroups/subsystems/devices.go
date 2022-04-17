package subsystems

type DevicesSubsystem struct {
}

func (d *DevicesSubsystem) Name() string {
	//TODO implement me
	panic("implement me")
}

func (d *DevicesSubsystem) Set(cgroupPath string, res *ResourceConfig) error {
	//TODO implement me
	panic("implement me")
}

func (d *DevicesSubsystem) Apply(cgroupPath string, pid int) error {
	//TODO implement me
	panic("implement me")
}

func (d *DevicesSubsystem) Remove(cgroupPath string) error {
	//TODO implement me
	panic("implement me")
}
