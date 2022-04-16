package examples

import (
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"os/exec"
	"path"
	"strconv"
	"syscall"
)

const cgroupMemoryHierarchyMount = "sys/fs/cgroup/memory"

func cgroup() {
	if os.Args[0] == "/proc/self/exc" {
		fmt.Printf("current pid %d /n", syscall.Getpid())
		cmd := exec.Command("sh", "-c", `stress --vm-bytes 200m --vm-keep -m 1`)
		cmd.SysProcAttr = &syscall.SysProcAttr{}

		cmd.Stdin = os.Stdin
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr
		if err := cmd.Run(); err != nil {
			fmt.Println(err)
			os.Exit(1)
		}
	}

	cmd := exec.Command("/proc/self/exe")
	cmd.SysProcAttr = &syscall.SysProcAttr{
		Cloneflags: syscall.CLONE_NEWUTS | syscall.CLONE_NEWPID | syscall.CLONE_NEWNS,
	}
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	if err := cmd.Run(); err != nil {
		log.Println("EORROR", err)
		os.Exit(1)
	} else {
		fmt.Printf("%v", cmd.Process.Pid)

		os.Mkdir(path.Join(cgroupMemoryHierarchyMount, "testmemorylimit"), 0755)
		ioutil.WriteFile(
			path.Join(cgroupMemoryHierarchyMount, "testmemorylimit", "tasks"),
			[]byte(strconv.Itoa(cmd.Process.Pid)),
			0644)
		ioutil.WriteFile(
			path.Join(cgroupMemoryHierarchyMount, "testmemorylimit", "memory.limit_in_bytes"),
			[]byte("100m"),
			0644)
	}
	cmd.Process.Wait()
}
