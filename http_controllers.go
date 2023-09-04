package webctl

import (
	"encoding/json"
	"net/http"
	"reflect"
	"strconv"
	"strings"
	"sync"
)

var httpControllerPools map[string]sync.Pool

func init() {
	httpControllerPools = map[string]sync.Pool{}
}

func RegistHttpController(name string, uci HttpUserControllerInterface) {

	name = strings.ToLower(name)
	_, ok := httpControllerPools[name]
	if ok {
		panic("已经存在叫做" + name + "的controller")
	}
	controllerType := reflect.Indirect(reflect.ValueOf(uci)).Type() // 必须是值类型
	controllerSize := controllerType.Size()
	pool := sync.Pool{
		New: func() interface{} {
			controllerReflect := reflect.New(controllerType) // New新的肯定是指针
			return (&httpControllerInstanceInfo{}).init(controllerReflect, controllerSize)
		},
	}

	httpControllerPools[name] = pool
}

func runHttpController(controllerName string, serverName string, rw http.ResponseWriter, rq *http.Request, data map[string]interface{}, bodyData []byte) {
	controllerName = strings.ToLower(controllerName)
	serverName = strings.ToLower(serverName)
	ci, ok := getHttpContorller(controllerName, rw, rq, data, bodyData)
	if !ok {
		rw.WriteHeader(404)
		return
	}
	server, ok := ci.methodsMap[serverName]
	if ok {
		server()
	}
	if !ci.contoller.isReplyed() {
		ci.contoller.replyState(403)
	}

	httpPutContorller(ci)
	return
}

func getHttpContorller(controllerName string, rw http.ResponseWriter, rq *http.Request, data map[string]interface{}, bodyData []byte) (ci *httpControllerInstanceInfo, ok bool) {
	controllerPool, ok := httpControllerPools[controllerName]
	if !ok {
		return
	}

	ci = controllerPool.Get().(*httpControllerInstanceInfo)
	ci.pool = controllerPool
	ci.contoller.init(data, rw, rq, bodyData)
	return ci, true
}

func httpPutContorller(uci *httpControllerInstanceInfo) {
	memsetZero(uci.pointer, uci.controllerSize)
	uci.pool.Put(uci)
}

// 组织实例后的信息
type httpControllerInstanceInfo struct {
	contoller      HttpUserControllerInterface
	pointer        uintptr
	controllerSize uintptr
	methodsMap     map[string]func()
	pool           sync.Pool
}

// 初始化组织实例信息
func (this *httpControllerInstanceInfo) init(controllerReflect reflect.Value, controllerSize uintptr) *httpControllerInstanceInfo {
	numMethod := controllerReflect.NumMethod()
	this.methodsMap = make(map[string]func())
	for i := 0; i < numMethod; i++ {
		methodName := controllerReflect.Type().Method(i).Name
		methodName = strings.ToLower(methodName)
		this.methodsMap[methodName] = controllerReflect.Method(i).Interface().(func())
	}
	this.contoller = controllerReflect.Interface().(HttpUserControllerInterface) // 必须是指针
	this.pointer = controllerReflect.Elem().UnsafeAddr()                         // 获取指针指向的的值的地址
	this.controllerSize = controllerSize                                         // conrtoller 内存尺寸
	return this
}

type HttpUserConnect struct {
	rw        http.ResponseWriter
	rq        *http.Request
	bodyData  []byte
	isReplyed bool
}

// 应答
func (this *HttpUserConnect) Response(data map[string]interface{}) {
	// 增加时间戳
	bytes, err := json.Marshal(data)
	if err != nil {
		return
	}
	this.rw.Write(bytes)
	this.isReplyed = true
}

// 应答字符串
func (this *HttpUserConnect) ResponseString(data string) {
	this.rw.Write([]byte(data))
	this.isReplyed = true
}

// 应答bytes
func (this *HttpUserConnect) ResponseBytes(data []byte) {
	this.rw.Write(data)
	this.isReplyed = true
}

func (this *HttpUserConnect) GetResponse() http.ResponseWriter {
	return this.rw
}

func (this *HttpUserConnect) GetRequest() *http.Request {
	return this.rq
}

func (this *HttpUserConnect) GetBodyData() []byte {
	return this.bodyData
}

type HttpUserController struct {
	ReceiveData  map[string]interface{}
	ResponseData map[string]interface{}
	Connect      HttpUserConnect
}

func (this *HttpUserController) init(receiveData map[string]interface{}, rw http.ResponseWriter, rq *http.Request, bodyData []byte) {
	this.ReceiveData = receiveData
	this.Connect.rw = rw
	this.Connect.rq = rq
	this.Connect.bodyData = bodyData
	this.ResponseData = map[string]interface{}{}
}

func (this *HttpUserController) isReplyed() bool {
	return this.Connect.isReplyed
}

func (this *HttpUserController) replyState(code int) {
	this.Connect.rw.WriteHeader(code)
	this.Connect.rw.Write([]byte(strconv.Itoa(code) + "\n"))
}

type HttpUserControllerInterface interface {
	init(data map[string]interface{}, rw http.ResponseWriter, rq *http.Request, bodyData []byte)
	isReplyed() bool
	replyState(code int)
}
