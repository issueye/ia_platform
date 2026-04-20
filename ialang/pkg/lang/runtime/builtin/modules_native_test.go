package builtin

import (
	"io"
	"net/http"
	"net/http/httptest"
	"path/filepath"
	goRuntime "runtime"
	"strings"
	"testing"
	"time"

	rt "ialang/pkg/lang/runtime"
)

func TestDefaultModulesContainNativeModules(t *testing.T) {
	modules := DefaultModules(rt.NewGoroutineRuntime())
	for _, name := range []string{
		"http", "websocket", "sse", "fs", "os", "process", "exec", "log", "path", "json", "time", "encoding", "crypto", "regexp", "uuid", "url", "strconv", "rand", "csv", "xml", "hex", "net", "mime", "hash", "compress", "hmac", "bytes", "sort", "set", "ipc", "socket", "iax", "interaction",
		"os/exec",
		"@std/http", "@std/websocket", "@std/sse", "@std/fs", "@std/os", "@std/process", "@std/exec", "@std/os/exec", "@std/log", "@std/path", "@std/json", "@std/time", "@std/encoding", "@std/crypto", "@std/regexp", "@std/uuid", "@std/url", "@std/strconv", "@std/rand", "@std/csv", "@std/xml", "@std/hex", "@std/net", "@std/mime", "@std/hash", "@std/compress", "@std/hmac", "@std/bytes", "@std/sort", "@std/set", "@std/ipc", "@std/socket", "@std/iax", "@std/interaction",
		"@stdlib/websocket",
		"@stdlib/sse",
		"@stdlib/rand",
		"@stdlib/csv",
		"@stdlib/xml",
		"@stdlib/hex",
		"@stdlib/net",
		"@stdlib/mime",
		"@stdlib/hash",
		"@stdlib/compress",
		"@stdlib/hmac",
		"@stdlib/bytes",
		"@stdlib/sort",
		"@stdlib/set",
		"@stdlib/ipc",
		"@stdlib/socket",
		"@stdlib/iax",
		"@stdlib/interaction",
	} {
		if _, ok := modules[name]; !ok {
			t.Fatalf("missing native module: %s", name)
		}
	}
}

func TestSSEModuleSyncAndAsync(t *testing.T) {
	modules := DefaultModules(rt.NewGoroutineRuntime())
	sseMod := mustModuleObject(t, modules, "sse")
	clientNS := mustObject(t, sseMod, "client")
	serverNS := mustObject(t, sseMod, "server")

	serverValue := callNative(t, serverNS, "serve", Object{
		"addr": "127.0.0.1:0",
		"path": "/events",
	})
	server := mustRuntimeObject(t, serverValue, "sse.server.serve return")
	serverURL, ok := server["url"].(string)
	if !ok || serverURL == "" {
		t.Fatalf("sse.server.serve url = %#v, want non-empty string", server["url"])
	}

	clientValue := callNative(t, clientNS, "connect", serverURL)
	client := mustRuntimeObject(t, clientValue, "sse.client.connect return")

	delivered := callNative(t, server, "send", "hello-sse", "message")
	if n, ok := delivered.(float64); !ok || n < 1 {
		t.Fatalf("sse.server.send delivered = %#v, want >= 1", delivered)
	}
	eventValue := callNative(t, client, "recv")
	eventObj := mustRuntimeObject(t, eventValue, "sse.client.recv return")
	if s, ok := eventObj["data"].(string); !ok || s != "hello-sse" {
		t.Fatalf("sse.client.recv data = %#v, want hello-sse", eventObj["data"])
	}
	if s, ok := eventObj["event"].(string); !ok || s != "message" {
		t.Fatalf("sse.client.recv event = %#v, want message", eventObj["event"])
	}

	_ = awaitValue(t, callNative(t, server, "sendAsync", "hello-async", "async"))
	asyncEventValue := awaitValue(t, callNative(t, client, "recvAsync"))
	asyncEventObj := mustRuntimeObject(t, asyncEventValue, "sse.client.recvAsync return")
	if s, ok := asyncEventObj["data"].(string); !ok || s != "hello-async" {
		t.Fatalf("sse.client.recvAsync data = %#v, want hello-async", asyncEventObj["data"])
	}

	_ = callNative(t, client, "close")
	_ = callNative(t, server, "close")

	serverValueAsync := awaitValue(t, callNative(t, serverNS, "serveAsync", Object{
		"addr": "127.0.0.1:0",
		"path": "/events-async",
	}))
	serverAsync := mustRuntimeObject(t, serverValueAsync, "sse.server.serveAsync return")
	asyncURL, ok := serverAsync["url"].(string)
	if !ok || asyncURL == "" {
		t.Fatalf("sse.server.serveAsync url = %#v, want non-empty string", serverAsync["url"])
	}
	clientValueAsync := awaitValue(t, callNative(t, clientNS, "connectAsync", asyncURL))
	clientAsync := mustRuntimeObject(t, clientValueAsync, "sse.client.connectAsync return")
	_ = callNative(t, serverAsync, "send", "from-connectAsync", "message")
	gotValue := callNative(t, clientAsync, "recv")
	gotObj := mustRuntimeObject(t, gotValue, "sse connectAsync recv return")
	if s, ok := gotObj["data"].(string); !ok || s != "from-connectAsync" {
		t.Fatalf("sse connectAsync recv data = %#v, want from-connectAsync", gotObj["data"])
	}
	_ = callNative(t, clientAsync, "close")
	_ = callNative(t, serverAsync, "close")
}

