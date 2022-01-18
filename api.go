package quark

import (
	"fmt"
	"io/ioutil"
	"net/http"
	"reflect"
	"runtime/debug"
	"strconv"
	"strings"
	"sync"
	"time"
	"unicode"

	"github.com/dovejb/quark/types"
	"github.com/dovejb/quark/util"
	"github.com/go-openapi/spec"
)

type Kind uint

const (
	Invalid Kind = iota
	Bool
	Int
	Float
	Big
	String
)

const (
	ROOT_SERVICE_NAME = "root"
)

var (
	validMethods = map[string]bool{
		http.MethodGet:     true,
		http.MethodPost:    true,
		http.MethodPatch:   true,
		http.MethodDelete:  true,
		http.MethodPut:     true,
		http.MethodOptions: true,
		http.MethodHead:    true,
	}
)

type Quark struct {
	lock      sync.Mutex
	Marshal   JsonMarshalFunc
	Unmarshal JsonUnmarshalFunc
	Services  []Service
	smap      map[string]int
	swagger   *spec.Swagger
}

func (q *Quark) SwaggerSpec() *spec.Swagger {
	q.lock.Lock()
	defer q.lock.Unlock()
	if q.swagger != nil {
		return q.swagger
	}
	q.swagger = new(spec.Swagger)
	q.swagger.Paths = new(spec.Paths)
	q.swagger.Paths.Paths = make(map[string]spec.PathItem)
	for _, service := range q.Services {
		q.swagger.SwaggerProps.Tags = append(q.swagger.SwaggerProps.Tags, spec.NewTag(service.Name, service.ServiceType.Name(), nil))
		for _, api := range service.Apis {
			item := api.SwaggerPathItem()
			if item == nil {
				continue
			}
			q.swagger.Paths.Paths[api.docPath] = *item
		}
	}
	q.swagger.Swagger = "2.0"
	q.swagger.Info = &spec.Info{
		InfoProps: spec.InfoProps{
			Title:       "Quark Service",
			Description: "This docunent is auto-generated by Quark",
			Version:     time.Now().Format("v2006.0102.150405"),
		},
	}
	return q.swagger
}

func (q *Quark) RegisterService(instances ...interface{}) {
	q.lock.Lock()
	defer q.lock.Unlock()
	if q.smap == nil {
		q.smap = make(map[string]int)
	}
	for _, inst := range instances {
		t := reflect.TypeOf(inst)
		s := q.newService(t)
		q.Services = append(q.Services, *s)
		q.smap[s.Name] = len(q.Services) - 1
	}
	q.swagger = nil
}

func (q *Quark) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	defer func() {
		if exception := recover(); exception != nil {
			t := reflect.TypeOf(exception)
			if t == haltPanicType {
				hp := exception.(haltPanic)
				w.WriteHeader(hp.Status)
				w.Write(hp.Body)
			} else {
				w.WriteHeader(http.StatusInternalServerError)
				w.Write([]byte(fmt.Sprintf("%v\n\n", exception)))
				w.Write([]byte(debug.Stack()))
			}
		}
	}()
	path := r.URL.Path
	if path != "" && path[0] == '/' {
		path = path[1:]
	}
	pathElems := strings.Split(path, "/")
	//firstly, find handler by leading service name
	for len(pathElems) > 0 {
		serviceIndex, ok := q.smap[pathElems[0]]
		if !ok {
			break
		}
		if q.Services[serviceIndex].Route(w, r, pathElems[1:]) {
			return
		}
		break
	}
	// if not hit, try root handler
	rootServiceIndex, ok := q.smap[ROOT_SERVICE_NAME]
	if !ok {
		w.WriteHeader(http.StatusGone)
		return
	}
	if q.Services[rootServiceIndex].Route(w, r, pathElems) {
		return
	}
	w.WriteHeader(http.StatusGone)
}

