package webctl

import (
	"encoding/json"
	"fmt"

	// "musdoor/models/db"

	// "net"
	"net/http"
	"strconv"
	"strings"

	// "time"

	"github.com/gorilla/websocket"

	"io/ioutil"

	"bytes"
	// "conf"
	// "musdoor/ini"
	// . "musdoor/msgCenter"
	// "musdoor/musdoorToken"
	"unsafe"
)

var (
	upgrader = websocket.Upgrader{
		ReadBufferSize:  2048,
		WriteBufferSize: 1024,
		CheckOrigin: func(r *http.Request) bool {
			return true
		},
	}
)

func Init() {
	fmt.Println("EditorNetwork启动")
	// http.HandleFunc("/ws", websocketServer)
	http.HandleFunc("/", httpServer)
	err := http.ListenAndServe(":"+strconv.Itoa(2023), nil)
	fmt.Println("等待")
	if err != nil {
		panic("WebSocket错误!地址 " + ":" + strconv.Itoa(2023) + " 不能监听")
	}
}

func httpServer(rw http.ResponseWriter, rq *http.Request) {

	if rq.Method == http.MethodHead {
		rw.WriteHeader(200)
		rw.Write([]byte{})
		return
	}

	rw.Header().Set("Access-Control-Allow-Origin", "*")
	rw.Header().Set("Access-Control-Allow-Methods", "POST, GET, OPTIONS, PUT, DELETE")
	rw.Header().Set("Access-Control-Allow-Headers", "Action, Module")

	index := strings.Index(rq.URL.Path, "/")
	server := rq.URL.Path[index+1:]

	controller_server := strings.Split(server, "/")
	if len(controller_server) != 2 {
		return
	}

	controller := controller_server[0]
	method := controller_server[1]

	defer rq.Body.Close()
	data, err := ioutil.ReadAll(rq.Body)
	fmt.Println("数据", string(data))
	if err != nil {
		return
	}

	jsonMap := map[string]interface{}{}
	if len(data) > 0 {
		err = json.Unmarshal(data, &jsonMap)
		if err != nil {
			rq.Body = ioutil.NopCloser(bytes.NewBuffer(data))
		}
	}

	runHttpController(controller, method, rw, rq, jsonMap, data)
}

// 初始化内存
func memsetZero(pointer uintptr, size uintptr) {

	if size < 1 { // 必须有要写0的内存尺寸
		return
	}

	tailPointer := pointer + size
	int64Size := unsafe.Sizeof(int64(0))

	if size < int64Size { // 不用加锁
		// 从头循到尾循环
		for ; pointer < tailPointer; pointer++ {
			pData := (*byte)(unsafe.Pointer(pointer))
			*pData = 0
		}
	} else { // 如果初始化的内存>=8字节才加速
		tail := size % int64Size              // 剩下的尾巴长度
		buttocksPointer := tailPointer - tail // 屁股的位置 = 尾巴位置 - 尾巴长度
		// 循环到屁股
		for ; pointer < buttocksPointer; pointer += int64Size {
			pData := (*int64)(unsafe.Pointer(pointer))
			*pData = 0
		}
		// 如果有尾巴就从屁股开始循环尾巴
		for ; buttocksPointer < tailPointer; buttocksPointer++ {
			pData := (*byte)(unsafe.Pointer(buttocksPointer))
			*pData = 0
		}
	}
}