func TestWebSocketModuleSyncAndAsync(t *testing.T) {
	modules := DefaultModules(rt.NewGoroutineRuntime())
	wsMod := mustModuleObject(t, modules, "websocket")
	clientNS := mustObject(t, wsMod, "client")
	serverNS := mustObject(t, wsMod, "server")

	serverValue := callNative(t, serverNS, "serve", Object{
		"addr": "127.0.0.1:0",
		"path": "/ws",
		"echo": true,
	})
	server := mustRuntimeObject(t, serverValue, "websocket.server.serve return")
	serverURL, ok := server["url"].(string)
	if !ok || serverURL == "" {
		t.Fatalf("websocket.server.serve url = %#v, want non-empty string", server["url"])
	}

	clientValue := callNative(t, clientNS, "connect", serverURL)
	client := mustRuntimeObject(t, clientValue, "websocket.client.connect return")
	_ = callNative(t, client, "send", "hello-websocket")
	echoed := callNative(t, client, "recv")
	if s, ok := echoed.(string); !ok || s != "hello-websocket" {
		t.Fatalf("websocket roundtrip recv = %#v, want hello-websocket", echoed)
	}

	_ = awaitValue(t, callNative(t, client, "sendAsync", "hello-async"))
	asyncEchoed := awaitValue(t, callNative(t, client, "recvAsync"))
	if s, ok := asyncEchoed.(string); !ok || s != "hello-async" {
		t.Fatalf("websocket async roundtrip recv = %#v, want hello-async", asyncEchoed)
	}
	_ = callNative(t, client, "close")
	_ = callNative(t, server, "close")

	serverValueAsync := awaitValue(t, callNative(t, serverNS, "serveAsync", Object{
		"addr": "127.0.0.1:0",
		"path": "/async",
	}))
	serverAsync := mustRuntimeObject(t, serverValueAsync, "websocket.server.serveAsync return")
	asyncURL, ok := serverAsync["url"].(string)
	if !ok || asyncURL == "" {
		t.Fatalf("websocket.server.serveAsync url = %#v, want non-empty string", serverAsync["url"])
	}
	clientValueAsync := awaitValue(t, callNative(t, clientNS, "connectAsync", asyncURL))
	clientAsync := mustRuntimeObject(t, clientValueAsync, "websocket.client.connectAsync return")
	_ = callNative(t, clientAsync, "send", "from-connectAsync")
	echoedAsync := callNative(t, clientAsync, "recv")
	if s, ok := echoedAsync.(string); !ok || s != "from-connectAsync" {
		t.Fatalf("websocket connectAsync recv = %#v, want from-connectAsync", echoedAsync)
	}
	_ = callNative(t, clientAsync, "close")
	_ = callNative(t, serverAsync, "close")
}

func TestFSModuleSyncAndAsync(t *testing.T) {
	modules := DefaultModules(rt.NewGoroutineRuntime())
	fsMod := mustModuleObject(t, modules, "fs")

	tmp := t.TempDir()
	file := filepath.Join(tmp, "native_fs_test.txt")

	_ = callNative(t, fsMod, "writeFile", file, "hello")
	got := callNative(t, fsMod, "readFile", file)
	if gs, ok := got.(string); !ok || gs != "hello" {
		t.Fatalf("fs.readFile = %#v, want hello", got)
	}

	exists := callNative(t, fsMod, "exists", file)
	if b, ok := exists.(bool); !ok || !b {
		t.Fatalf("fs.exists = %#v, want true", exists)
	}

	stat := callNative(t, fsMod, "stat", file)
	statObj, ok := stat.(Object)
	if !ok {
		t.Fatalf("fs.stat type = %T, want Object", stat)
	}
	if size, ok := statObj["size"].(float64); !ok || size <= 0 {
		t.Fatalf("fs.stat.size = %#v, want > 0", statObj["size"])
	}

	_ = awaitValue(t, callNative(t, fsMod, "writeFileAsync", file, "world"))
	asyncRead := awaitValue(t, callNative(t, fsMod, "readFileAsync", file))
	if gs, ok := asyncRead.(string); !ok || gs != "world" {
		t.Fatalf("fs.readFileAsync result = %#v, want world", asyncRead)
	}

	dirEntries := callNative(t, fsMod, "readDir", tmp)
	arr, ok := dirEntries.(Array)
	if !ok {
		t.Fatalf("fs.readDir type = %T, want Array", dirEntries)
	}
	found := false
	for _, v := range arr {
		if s, ok := v.(string); ok && s == "native_fs_test.txt" {
			found = true
			break
		}
	}
	if !found {
		t.Fatalf("fs.readDir does not include target file: %v", arr)
	}

	// namespace export
	fsNS, ok := fsMod["fs"].(Object)
	if !ok {
		t.Fatalf("fs namespace export type = %T, want Object", fsMod["fs"])
	}
	gotNS := callNative(t, fsNS, "readFile", file)
	if gs, ok := gotNS.(string); !ok || gs != "world" {
		t.Fatalf("fs namespace readFile = %#v, want world", gotNS)
	}
}