func (q *Quark) addModel(t reflect.Type) {
	if q.swagger.Definitions == nil {
		q.swagger.Definitions = make(spec.Definitions)
	}
	if _, exists := q.swagger.Definitions[t.Name()]; exists {
		return
	}
	q.swagger.Definitions[t.Name()] = q.SwaggerSchemaFromType(t, true)
}

type Service struct {
	Name          string
	ServiceType   reflect.Type
	Apis          []Api
	atrie         util.Trie // path format with {%} mark; there's a final tire indicating method, by :GET, :POST or : for ANY
	quarkInstance *Quark
}

func (s Service) DumpPaths() {
	s.atrie.Dump()
}

func (s *Service) route(pathElems []string, trie util.Trie) (methodTrie util.Trie) {
	if len(pathElems) == 0 {
		return trie
	}
	choices := PathElementChoices(pathElems[0])
	for i := range choices {
		sub := trie.Sub(choices[i])
		if sub.Valid() {
			candidate := s.route(pathElems[1:], sub)
			if candidate.Valid() {
				return candidate
			}
		}
	}
	return util.NilTrie
}

func (s *Service) Route(w http.ResponseWriter, r *http.Request, pathElems []string) (accepted bool) {
	methodTrie := s.route(pathElems, s.atrie)
	if !methodTrie.Valid() {
		return false
	}
	apiIndexTrie, ok := methodTrie.Find([]string{":" + r.Method})
	if !ok {
		apiIndexTrie, ok = methodTrie.Find([]string{":"})
	}
	if !ok {
		return false
	}
	if index := apiIndexTrie.Value(); index == nil {
		return false
	} else {
		s.Apis[*index].Run(w, r, pathElems)
	}
	return true
}

func (s *Service) Quark() *Quark {
	return s.quarkInstance
}

type Api struct {
	Method          string // blank means any, or http.MethodXXX
	docMethod       string // no req or only Parameter type in req, then is GET,  or POST
	Path            string
	docPath         string
	Request         reflect.Type
	Response        reflect.Type
	ReflectMethod   reflect.Method
	PathVars        []util.PathVar // key: path element pos, value: i/s/f.{varname}
	serviceInstance *Service
}

func (a *Api) SwaggerPathItem() *spec.PathItem {
	pi := new(spec.PathItem)
	switch a.docMethod {
	case http.MethodGet:
		pi.Get = a.SwaggerOperations()
	case http.MethodPost:
		pi.Post = a.SwaggerOperations()
	case http.MethodDelete:
		pi.Delete = a.SwaggerOperations()
	case http.MethodPatch:
		pi.Patch = a.SwaggerOperations()
	case http.MethodPut:
		pi.Put = a.SwaggerOperations()
	case http.MethodHead:
		pi.Head = a.SwaggerOperations()
	case http.MethodOptions:
		pi.Options = a.SwaggerOperations()
	}

	return pi
}

