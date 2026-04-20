package builtin

import (
	goJSON "encoding/json"
	"fmt"
	"reflect"
	"sort"
	"strings"
	"sync"
	"time"

	"github.com/syndtr/goleveldb/leveldb"
)

const (
	iaxProtocolValue    = "iax/1"
	iaxDefaultFrom      = "ialang-app"
	iaxRouteModeDot     = "dot"
	iaxRouteModeSlash   = "slash"
	iaxRouteModeColon   = "colon"
	iaxRouteModeExpress = "express"
	iaxMessageKindEvent = "event"
)

const (
	iaxSendModeUnknown = iota
	iaxSendModeObject
	iaxSendModeString
)

type iaxTransport struct {
	send  NativeFunction
	recv  NativeFunction
	call  NativeFunction
	reply NativeFunction

	mu       sync.Mutex
	sendMode int

	writerOnce sync.Once
	writeCh    chan iaxWriteRequest

	readerOnce sync.Once
	readMu     sync.Mutex
	readCond   *sync.Cond
	readErr    error
	inbox      []Object
}

type iaxWriteRequest struct {
	message Object
	done    chan error
}

type iaxPersistenceConfig struct {
	Enabled bool
	Path    string
}

type iaxPersistenceState struct {
	mu sync.RWMutex

	cfg  iaxPersistenceConfig
	dbMu sync.Mutex
	dbs  map[string]*iaxPersistenceDB
}

type iaxPersistenceDB struct {
	db *leveldb.DB

	mu         sync.Mutex
	batch      *leveldb.Batch
	batchCount int
	closed     bool

	flushStop chan struct{}
}

const (
	iaxPersistenceBatchSize     = 32
	iaxPersistenceFlushInterval = 5 * time.Millisecond
)

var globalIAXPersistence = &iaxPersistenceState{}
var iaxTransportCache sync.Map

