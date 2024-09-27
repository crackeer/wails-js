package bind

import (
	"context"
	"fmt"
	"log"
	"net"
	"os/user"
	"path/filepath"
)

// System App struct
type System struct {
	ctx context.Context
}

// NewApp creates a new App application struct
func NewSystem() *System {
	return &System{}
}

// GetInnerIP ...
//
//	@receiver a
//	@param name
//	@return string
func (a *System) GetInnerIP() []string {
	ipList := []string{}
	interfaces, err := net.Interfaces()
	if err != nil {
		fmt.Println("Error getting network interfaces:", err)
		return ipList
	}

	for _, iface := range interfaces {
		addrs, err := iface.Addrs()
		if err != nil {
			fmt.Printf("Error fetching addresses for %q: %v\n", iface.Name, err)
			continue
		}

		for _, addr := range addrs {
			var ip net.IP
			switch v := addr.(type) {
			case *net.IPNet:
				ip = v.IP
			case *net.IPAddr:
				ip = v.IP
			}

			if ip.IsLoopback() || ip.IsLinkLocalUnicast() || ip.IsGlobalUnicast() == false {
				// 忽略回环地址、链路本地地址和非全局单播地址
				continue
			}

			if ip.To4() != nil && ip.IsPrivate() {
				ipList = append(ipList, ip.String())
			}
			// 注意：如果你也需要IPv6私有地址，可以取消上面的IP版本检查
		}
	}
	return ipList
}

func (a *System) GetDownloadDir() string {
	u, err := user.Current()
	if err != nil {
		log.Fatal(err)
		return ""
	}

	return filepath.Join(u.HomeDir, "Downloads")
	/*
		// 根据操作系统获取默认下载目录
		switch runtime.GOOS {
		case "windows":
			// 在Windows上通常下载目录在用户的文档文件夹内
			return filepath.Join(os.Getenv("USERPROFILE"), "Downloads")
		case "linux":
			// 在Linux上，下载目录可能在用户的文档文件夹内，或者通过环境变量获取
			xdgDownloadDir := os.Getenv("XDG_DOWNLOAD_DIR")
			if xdgDownloadDir != "" {
				return xdgDownloadDir
			}
			return filepath.Join(os.Getenv("HOME"), "Downloads")
		case "darwin":
			// 在macOS上，下载目录可能在用户的文档文件夹内，或者通过环境变量获取
			return filepath.Join(os.Getenv("HOME"), "Downloads")
		}

		return ""
	*/
}
