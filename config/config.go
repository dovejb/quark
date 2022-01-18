package config

import (
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"reflect"
	"strconv"
	"strings"

	"github.com/dovejb/quark/util"
	"gopkg.in/yaml.v3"
)

type Configuration struct {
	template propertyTemplate
}

func (c *Configuration) Load(tmplFile string, cfgObj interface{}) error {
	if e := c.LoadTemplate(tmplFile); e != nil {
		return e
	}
	if e := c.LoadEnvVars(); e != nil {
		return e
	}
	if e := c.Extract(cfgObj); e != nil {
		return e
	}
	return nil
}

func (c *Configuration) LoadTemplate(filename string) error {
	tmpl, e := buildPropertyTemplateFromFile(filename)
	if e != nil {
		return e
	}
	c.template = *tmpl
	return nil
}

func (c Configuration) LoadEnvVars() error {
	c.template.loadEnvVars()
	return nil
}

func (c Configuration) Extract(v interface{}) error {
	return c.template.ExtractTo(v)
}

type property struct {
	path        []string
	repeated    bool
	fileValue   interface{}
	envValue    string
	hasEnvValue bool
}

func (p property) Value() interface{} {
	if p.hasEnvValue {
		return p.envValue
	}
	return p.fileValue
}

func (p property) extract_value(value reflect.Value) error {
	if !p.hasEnvValue && p.fileValue == nil {
		return nil
	}
	v := value
	for v.Type().Kind() == reflect.Ptr {
		v.Set(reflect.New(v.Type()))
		v = v.Elem()
	}
	if p.hasEnvValue {
		return p.extract_env_value(v)
	} else {
		return p.extract_file_value(v)
	}
}

func (p property) extract_env_value(value reflect.Value) error {
	fmt.Println("in extract_env_value", p.path, p.hasEnvValue, p.envValue)
	if !p.hasEnvValue {
		return nil
	}
	kind := value.Type().Kind()
	switch {
	case reflect.Int <= kind && kind <= reflect.Int64:
		i, e := strconv.ParseInt(p.envValue, 10, 64)
		if e != nil {
			return fmt.Errorf("invalid env value(%s) as int, %v", p.envValue, e)
		}
		value.SetInt(i)
	case reflect.Uint <= kind && kind <= reflect.Uintptr:
		u, e := strconv.ParseUint(p.envValue, 10, 64)
		if e != nil {
			return fmt.Errorf("invalid env value(%s) as uint, %v", p.envValue, e)
		}
		value.SetUint(u)
	case reflect.Float32 == kind || reflect.Float64 == kind:
		f, e := strconv.ParseFloat(p.envValue, 64)
		if e != nil {
			return fmt.Errorf("invalid env value(%s) as float, %v", p.envValue, e)
		}
		value.SetFloat(f)
	case reflect.String == kind:
		value.SetString(p.envValue)
	case reflect.Slice == kind || reflect.Array == kind: //TODO
		return fmt.Errorf("slice or array is not supported in env vars extraction now")
	case reflect.Bool == kind:
		if strings.ToLower(p.envValue) == "true" {
			value.SetBool(true)
		} else {
			value.SetBool(false)
		}
	default:
		return fmt.Errorf("unsupported property type kind[%v]", kind)
	}
	return nil
}

func (p property) extract_file_value(value reflect.Value) error {
	kind := value.Type().Kind()
	switch {
	case reflect.Bool == kind:
		value.SetBool(p.fileValue.(bool))
	case reflect.Int <= kind && kind <= reflect.Int64:
		value.SetInt(int64(p.fileValue.(int)))
	case reflect.Uint <= kind && kind <= reflect.Uintptr:
		value.SetUint(uint64(p.fileValue.(int)))
	case reflect.Float32 == kind || reflect.Float64 == kind:
		value.SetFloat(p.fileValue.(float64))
	case reflect.String == kind:
		value.SetString((p.fileValue.(string)))
	case reflect.Slice == kind || reflect.Array == kind: //TODO
		return fmt.Errorf("slice or array is not supported in file vars extraction now")
	default:
		return fmt.Errorf("unsupported property type kind[%v]", kind)
	}
	return nil
}

func (p *propertyTemplate) ExtractTo(x interface{}) error {
	v := reflect.ValueOf(x)
	if v.Kind() != reflect.Ptr {
		return fmt.Errorf("extract argument must be a pointer of struct")
	}
	v = v.Elem()
	return p.extract_to(v, p.items, p.trie, make([]string, 0))
}