func newIAXModule(asyncRuntime AsyncRuntime) Object {
	versionFn := NativeFunction(func(args []Value) (Value, error) {
		if len(args) != 0 {
			return nil, fmt.Errorf("iax.version expects 0 args, got %d", len(args))
		}
		return iaxProtocolValue, nil
	})

	buildRequestFn := NativeFunction(func(args []Value) (Value, error) {
		if len(args) < 3 || len(args) > 4 {
			return nil, fmt.Errorf("iax.buildRequest expects 3-4 args: service, action, payload, [options]")
		}
		service, err := asStringArg("iax.buildRequest", args, 0)
		if err != nil {
			return nil, err
		}
		action, err := asStringArg("iax.buildRequest", args, 1)
		if err != nil {
			return nil, err
		}
		service = strings.TrimSpace(service)
		action = strings.TrimSpace(action)
		if service == "" {
			return nil, fmt.Errorf("iax.buildRequest service cannot be empty")
		}
		if action == "" {
			return nil, fmt.Errorf("iax.buildRequest action cannot be empty")
		}

		options := Object{}
		if len(args) == 4 && args[3] != nil {
			optObj, ok := args[3].(Object)
			if !ok {
				return nil, fmt.Errorf("iax.buildRequest arg[3] expects object options, got %T", args[3])
			}
			options = optObj
		}

		from := iaxDefaultFrom
		if v, ok := options["from"]; ok && v != nil {
			parsed, err := asStringValue("iax.buildRequest options.from", v)
			if err != nil {
				return nil, err
			}
			if strings.TrimSpace(parsed) != "" {
				from = strings.TrimSpace(parsed)
			}
		}

		requestID := ""
		if v, ok := options["requestId"]; ok && v != nil {
			parsed, err := asStringValue("iax.buildRequest options.requestId", v)
			if err != nil {
				return nil, err
			}
			requestID = strings.TrimSpace(parsed)
		}
		if requestID == "" {
			requestID, err = ipcNewID()
			if err != nil {
				return nil, fmt.Errorf("iax.buildRequest generate requestId error: %w", err)
			}
		}

		timestampMs := time.Now().UnixMilli()
		if v, ok := options["timestampMs"]; ok && v != nil {
			ts, err := asIntValue("iax.buildRequest options.timestampMs", v)
			if err != nil {
				return nil, err
			}
			timestampMs = int64(ts)
		}

		traceID := ""
		if v, ok := options["traceId"]; ok && v != nil {
			parsed, err := asStringValue("iax.buildRequest options.traceId", v)
			if err != nil {
				return nil, err
			}
			traceID = strings.TrimSpace(parsed)
		}
		if traceID == "" {
			traceID = from + "-" + requestID
		}

		routeMode, err := iaxRouteModeFromOptions("iax.buildRequest", options)
		if err != nil {
			return nil, err
		}
		route, err := iaxResolveRoute("iax.buildRequest", service, action, routeMode, options)
		if err != nil {
			return nil, err
		}

		return Object{
			"protocol":    iaxProtocolValue,
			"from":        from,
			"service":     service,
			"action":      action,
			"routeMode":   routeMode,
			"route":       route,
			"requestId":   requestID,
			"traceId":     traceID,
			"timestampMs": float64(timestampMs),
			"payload":     args[2],
		}, nil
	})

	buildEventFn := NativeFunction(func(args []Value) (Value, error) {
		if len(args) < 2 || len(args) > 3 {
			return nil, fmt.Errorf("iax.buildEvent expects 2-3 args: topic, payload, [options]")
		}
		topic, err := asStringArg("iax.buildEvent", args, 0)
		if err != nil {
			return nil, err
		}
		topic = strings.TrimSpace(topic)
		if topic == "" {
			return nil, fmt.Errorf("iax.buildEvent topic cannot be empty")
		}

		options := Object{}
		if len(args) == 3 && args[2] != nil {
			optObj, ok := args[2].(Object)
			if !ok {
				return nil, fmt.Errorf("iax.buildEvent arg[2] expects object options, got %T", args[2])
			}
			options = optObj
		}

		from := iaxDefaultFrom
		if v, ok := options["from"]; ok && v != nil {
			parsed, err := asStringValue("iax.buildEvent options.from", v)
			if err != nil {
				return nil, err
			}
			if strings.TrimSpace(parsed) != "" {
				from = strings.TrimSpace(parsed)
			}
		}

		eventID := ""
		if v, ok := options["eventId"]; ok && v != nil {
			parsed, err := asStringValue("iax.buildEvent options.eventId", v)
			if err != nil {
				return nil, err
			}
			eventID = strings.TrimSpace(parsed)
		}
		if eventID == "" {
			eventID, err = ipcNewID()
			if err != nil {
				return nil, fmt.Errorf("iax.buildEvent generate eventId error: %w", err)
			}
		}

		timestampMs := time.Now().UnixMilli()
		if v, ok := options["timestampMs"]; ok && v != nil {
			ts, err := asIntValue("iax.buildEvent options.timestampMs", v)
			if err != nil {
				return nil, err
			}
			timestampMs = int64(ts)
		}

		traceID := ""
		if v, ok := options["traceId"]; ok && v != nil {
			parsed, err := asStringValue("iax.buildEvent options.traceId", v)
			if err != nil {
				return nil, err
			}
			traceID = strings.TrimSpace(parsed)
		}
		if traceID == "" {
			traceID = from + "-" + eventID
		}

		return Object{
			"protocol":    iaxProtocolValue,
			"type":        iaxMessageKindEvent,
			"topic":       topic,
			"eventId":     eventID,
			"traceId":     traceID,
			"from":        from,
			"timestampMs": float64(timestampMs),
			"payload":     args[1],
		}, nil
	})

	configurePersistenceFn := NativeFunction(func(args []Value) (Value, error) {
		if len(args) != 1 {
			return nil, fmt.Errorf("iax.configurePersistence expects 1 arg: options")
		}
		options, ok := args[0].(Object)
		if !ok {
			return nil, fmt.Errorf("iax.configurePersistence arg[0] expects object options, got %T", args[0])
		}
		cfg, err := iaxParsePersistenceConfigOptions("iax.configurePersistence", options)
		if err != nil {
			return nil, err
		}
		globalIAXPersistence.Set(cfg)
		if !cfg.Enabled {
			globalIAXPersistence.closeAllDB()
		}
		return iaxPersistenceConfigToObject(cfg), nil
	})

	getPersistenceFn := NativeFunction(func(args []Value) (Value, error) {
		if len(args) != 0 {
			return nil, fmt.Errorf("iax.getPersistence expects 0 args, got %d", len(args))
		}
		return iaxPersistenceConfigToObject(globalIAXPersistence.Get()), nil
	})

	loadEventsFn := NativeFunction(func(args []Value) (Value, error) {
		if len(args) > 1 {
			return nil, fmt.Errorf("iax.loadEvents expects 0-1 args: [options]")
		}
		options := Object{}
		if len(args) == 1 && args[0] != nil {
			optObj, ok := args[0].(Object)
			if !ok {
				return nil, fmt.Errorf("iax.loadEvents arg[0] expects object options, got %T", args[0])
			}
			options = optObj
		}
		events, err := iaxLoadPersistedEvents(options)
		if err != nil {
			return nil, err
		}
		return events, nil
	})

	replayFn := NativeFunction(func(args []Value) (Value, error) {
		if len(args) < 1 || len(args) > 2 {
			return nil, fmt.Errorf("iax.replay expects 1-2 args: conn, [options]")
		}
		connObj, ok := args[0].(Object)
		if !ok {
			return nil, fmt.Errorf("iax.replay arg[0] expects connection object, got %T", args[0])
		}
		options := Object{}
		if len(args) == 2 && args[1] != nil {
			optObj, ok := args[1].(Object)
			if !ok {
				return nil, fmt.Errorf("iax.replay arg[1] expects object options, got %T", args[1])
			}
			options = optObj
		}
		events, err := iaxLoadPersistedEvents(options)
		if err != nil {
			return nil, err
		}
		sent := 0
		for _, v := range events {
			event, ok := v.(Object)
			if !ok {
				continue
			}
			topic, _ := event["topic"].(string)
			eventID, _ := event["eventId"].(string)
			message := Object{
				"kind":    iaxMessageKindEvent,
				"id":      eventID,
				"method":  topic,
				"topic":   topic,
				"payload": event,
			}
			if err := iaxConnSend(connObj, message); err != nil {
				return nil, err
			}
			sent++
		}
		return Object{
			"ok":    true,
			"count": float64(sent),
		}, nil
	})

	callFn := NativeFunction(func(args []Value) (Value, error) {
		if len(args) < 4 || len(args) > 5 {
			return nil, fmt.Errorf("iax.call expects 4-5 args: conn, service, action, payload, [options]")
		}
		connObj, ok := args[0].(Object)
		if !ok {
			return nil, fmt.Errorf("iax.call arg[0] expects connection object, got %T", args[0])
		}
		service, err := asStringArg("iax.call", args, 1)
		if err != nil {
			return nil, err
		}
		action, err := asStringArg("iax.call", args, 2)
		if err != nil {
			return nil, err
		}

		options := Object{}
		if len(args) == 5 && args[4] != nil {
			optObj, ok := args[4].(Object)
			if !ok {
				return nil, fmt.Errorf("iax.call arg[4] expects object options, got %T", args[4])
			}
			options = optObj
		}

		envelopeOptions, err := iaxNestedObjectOption("iax.call", options, "requestOptions")
		if err != nil {
			return nil, err
		}
		callOptions, err := iaxNestedObjectOption("iax.call", options, "callOptions")
		if err != nil {
			return nil, err
		}

		requestArgs := []Value{service, action, args[3]}
		if envelopeOptions != nil {
			requestArgs = append(requestArgs, envelopeOptions)
		}
		envelopeVal, err := buildRequestFn(requestArgs)
		if err != nil {
			return nil, err
		}
		envelope := envelopeVal.(Object)
		route, _ := envelope["route"].(string)
		if strings.TrimSpace(route) == "" {
			route = strings.TrimSpace(service) + "." + strings.TrimSpace(action)
		}
		if v, ok := options["route"]; ok && v != nil {
			override, err := asStringValue("iax.call options.route", v)
			if err != nil {
				return nil, err
			}
			if strings.TrimSpace(override) != "" {
				route = strings.TrimSpace(override)
			}
		}

		respVal, err := iaxConnCall(connObj, route, envelope, callOptions)
		if err != nil {
			return nil, err
		}
		respObj, ok := respVal.(Object)
		if !ok {
			return nil, fmt.Errorf("iax.call expects ipc response object, got %T", respVal)
		}
		transportOK, _ := respObj["ok"].(bool)
		if !transportOK {
			msg, _ := respObj["error"].(string)
			if strings.TrimSpace(msg) == "" {
				msg = "ipc call failed"
			}
			return iaxCallResult(false, "IPC_CALL_FAILED", msg, nil, nil, respObj), nil
		}

		remoteEnvelope, ok := respObj["payload"].(Object)
		if !ok {
			return nil, fmt.Errorf("iax.call expects response payload object, got %T", respObj["payload"])
		}
		if protocol, _ := remoteEnvelope["protocol"].(string); protocol != "" && protocol != iaxProtocolValue {
			return iaxCallResult(false, "BAD_PROTOCOL", "response protocol is invalid", nil, remoteEnvelope, respObj), nil
		}

		remoteOK, _ := remoteEnvelope["ok"].(bool)
		code, _ := remoteEnvelope["code"].(string)
		message, _ := remoteEnvelope["message"].(string)
		if code == "" {
			if remoteOK {
				code = "OK"
			} else {
				code = "REMOTE_ERROR"
			}
		}
		if !remoteOK && strings.TrimSpace(message) == "" {
			message = "remote handler failed"
		}
		return iaxCallResult(remoteOK, code, message, remoteEnvelope["data"], remoteEnvelope, respObj), nil
	})

	callAsyncFn := NativeFunction(func(args []Value) (Value, error) {
		copiedArgs := append([]Value(nil), args...)
		return asyncRuntime.Spawn(func() (Value, error) {
			return callFn(copiedArgs)
		}), nil
	})

	publishFn := NativeFunction(func(args []Value) (Value, error) {
		if len(args) < 3 || len(args) > 4 {
			return nil, fmt.Errorf("iax.publish expects 3-4 args: conn, topic, payload, [options]")
		}
		connObj, ok := args[0].(Object)
		if !ok {
			return nil, fmt.Errorf("iax.publish arg[0] expects connection object, got %T", args[0])
		}

		options := Object{}
		if len(args) == 4 && args[3] != nil {
			optObj, ok := args[3].(Object)
			if !ok {
				return nil, fmt.Errorf("iax.publish arg[3] expects object options, got %T", args[3])
			}
			options = optObj
		}
		persistCfg, err := iaxResolvePublishPersistence(options)
		if err != nil {
			return nil, err
		}

		eventVal, err := buildEventFn([]Value{args[1], args[2], options})
		if err != nil {
			return nil, err
		}
		event := eventVal.(Object)
		topic, _ := event["topic"].(string)
		eventID, _ := event["eventId"].(string)

		message := Object{
			"kind":    iaxMessageKindEvent,
			"id":      eventID,
			"method":  topic,
			"topic":   topic,
			"payload": event,
		}
		if err := iaxConnSend(connObj, message); err != nil {
			return nil, err
		}
		if persistCfg.Enabled {
			if err := globalIAXPersistence.AppendEvent(persistCfg.Path, event); err != nil {
				return nil, err
			}
		}
		return Object{
			"ok":     true,
			"topic":  topic,
			"event":  event,
			"sentAt": float64(time.Now().UnixMilli()),
		}, nil
	})

	publishAsyncFn := NativeFunction(func(args []Value) (Value, error) {
		copiedArgs := append([]Value(nil), args...)
		return asyncRuntime.Spawn(func() (Value, error) {
			return publishFn(copiedArgs)
		}), nil
	})

	subscribeFn := NativeFunction(func(args []Value) (Value, error) {
		if len(args) < 1 || len(args) > 3 {
			return nil, fmt.Errorf("iax.subscribe expects 1-3 args: conn, [topics], [options]")
		}
		connObj, ok := args[0].(Object)
		if !ok {
			return nil, fmt.Errorf("iax.subscribe arg[0] expects connection object, got %T", args[0])
		}

		topics := []string{"*"}
		options := Object{}
		if len(args) >= 2 && args[1] != nil {
			parsedTopics, err := iaxParseTopicsArg("iax.subscribe", args[1])
			if err != nil {
				return nil, err
			}
			topics = parsedTopics
		}
		if len(args) == 3 && args[2] != nil {
			optObj, ok := args[2].(Object)
			if !ok {
				return nil, fmt.Errorf("iax.subscribe arg[2] expects object options, got %T", args[2])
			}
			options = optObj
		}
		strictProtocol := true
		if v, ok := options["strictProtocol"]; ok && v != nil {
			parsed, err := iaxAsBoolValue("iax.subscribe options.strictProtocol", v)
			if err != nil {
				return nil, err
			}
			strictProtocol = parsed
		}
		subMu := sync.Mutex{}
		closed := false

		nextFn := NativeFunction(func(nextArgs []Value) (Value, error) {
			if len(nextArgs) != 0 {
				return nil, fmt.Errorf("iax.subscription.next expects 0 args, got %d", len(nextArgs))
			}
			subMu.Lock()
			if closed {
				subMu.Unlock()
				return nil, fmt.Errorf("iax.subscription.next subscription is closed")
			}
			subMu.Unlock()

			msgVal, err := iaxConnRecvMatch(connObj, func(msgObj Object) bool {
				kind, _ := msgObj["kind"].(string)
				if kind != iaxMessageKindEvent {
					return false
				}
				topic, _ := msgObj["topic"].(string)
				if strings.TrimSpace(topic) == "" {
					topic, _ = msgObj["method"].(string)
				}
				if !iaxTopicMatchAny(topics, topic) {
					return false
				}
				event, _ := msgObj["payload"].(Object)
				if strictProtocol && event != nil {
					if protocol, _ := event["protocol"].(string); protocol != "" && protocol != iaxProtocolValue {
						return false
					}
				}
				return true
			})
			if err != nil {
				return nil, err
			}
			msgObj, ok := msgVal.(Object)
			if !ok {
				return nil, fmt.Errorf("iax.subscription.next expects event object, got %T", msgVal)
			}
			topic, _ := msgObj["topic"].(string)
			if strings.TrimSpace(topic) == "" {
				topic, _ = msgObj["method"].(string)
			}
			event, _ := msgObj["payload"].(Object)
			out := Object{
				"ok":      true,
				"topic":   topic,
				"event":   event,
				"message": msgObj,
			}
			if event != nil {
				out["payload"] = event["payload"]
				if eventID, ok := event["eventId"]; ok {
					out["eventId"] = eventID
				}
			}
			return out, nil
		})

		nextAsyncFn := NativeFunction(func(nextArgs []Value) (Value, error) {
			copiedArgs := append([]Value(nil), nextArgs...)
			return asyncRuntime.Spawn(func() (Value, error) {
				return nextFn(copiedArgs)
			}), nil
		})

		closeFn := NativeFunction(func(closeArgs []Value) (Value, error) {
			if len(closeArgs) != 0 {
				return nil, fmt.Errorf("iax.subscription.close expects 0 args, got %d", len(closeArgs))
			}
			subMu.Lock()
			closed = true
			subMu.Unlock()
			return true, nil
		})

		return Object{
			"topics":    iaxTopicsToArray(topics),
			"next":      nextFn,
			"nextAsync": nextAsyncFn,
			"close":     closeFn,
		}, nil
	})

	receiveFn := NativeFunction(func(args []Value) (Value, error) {
		if len(args) < 1 || len(args) > 2 {
			return nil, fmt.Errorf("iax.receive expects 1-2 args: conn, [options]")
		}
		connObj, ok := args[0].(Object)
		if !ok {
			return nil, fmt.Errorf("iax.receive arg[0] expects connection object, got %T", args[0])
		}

		options := Object{}
		if len(args) == 2 && args[1] != nil {
			optObj, ok := args[1].(Object)
			if !ok {
				return nil, fmt.Errorf("iax.receive arg[1] expects object options, got %T", args[1])
			}
			options = optObj
		}
		requireProtocol := true
		if v, ok := options["requireProtocol"]; ok && v != nil {
			parsed, err := iaxAsBoolValue("iax.receive options.requireProtocol", v)
			if err != nil {
				return nil, err
			}
			requireProtocol = parsed
		}

		reqVal, err := iaxConnRecvMatch(connObj, func(reqObj Object) bool {
			kind, _ := reqObj["kind"].(string)
			return kind == ipcMessageKindReq
		})
		if err != nil {
			return nil, err
		}
		reqObj, ok := reqVal.(Object)
		if !ok {
			return nil, fmt.Errorf("iax.receive expects ipc request object, got %T", reqVal)
		}

		envelope, ok := reqObj["payload"].(Object)
		if !ok {
			return iaxReceiveResult(false, "BAD_IPC_PAYLOAD", "ipc request payload is not object", reqObj, nil), nil
		}
		if requireProtocol {
			if protocol, _ := envelope["protocol"].(string); protocol != iaxProtocolValue {
				return iaxReceiveResult(false, "BAD_PROTOCOL", "request protocol is missing or invalid", reqObj, envelope), nil
			}
		}

		service, _ := envelope["service"].(string)
		action, _ := envelope["action"].(string)
		route, _ := reqObj["method"].(string)
		if strings.TrimSpace(route) == "" {
			route, _ = envelope["route"].(string)
		}

		out := iaxReceiveResult(true, "OK", "", reqObj, envelope)
		out["service"] = service
		out["action"] = action
		out["payload"] = envelope["payload"]
		out["route"] = route
		if routeMethod, routePath, hasExpress := iaxSplitExpressRoute(route); hasExpress {
			out["routeMethod"] = routeMethod
			out["routePath"] = routePath
		}
		return out, nil
	})

	receiveAsyncFn := NativeFunction(func(args []Value) (Value, error) {
		copiedArgs := append([]Value(nil), args...)
		return asyncRuntime.Spawn(func() (Value, error) {
			return receiveFn(copiedArgs)
		}), nil
	})

	replyFn := NativeFunction(func(args []Value) (Value, error) {
		if len(args) < 4 || len(args) > 5 {
			return nil, fmt.Errorf("iax.reply expects 4-5 args: conn, receiveResult, ok, data, [error]")
		}
		connObj, ok := args[0].(Object)
		if !ok {
			return nil, fmt.Errorf("iax.reply arg[0] expects connection object, got %T", args[0])
		}
		receiveResult, ok := args[1].(Object)
		if !ok {
			return nil, fmt.Errorf("iax.reply arg[1] expects receiveResult object, got %T", args[1])
		}
		okFlag, ok := args[2].(bool)
		if !ok {
			return nil, fmt.Errorf("iax.reply arg[2] expects bool, got %T", args[2])
		}

		request, ok := receiveResult["request"].(Object)
		if !ok {
			return nil, fmt.Errorf("iax.reply receiveResult.request type = %T, want Object", receiveResult["request"])
		}
		requestEnvelope, _ := receiveResult["envelope"].(Object)

		code := "OK"
		message := ""
		if !okFlag {
			code = "REMOTE_ERROR"
			message = "remote handler failed"
			if len(args) == 5 && args[4] != nil {
				text, err := asStringValue("iax.reply arg[4]", args[4])
				if err != nil {
					return nil, err
				}
				if strings.TrimSpace(text) != "" {
					message = strings.TrimSpace(text)
				}
			}
		}

		route, _ := request["method"].(string)
		service, _ := receiveResult["service"].(string)
		action, _ := receiveResult["action"].(string)
		requestID, _ := request["id"].(string)
		traceID := ""
		from := iaxDefaultFrom
		if requestEnvelope != nil {
			if v, ok := requestEnvelope["traceId"].(string); ok {
				traceID = v
			}
			if v, ok := requestEnvelope["from"].(string); ok && strings.TrimSpace(v) != "" {
				from = v
			}
		}

		responseEnvelope := Object{
			"protocol":    iaxProtocolValue,
			"from":        from,
			"service":     service,
			"action":      action,
			"route":       route,
			"requestId":   requestID,
			"traceId":     traceID,
			"timestampMs": float64(time.Now().UnixMilli()),
			"ok":          okFlag,
			"code":        code,
			"message":     message,
			"data":        args[3],
		}

		if err := iaxConnReply(connObj, request, true, responseEnvelope); err != nil {
			return nil, err
		}
		return true, nil
	})

	replyAsyncFn := NativeFunction(func(args []Value) (Value, error) {
		copiedArgs := append([]Value(nil), args...)
		return asyncRuntime.Spawn(func() (Value, error) {
			return replyFn(copiedArgs)
		}), nil
	})

	namespace := Object{
		"version":              versionFn,
		"buildRequest":         buildRequestFn,
		"buildEvent":           buildEventFn,
		"configurePersistence": configurePersistenceFn,
		"getPersistence":       getPersistenceFn,
		"loadEvents":           loadEventsFn,
		"replay":               replayFn,
		"call":                 callFn,
		"callAsync":            callAsyncFn,
		"publish":              publishFn,
		"publishAsync":         publishAsyncFn,
		"subscribe":            subscribeFn,
		"receive":              receiveFn,
		"receiveAsync":         receiveAsyncFn,
		"reply":                replyFn,
		"replyAsync":           replyAsyncFn,
	}
	module := cloneObject(namespace)
	module["iax"] = namespace
	return module
}