func (a *Api) SwaggerOperations() *spec.Operation {
	op := new(spec.Operation)
	op.Tags = []string{a.Service().Name}
	for _, pv := range a.PathVars {
		ss := strings.SplitN(pv.Var, ".", 2)
		var typ string
		switch ss[0] {
		case "s":
			typ = "string"
		case "i":
			typ = "integer"
		case "f":
			typ = "number"
		}
		op.Parameters = append(op.Parameters, spec.Parameter{
			ParamProps: spec.ParamProps{
				Name:     ss[1],
				In:       "path",
				Required: true,
			},
			SimpleSchema: spec.SimpleSchema{
				Type: typ,
			},
		})
	}
	if a.Request != nil {
		hasBody := false
		for i := 0; i < a.Request.NumField(); i++ {
			f := a.Request.Field(i)
			var dataType reflect.Type = f.Type
			var nullable bool
			for dataType.Kind() == reflect.Ptr {
				dataType = dataType.Elem()
				nullable = true
			}
			var typ string
			var in string
			switch dataType {
			case types.URLIntType:
				typ = "integer"
				in = "query"
			case types.URLNumberType:
				typ = "number"
				in = "query"
			case types.URLStringType:
				typ = "string"
				in = "query"
			default: //body field
				hasBody = true
				continue
			}
			op.Parameters = append(op.Parameters, spec.Parameter{
				ParamProps: spec.ParamProps{
					Name:     f.Name,
					In:       in,
					Required: !nullable,
				},
				SimpleSchema: spec.SimpleSchema{
					Type:     typ,
					Nullable: nullable,
				},
			})
		}
		if hasBody {
			var schema spec.Schema
			if a.Request.Name() != "" { //Public model
				schema.Ref, _ = spec.NewRef("#/definitions/" + a.Request.Name())
				a.Service().Quark().addModel(a.Request)
			} else { //Anonymous local schema
				schema = a.Service().Quark().SwaggerSchemaFromType(a.Request, true)
			}
			op.Parameters = append(op.Parameters, spec.Parameter{
				ParamProps: spec.ParamProps{
					Name:     "request-body",
					In:       "body",
					Required: true,
					Schema:   &schema,
				},
			})
		}
	}
	var rsp200 *spec.Schema
	if a.Response != nil {
		rsp200 = new(spec.Schema)
		*rsp200 = a.Service().Quark().SwaggerSchemaFromType(a.Response, false)
	}
	op.Responses = &spec.Responses{
		ResponsesProps: spec.ResponsesProps{
			StatusCodeResponses: map[int]spec.Response{
				200: {
					ResponseProps: spec.ResponseProps{
						Schema: rsp200,
					},
				},
				/*
					400: {},
					401: {},
					404: {},
					500: {},
				*/
			},
		},
	}
	return op
}

func (a *Api) Run(w http.ResponseWriter, r *http.Request, pathElems []string) {
	body, e := ioutil.ReadAll(r.Body)
	if e != nil {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(fmt.Sprintf("read request body fail, %v", e)))
		return
	}
	r.Body.Close()
	r.ParseForm()
	objV := reflect.New(a.ReflectMethod.Type.In(0)).Elem()
	if objV.Kind() == reflect.Struct && objV.NumField() > 0 {
		if consoleValue := objV.Field(0); consoleValue.Type() == consoleType {
			consoleValue.Set(reflect.ValueOf(NewConsole(w, r, body)).Elem())
		}
	}
	in := []reflect.Value{objV}
	for _, pv := range a.PathVars {
		argV := reflect.New(a.ReflectMethod.Type.In(len(in))).Elem()
		arg := pathElems[pv.Pos]
		switch pv.Var[0] {
		case 'i':
			i, e := strconv.ParseInt(arg, 10, 64)
			if e != nil {
				w.WriteHeader(http.StatusBadRequest)
				w.Write([]byte(e.Error()))
				return
			}
			argV.SetInt(i)
		case 'f':
			f, e := strconv.ParseFloat(arg, 64)
			if e != nil {
				w.WriteHeader(http.StatusBadRequest)
				w.Write([]byte(e.Error()))
				return
			}
			argV.SetFloat(f)
		case 's':
			argV.SetString(arg)
		}
		in = append(in, argV)
	}
	if a.Request != nil {
		reqV := reflect.New(a.Request)
		if len(body) > 0 {
			e := a.Service().Quark().Unmarshal(body, reqV.Interface())
			if e != nil {
				w.WriteHeader(http.StatusBadRequest)
				w.Write([]byte(e.Error()))
				return
			}
		}
		in = append(in, reqV.Elem())
	}
	out := a.ReflectMethod.Func.Call(in)
	if len(out) > 0 {
		b, e := a.Service().Quark().Marshal(out[0].Interface())
		if e != nil {
			w.WriteHeader(http.StatusInternalServerError)
			w.Write([]byte(e.Error()))
			return
		}
		w.Write(b)
	}
	return
}

func (api *Api) Service() *Service {
	return api.serviceInstance
}

