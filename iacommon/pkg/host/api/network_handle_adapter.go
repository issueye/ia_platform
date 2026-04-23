package api

import hostnet "iacommon/pkg/host/network"

type NetworkDialRequest struct {
	Endpoint hostnet.Endpoint
	Opts     hostnet.DialOptions
}

type NetworkDialResponse struct {
	Handle uint64
}

type NetworkListenRequest struct {
	Endpoint hostnet.Endpoint
	Opts     hostnet.ListenOptions
}

type NetworkListenResponse struct {
	Handle uint64
}

type NetworkAcceptRequest struct {
	Handle uint64
}

type NetworkAcceptResponse struct {
	Handle uint64
}

type NetworkSendRequest struct {
	Handle uint64
	Data   []byte
}

type NetworkSendResponse struct {
	N int64
}

type NetworkRecvRequest struct {
	Handle uint64
	Size   int64
}

type NetworkRecvResponse struct {
	Data []byte
	N    int64
}

type NetworkCloseRequest struct {
	Handle uint64
}

func decodeNetworkDialRequest(args map[string]any) (NetworkDialRequest, error) {
	networkName, err := readOptionalStringAny(args, "network")
	if err != nil {
		return NetworkDialRequest{}, err
	}
	if networkName == "" {
		networkName = "tcp"
	}
	host, err := readString(args, "host")
	if err != nil {
		return NetworkDialRequest{}, err
	}
	port, err := readOptionalInt64Any(args, "port")
	if err != nil {
		return NetworkDialRequest{}, err
	}
	timeoutMS, err := readOptionalInt64Any(args, "timeout_ms", "timeoutMS")
	if err != nil {
		return NetworkDialRequest{}, err
	}
	return NetworkDialRequest{
		Endpoint: hostnet.Endpoint{
			Network: networkName,
			Host:    host,
			Port:    int(port),
		},
		Opts: hostnet.DialOptions{
			TimeoutMS: timeoutMS,
		},
	}, nil
}

func encodeNetworkDialResponse(handle uint64) CallResult {
	return CallResult{Value: map[string]any{"handle": handle}}
}

func decodeNetworkListenRequest(args map[string]any) (NetworkListenRequest, error) {
	networkName, err := readOptionalStringAny(args, "network")
	if err != nil {
		return NetworkListenRequest{}, err
	}
	if networkName == "" {
		networkName = "tcp"
	}
	host, err := readString(args, "host")
	if err != nil {
		return NetworkListenRequest{}, err
	}
	port, err := readOptionalInt64Any(args, "port")
	if err != nil {
		return NetworkListenRequest{}, err
	}
	backlog, err := readOptionalInt64Any(args, "backlog")
	if err != nil {
		return NetworkListenRequest{}, err
	}
	return NetworkListenRequest{
		Endpoint: hostnet.Endpoint{
			Network: networkName,
			Host:    host,
			Port:    int(port),
		},
		Opts: hostnet.ListenOptions{
			Backlog: int(backlog),
		},
	}, nil
}

func encodeNetworkListenResponse(handle uint64) CallResult {
	return CallResult{Value: map[string]any{"handle": handle}}
}

func decodeNetworkAcceptRequest(args map[string]any) (NetworkAcceptRequest, error) {
	handle, err := readRequiredUint64(args, "handle")
	if err != nil {
		return NetworkAcceptRequest{}, err
	}
	return NetworkAcceptRequest{Handle: handle}, nil
}

func encodeNetworkAcceptResponse(handle uint64) CallResult {
	return CallResult{Value: map[string]any{"handle": handle}}
}

func decodeNetworkSendRequest(args map[string]any) (NetworkSendRequest, error) {
	handle, err := readRequiredUint64(args, "handle")
	if err != nil {
		return NetworkSendRequest{}, err
	}
	data, err := readBytes(args, "data")
	if err != nil {
		return NetworkSendRequest{}, err
	}
	return NetworkSendRequest{Handle: handle, Data: data}, nil
}

func encodeNetworkSendResponse(n int64) CallResult {
	return CallResult{Value: map[string]any{"n": n}}
}

func decodeNetworkRecvRequest(args map[string]any) (NetworkRecvRequest, error) {
	handle, err := readRequiredUint64(args, "handle")
	if err != nil {
		return NetworkRecvRequest{}, err
	}
	size, err := readOptionalInt64Any(args, "size")
	if err != nil {
		return NetworkRecvRequest{}, err
	}
	if size <= 0 {
		size = 4096
	}
	return NetworkRecvRequest{Handle: handle, Size: size}, nil
}

func encodeNetworkRecvResponse(data []byte, n int64) CallResult {
	return CallResult{Value: map[string]any{
		"data": data,
		"n":    n,
	}}
}

func decodeNetworkCloseRequest(args map[string]any) (NetworkCloseRequest, error) {
	handle, err := readRequiredUint64(args, "handle")
	if err != nil {
		return NetworkCloseRequest{}, err
	}
	return NetworkCloseRequest{Handle: handle}, nil
}