func iaxConnSend(conn Object, message Object) error {
	transport, err := iaxResolveTransport(conn)
	if err != nil {
		return err
	}
	iaxEnsureWriterLoop(transport)

	req := iaxWriteRequest{
		message: message,
		done:    make(chan error, 1),
	}
	transport.writeCh <- req
	return <-req.done
}

func iaxConnRecv(conn Object) (Value, error) {
	return iaxConnRecvMatch(conn, nil)
}

func iaxConnRecvMatch(conn Object, matcher func(Object) bool) (Value, error) {
	transport, err := iaxResolveTransport(conn)
	if err != nil {
		return nil, err
	}
	iaxEnsureReaderLoop(transport)

	transport.readMu.Lock()
	defer transport.readMu.Unlock()
	for {
		if cached, ok := iaxConnTakeInboxMatchLocked(transport, matcher); ok {
			return cached, nil
		}
		if transport.readErr != nil {
			return nil, transport.readErr
		}
		transport.readCond.Wait()
	}
}

func iaxConnCall(conn Object, route string, envelope Object, callOptions Object) (Value, error) {
	transport, err := iaxResolveTransport(conn)
	if err != nil {
		return nil, err
	}
	if transport.call != nil {
		callArgs := []Value{route, envelope}
		if callOptions != nil {
			callArgs = append(callArgs, callOptions)
		}
		return transport.call(callArgs)
	}

	requestID, _ := envelope["requestId"].(string)
	if strings.TrimSpace(requestID) == "" {
		return nil, fmt.Errorf("iax.call fallback requires envelope.requestId")
	}
	req := Object{
		"kind":    ipcMessageKindReq,
		"id":      requestID,
		"method":  route,
		"payload": envelope,
	}
	if err := iaxConnSend(conn, req); err != nil {
		return nil, err
	}
	respVal, err := iaxConnRecvMatch(conn, func(respObj Object) bool {
		kind, _ := respObj["kind"].(string)
		if kind != ipcMessageKindResp {
			return false
		}
		responseReqID, _ := respObj["requestId"].(string)
		return responseReqID == requestID
	})
	if err != nil {
		return nil, err
	}
	return respVal, nil
}