func TestOSAndProcessModules(t *testing.T) {
	modules := DefaultModules(rt.NewGoroutineRuntime())
	osMod := mustModuleObject(t, modules, "os")
	processMod := mustModuleObject(t, modules, "process")

	platform := callNative(t, osMod, "platform")
	if s, ok := platform.(string); !ok || s == "" {
		t.Fatalf("os.platform = %#v, want non-empty string", platform)
	}
	arch := callNative(t, osMod, "arch")
	if s, ok := arch.(string); !ok || s == "" {
		t.Fatalf("os.arch = %#v, want non-empty string", arch)
	}
	userDir := callNative(t, osMod, "userDir")
	userDirStr, ok := userDir.(string)
	if !ok || userDirStr == "" || !filepath.IsAbs(userDirStr) {
		t.Fatalf("os.userDir = %#v, want non-empty absolute path", userDir)
	}
	dataDir := callNative(t, osMod, "dataDir")
	dataDirStr, ok := dataDir.(string)
	if !ok || dataDirStr == "" || !filepath.IsAbs(dataDirStr) {
		t.Fatalf("os.dataDir = %#v, want non-empty absolute path", dataDir)
	}
	tmpDir := callNative(t, osMod, "tmpDir")
	tempDir := callNative(t, osMod, "tempDir")
	tmpDirStr, okTmp := tmpDir.(string)
	tempDirStr, okTemp := tempDir.(string)
	if !okTmp || !okTemp || tmpDirStr == "" || tempDirStr == "" {
		t.Fatalf("os.tmpDir/os.tempDir = %#v / %#v, want non-empty strings", tmpDir, tempDir)
	}
	if filepath.Clean(tmpDirStr) != filepath.Clean(tempDirStr) {
		t.Fatalf("os.tmpDir != os.tempDir: %q vs %q", tmpDirStr, tempDirStr)
	}
	configDir := callNative(t, osMod, "configDir")
	if s, ok := configDir.(string); !ok || s == "" || !filepath.IsAbs(s) {
		t.Fatalf("os.configDir = %#v, want non-empty absolute path", configDir)
	}
	cacheDir := callNative(t, osMod, "cacheDir")
	if s, ok := cacheDir.(string); !ok || s == "" || !filepath.IsAbs(s) {
		t.Fatalf("os.cacheDir = %#v, want non-empty absolute path", cacheDir)
	}

	key := "IALANG_NATIVE_TEST_" + time.Now().Format("150405")
	_ = callNative(t, osMod, "setEnv", key, "ok")
	getEnv := callNative(t, osMod, "getEnv", key)
	if s, ok := getEnv.(string); !ok || s != "ok" {
		t.Fatalf("os.getEnv = %#v, want ok", getEnv)
	}

	pid := callNative(t, processMod, "pid")
	if n, ok := pid.(float64); !ok || n <= 0 {
		t.Fatalf("process.pid = %#v, want > 0", pid)
	}
	args := callNative(t, processMod, "args")
	argArr, ok := args.(Array)
	if !ok || len(argArr) == 0 {
		t.Fatalf("process.args = %#v, want non-empty Array", args)
	}
	if _, err := callNativeWithError(processMod, "exit", float64(0)); err == nil {
		t.Fatal("process.exit expected disabled error, got nil")
	}
}

func TestHTTPModuleSyncAndAsync(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch {
		case r.Method == http.MethodGet && r.URL.Path == "/hello":
			w.WriteHeader(http.StatusOK)
			_, _ = w.Write([]byte("world"))
		case r.Method == http.MethodGet && r.URL.Path == "/stream":
			flusher, ok := w.(http.Flusher)
			if !ok {
				w.WriteHeader(http.StatusInternalServerError)
				return
			}
			w.Header().Set("Content-Type", "text/plain; charset=utf-8")
			w.WriteHeader(http.StatusOK)
			for _, chunk := range []string{"a", "b", "c"} {
				_, _ = w.Write([]byte(chunk))
				flusher.Flush()
			}
		case r.Method == http.MethodPost && r.URL.Path == "/echo":
			raw, _ := io.ReadAll(r.Body)
			w.WriteHeader(http.StatusCreated)
			_, _ = w.Write([]byte("echo:" + string(raw)))
		default:
			w.WriteHeader(http.StatusNotFound)
		}
	}))
	defer srv.Close()

	modules := DefaultModules(rt.NewGoroutineRuntime())
	httpMod := mustModuleObject(t, modules, "http")
	clientNS := mustObject(t, httpMod, "client")
	serverNS := mustObject(t, httpMod, "server")

	respGet := callNative(t, clientNS, "get", srv.URL+"/hello")
	objGet, ok := respGet.(Object)
	if !ok {
		t.Fatalf("http.client.get response type = %T, want Object", respGet)
	}
	if code, ok := objGet["statusCode"].(float64); !ok || int(code) != http.StatusOK {
		t.Fatalf("http.client.get statusCode = %#v, want 200", objGet["statusCode"])
	}
	if body, ok := objGet["body"].(string); !ok || body != "world" {
		t.Fatalf("http.client.get body = %#v, want world", objGet["body"])
	}

	respPost := callNative(t, clientNS, "post", srv.URL+"/echo", Object{
		"body": "abc",
	})
	objPost, ok := respPost.(Object)
	if !ok {
		t.Fatalf("http.client.post response type = %T, want Object", respPost)
	}
	if code, ok := objPost["statusCode"].(float64); !ok || int(code) != http.StatusCreated {
		t.Fatalf("http.client.post statusCode = %#v, want 201", objPost["statusCode"])
	}
	if body, ok := objPost["body"].(string); !ok || body != "echo:abc" {
		t.Fatalf("http.client.post body = %#v, want echo:abc", objPost["body"])
	}

	respReq := callNative(t, clientNS, "request", srv.URL+"/echo", Object{
		"method":  "POST",
		"body":    "xyz",
		"headers": Object{"X-Test": "1"},
	})
	objReq, ok := respReq.(Object)
	if !ok {
		t.Fatalf("http.client.request response type = %T, want Object", respReq)
	}
	if code, ok := objReq["statusCode"].(float64); !ok || int(code) != http.StatusCreated {
		t.Fatalf("http.client.request statusCode = %#v, want 201", objReq["statusCode"])
	}
	if body, ok := objReq["body"].(string); !ok || body != "echo:xyz" {
		t.Fatalf("http.client.request body = %#v, want echo:xyz", objReq["body"])
	}

	respGetAsync := awaitValue(t, callNative(t, clientNS, "getAsync", srv.URL+"/hello"))
	objAsync, ok := respGetAsync.(Object)
	if !ok {
		t.Fatalf("http.client.getAsync response type = %T, want Object", respGetAsync)
	}
	if body, ok := objAsync["body"].(string); !ok || body != "world" {
		t.Fatalf("http.client.getAsync body = %#v, want world", objAsync["body"])
	}

	streamResp := callNative(t, clientNS, "stream", srv.URL+"/stream", Object{
		"chunkSize": float64(1),
	})
	streamObj := mustRuntimeObject(t, streamResp, "http.client.stream return")
	if code, ok := streamObj["statusCode"].(float64); !ok || int(code) != http.StatusOK {
		t.Fatalf("http.client.stream statusCode = %#v, want 200", streamObj["statusCode"])
	}
	p1 := mustRuntimeObject(t, callNative(t, streamObj, "recv"), "http.stream.recv #1")
	p2 := mustRuntimeObject(t, awaitValue(t, callNative(t, streamObj, "recvAsync")), "http.stream.recvAsync #2")
	p3 := mustRuntimeObject(t, callNative(t, streamObj, "recv"), "http.stream.recv #3")
	if s, _ := p1["chunk"].(string); s != "a" {
		t.Fatalf("http.stream chunk1 = %#v, want a", p1["chunk"])
	}
	if s, _ := p2["chunk"].(string); s != "b" {
		t.Fatalf("http.stream chunk2 = %#v, want b", p2["chunk"])
	}
	if s, _ := p3["chunk"].(string); s != "c" {
		t.Fatalf("http.stream chunk3 = %#v, want c", p3["chunk"])
	}
	done := mustRuntimeObject(t, callNative(t, streamObj, "recv"), "http.stream.recv done")
	if d, ok := done["done"].(bool); !ok || !d {
		t.Fatalf("http.stream done = %#v, want true", done["done"])
	}
	_ = callNative(t, streamObj, "close")

	localServerValue := callNative(t, serverNS, "serve", Object{
		"addr":       "127.0.0.1:0",
		"statusCode": float64(http.StatusCreated),
		"body":       "from-server",
	})
	localServer := mustRuntimeObject(t, localServerValue, "http.server.serve return")
	addr, ok := localServer["addr"].(string)
	if !ok || addr == "" {
		t.Fatalf("http.server.serve addr = %#v, want non-empty string", localServer["addr"])
	}
	resp, err := http.Get("http://" + addr + "/")
	if err != nil {
		t.Fatalf("probe http.server.serve failed: %v", err)
	}
	raw, err := io.ReadAll(resp.Body)
	_ = resp.Body.Close()
	if err != nil {
		t.Fatalf("probe http.server.serve read body failed: %v", err)
	}
	if resp.StatusCode != http.StatusCreated {
		t.Fatalf("http.server.serve status = %d, want 201", resp.StatusCode)
	}
	if string(raw) != "from-server" {
		t.Fatalf("http.server.serve body = %q, want from-server", string(raw))
	}
	_ = callNative(t, localServer, "close")
}

