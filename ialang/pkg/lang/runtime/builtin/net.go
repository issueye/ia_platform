package builtin

import (
	"fmt"
	goNet "net"
	"strconv"
)

func newNetModule() Object {
	isIPFn := NativeFunction(func(args []Value) (Value, error) {
		if len(args) != 1 {
			return nil, fmt.Errorf("net.isIP expects 1 arg, got %d", len(args))
		}
		text, err := asStringArg("net.isIP", args, 0)
		if err != nil {
			return nil, err
		}
		return goNet.ParseIP(text) != nil, nil
	})

	isIPv4Fn := NativeFunction(func(args []Value) (Value, error) {
		if len(args) != 1 {
			return nil, fmt.Errorf("net.isIPv4 expects 1 arg, got %d", len(args))
		}
		text, err := asStringArg("net.isIPv4", args, 0)
		if err != nil {
			return nil, err
		}
		ip := goNet.ParseIP(text)
		return ip != nil && ip.To4() != nil, nil
	})

	isIPv6Fn := NativeFunction(func(args []Value) (Value, error) {
		if len(args) != 1 {
			return nil, fmt.Errorf("net.isIPv6 expects 1 arg, got %d", len(args))
		}
		text, err := asStringArg("net.isIPv6", args, 0)
		if err != nil {
			return nil, err
		}
		ip := goNet.ParseIP(text)
		return ip != nil && ip.To4() == nil, nil
	})

	parseHostPortFn := NativeFunction(func(args []Value) (Value, error) {
		if len(args) != 1 {
			return nil, fmt.Errorf("net.parseHostPort expects 1 arg, got %d", len(args))
		}
		addr, err := asStringArg("net.parseHostPort", args, 0)
		if err != nil {
			return nil, err
		}
		host, port, err := goNet.SplitHostPort(addr)
		if err != nil {
			return nil, err
		}
		portNum, convErr := strconv.Atoi(port)
		if convErr != nil {
			portNum = 0
		}
		return Object{
			"host": host,
			"port": float64(portNum),
		}, nil
	})

	joinHostPortFn := NativeFunction(func(args []Value) (Value, error) {
		if len(args) != 2 {
			return nil, fmt.Errorf("net.joinHostPort expects 2 args, got %d", len(args))
		}
		host, err := asStringValue("net.joinHostPort arg[0]", args[0])
		if err != nil {
			return nil, err
		}
		port, err := asIntValue("net.joinHostPort arg[1]", args[1])
		if err != nil {
			return nil, err
		}
		if port < 0 || port > 65535 {
			return nil, fmt.Errorf("net.joinHostPort port out of range: %d", port)
		}
		return goNet.JoinHostPort(host, strconv.Itoa(port)), nil
	})

	lookupIPFn := NativeFunction(func(args []Value) (Value, error) {
		if len(args) != 1 {
			return nil, fmt.Errorf("net.lookupIP expects 1 arg, got %d", len(args))
		}
		host, err := asStringArg("net.lookupIP", args, 0)
		if err != nil {
			return nil, err
		}
		ips, err := goNet.LookupIP(host)
		if err != nil {
			return nil, err
		}
		out := make(Array, 0, len(ips))
		for _, ip := range ips {
			out = append(out, ip.String())
		}
		return out, nil
	})

	parseCIDRFn := NativeFunction(func(args []Value) (Value, error) {
		if len(args) != 1 {
			return nil, fmt.Errorf("net.parseCIDR expects 1 arg, got %d", len(args))
		}
		cidrText, err := asStringArg("net.parseCIDR", args, 0)
		if err != nil {
			return nil, err
		}
		ip, ipnet, err := goNet.ParseCIDR(cidrText)
		if err != nil {
			return nil, err
		}
		ones, bits := ipnet.Mask.Size()
		return Object{
			"ip":       ip.String(),
			"network":  ipnet.IP.String(),
			"maskOnes": float64(ones),
			"maskBits": float64(bits),
		}, nil
	})

	containsCIDRFn := NativeFunction(func(args []Value) (Value, error) {
		if len(args) != 2 {
			return nil, fmt.Errorf("net.containsCIDR expects 2 args, got %d", len(args))
		}
		cidrText, err := asStringValue("net.containsCIDR arg[0]", args[0])
		if err != nil {
			return nil, err
		}
		ipText, err := asStringValue("net.containsCIDR arg[1]", args[1])
		if err != nil {
			return nil, err
		}
		ip := goNet.ParseIP(ipText)
		if ip == nil {
			return nil, fmt.Errorf("net.containsCIDR invalid ip: %q", ipText)
		}
		_, ipnet, err := goNet.ParseCIDR(cidrText)
		if err != nil {
			return nil, err
		}
		return ipnet.Contains(ip), nil
	})

	namespace := Object{
		"isIP":          isIPFn,
		"isIPv4":        isIPv4Fn,
		"isIPv6":        isIPv6Fn,
		"parseHostPort": parseHostPortFn,
		"joinHostPort":  joinHostPortFn,
		"lookupIP":      lookupIPFn,
		"parseCIDR":     parseCIDRFn,
		"containsCIDR":  containsCIDRFn,
	}
	module := cloneObject(namespace)
	module["net"] = namespace
	return module
}