func iaxConnReply(conn Object, request Object, okFlag bool, payload Value) error {
	transport, err := iaxResolveTransport(conn)
	if err != nil {
		return err
	}
	if transport.reply != nil {
		_, callErr := transport.reply([]Value{request, okFlag, payload})
		return callErr
	}
	reqID, _ := request["id"].(string)
	if strings.TrimSpace(reqID) == "" {
		return fmt.Errorf("iax.reply request.id is required")
	}
	resp := Object{
		"kind":      ipcMessageKindResp,
		"requestId": reqID,
		"ok":        okFlag,
		"payload":   payload,
	}
	return iaxConnSend(conn, resp)
}

func iaxResolveTransport(conn Object) (*iaxTransport, error) {
	key := iaxConnCacheKey(conn)
	if cached, ok := iaxTransportCache.Load(key); ok && cached != nil {
		if transport, ok := cached.(*iaxTransport); ok {
			return transport, nil
		}
	}
	getFn := func(name string) (NativeFunction, bool, error) {
		raw, ok := conn[name]
		if !ok || raw == nil {
			return nil, false, nil
		}
		fn, ok := raw.(NativeFunction)
		if !ok {
			return nil, false, fmt.Errorf("iax connection field %q is not function: %T", name, raw)
		}
		return fn, true, nil
	}

	sendFn, hasSend, err := getFn("send")
	if err != nil {
		return nil, err
	}
	recvFn, hasRecv, err := getFn("recv")
	if err != nil {
		return nil, err
	}
	if !hasSend || !hasRecv {
		return nil, fmt.Errorf("iax connection requires send/recv functions")
	}
	callFn, _, err := getFn("call")
	if err != nil {
		return nil, err
	}
	replyFn, _, err := getFn("reply")
	if err != nil {
		return nil, err
	}
	transport := &iaxTransport{
		send:     sendFn,
		recv:     recvFn,
		call:     callFn,
		reply:    replyFn,
		sendMode: iaxSendModeUnknown,
	}
	transport.readCond = sync.NewCond(&transport.readMu)
	actual, loaded := iaxTransportCache.LoadOrStore(key, transport)
	if loaded {
		if cached, ok := actual.(*iaxTransport); ok {
			return cached, nil
		}
	}
	return transport, nil
}

