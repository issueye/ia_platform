package lang

import commonmodule "iacommon/pkg/ialang/module"

type ModuleResolverOptions = commonmodule.ModuleResolverOptions

type ProjectConfig = commonmodule.ProjectConfig

func ProjectConfigFileName() string {
	return commonmodule.ProjectConfigFileName()
}

func DiscoverModuleResolverOptions(startPath string) (ModuleResolverOptions, error) {
	return commonmodule.DiscoverModuleResolverOptions(startPath)
}

func ReadProjectConfig(path string) (*ProjectConfig, error) {
	return commonmodule.ReadProjectConfig(path)
}

func ResolveModulePathWithOptions(fromPath, moduleName string, opts ModuleResolverOptions) (string, error) {
	return commonmodule.ResolveModulePathWithOptions(fromPath, moduleName, opts)
}
