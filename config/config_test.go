package config

import (
	"testing"

	"github.com/dovejb/quark/util"
)

var (
	yamldata = []byte(`service:
  array:
    - s1
    - s2
    - s3
    - abc
  int_array:
    - 1
    - 2
    - 3
  int: 123
  float: 4.56
  bool: true
  object:
    obj_prop_a: key
    obj_prop_b: value
  blank:
extra:
  another-style-string-slice: hello,world,?`)
)

func TestYAML(t *testing.T) {
	p, e := buildPropertyTemplate(yamldata)
	if e != nil {
		t.Error(e)
	}
	t.Log(p)
}

func TestExtractWithoutArray(t *testing.T) {
	p, e := buildPropertyTemplate(yamldata)
	if e != nil {
		t.Error(e)
	}
	var conf struct {
		Service struct {
			//Array    []string
			//IntArray []int
			Int    int
			Float  float64
			Bool   bool
			Object struct {
				ObjPropA string
				ObjPropB string
			}
			Blank string
		}
		Extra struct {
			AnotherStyleStringSlice string
		}
	}
	if e := p.ExtractTo(&conf); e != nil {
		t.Error(e)
	}
	t.Log(util.Js(conf))
}

func TestConfiguration(t *testing.T) {
	var c Configuration
	if e := c.LoadTemplate("config.yaml"); e != nil {
		t.Errorf("Load config fail %v", e)
	}
	var config struct {
		Hello string
		World int
	}
	if e := c.Extract(&config); e != nil {
		t.Errorf("Extract fail %v", e)
	}
	t.Log(config)
}

func TestConfigurationWithEnv(t *testing.T) {
	var c struct {
		Host  string
		Value int
		Data  float64
	}
	cf := Configuration{}
	if e := cf.Load("config.yaml", &c); e != nil {
		t.Fatal(e)
	}
	t.Log(c)
}
