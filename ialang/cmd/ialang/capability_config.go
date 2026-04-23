package main

import (
	"fmt"
	"os"

	"github.com/BurntSushi/toml"

	"iacommon/pkg/host/api"
	hostfs "iacommon/pkg/host/fs"
	hostnet "iacommon/pkg/host/network"
	"iavm/pkg/module"
)

type capabilityConfigFile struct {
	FS      capabilityFSConfig      `toml:"fs"`
	Network capabilityNetworkConfig `toml:"network"`
}

type capabilityFSConfig struct {
	Rights   []string                  `toml:"rights"`
	Preopens []capabilityFSPreopenDecl `toml:"preopens"`
}

type capabilityFSPreopenDecl struct {
	VirtualPath string `toml:"virtual_path"`
	RealPath    string `toml:"real_path"`
	ReadOnly    bool   `toml:"read_only"`
}

type capabilityNetworkConfig struct {
	Rights             []string `toml:"rights"`
	AllowHosts         []string `toml:"allow_hosts"`
	AllowPorts         []int    `toml:"allow_ports"`
	AllowSchemes       []string `toml:"allow_schemes"`
	AllowCIDRs         []string `toml:"allow_cidrs"`
	MaxConnections     int      `toml:"max_connections"`
	MaxInflightRequest int      `toml:"max_inflight_request"`
	MaxBytesPerRequest int64    `toml:"max_bytes_per_request"`
}

func loadCapabilityConfig(path string) (*capabilityConfigFile, error) {
	if path == "" {
		return &capabilityConfigFile{}, nil
	}

	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("[cap-config] read file error: %w", err)
	}

	var cfg capabilityConfigFile
	if err := toml.Unmarshal(data, &cfg); err != nil {
		return nil, fmt.Errorf("[cap-config] decode error: %w", err)
	}
	return &cfg, nil
}

func applyCapabilityConfig(mod *module.Module, cfg *capabilityConfigFile) {
	if mod == nil || cfg == nil {
		return
	}

	overrides := map[module.CapabilityKind]map[string]any{}
	if values := cfg.FS.toCapabilityConfigMap(); len(values) > 0 {
		overrides[module.CapabilityFS] = values
	}
	if values := cfg.Network.toCapabilityConfigMap(); len(values) > 0 {
		overrides[module.CapabilityNetwork] = values
	}

	if len(overrides) == 0 {
		return
	}

	for i := range mod.Capabilities {
		override, ok := overrides[mod.Capabilities[i].Kind]
		if !ok {
			continue
		}
		mod.Capabilities[i].Config = mergeCapabilityConfigMaps(mod.Capabilities[i].Config, override)
	}
}

func buildRunIavmHost(cfg *capabilityConfigFile) (*api.DefaultHost, error) {
	networkPolicy := hostnet.Policy{}
	if cfg != nil {
		networkPolicy = hostnet.Policy{
			Rights:             copyStringSlice(cfg.Network.Rights),
			AllowHosts:         copyStringSlice(cfg.Network.AllowHosts),
			AllowPorts:         copyIntSlice(cfg.Network.AllowPorts),
			AllowSchemes:       copyStringSlice(cfg.Network.AllowSchemes),
			AllowCIDRs:         copyStringSlice(cfg.Network.AllowCIDRs),
			MaxConnections:     cfg.Network.MaxConnections,
			MaxInflightRequest: cfg.Network.MaxInflightRequest,
			MaxBytesPerRequest: cfg.Network.MaxBytesPerRequest,
		}
	}
	host := &api.DefaultHost{
		FS: &hostfs.MemFSProvider{},
		Network: &hostnet.CompositeProvider{
			HTTP:   &hostnet.HTTPProvider{Policy: networkPolicy},
			Socket: &hostnet.SocketProvider{Policy: networkPolicy},
		},
	}
	if cfg == nil {
		return host, nil
	}

	if len(cfg.FS.Preopens) > 0 {
		preopens := make([]hostfs.Preopen, 0, len(cfg.FS.Preopens))
		for _, entry := range cfg.FS.Preopens {
			preopens = append(preopens, hostfs.Preopen{
				VirtualPath: entry.VirtualPath,
				RealPath:    entry.RealPath,
				ReadOnly:    entry.ReadOnly,
			})
		}
		mapper, err := hostfs.NewPreopenPathMapper(preopens)
		if err != nil {
			return nil, fmt.Errorf("[cap-config] fs preopen error: %w", err)
		}
		host.FS = &hostfs.LocalFSProvider{Mapper: mapper}
	}
	return host, nil
}

func (cfg capabilityFSConfig) toCapabilityConfigMap() map[string]any {
	result := map[string]any{}
	if len(cfg.Rights) > 0 {
		result["rights"] = copyStringSlice(cfg.Rights)
	}
	if len(cfg.Preopens) > 0 {
		preopens := make([]any, 0, len(cfg.Preopens))
		for _, entry := range cfg.Preopens {
			preopens = append(preopens, map[string]any{
				"virtual_path": entry.VirtualPath,
				"real_path":    entry.RealPath,
				"read_only":    entry.ReadOnly,
			})
		}
		result["preopens"] = preopens
	}
	return result
}

func (cfg capabilityNetworkConfig) toCapabilityConfigMap() map[string]any {
	result := map[string]any{}
	if len(cfg.Rights) > 0 {
		result["rights"] = copyStringSlice(cfg.Rights)
	}
	if len(cfg.AllowHosts) > 0 {
		result["allow_hosts"] = copyStringSlice(cfg.AllowHosts)
	}
	if len(cfg.AllowPorts) > 0 {
		result["allow_ports"] = copyIntSlice(cfg.AllowPorts)
	}
	if len(cfg.AllowSchemes) > 0 {
		result["allow_schemes"] = copyStringSlice(cfg.AllowSchemes)
	}
	if len(cfg.AllowCIDRs) > 0 {
		result["allow_cidrs"] = copyStringSlice(cfg.AllowCIDRs)
	}
	if cfg.MaxConnections > 0 {
		result["max_connections"] = cfg.MaxConnections
	}
	if cfg.MaxInflightRequest > 0 {
		result["max_inflight_request"] = cfg.MaxInflightRequest
	}
	if cfg.MaxBytesPerRequest > 0 {
		result["max_bytes_per_request"] = cfg.MaxBytesPerRequest
	}
	return result
}

func mergeCapabilityConfigMaps(base, overlay map[string]any) map[string]any {
	if len(base) == 0 && len(overlay) == 0 {
		return nil
	}
	result := make(map[string]any, len(base)+len(overlay))
	for key, value := range base {
		result[key] = value
	}
	for key, value := range overlay {
		result[key] = value
	}
	return result
}

func copyStringSlice(values []string) []string {
	if len(values) == 0 {
		return nil
	}
	result := make([]string, len(values))
	copy(result, values)
	return result
}

func copyIntSlice(values []int) []int {
	if len(values) == 0 {
		return nil
	}
	result := make([]int, len(values))
	copy(result, values)
	return result
}