func iaxConnCacheKey(conn Object) uintptr {
	// Object is map-backed; use map identity pointer as cache key.
	return reflect.ValueOf(conn).Pointer()
}

func iaxConnRecvRawObject(transport *iaxTransport) (Object, error) {
	val, err := transport.recv([]Value{})
	if err != nil {
		return nil, err
	}
	if msgObj, ok := val.(Object); ok {
		return msgObj, nil
	}
	text, ok := val.(string)
	if !ok {
		return nil, fmt.Errorf("iax.recv unsupported message type: %T", val)
	}
	var parsed any
	if err := goJSON.Unmarshal([]byte(text), &parsed); err != nil {
		return nil, fmt.Errorf("iax.recv json decode error: %w", err)
	}
	out := toRuntimeJSONValue(parsed)
	msgObj, ok := out.(Object)
	if !ok {
		return nil, fmt.Errorf("iax.recv expects json object, got %T", out)
	}
	return msgObj, nil
}

func iaxConnTakeInboxMatchLocked(transport *iaxTransport, matcher func(Object) bool) (Object, bool) {
	if len(transport.inbox) == 0 {
		return nil, false
	}
	if matcher == nil {
		msg := transport.inbox[0]
		transport.inbox = transport.inbox[1:]
		return msg, true
	}
	for i := 0; i < len(transport.inbox); i++ {
		msg := transport.inbox[i]
		if matcher(msg) {
			transport.inbox = append(transport.inbox[:i], transport.inbox[i+1:]...)
			return msg, true
		}
	}
	return nil, false
}

func iaxEnsureWriterLoop(transport *iaxTransport) {
	transport.writerOnce.Do(func() {
		transport.writeCh = make(chan iaxWriteRequest, 256)
		go func() {
			for req := range transport.writeCh {
				req.done <- iaxConnSendRaw(transport, req.message)
			}
		}()
	})
}

