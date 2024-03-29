package network

import (
	"encoding/json"
	"fmt"
	"github.com/sirupsen/logrus"
	"net"
	"os"
	"path"
	"strings"
)

const (
	ipamDefaultAllocatorPath = "/var/run/myrunc/network/ipam/subnet.json"
)

type IPAM struct {
	SubnetAllocatorPath string
	Subnets             *map[string]string
}

var ipAllocator = &IPAM{
	SubnetAllocatorPath: ipamDefaultAllocatorPath,
}

func (ipam *IPAM) load() error {
	if _, err := os.Stat(ipam.SubnetAllocatorPath); err != nil {
		if os.IsNotExist(err) {
			return nil
		} else {
			return err
		}
	}
	subnetConfigFile, err := os.Open(ipam.SubnetAllocatorPath)
	defer subnetConfigFile.Close()
	if err != nil {
		return err
	}
	subnetJson := make([]byte, 2000)
	n, err := subnetConfigFile.Read(subnetJson)
	if err != nil {
		return err
	}
	err = json.Unmarshal(subnetJson[:n], ipam.Subnets)
	if err != nil {
		logrus.Errorf("error dump allocation info; %v", err)
		return err
	}
	return nil
}

func (ipam *IPAM) dump() error {
	ipamConfigFileDir, _ := path.Split(ipam.SubnetAllocatorPath)
	if _, err := os.Stat(ipamConfigFileDir); err != nil {
		if os.IsNotExist(err) {
			os.MkdirAll(ipamConfigFileDir, 0644)
		} else {
			return err
		}
	}
	subnetConfigFile, err := os.OpenFile(ipam.SubnetAllocatorPath, os.O_TRUNC|os.O_WRONLY|os.O_CREATE, 0644)
	defer subnetConfigFile.Close()
	if err != nil {
		return err
	}

	ipamConfigJson, err := json.Marshal(ipam.Subnets)
	if err != nil {
		return err
	}

	_, err = subnetConfigFile.Write(ipamConfigJson)
	if err != nil {
		return err
	}

	return nil
}

func (ipam *IPAM) CreateSubnet(subnet *net.IPNet) (net.IP, error) {
	ipam.Subnets = &map[string]string{}
	err := ipam.load()
	if err != nil {
		logrus.Errorf("error load allocation info: %v", err)
		return nil, err
	}

	ones, size := subnet.Mask.Size()
	subnetStr := subnet.String()

	_, exist := (*ipam.Subnets)[subnetStr]
	if exist {
		return nil, fmt.Errorf("pool overlaps with other one on this address space: %s", subnetStr)
	}

	// 如果之前没有分配过指定网段，则初始化网段的分配配置
	// 用 0 填满网段配置，1<<uint8(size-ones)表示网段中的可用地址数目
	// size-ones表示子网掩码后买呢的网络位数，2^(size-ones)表示可用ip数目
	(*ipam.Subnets)[subnetStr] = strings.Repeat("0", 1<<uint8(size-ones))

	ip, err := ipam.doAllocate(subnetStr)
	if err != nil {
		return nil, fmt.Errorf("allocate ip addr error: %v", err)
	}

	err = ipam.dump()
	if err != nil {
		logrus.Errorf("error dump allocation info: %v", err)
		return nil, err
	}

	return ip, nil

}

func (ipam *IPAM) Allocate(subnet string) (net.IP, error) {
	ipam.Subnets = &map[string]string{}
	err := ipam.load()
	if err != nil {
		logrus.Errorf("error load allocation info: %v", err)
		return nil, err
	}

	ip, err := ipam.doAllocate(subnet)
	if err != nil {
		return nil, fmt.Errorf("allocate ip addr error: %v", err)
	}

	err = ipam.dump()
	if err != nil {
		logrus.Errorf("error dump allocation info: %v", err)
		return nil, err
	}

	return ip, nil
}

func (ipam *IPAM) doAllocate(subnet string) (net.IP, error) {
	var ip net.IP

	bitMap := (*ipam.Subnets)[subnet]

	for idx, ch := range bitMap {
		if ch == '0' {
			// 计算当前数组偏移量对应的ip
			// 示例：
			// 原始数组[172，16，0，0]，偏移量idx=65555
			// 则需要在各个部分依次加上 [uint8(65555>>24),uint8(65555>>16)，uint8(65555>>8)，uint8(65555>>0)]
			// 结果为[0,1,0,19] 则偏移后的ip为[172,17,0,19]
			ip, _, _ = net.ParseCIDR(subnet)
			ip = ip.To4()
			for t := uint(4); t > 0; t -= 1 {
				ip[4-t] += uint8(idx >> ((t - 1) * 8))
			}

			// 计算下一个可用的ip

			ip[3] += uint8(1)

			ipalloc := []byte((*ipam.Subnets)[subnet])
			ipalloc[idx] = '1'
			(*ipam.Subnets)[subnet] = string(ipalloc)
			break
		}
	}

	if ip == nil {
		return nil, fmt.Errorf("no available ip for subnet %s", subnet)
	}

	return ip, nil
}

func (ipam *IPAM) Release(subnet *net.IPNet, ipaddr *net.IP) interface{} {
	ipam.Subnets = &map[string]string{}

	_, subnet, _ = net.ParseCIDR(subnet.String())

	err := ipam.load()
	if err != nil {
		logrus.Errorf("Error dump allocation info, %v", err)
	}

	c := 0
	releaseIP := ipaddr.To4()
	releaseIP[3] -= 1
	for t := uint(4); t > 0; t -= 1 {
		c += int(releaseIP[t-1]-subnet.IP[t-1]) << ((4 - t) * 8)
	}

	ipalloc := []byte((*ipam.Subnets)[subnet.String()])
	ipalloc[c] = '0'
	(*ipam.Subnets)[subnet.String()] = string(ipalloc)

	ipam.dump()
	return nil
}

func (ipam *IPAM) Delete(subnetStr string) error {
	ipam.Subnets = &map[string]string{}
	err := ipam.load()
	if err != nil {
		logrus.Errorf("error load allocation info: %v", err)
		return err
	}

	// 检查是否存在除网关外仍在使用的ip，如果存在，不允许释放
	ipalloc := []byte((*ipam.Subnets)[subnetStr])
	for i := 1; i < len(ipalloc); i++ {
		if ipalloc[i] == '1' {
			return fmt.Errorf("exist used ip addr in subnet %s", subnetStr)
		}
	}

	delete(*ipam.Subnets, subnetStr)

	err = ipam.dump()
	if err != nil {
		logrus.Errorf("error dump allocation info: %v", err)
		return err
	}

	return nil

}