func (p *propertyTemplate) extract_to(x reflect.Value, ps []property, trie util.Trie, prefix []string) error {
	for i := 0; i < x.NumField(); i++ {
		fv := x.Field(i)
		ft := x.Type().Field(i)
		fp := util.PascalSplit(ft.Name)
		next, found := trie.Find(fp)
		if !found {
			continue
		}
		if pos := next.Value(); pos == nil { // path exists partially
			if ft.Type.Kind() == reflect.Struct {
				if e := p.extract_to(fv, ps, next, append(prefix, fp...)); e != nil {
					return e
				}
			} else if ft.Type.Kind() == reflect.Ptr && ft.Type.Elem().Kind() == reflect.Struct {
				fv.Set(reflect.New(ft.Type))
				if e := p.extract_to(fv.Elem(), ps, next, append(prefix, fp...)); e != nil {
					return e
				}
			} else {
				return fmt.Errorf("property[%s.%s] wants object in template, but receiver is %v", strings.Join(prefix, "_"), strings.Join(fp, "_"), ft.Type)
			}
		} else { // property exists
			prop := ps[*pos]
			if e := prop.extract_value(fv); e != nil {
				return fmt.Errorf("property[%s.%s] extract value fail, value is [%v] but receiver type is [%v], %v", strings.Join(prefix, "_"), strings.Join(fp, "_"), prop.Value(), ft.Type, e)
			}
		}
	}
	return nil
}

func (p *propertyTemplate) loadEnvVars() {
	for _, ev := range os.Environ() {
		pair := strings.SplitN(ev, "=", 2)
		propertyPath := standardSplit(pair[0])
		if t, found := p.trie.Find(propertyPath); found {
			pos := t.Value() // found means having the path, not sure yet about having a value
			if pos != nil {
				p.items[*pos].envValue = pair[1]
				p.items[*pos].hasEnvValue = true
			}
		}
	}
}

type propertyTemplate struct {
	items []property
	trie  util.Trie
}

func buildPropertyTemplateFromFile(yaml_path string) (*propertyTemplate, error) {
	f, e := os.Open(yaml_path)
	if e != nil {
		return nil, e
	}
	defer f.Close()
	content, _ := ioutil.ReadAll(f)
	return buildPropertyTemplate(content)
}

func buildPropertyTemplate(yaml_content []byte) (*propertyTemplate, error) {
	var m map[string]interface{}
	if e := yaml.Unmarshal(yaml_content, &m); e != nil {
		return nil, e
	}
	tmpl := &propertyTemplate{
		items: make([]property, 0, 8),
		trie:  util.NewTrie(),
	}
	if e := buildTemplate(tmpl, m, make([]string, 0)); e != nil {
		return nil, e
	}
	return tmpl, nil
}

func buildTemplate(tmpl *propertyTemplate, m map[string]interface{}, prefix []string) error {
	for k, value := range m {
		ks := standardSplit(k)
		prefix = append(prefix, ks...)
		switch v := value.(type) {
		case map[string]interface{}:
			buildTemplate(tmpl, v, prefix)
		case []interface{}:
			if sub, found := tmpl.trie.Find(prefix); found && !sub.Empty() {
				log.Printf("WARN: template has duplicated property names! %s", strings.Join(prefix, "_"))
				continue
			}
			tmpl.items = append(tmpl.items, property{
				path:      copySlice(prefix),
				repeated:  true,
				fileValue: v,
			})
			tmpl.trie.Add(prefix, len(tmpl.items)-1)
		default:
			if sub, found := tmpl.trie.Find(prefix); found && !sub.Empty() {
				log.Printf("WARN: template has duplicated property names! %s", strings.Join(prefix, "_"))
				continue
			}
			tmpl.items = append(tmpl.items, property{
				path:      copySlice(prefix),
				fileValue: v,
			})
			tmpl.trie.Add(prefix, len(tmpl.items)-1)
		}
		prefix = prefix[:len(prefix)-len(ks)]
	}
	return nil
}

//split by _.- all ToLower
func standardSplit(s string) []string {
	return strings.FieldsFunc(strings.ToLower(s), func(r rune) bool {
		if r == '_' || r == '.' || r == '-' {
			return true
		}
		return false
	})
}

func copySlice(s []string) []string {
	b := make([]string, len(s))
	copy(b, s)
	return b
}