func iaxConnSendRaw(transport *iaxTransport, message Object) error {
	transport.mu.Lock()
	mode := transport.sendMode
	transport.mu.Unlock()

	switch mode {
	case iaxSendModeObject:
		_, err := transport.send([]Value{message})
		return err
	case iaxSendModeString:
		raw, marshalErr := goJSON.Marshal(message)
		if marshalErr != nil {
			return marshalErr
		}
		_, err := transport.send([]Value{string(raw)})
		return err
	default:
		if _, err := transport.send([]Value{message}); err == nil {
			transport.mu.Lock()
			transport.sendMode = iaxSendModeObject
			transport.mu.Unlock()
			return nil
		}
		raw, marshalErr := goJSON.Marshal(message)
		if marshalErr != nil {
			return marshalErr
		}
		if _, err := transport.send([]Value{string(raw)}); err != nil {
			return err
		}
		transport.mu.Lock()
		transport.sendMode = iaxSendModeString
		transport.mu.Unlock()
		return nil
	}
}

func iaxEnsureReaderLoop(transport *iaxTransport) {
	transport.readerOnce.Do(func() {
		go func() {
			for {
				msgObj, err := iaxConnRecvRawObject(transport)

				transport.readMu.Lock()
				if err != nil {
					if transport.readErr == nil {
						transport.readErr = err
					}
					transport.readCond.Broadcast()
					transport.readMu.Unlock()
					return
				}
				transport.inbox = append(transport.inbox, msgObj)
				transport.readCond.Broadcast()
				transport.readMu.Unlock()
			}
		}()
	})
}

func iaxCallResult(ok bool, code, message string, data Value, envelope Object, response Object) Object {
	return Object{
		"ok":       ok,
		"code":     code,
		"message":  message,
		"data":     data,
		"envelope": envelope,
		"response": response,
	}
}

func iaxReceiveResult(ok bool, code, message string, request Object, envelope Object) Object {
	return Object{
		"ok":       ok,
		"code":     code,
		"message":  message,
		"request":  request,
		"envelope": envelope,
	}
}

func iaxNestedObjectOption(label string, options Object, key string) (Object, error) {
	if len(options) == 0 {
		return nil, nil
	}
	v, ok := options[key]
	if !ok || v == nil {
		return nil, nil
	}
	obj, ok := v.(Object)
	if !ok {
		return nil, fmt.Errorf("%s options.%s expects object, got %T", label, key, v)
	}
	return obj, nil
}

func iaxRouteModeFromOptions(label string, options Object) (string, error) {
	mode := iaxRouteModeDot
	if len(options) == 0 {
		return mode, nil
	}
	if v, ok := options["routeMode"]; ok && v != nil {
		parsed, err := asStringValue(label+" options.routeMode", v)
		if err != nil {
			return "", err
		}
		switch strings.ToLower(strings.TrimSpace(parsed)) {
		case "", iaxRouteModeDot:
			mode = iaxRouteModeDot
		case iaxRouteModeSlash:
			mode = iaxRouteModeSlash
		case iaxRouteModeColon:
			mode = iaxRouteModeColon
		case iaxRouteModeExpress:
			mode = iaxRouteModeExpress
		default:
			return "", fmt.Errorf("%s options.routeMode expects one of [dot, slash, colon, express], got %q", label, parsed)
		}
	}
	return mode, nil
}

func iaxResolveRoute(label, service, action, routeMode string, options Object) (string, error) {
	if len(options) > 0 {
		if v, ok := options["route"]; ok && v != nil {
			route, err := asStringValue(label+" options.route", v)
			if err != nil {
				return "", err
			}
			if strings.TrimSpace(route) != "" {
				return route, nil
			}
		}
	}

	switch routeMode {
	case iaxRouteModeDot:
		return service + "." + action, nil
	case iaxRouteModeSlash:
		return "/" + strings.Trim(service, "/") + "/" + strings.Trim(action, "/"), nil
	case iaxRouteModeColon:
		return service + ":" + action, nil
	case iaxRouteModeExpress:
		path := "/" + strings.Trim(service, "/") + "/" + strings.Trim(action, "/")
		if len(options) > 0 {
			if v, ok := options["routeTemplate"]; ok && v != nil {
				template, err := asStringValue(label+" options.routeTemplate", v)
				if err != nil {
					return "", err
				}
				if strings.TrimSpace(template) != "" {
					path = strings.ReplaceAll(template, ":service", strings.Trim(service, "/"))
					path = strings.ReplaceAll(path, ":action", strings.Trim(action, "/"))
				}
			}
			if v, ok := options["routePrefix"]; ok && v != nil {
				prefix, err := asStringValue(label+" options.routePrefix", v)
				if err != nil {
					return "", err
				}
				prefix = strings.TrimSpace(prefix)
				if prefix != "" {
					prefix = "/" + strings.Trim(prefix, "/")
					path = strings.TrimSuffix(prefix, "/") + "/" + strings.TrimPrefix(path, "/")
				}
			}
		}
		verb := "POST"
		if len(options) > 0 {
			if v, ok := options["routeMethod"]; ok && v != nil {
				methodText, err := asStringValue(label+" options.routeMethod", v)
				if err != nil {
					return "", err
				}
				methodText = strings.TrimSpace(methodText)
				if methodText != "" {
					verb = strings.ToUpper(methodText)
				}
			}
		}
		return verb + " " + path, nil
	default:
		return "", fmt.Errorf("%s options.routeMode is unsupported: %q", label, routeMode)
	}
}

func iaxSplitExpressRoute(route string) (string, string, bool) {
	parts := strings.SplitN(strings.TrimSpace(route), " ", 2)
	if len(parts) != 2 {
		return "", "", false
	}
	method := strings.TrimSpace(parts[0])
	path := strings.TrimSpace(parts[1])
	if method == "" || path == "" || !strings.HasPrefix(path, "/") {
		return "", "", false
	}
	return method, path, true
}

func iaxAsBoolValue(label string, v Value) (bool, error) {
	b, ok := v.(bool)
	if !ok {
		return false, fmt.Errorf("%s expects bool, got %T", label, v)
	}
	return b, nil
}

