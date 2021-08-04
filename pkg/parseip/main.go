package parseip

import (
	"errors"
	"fmt"
	"net"
	"strconv"
	"strings"
)

// 支持常见的ip格式
// 192.168.113.159
// 192.168.113.159-254
// 192.168.113.159-192.168.113.254
// 192.168.113.0/24
// 191.168.113.159-192.168.114.254
// 192.167.113.159-192.168.114.254
// 192.168.113.159-192.168.114.254
func ParseIP(ip string) (startx uint32, endx uint32, err error) {
	if strings.Contains(ip, "-") {
		if len(strings.Split(ip, "-")[1]) <= 3 {
			return multipleip(ip)
		} else {
			return multipleip2(ip)
		}
	} else if strings.Contains(ip, "/") {
		return multipleip3(ip)
	} else {
		return singleip(ip)
	}
}

// 192.168.113.159
func singleip(ip string) (startx uint32, endx uint32, err error) {
	for _, val := range strings.Split(ip, ".") {
		ips, err := strconv.Atoi(val)
		if err != nil {
			return 0, 0, errors.New(ip + " " + err.Error() + " ip parse error")
		}
		if ips > 255 {
			return 0, 0, errors.New(ip + " ip parse error")
		}
	}
	startx, err = ip2UInt32(ip)
	if err != nil {
		return 0, 0, err
	}
	endx, err = ip2UInt32(ip)
	if err != nil {
		return 0, 0, err
	}
	return startx, endx, nil
}

// 192.168.113.159-255
func multipleip(ips string) (startx uint32, endx uint32, err error) {
	host := strings.Split(ips, "-")
	ip := host[0]
	if len(strings.Split(ip, ".")) != 4 {
		return 0, 0, errors.New("multipleip error " + ips)
	}
	start, err := strconv.Atoi(strings.Split(ip, ".")[3])
	if err != nil {
		return 0, 0, errors.New(ips + " " + err.Error() + " ip parse error")
	}
	end, err := strconv.Atoi(host[1])
	if err != nil {
		return 0, 0, errors.New(ips + " " + err.Error() + " ip parse error")
	}
	if start > end {
		return 0, 0, errors.New(ips + " ip parse error")
	}
	if start < 0 {
		start = 0
	}
	if end > 255 {
		end = 255
	}
	temp := strings.Split(ip, ".")
	start_t, err := ip2UInt32(temp[0] + "." + temp[1] + "." + temp[2] + "." + strconv.Itoa(start))
	if err != nil {
		return 0, 0, err
	}
	end_t, err := ip2UInt32(temp[0] + "." + temp[1] + "." + temp[2] + "." + strconv.Itoa(end))
	if err != nil {
		return 0, 0, err
	}
	return start_t, end_t, nil
}

// 192.168.113.159-192.168.113.254
func multipleip2(ips string) (startx uint32, endx uint32, err error) {
	start, err := ip2UInt32(strings.Split(ips, "-")[0])
	if err != nil {
		return 0, 0, err
	}
	end, err := ip2UInt32(strings.Split(ips, "-")[1])
	if err != nil {
		return 0, 0, err
	}
	if start > end {
		return 0, 0, errors.New(ips + " error")
	}
	return start, end, nil
}

// 192.168.113.0/24
func multipleip3(ips string) (startx uint32, endx uint32, err error) {
	host := strings.Split(ips, "/")[0]
	mask, err := strconv.Atoi(strings.Split(ips, "/")[1])
	if err != nil {
		return 0, 0, errors.New(ips + " " + err.Error() + " ip parse error")
	}
	if len(strings.Split(host, ".")) != 4 {
		return 0, 0, errors.New(ips + " ip parse error")
	}
	a, err := strconv.Atoi(strings.Split(host, ".")[0])
	b, err := strconv.Atoi(strings.Split(host, ".")[1])
	c, err := strconv.Atoi(strings.Split(host, ".")[2])
	d, err := strconv.Atoi(strings.Split(host, ".")[3])
	if err != nil {
		return 0, 0, errors.New(ips + " ip parse error")
	}
	ipbin := fmt.Sprintf("%08s", strconv.FormatInt(int64(a), 2)) +
		fmt.Sprintf("%08s", strconv.FormatInt(int64(b), 2)) +
		fmt.Sprintf("%08s", strconv.FormatInt(int64(c), 2)) +
		fmt.Sprintf("%08s", strconv.FormatInt(int64(d), 2))

	start := ipbin[:mask]
	end := ipbin[:mask]
	for i := 0; i < len(ipbin)-mask; i++ {
		start += "0"
		end += "1"
	}
	start1, err := strconv.ParseUint(start, 2, 32)
	if err != nil {
		return 0, 0, errors.New(ips + " ip parse error: " + err.Error())
	}
	end2, err := strconv.ParseUint(end, 2, 32)
	if err != nil {
		return 0, 0, errors.New(ips + " ip parse error: " + err.Error())
	}
	return uint32(start1), uint32(end2), nil
}

func ip2UInt32(ipnr string) (uint32, error) {
	bits := strings.Split(ipnr, ".")
	if len(bits) != 4 {
		return 0, errors.New("ip2Uint32 error " + ipnr)
	}

	b0, err := strconv.Atoi(bits[0])
	if err != nil {
		return 0, err
	}
	b1, err := strconv.Atoi(bits[1])
	if err != nil {
		return 0, err
	}
	b2, err := strconv.Atoi(bits[2])
	if err != nil {
		return 0, err
	}
	b3, err := strconv.Atoi(bits[3])
	if err != nil {
		return 0, err
	}

	var sum uint32
	sum += uint32(b0) << 24
	sum += uint32(b1) << 16
	sum += uint32(b2) << 8
	sum += uint32(b3)
	return sum, nil
}

func UInt32ToIP(intIP uint32) string {
	var bytes [4]byte
	bytes[0] = byte(intIP & 0xFF)
	bytes[1] = byte((intIP >> 8) & 0xFF)
	bytes[2] = byte((intIP >> 16) & 0xFF)
	bytes[3] = byte((intIP >> 24) & 0xFF)

	return net.IPv4(bytes[3], bytes[2], bytes[1], bytes[0]).String()
}