func (q *Quark) newService(t reflect.Type) (s *Service) {
	s = new(Service)
	s.Name = t.Name()
	s.ServiceType = t
	s.quarkInstance = q
	s.atrie = util.NewTrie()
	if t.Kind() != reflect.Struct {
		panic(fmt.Errorf("only allow struct type, but receive [%s-%s]", t.Name(), t.Kind()))
	}
	for i := 0; i < t.NumMethod(); i++ {
		method := t.Method(i)
		if isConsoleMethod(method) {
			continue
		}
		if api, e := s.newApi(method); e != nil {
			panic(fmt.Errorf("newApi fail %v", e))
		} else {
			s.Apis = append(s.Apis, *api)
			triePath := append(strings.Split(api.Path[1:], "/"), ":"+api.Method)
			s.atrie.Add(triePath, len(s.Apis)-1)
		}
	}
	return
}

func (s *Service) newApi(method reflect.Method) (api *Api, e error) {
	name := method.Name
	api = new(Api)
	api.serviceInstance = s
	api.ReflectMethod = method
	if firstUnderlinePos := strings.Index(name, "_"); firstUnderlinePos >= 0 {
		if methodCandidate := name[:firstUnderlinePos]; validMethods[methodCandidate] {
			api.Method = methodCandidate
			name = name[firstUnderlinePos+1:]
		}
	}
	var varsType []reflect.Kind
	mtype := method.Type
	switch mtype.NumIn() {
	case 1: // no input
	default: // path vars, with or without final request struct
		for i := 1; i < mtype.NumIn(); i++ {
			argType := mtype.In(i)
			argKind := argType.Kind()
			switch {
			case argKind == reflect.String:
				varsType = append(varsType, reflect.String)
			case reflect.Int <= argKind && argKind <= reflect.Uint64:
				varsType = append(varsType, reflect.Int)
			case reflect.Float32 <= argKind && argKind <= reflect.Float64:
				varsType = append(varsType, reflect.Float64)
			case argKind == reflect.Struct:
				if i == mtype.NumIn()-1 {
					api.Request = argType
				}
			default:
				api, e = nil, fmt.Errorf("newApi invalid func format[%s], path vars only accept number or string type, but appear %v", method.Name, argKind)
				return
			}
		}
	}
	api.Path, api.PathVars, e = util.FuncNameToPathWithVars(name, varsType)
	if e != nil {
		e = fmt.Errorf("newApi parse path vars fail, %v", e)
		return
	}
	if mtype.NumOut() > 0 {
		api.Response = mtype.Out(0)
	}
	api.docMethod = api.Method
	if api.docMethod == "" {
		if api.Request != nil {
			for i := 0; i < api.Request.NumField(); i++ {
				f := api.Request.Field(i)
				if !types.IsUrlType(f.Type) {
					api.docMethod = http.MethodPost
				}
			}
		}
		if api.docMethod == "" {
			api.docMethod = http.MethodGet
		}
	}

	if len(api.PathVars) == 0 {
		api.docPath = api.Path
	} else {
		pe := strings.Split(api.Path, "/")
		for _, pv := range api.PathVars {
			pe[pv.Pos+1] = "{" + strings.Split(pv.Var, ".")[1] + "}"
		}
		api.docPath = strings.Join(pe, "/")
	}

	return
}

func PathElementChoices(elem string) (results []string) {
	choices := []string{elem, "{i}", "{f}", "{s}"}
	POS_SAME := 0
	POS_INT := 1
	POS_FLOAT := 2
	dotNum := 0
	for _, r := range elem {
		if unicode.IsLetter(r) {
			choices[POS_FLOAT] = ""
			choices[POS_INT] = ""
		} else if unicode.IsSymbol(r) || unicode.IsPunct(r) || unicode.IsSpace(r) {
			choices[POS_INT] = ""
			switch r {
			case '.':
				choices[POS_SAME] = ""
				dotNum += 1
				if dotNum > 1 {
					choices[POS_FLOAT] = ""
				}
			case '_':
			default:
				choices[POS_SAME] = ""
			}
		} else if !unicode.IsNumber(r) {
			choices[POS_FLOAT] = ""
			choices[POS_INT] = ""
		}
	}
	for i := range choices {
		if choices[i] != "" {
			results = append(results, choices[i])
		}
	}
	return results
}