func iaxParseTopicsArg(label string, v Value) ([]string, error) {
	switch raw := v.(type) {
	case string:
		topic := strings.TrimSpace(raw)
		if topic == "" {
			return nil, fmt.Errorf("%s topics cannot be empty", label)
		}
		return []string{topic}, nil
	case Array:
		if len(raw) == 0 {
			return nil, fmt.Errorf("%s topics array cannot be empty", label)
		}
		out := make([]string, 0, len(raw))
		for i, item := range raw {
			s, err := asStringValue(fmt.Sprintf("%s topics[%d]", label, i), item)
			if err != nil {
				return nil, err
			}
			s = strings.TrimSpace(s)
			if s == "" {
				return nil, fmt.Errorf("%s topics[%d] cannot be empty", label, i)
			}
			out = append(out, s)
		}
		return out, nil
	default:
		return nil, fmt.Errorf("%s topics expects string or array, got %T", label, v)
	}
}

func iaxTopicMatchAny(patterns []string, topic string) bool {
	topic = strings.TrimSpace(topic)
	for _, pattern := range patterns {
		if iaxTopicMatch(pattern, topic) {
			return true
		}
	}
	return false
}

func iaxTopicMatch(pattern, topic string) bool {
	pattern = strings.TrimSpace(pattern)
	if pattern == "*" {
		return true
	}
	if !strings.Contains(pattern, "*") {
		return pattern == topic
	}
	parts := strings.Split(pattern, "*")
	if len(parts) == 2 {
		prefix := parts[0]
		suffix := parts[1]
		return strings.HasPrefix(topic, prefix) && strings.HasSuffix(topic, suffix)
	}
	cursor := 0
	for i, part := range parts {
		if part == "" {
			continue
		}
		idx := strings.Index(topic[cursor:], part)
		if idx < 0 {
			return false
		}
		if i == 0 && !strings.HasPrefix(topic, part) {
			return false
		}
		cursor += idx + len(part)
	}
	last := parts[len(parts)-1]
	if last != "" && !strings.HasSuffix(topic, last) {
		return false
	}
	return true
}

func iaxTopicsToArray(topics []string) Array {
	out := make(Array, 0, len(topics))
	for _, topic := range topics {
		out = append(out, topic)
	}
	return out
}

func (s *iaxPersistenceState) Get() iaxPersistenceConfig {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.cfg
}

func (s *iaxPersistenceState) Set(cfg iaxPersistenceConfig) {
	s.mu.Lock()
	prev := s.cfg
	defer s.mu.Unlock()
	s.cfg = cfg
	prevPath := strings.TrimSpace(prev.Path)
	nextPath := strings.TrimSpace(cfg.Path)
	if prevPath != "" && (nextPath == "" || prevPath != nextPath || !cfg.Enabled) {
		s.closeDB(prevPath)
	}
	if !cfg.Enabled && nextPath == "" {
		s.closeAllDB()
	}
}

func (s *iaxPersistenceState) AppendEvent(path string, event Object) error {
	if strings.TrimSpace(path) == "" {
		return fmt.Errorf("iax persistence path cannot be empty")
	}
	pdb, err := s.openDB(path)
	if err != nil {
		return err
	}
	key := iaxPersistenceEventKey(event)
	raw, err := goJSON.Marshal(event)
	if err != nil {
		return fmt.Errorf("iax persistence encode event error: %w", err)
	}
	return pdb.append([]byte(key), raw)
}

func iaxPersistenceConfigToObject(cfg iaxPersistenceConfig) Object {
	return Object{
		"enabled": cfg.Enabled,
		"path":    cfg.Path,
	}
}

func iaxParsePersistenceConfigOptions(label string, options Object) (iaxPersistenceConfig, error) {
	cfg := globalIAXPersistence.Get()
	if v, ok := options["enabled"]; ok && v != nil {
		enabled, err := iaxAsBoolValue(label+" options.enabled", v)
		if err != nil {
			return iaxPersistenceConfig{}, err
		}
		cfg.Enabled = enabled
	}
	if v, ok := options["path"]; ok && v != nil {
		path, err := asStringValue(label+" options.path", v)
		if err != nil {
			return iaxPersistenceConfig{}, err
		}
		cfg.Path = strings.TrimSpace(path)
	}
	if cfg.Enabled && strings.TrimSpace(cfg.Path) == "" {
		return iaxPersistenceConfig{}, fmt.Errorf("%s requires non-empty options.path when enabled=true", label)
	}
	return cfg, nil
}

func iaxResolvePublishPersistence(options Object) (iaxPersistenceConfig, error) {
	cfg := globalIAXPersistence.Get()
	if len(options) == 0 {
		return cfg, nil
	}
	if v, ok := options["persist"]; ok && v != nil {
		switch pv := v.(type) {
		case bool:
			cfg.Enabled = pv
		case Object:
			parsed, err := iaxParsePersistenceConfigOptions("iax.publish options.persist", pv)
			if err != nil {
				return iaxPersistenceConfig{}, err
			}
			cfg = parsed
		default:
			return iaxPersistenceConfig{}, fmt.Errorf("iax.publish options.persist expects bool or object, got %T", v)
		}
	}
	if v, ok := options["persistPath"]; ok && v != nil {
		path, err := asStringValue("iax.publish options.persistPath", v)
		if err != nil {
			return iaxPersistenceConfig{}, err
		}
		cfg.Path = strings.TrimSpace(path)
	}
	if cfg.Enabled && strings.TrimSpace(cfg.Path) == "" {
		return iaxPersistenceConfig{}, fmt.Errorf("iax.publish persistence enabled but path is empty")
	}
	return cfg, nil
}