func TestAdditionalNativeModules(t *testing.T) {
	modules := DefaultModules(rt.NewGoroutineRuntime())

	execMod := mustModuleObject(t, modules, "exec")
	logMod := mustModuleObject(t, modules, "log")
	pathMod := mustModuleObject(t, modules, "path")
	jsonMod := mustModuleObject(t, modules, "json")
	timeMod := mustModuleObject(t, modules, "time")
	encodingMod := mustModuleObject(t, modules, "encoding")
	cryptoMod := mustModuleObject(t, modules, "crypto")
	regexpMod := mustModuleObject(t, modules, "regexp")
	uuidMod := mustModuleObject(t, modules, "uuid")
	urlMod := mustModuleObject(t, modules, "url")
	strconvMod := mustModuleObject(t, modules, "strconv")
	randMod := mustModuleObject(t, modules, "rand")
	csvMod := mustModuleObject(t, modules, "csv")
	xmlMod := mustModuleObject(t, modules, "xml")
	hexMod := mustModuleObject(t, modules, "hex")
	netMod := mustModuleObject(t, modules, "net")
	mimeMod := mustModuleObject(t, modules, "mime")
	hashMod := mustModuleObject(t, modules, "hash")
	compressMod := mustModuleObject(t, modules, "compress")
	hmacMod := mustModuleObject(t, modules, "hmac")
	bytesMod := mustModuleObject(t, modules, "bytes")
	sortMod := mustModuleObject(t, modules, "sort")
	setMod := mustModuleObject(t, modules, "set")

	joined := callNative(t, pathMod, "join", "a", "b", "c")
	if s, ok := joined.(string); !ok || s != filepath.Join("a", "b", "c") {
		t.Fatalf("path.join = %#v, want %q", joined, filepath.Join("a", "b", "c"))
	}
	base := callNative(t, pathMod, "base", filepath.Join("a", "b.txt"))
	if s, ok := base.(string); !ok || s != "b.txt" {
		t.Fatalf("path.base = %#v, want b.txt", base)
	}

	parsed := callNative(t, jsonMod, "parse", `{"a":1,"b":[true,"x"]}`)
	parsedObj, ok := parsed.(Object)
	if !ok {
		t.Fatalf("json.parse type = %T, want Object", parsed)
	}
	if n, ok := parsedObj["a"].(float64); !ok || n != 1 {
		t.Fatalf("json.parse.a = %#v, want 1", parsedObj["a"])
	}
	stringified := callNative(t, jsonMod, "stringify", parsedObj)
	if s, ok := stringified.(string); !ok || !strings.Contains(s, `"a":1`) {
		t.Fatalf("json.stringify = %#v, want contains \"a\":1", stringified)
	}
	valid := callNative(t, jsonMod, "valid", `{"x":1}`)
	if b, ok := valid.(bool); !ok || !b {
		t.Fatalf("json.valid = %#v, want true", valid)
	}

	nowUnix := callNative(t, timeMod, "nowUnix")
	if n, ok := nowUnix.(float64); !ok || n <= 0 {
		t.Fatalf("time.nowUnix = %#v, want > 0", nowUnix)
	}
	nowISO := callNative(t, timeMod, "nowISO")
	isoText, ok := nowISO.(string)
	if !ok || isoText == "" {
		t.Fatalf("time.nowISO = %#v, want non-empty string", nowISO)
	}
	if _, err := time.Parse(time.RFC3339Nano, isoText); err != nil {
		t.Fatalf("time.nowISO parse failed: %v", err)
	}
	_ = callNative(t, timeMod, "sleep", float64(1))
	_ = awaitValue(t, callNative(t, timeMod, "sleepAsync", float64(1)))

	encoded := callNative(t, encodingMod, "base64Encode", "abc")
	decoded := callNative(t, encodingMod, "base64Decode", encoded)
	if s, ok := decoded.(string); !ok || s != "abc" {
		t.Fatalf("encoding base64 roundtrip = %#v, want abc", decoded)
	}
	urlEncoded := callNative(t, encodingMod, "urlEncode", "a b")
	urlDecoded := callNative(t, encodingMod, "urlDecode", urlEncoded)
	if s, ok := urlDecoded.(string); !ok || s != "a b" {
		t.Fatalf("encoding url roundtrip = %#v, want 'a b'", urlDecoded)
	}

	sha256Value := callNative(t, cryptoMod, "sha256", "abc")
	if s, ok := sha256Value.(string); !ok || s != "ba7816bf8f01cfea414140de5dae2223b00361a396177a9cb410ff61f20015ad" {
		t.Fatalf("crypto.sha256 = %#v, want known hash", sha256Value)
	}
	md5Value := callNative(t, cryptoMod, "md5", "abc")
	if s, ok := md5Value.(string); !ok || s != "900150983cd24fb0d6963f7d28e17f72" {
		t.Fatalf("crypto.md5 = %#v, want known hash", md5Value)
	}

	match := callNative(t, regexpMod, "test", "h.llo", "hello")
	if b, ok := match.(bool); !ok || !b {
		t.Fatalf("regexp.test = %#v, want true", match)
	}
	found := callNative(t, regexpMod, "find", "\\d+", "abc123xyz")
	if s, ok := found.(string); !ok || s != "123" {
		t.Fatalf("regexp.find = %#v, want 123", found)
	}
	replaced := callNative(t, regexpMod, "replaceAll", "\\s+", "a  b   c", "-")
	if s, ok := replaced.(string); !ok || s != "a-b-c" {
		t.Fatalf("regexp.replaceAll = %#v, want a-b-c", replaced)
	}
	split := callNative(t, regexpMod, "split", "\\s+", "a b  c")
	splitArr, ok := split.(Array)
	if !ok || len(splitArr) != 3 {
		t.Fatalf("regexp.split = %#v, want len 3", split)
	}

	id := callNative(t, uuidMod, "v4")
	idText, ok := id.(string)
	if !ok || idText == "" {
		t.Fatalf("uuid.v4 = %#v, want non-empty string", id)
	}
	uuidValid := callNative(t, uuidMod, "isValid", idText)
	if b, ok := uuidValid.(bool); !ok || !b {
		t.Fatalf("uuid.isValid(v4) = %#v, want true", uuidValid)
	}
	invalid := callNative(t, uuidMod, "isValid", "not-a-uuid")
	if b, ok := invalid.(bool); !ok || b {
		t.Fatalf("uuid.isValid(invalid) = %#v, want false", invalid)
	}

	urlParsed := callNative(t, urlMod, "parse", "https://example.com/a/b?q=1#frag")
	urlObj, ok := urlParsed.(Object)
	if !ok {
		t.Fatalf("url.parse type = %T, want Object", urlParsed)
	}
	if s, ok := urlObj["host"].(string); !ok || s != "example.com" {
		t.Fatalf("url.parse.host = %#v, want example.com", urlObj["host"])
	}
	escaped := callNative(t, urlMod, "escape", "a b")
	unescaped := callNative(t, urlMod, "unescape", escaped)
	if s, ok := unescaped.(string); !ok || s != "a b" {
		t.Fatalf("url escape/unescape roundtrip = %#v, want 'a b'", unescaped)
	}
	query := callNative(t, urlMod, "queryEncode", Object{"a": "1", "b": "x y"})
	queryObj := callNative(t, urlMod, "queryDecode", query)
	queryMap, ok := queryObj.(Object)
	if !ok {
		t.Fatalf("url.queryDecode type = %T, want Object", queryObj)
	}
	if s, ok := queryMap["a"].(string); !ok || s != "1" {
		t.Fatalf("url.queryDecode.a = %#v, want 1", queryMap["a"])
	}

	atoi := callNative(t, strconvMod, "atoi", "42")
	if n, ok := atoi.(float64); !ok || n != 42 {
		t.Fatalf("strconv.atoi = %#v, want 42", atoi)
	}
	itoa := callNative(t, strconvMod, "itoa", float64(42))
	if s, ok := itoa.(string); !ok || s != "42" {
		t.Fatalf("strconv.itoa = %#v, want 42", itoa)
	}
	parsedFloat := callNative(t, strconvMod, "parseFloat", "3.5")
	if n, ok := parsedFloat.(float64); !ok || n != 3.5 {
		t.Fatalf("strconv.parseFloat = %#v, want 3.5", parsedFloat)
	}
	formattedFloat := callNative(t, strconvMod, "formatFloat", float64(3.5), float64(2))
	if s, ok := formattedFloat.(string); !ok || s != "3.50" {
		t.Fatalf("strconv.formatFloat = %#v, want 3.50", formattedFloat)
	}
	parsedBool := callNative(t, strconvMod, "parseBool", "true")
	if b, ok := parsedBool.(bool); !ok || !b {
		t.Fatalf("strconv.parseBool = %#v, want true", parsedBool)
	}
	formattedBool := callNative(t, strconvMod, "formatBool", true)
	if s, ok := formattedBool.(string); !ok || s != "true" {
		t.Fatalf("strconv.formatBool = %#v, want true", formattedBool)
	}

	randInt := callNative(t, randMod, "int", float64(10))
	if n, ok := randInt.(float64); !ok || n < 0 || n >= 10 {
		t.Fatalf("rand.int(10) = %#v, want in [0,10)", randInt)
	}
	randFloat := callNative(t, randMod, "float", float64(1), float64(2))
	if n, ok := randFloat.(float64); !ok || n < 1 || n >= 2 {
		t.Fatalf("rand.float(1,2) = %#v, want in [1,2)", randFloat)
	}
	picked := callNative(t, randMod, "pick", Array{"x", "y", "z"})
	if s, ok := picked.(string); !ok || (s != "x" && s != "y" && s != "z") {
		t.Fatalf("rand.pick = %#v, want one of x/y/z", picked)
	}
	randText := callNative(t, randMod, "string", float64(8), "ab")
	if s, ok := randText.(string); !ok || len(s) != 8 {
		t.Fatalf("rand.string = %#v, want len 8", randText)
	}

	csvText := callNative(t, csvMod, "stringify", Array{
		Array{"name", "age"},
		Array{"alice", float64(20)},
	})
	csvParsed := callNative(t, csvMod, "parse", csvText)
	csvRows, ok := csvParsed.(Array)
	if !ok || len(csvRows) != 2 {
		t.Fatalf("csv.parse = %#v, want 2 rows", csvParsed)
	}
	secondRow, ok := csvRows[1].(Array)
	if !ok || len(secondRow) != 2 {
		t.Fatalf("csv second row = %#v, want len 2", csvRows[1])
	}
	if s, ok := secondRow[0].(string); !ok || s != "alice" {
		t.Fatalf("csv second row first col = %#v, want alice", secondRow[0])
	}

	xmlParsed := callNative(t, xmlMod, "parse", `<root id="1"><item>ok</item></root>`)
	xmlObj := mustRuntimeObject(t, xmlParsed, "xml.parse return")
	if s, ok := xmlObj["name"].(string); !ok || s != "root" {
		t.Fatalf("xml.parse name = %#v, want root", xmlObj["name"])
	}
	xmlChildren, ok := xmlObj["children"].(Array)
	if !ok || len(xmlChildren) != 1 {
		t.Fatalf("xml.parse children = %#v, want len 1", xmlObj["children"])
	}
	xmlTextNode := mustRuntimeObject(t, xmlChildren[0], "xml child")
	if s, ok := xmlTextNode["text"].(string); !ok || s != "ok" {
		t.Fatalf("xml child text = %#v, want ok", xmlTextNode["text"])
	}
	xmlStr := callNative(t, xmlMod, "stringify", xmlObj)
	if s, ok := xmlStr.(string); !ok || !strings.Contains(s, "<root") || !strings.Contains(s, "<item>ok</item>") {
		t.Fatalf("xml.stringify = %#v, want root/item xml", xmlStr)
	}
	xmlValid := callNative(t, xmlMod, "valid", `<a><b/></a>`)
	if b, ok := xmlValid.(bool); !ok || !b {
		t.Fatalf("xml.valid = %#v, want true", xmlValid)
	}

	hexEncoded := callNative(t, hexMod, "encode", "Hi")
	if s, ok := hexEncoded.(string); !ok || s != "4869" {
		t.Fatalf("hex.encode = %#v, want 4869", hexEncoded)
	}
	hexDecoded := callNative(t, hexMod, "decode", hexEncoded)
	if s, ok := hexDecoded.(string); !ok || s != "Hi" {
		t.Fatalf("hex.decode = %#v, want Hi", hexDecoded)
	}
	hexBytes := callNative(t, hexMod, "decodeBytes", "4869")
	byteArr, ok := hexBytes.(Array)
	if !ok || len(byteArr) != 2 {
		t.Fatalf("hex.decodeBytes = %#v, want len 2", hexBytes)
	}

	isIP := callNative(t, netMod, "isIP", "127.0.0.1")
	if b, ok := isIP.(bool); !ok || !b {
		t.Fatalf("net.isIP = %#v, want true", isIP)
	}
	joinedHostPort := callNative(t, netMod, "joinHostPort", "127.0.0.1", float64(8080))
	parsedHostPort := callNative(t, netMod, "parseHostPort", joinedHostPort)
	hostPortObj := mustRuntimeObject(t, parsedHostPort, "net.parseHostPort")
	if s, ok := hostPortObj["host"].(string); !ok || s != "127.0.0.1" {
		t.Fatalf("net.parseHostPort.host = %#v, want 127.0.0.1", hostPortObj["host"])
	}
	containsCIDR := callNative(t, netMod, "containsCIDR", "10.0.0.0/8", "10.2.3.4")
	if b, ok := containsCIDR.(bool); !ok || !b {
		t.Fatalf("net.containsCIDR = %#v, want true", containsCIDR)
	}

	mimeByExt := callNative(t, mimeMod, "typeByExt", ".json")
	if s, ok := mimeByExt.(string); !ok || !strings.Contains(s, "application/json") {
		t.Fatalf("mime.typeByExt = %#v, want application/json", mimeByExt)
	}
	detected := callNative(t, mimeMod, "detectByPath", "a.txt")
	if s, ok := detected.(string); !ok || !strings.Contains(s, "text/plain") {
		t.Fatalf("mime.detectByPath = %#v, want text/plain", detected)
	}

	sha1Value := callNative(t, hashMod, "sha1", "abc")
	if s, ok := sha1Value.(string); !ok || s != "a9993e364706816aba3e25717850c26c9cd0d89d" {
		t.Fatalf("hash.sha1 = %#v, want known value", sha1Value)
	}
	crcValue := callNative(t, hashMod, "crc32", "abc")
	if n, ok := crcValue.(float64); !ok || n <= 0 {
		t.Fatalf("hash.crc32 = %#v, want positive", crcValue)
	}

	if level, ok := compressMod["bestCompression"].(float64); !ok || level != 9 {
		t.Fatalf("compress.bestCompression = %#v, want 9", compressMod["bestCompression"])
	}
	gzipCompressed := callNative(t, compressMod, "gzipCompress", "hello compress", float64(9))
	gzipRaw := callNative(t, compressMod, "gzipDecompress", gzipCompressed)
	if s, ok := gzipRaw.(string); !ok || s != "hello compress" {
		t.Fatalf("compress gzip roundtrip = %#v, want hello compress", gzipRaw)
	}
	gzipAliasRaw := callNative(t, compressMod, "gunzip", callNative(t, compressMod, "gzip", "hello alias"))
	if s, ok := gzipAliasRaw.(string); !ok || s != "hello alias" {
		t.Fatalf("compress gzip alias roundtrip = %#v, want hello alias", gzipAliasRaw)
	}
	zlibCompressed := callNative(t, compressMod, "zlibCompress", "hello zlib", float64(1))
	zlibRaw := callNative(t, compressMod, "zlibDecompress", zlibCompressed)
	if s, ok := zlibRaw.(string); !ok || s != "hello zlib" {
		t.Fatalf("compress zlib roundtrip = %#v, want hello zlib", zlibRaw)
	}
	zlibAliasRaw := callNative(t, compressMod, "inflate", callNative(t, compressMod, "deflate", "hello deflate"))
	if s, ok := zlibAliasRaw.(string); !ok || s != "hello deflate" {
		t.Fatalf("compress zlib alias roundtrip = %#v, want hello deflate", zlibAliasRaw)
	}
	compressedBytes := callNative(t, compressMod, "gzipCompressBytes", Array{float64(0), float64(1), float64(2), float64(255)}, float64(1))
	decompressedBytes := callNative(t, compressMod, "gzipDecompressBytes", compressedBytes)
	byteValues, ok := decompressedBytes.(Array)
	if !ok || len(byteValues) != 4 {
		t.Fatalf("compress gzip bytes roundtrip = %#v, want len 4", decompressedBytes)
	}
	for i, want := range []float64{0, 1, 2, 255} {
		if byteValues[i] != want {
			t.Fatalf("compress gzip bytes[%d] = %#v, want %v", i, byteValues[i], want)
		}
	}
	if _, err := callNativeWithError(compressMod, "gzipCompress", "bad-level", float64(10)); err == nil || !strings.Contains(err.Error(), "[-2, 9]") {
		t.Fatalf("compress invalid level error = %v, want range validation", err)
	}

	hmacSig := callNative(t, hmacMod, "sha256", "k", "payload")
	okVerify := callNative(t, hmacMod, "verifySha256", "k", "payload", hmacSig)
	if b, ok := okVerify.(bool); !ok || !b {
		t.Fatalf("hmac.verifySha256(valid) = %#v, want true", okVerify)
	}
	badVerify := callNative(t, hmacMod, "verifySha256", "k", "payload-x", hmacSig)
	if b, ok := badVerify.(bool); !ok || b {
		t.Fatalf("hmac.verifySha256(invalid) = %#v, want false", badVerify)
	}

	byteArrVal := callNative(t, bytesMod, "fromString", "Hi")
	byteArrObj, ok := byteArrVal.(Array)
	if !ok || len(byteArrObj) != 2 {
		t.Fatalf("bytes.fromString = %#v, want len 2", byteArrVal)
	}
	b64 := callNative(t, bytesMod, "toBase64", byteArrObj)
	decodedArr := callNative(t, bytesMod, "fromBase64", b64)
	backText := callNative(t, bytesMod, "toString", decodedArr)
	if s, ok := backText.(string); !ok || s != "Hi" {
		t.Fatalf("bytes base64 roundtrip = %#v, want Hi", backText)
	}
	sliced := callNative(t, bytesMod, "slice", byteArrObj, float64(1))
	slicedText := callNative(t, bytesMod, "toString", sliced)
	if s, ok := slicedText.(string); !ok || s != "i" {
		t.Fatalf("bytes.slice = %#v, want i", slicedText)
	}

	sortedAsc := callNative(t, sortMod, "asc", Array{float64(3), float64(1), float64(2)})
	sortedAscArr, ok := sortedAsc.(Array)
	if !ok || len(sortedAscArr) != 3 || sortedAscArr[0].(float64) != 1 || sortedAscArr[2].(float64) != 3 {
		t.Fatalf("sort.asc = %#v, want [1,2,3]", sortedAsc)
	}
	reversed := callNative(t, sortMod, "reverse", Array{"a", "b", "c"})
	reversedArr, ok := reversed.(Array)
	if !ok || len(reversedArr) != 3 || reversedArr[0].(string) != "c" {
		t.Fatalf("sort.reverse = %#v, want starts with c", reversed)
	}

	union := callNative(t, setMod, "union", Array{"a", "b"}, Array{"b", "c"})
	unionArr, ok := union.(Array)
	if !ok || len(unionArr) != 3 {
		t.Fatalf("set.union = %#v, want len 3", union)
	}
	inter := callNative(t, setMod, "intersect", Array{"a", "b"}, Array{"b", "c"})
	interArr, ok := inter.(Array)
	if !ok || len(interArr) != 1 || interArr[0].(string) != "b" {
		t.Fatalf("set.intersect = %#v, want [b]", inter)
	}

	command := "sh"
	args := Array{"-c", "echo hello"}
	if goRuntime.GOOS == "windows" {
		command = "cmd"
		args = Array{"/C", "echo hello"}
	}
	runRes := callNative(t, execMod, "run", command, Object{
		"args": args,
	})
	runObj := mustRuntimeObject(t, runRes, "exec.run return")
	if ok, _ := runObj["ok"].(bool); !ok {
		t.Fatalf("exec.run ok = %#v, want true", runObj["ok"])
	}
	if code, ok := runObj["code"].(float64); !ok || int(code) != 0 {
		t.Fatalf("exec.run code = %#v, want 0", runObj["code"])
	}
	if out, ok := runObj["stdout"].(string); !ok || !strings.Contains(strings.ToLower(out), "hello") {
		t.Fatalf("exec.run stdout = %#v, want contains hello", runObj["stdout"])
	}

	runAsyncRes := awaitValue(t, callNative(t, execMod, "runAsync", command, Object{
		"args": args,
	}))
	runAsyncObj := mustRuntimeObject(t, runAsyncRes, "exec.runAsync return")
	if ok, _ := runAsyncObj["ok"].(bool); !ok {
		t.Fatalf("exec.runAsync ok = %#v, want true", runAsyncObj["ok"])
	}

	startRes := callNative(t, execMod, "start", command, Object{
		"args": args,
	})
	startObj := mustRuntimeObject(t, startRes, "exec.start return")
	if pid, ok := callNative(t, startObj, "pid").(float64); !ok || pid <= 0 {
		t.Fatalf("exec.start pid = %#v, want > 0", pid)
	}
	startWaitObj := mustRuntimeObject(t, callNative(t, startObj, "wait"), "exec.start wait return")
	if ok, _ := startWaitObj["ok"].(bool); !ok {
		t.Fatalf("exec.start wait ok = %#v, want true", startWaitObj["ok"])
	}
	if out, ok := startWaitObj["stdout"].(string); !ok || !strings.Contains(strings.ToLower(out), "hello") {
		t.Fatalf("exec.start wait stdout = %#v, want contains hello", startWaitObj["stdout"])
	}

	longRunningLine := "sleep 5"
	if goRuntime.GOOS == "windows" {
		longRunningLine = "ping -n 6 127.0.0.1 > NUL"
	}
	longRunningRes := callNative(t, execMod, "start", longRunningLine, Object{
		"shell": true,
	})
	longRunningObj := mustRuntimeObject(t, longRunningRes, "exec.start long-running return")
	if running, ok := callNative(t, longRunningObj, "isRunning").(bool); !ok || !running {
		t.Fatalf("exec.start isRunning = %#v, want true", running)
	}
	_ = callNative(t, longRunningObj, "kill")
	killedWaitObj := mustRuntimeObject(t, awaitValue(t, callNative(t, longRunningObj, "waitAsync")), "exec.start waitAsync return")
	if ok, _ := killedWaitObj["ok"].(bool); ok {
		t.Fatalf("exec.start killed wait ok = %#v, want false", killedWaitObj["ok"])
	}
	if running, ok := callNative(t, longRunningObj, "isRunning").(bool); !ok || running {
		t.Fatalf("exec.start isRunning after wait = %#v, want false", running)
	}

	lookPathRes := callNative(t, execMod, "lookPath", command)
	if s, ok := lookPathRes.(string); !ok || s == "" {
		t.Fatalf("exec.lookPath = %#v, want non-empty string", lookPathRes)
	}

	level := callNative(t, logMod, "setLevel", "debug")
	if s, ok := level.(string); !ok || s != "debug" {
		t.Fatalf("log.setLevel = %#v, want debug", level)
	}
	gotLevel := callNative(t, logMod, "getLevel")
	if s, ok := gotLevel.(string); !ok || s != "debug" {
		t.Fatalf("log.getLevel = %#v, want debug", gotLevel)
	}
	setJSON := callNative(t, logMod, "setJSON", true)
	if b, ok := setJSON.(bool); !ok || !b {
		t.Fatalf("log.setJSON = %#v, want true", setJSON)
	}
	isJSON := callNative(t, logMod, "isJSON")
	if b, ok := isJSON.(bool); !ok || !b {
		t.Fatalf("log.isJSON = %#v, want true", isJSON)
	}
	_ = callNative(t, logMod, "info", "hello", Object{"k": "v"})
	_ = callNative(t, logMod, "log", "warn", "custom warn", Object{"a": float64(1)})
	scoped := mustRuntimeObject(t, callNative(t, logMod, "with", Object{"scope": "test"}), "log.with return")
	_ = callNative(t, scoped, "debug", "scoped message", Object{"x": "y"})
}