type TypeAndFormat struct {
	T string
	F string
}

var (
	reservedStructTypes = map[reflect.Type]TypeAndFormat{
		reflect.TypeOf(time.Time{}): {"string", "date-time"},
		types.URLIntType:            {"integer", "int64"},
		types.URLNumberType:         {"number", "double"},
		types.URLStringType:         {"string", ""},
	}
	intSchema = spec.Schema{
		SchemaProps: spec.SchemaProps{
			Type:   spec.StringOrArray{"integer"},
			Format: "int64",
		},
	}
	doubleSchema = spec.Schema{
		SchemaProps: spec.SchemaProps{
			Type:   spec.StringOrArray{"number"},
			Format: "double",
		},
	}
	stringSchema = spec.Schema{
		SchemaProps: spec.SchemaProps{
			Type: spec.StringOrArray{"string"},
		},
	}
)

func (q *Quark) SwaggerSchemaFromType(t reflect.Type, omit_url_parameters bool) (schema spec.Schema) {
	repeated := 0
	nullable := false
UNWRAP:
	for {
		switch t.Kind() {
		case reflect.Ptr:
			nullable = true
		case reflect.Slice, reflect.Array:
			repeated++
		default:
			break UNWRAP
		}
		t = t.Elem()
	}
	kind := t.Kind()
	switch {
	case reflect.Int <= kind && kind <= reflect.Uint64:
		schema = intSchema
	case reflect.Float32 <= kind && kind <= reflect.Float64:
		schema = doubleSchema
	case reflect.String == kind:
		schema = stringSchema
	case reflect.Struct == kind:
		schema = q.SwaggerSchemaFromStruct(t, omit_url_parameters)
	default:
		panic(fmt.Errorf("unsupported Type kind[%v] for swagger schema", kind))
	}
	for i := 0; i < repeated; i++ {
		schema = arraySchemaWrap(schema)
	}
	if repeated == 0 {
		schema.Nullable = nullable
	}
	return
}

func (q *Quark) SwaggerSchemaFromStruct(t reflect.Type, omit_url_parameters bool) (schema spec.Schema) {
	if !omit_url_parameters && t.Name() != "" {
		ref, _ := spec.NewRef("#/definitions/" + t.Name())
		schema = spec.Schema{
			SchemaProps: spec.SchemaProps{
				Ref: ref,
			},
		}
		q.addModel(t)
		return
	}
	schema.Properties = make(spec.SchemaProperties)
	for i := 0; i < t.NumField(); i++ {
		f := t.Field(i)
		if omit_url_parameters && types.IsUrlType(f.Type) {
			continue
		}
		nullable := false
		typ := f.Type
		for typ.Kind() == reflect.Ptr {
			typ = typ.Elem()
			nullable = true
		}
		if taf, ok := reservedStructTypes[typ]; ok {
			schema.Properties[f.Name] = spec.Schema{
				SchemaProps: spec.SchemaProps{
					Type:     spec.StringOrArray{taf.T},
					Format:   taf.F,
					Nullable: nullable,
				},
			}
			continue
		}
		var sub spec.Schema = q.SwaggerSchemaFromType(typ, false)
		schema.Properties[f.Name] = sub
	}
	return
}

func arraySchemaWrap(schema spec.Schema) spec.Schema {
	return spec.Schema{
		SchemaProps: spec.SchemaProps{
			Type: []string{"array"},
			Items: &spec.SchemaOrArray{
				Schema: &schema,
			},
		},
	}
}