func iaxLoadPersistedEvents(options Object) (Array, error) {
	cfg := globalIAXPersistence.Get()
	if v, ok := options["path"]; ok && v != nil {
		path, err := asStringValue("iax.loadEvents options.path", v)
		if err != nil {
			return nil, err
		}
		cfg.Path = strings.TrimSpace(path)
	}
	path := strings.TrimSpace(cfg.Path)
	if path == "" {
		return nil, fmt.Errorf("iax.loadEvents path is empty; configure persistence first or pass options.path")
	}

	var patterns []string
	if v, ok := options["topics"]; ok && v != nil {
		parsed, err := iaxParseTopicsArg("iax.loadEvents", v)
		if err != nil {
			return nil, err
		}
		patterns = parsed
	}
	if v, ok := options["topic"]; ok && v != nil {
		parsed, err := iaxParseTopicsArg("iax.loadEvents", v)
		if err != nil {
			return nil, err
		}
		patterns = parsed
	}
	sinceMs := int64(0)
	if v, ok := options["sinceMs"]; ok && v != nil {
		parsed, err := asIntValue("iax.loadEvents options.sinceMs", v)
		if err != nil {
			return nil, err
		}
		sinceMs = int64(parsed)
	}
	limit := 0
	if v, ok := options["limit"]; ok && v != nil {
		parsed, err := asIntValue("iax.loadEvents options.limit", v)
		if err != nil {
			return nil, err
		}
		if parsed > 0 {
			limit = parsed
		}
	}

	pdb, err := globalIAXPersistence.openDB(path)
	if err != nil {
		return nil, err
	}
	if err := pdb.flushNow(); err != nil {
		return nil, err
	}
	iter := pdb.db.NewIterator(nil, nil)
	defer iter.Release()

	events := make(Array, 0, 64)
	staged := make([]Object, 0, 64)
	for iter.Next() {
		var parsed any
		if err := goJSON.Unmarshal(iter.Value(), &parsed); err != nil {
			continue
		}
		eventVal := toRuntimeJSONValue(parsed)
		event, ok := eventVal.(Object)
		if !ok {
			continue
		}
		if len(patterns) > 0 {
			topic, _ := event["topic"].(string)
			if !iaxTopicMatchAny(patterns, topic) {
				continue
			}
		}
		if sinceMs > 0 {
			tsFloat, _ := event["timestampMs"].(float64)
			if int64(tsFloat) < sinceMs {
				continue
			}
		}
		staged = append(staged, event)
	}
	if err := iter.Error(); err != nil {
		return nil, fmt.Errorf("iax.loadEvents leveldb iterate error: %w", err)
	}
	// Keep deterministic order by timestamp then eventId.
	sort.SliceStable(staged, func(i, j int) bool {
		it, _ := staged[i]["timestampMs"].(float64)
		jt, _ := staged[j]["timestampMs"].(float64)
		if it != jt {
			return it < jt
		}
		ie, _ := staged[i]["eventId"].(string)
		je, _ := staged[j]["eventId"].(string)
		return ie < je
	})
	if limit > 0 && len(staged) > limit {
		staged = staged[:limit]
	}
	for _, event := range staged {
		events = append(events, event)
	}
	return events, nil
}

func (s *iaxPersistenceState) openDB(path string) (*iaxPersistenceDB, error) {
	cleanPath := strings.TrimSpace(path)
	if cleanPath == "" {
		return nil, fmt.Errorf("iax persistence path cannot be empty")
	}
	s.dbMu.Lock()
	defer s.dbMu.Unlock()
	if s.dbs == nil {
		s.dbs = make(map[string]*iaxPersistenceDB)
	}
	if pdb, ok := s.dbs[cleanPath]; ok && pdb != nil {
		return pdb, nil
	}
	db, err := leveldb.OpenFile(cleanPath, nil)
	if err != nil {
		return nil, fmt.Errorf("iax persistence open leveldb error: %w", err)
	}
	pdb := &iaxPersistenceDB{
		db:        db,
		batch:     &leveldb.Batch{},
		flushStop: make(chan struct{}),
	}
	s.dbs[cleanPath] = pdb
	go pdb.flushLoop()
	return pdb, nil
}

func (s *iaxPersistenceState) closeDB(path string) {
	cleanPath := strings.TrimSpace(path)
	if cleanPath == "" {
		return
	}
	s.dbMu.Lock()
	pdb := s.dbs[cleanPath]
	if pdb != nil {
		delete(s.dbs, cleanPath)
	}
	s.dbMu.Unlock()
	if pdb != nil {
		_ = pdb.close()
	}
}

func (s *iaxPersistenceState) closeAllDB() {
	s.dbMu.Lock()
	dbs := make([]*iaxPersistenceDB, 0, len(s.dbs))
	for key, pdb := range s.dbs {
		if pdb != nil {
			dbs = append(dbs, pdb)
		}
		delete(s.dbs, key)
	}
	s.dbMu.Unlock()
	for _, pdb := range dbs {
		_ = pdb.close()
	}
}

func (pdb *iaxPersistenceDB) append(key []byte, value []byte) error {
	pdb.mu.Lock()
	defer pdb.mu.Unlock()
	if pdb.closed {
		return fmt.Errorf("iax persistence db is closed")
	}
	pdb.batch.Put(key, value)
	pdb.batchCount++
	if pdb.batchCount < iaxPersistenceBatchSize {
		return nil
	}
	return pdb.flushLocked()
}

func (pdb *iaxPersistenceDB) flushNow() error {
	pdb.mu.Lock()
	defer pdb.mu.Unlock()
	if pdb.closed {
		return nil
	}
	return pdb.flushLocked()
}

func (pdb *iaxPersistenceDB) flushLocked() error {
	if pdb.batchCount == 0 {
		return nil
	}
	batch := pdb.batch
	pdb.batch = &leveldb.Batch{}
	pdb.batchCount = 0
	if err := pdb.db.Write(batch, nil); err != nil {
		return fmt.Errorf("iax persistence write leveldb batch error: %w", err)
	}
	return nil
}

func (pdb *iaxPersistenceDB) flushLoop() {
	ticker := time.NewTicker(iaxPersistenceFlushInterval)
	defer ticker.Stop()
	for {
		select {
		case <-ticker.C:
			_ = pdb.flushNow()
		case <-pdb.flushStop:
			return
		}
	}
}

func (pdb *iaxPersistenceDB) close() error {
	pdb.mu.Lock()
	if pdb.closed {
		pdb.mu.Unlock()
		return nil
	}
	pdb.closed = true
	close(pdb.flushStop)
	err := pdb.flushLocked()
	pdb.mu.Unlock()
	if err != nil {
		_ = pdb.db.Close()
		return err
	}
	return pdb.db.Close()
}

func iaxPersistenceEventKey(event Object) string {
	ts := time.Now().UnixMilli()
	if raw, ok := event["timestampMs"]; ok && raw != nil {
		if n, err := asIntValue("iax.persistence event.timestampMs", raw); err == nil {
			ts = int64(n)
		}
	}
	eventID := ""
	if v, ok := event["eventId"].(string); ok {
		eventID = strings.TrimSpace(v)
	}
	if eventID == "" {
		eventID = fmt.Sprintf("evt-%d", time.Now().UnixNano())
	}
	return fmt.Sprintf("%020d:%s", ts, eventID)
}
