package builtin

import (
	common "iacommon/pkg/ialang/value"
	"ialang/pkg/lang/runtime"
	rttypes "ialang/pkg/lang/runtime/types"
)

func DefaultModules(asyncRuntime rttypes.AsyncRuntime) map[string]common.Value {
	if asyncRuntime == nil {
		asyncRuntime = runtime.NewGoroutineRuntime()
	}

	httpModule := newHTTPModule(asyncRuntime)
	websocketModule := newWebSocketModule(asyncRuntime)
	sseModule := newSSEModule(asyncRuntime)
	expressModule := newExpressModule(asyncRuntime)
	databaseModule := newDatabaseModule(asyncRuntime)
	yamlModule := newYAMLModule()
	tomlModule := newTOMLModule()
	assetBundleModule := newAssetBundleModule()
	ormModule := newORMModule()
	timerModule := newTimerModule(asyncRuntime)
	goroutinePoolModule := newGoroutinePoolModule()
	fsModule := newFSModule(asyncRuntime)
	osModule := newOSModule()
	processModule := newProcessModule()
	signalModule := newSignalModule()
	execModule := newExecModule(asyncRuntime)
	logModule := newLogModule()
	pathModule := newPathModule()
	jsonModule := newJSONModule(asyncRuntime)
	timeModule := newTimeModule(asyncRuntime)
	encodingModule := newEncodingModule()
	cryptoModule := newCryptoModule()
	regexpModule := newRegexpModule()
	uuidModule := newUUIDModule()
	urlModule := newURLModule()
	strconvModule := newStrconvModule()
	randModule := newRandModule()
	csvModule := newCSVModule()
	xmlModule := newXMLModule()
	hexModule := newHexModule()
	netModule := newNetModule()
	mimeModule := newMIMEModule()
	hashModule := newHashModule()
	compressModule := newCompressModule()
	hmacModule := newHMACModule()
	bytesModule := newBytesModule()
	sortModule := newSortModule()
	setModule := newSetModule()
	ipcModule := newIPCModule(asyncRuntime)
	socketModule := newSocketModule(asyncRuntime)
	iaxModule := newIAXModule(asyncRuntime)
	mathModule := newMathModule()
	stringModule := newStringModule()
	arrayModule := newArrayModule()
	promiseModule := newPromiseModule()

	modules := map[string]common.Value{
		"@agent/sdk": newAgentSDKModule(asyncRuntime),
	}

	groups := []moduleAliasGroup{
		{plain: []string{"http"}, value: httpModule},
		{plain: []string{"websocket"}, value: websocketModule},
		{plain: []string{"sse"}, value: sseModule},
		{plain: []string{"express"}, value: expressModule},
		{plain: []string{"db", "database"}, value: databaseModule},
		{plain: []string{"yaml"}, value: yamlModule},
		{plain: []string{"toml"}, value: tomlModule},
		{plain: []string{"asset", "bundle"}, value: assetBundleModule},
		// setTimeout 保持 plain 兼容，不扩展到 @std/@stdlib。
		{plain: []string{"timer", "setTimeout"}, prefixed: []string{"timer"}, value: timerModule},
		{plain: []string{"orm"}, value: ormModule},
		{plain: []string{"pool"}, value: goroutinePoolModule},
		{plain: []string{"fs"}, value: fsModule},
		{plain: []string{"os"}, value: osModule},
		{plain: []string{"process"}, value: processModule},
		{plain: []string{"signal"}, value: signalModule},
		{plain: []string{"exec", "os/exec"}, value: execModule},
		{plain: []string{"log"}, value: logModule},
		{plain: []string{"path"}, value: pathModule},
		{plain: []string{"json"}, value: jsonModule},
		{plain: []string{"time"}, value: timeModule},
		{plain: []string{"encoding"}, value: encodingModule},
		{plain: []string{"crypto"}, value: cryptoModule},
		{plain: []string{"regexp"}, value: regexpModule},
		{plain: []string{"uuid"}, value: uuidModule},
		{plain: []string{"url"}, value: urlModule},
		{plain: []string{"strconv"}, value: strconvModule},
		{plain: []string{"rand"}, value: randModule},
		{plain: []string{"csv"}, value: csvModule},
		{plain: []string{"xml"}, value: xmlModule},
		{plain: []string{"hex"}, value: hexModule},
		{plain: []string{"net"}, value: netModule},
		{plain: []string{"mime"}, value: mimeModule},
		{plain: []string{"hash"}, value: hashModule},
		{plain: []string{"compress"}, value: compressModule},
		{plain: []string{"hmac"}, value: hmacModule},
		{plain: []string{"bytes"}, value: bytesModule},
		{plain: []string{"sort"}, value: sortModule},
		{plain: []string{"set"}, value: setModule},
		{plain: []string{"ipc"}, value: ipcModule},
		{plain: []string{"socket"}, value: socketModule},
		{plain: []string{"iax", "interaction"}, value: iaxModule},
		{plain: []string{"math"}, value: mathModule},
		{plain: []string{"string"}, value: stringModule},
		{plain: []string{"array"}, value: arrayModule},
		{plain: []string{"Promise"}, value: promiseModule},
	}

	registerModuleAliasGroups(modules, groups, []string{"@std/", "@stdlib/"})
	return modules
}

type moduleAliasGroup struct {
	plain    []string
	prefixed []string
	value    common.Value
}

func registerModuleAliasGroups(modules map[string]common.Value, groups []moduleAliasGroup, prefixes []string) {
	for _, group := range groups {
		for _, key := range group.plain {
			modules[key] = group.value
		}

		keysForPrefixes := group.prefixed
		if len(keysForPrefixes) == 0 {
			keysForPrefixes = group.plain
		}
		for _, prefix := range prefixes {
			for _, key := range keysForPrefixes {
				modules[prefix+key] = group.value
			}
		}
	}
}