func mustModuleObject(t *testing.T, modules map[string]Value, moduleName string) Object {
	t.Helper()
	raw, ok := modules[moduleName]
	if !ok {
		t.Fatalf("module %s not found", moduleName)
	}
	obj, ok := raw.(Object)
	if !ok {
		t.Fatalf("module %s type = %T, want Object", moduleName, raw)
	}
	return obj
}

func callNative(t *testing.T, module Object, name string, args ...Value) Value {
	t.Helper()
	ret, err := callNativeWithError(module, name, args...)
	if err != nil {
		t.Fatalf("%s() unexpected error: %v", name, err)
	}
	return ret
}

func callNativeWithError(module Object, name string, args ...Value) (Value, error) {
	raw, ok := module[name]
	if !ok {
		return nil, &testError{msg: "native function not found: " + name}
	}
	fn, ok := raw.(NativeFunction)
	if !ok {
		return nil, &testError{msg: "value is not native function: " + name}
	}
	return fn(args)
}

func awaitValue(t *testing.T, v Value) Value {
	t.Helper()
	awaitable, ok := v.(rt.Awaitable)
	if !ok {
		t.Fatalf("value is not Awaitable: %T", v)
	}
	resolved, err := awaitable.Await()
	if err != nil {
		t.Fatalf("await error: %v", err)
	}
	return resolved
}

func mustObject(t *testing.T, obj Object, key string) Object {
	t.Helper()
	raw, ok := obj[key]
	if !ok {
		t.Fatalf("object key %s not found", key)
	}
	out, ok := raw.(Object)
	if !ok {
		t.Fatalf("object key %s type = %T, want Object", key, raw)
	}
	return out
}

func mustRuntimeObject(t *testing.T, v Value, label string) Object {
	t.Helper()
	obj, ok := v.(Object)
	if !ok {
		t.Fatalf("%s type = %T, want Object", label, v)
	}
	return obj
}

type testError struct {
	msg string
}

func (e *testError) Error() string {
	return strings.TrimSpace(e.msg)
}

