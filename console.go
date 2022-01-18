package quark

import (
	"encoding/json"
	"fmt"
	"net/http"
	"reflect"
	"sync"
	"unsafe"
)

var (
	consoleMethodMap  map[string]reflect.Method
	consoleMethodLock sync.Mutex
	consoleType       = reflect.TypeOf(Console{})
	haltPanicType     = reflect.TypeOf(haltPanic{})
)

type Console struct {
	w     http.ResponseWriter
	r     *http.Request
	body  []byte
	quark uintptr
}

func NewConsole(w http.ResponseWriter, r *http.Request, body []byte) *Console {
	return &Console{
		w:    w,
		r:    r,
		body: body,
	}
}

func (c Console) Halt(status int, e interface{}) {
	if e != nil {
		var b []byte
		switch v := e.(type) {
		case error:
			b = []byte(fmt.Sprintf("%v", v))
		default:
			marshal := json.Marshal
			if c.quark != 0 {
				marshal = (*Quark)(unsafe.Pointer(c.quark)).Marshal
			}
			var err error
			b, err = marshal(e)
			if err != nil {
				panic(fmt.Errorf("Halt marshal error, %v", err))
			}
		}
		panic(haltPanic{status, b})
	}
	panic(haltPanic{Status: status})
}

func (c Console) Request() *http.Request {
	return c.r
}

func (c Console) ResponseWriter() http.ResponseWriter {
	return c.w
}

func (c Console) Body() []byte {
	return c.body
}

type haltPanic struct {
	Status int
	Body   []byte
}

func isConsoleMethod(m reflect.Method) bool {
	consoleMethodLock.Lock()
	if consoleMethodMap == nil {
		consoleMethodMap = make(map[string]reflect.Method)
		t := reflect.TypeOf(Console{})
		for i := 0; i < t.NumMethod(); i++ {
			method := t.Method(i)
			consoleMethodMap[method.Name] = method
		}
	}
	consoleMethodLock.Unlock()
	cm, ok := consoleMethodMap[m.Name]
	if !ok {
		return false
	}
	t, ct := m.Type, cm.Type
	if t.NumIn() != ct.NumIn() || t.NumOut() != ct.NumOut() {
		return false
	}
	if t.NumIn() > 1 {
		for i := 1; i < t.NumIn(); i++ {
			if t.In(i) != ct.In(i) {
				return false
			}
		}
	}
	if t.NumOut() > 0 {
		for i := 0; i < t.NumOut(); i++ {
			if t.Out(i) != ct.Out(i) {
				return false
			}
		}
	}
	return true
}
