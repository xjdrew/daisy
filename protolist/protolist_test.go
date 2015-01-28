package protolist

import (
	"fmt"
	"io/ioutil"
	"os"
	"strings"
	"testing"
)

func TestCamelCase(t *testing.T) {
	var m = map[string]string{
		"lua_out":            "LuaOut",
		"query_hirable_hero": "QueryHirableHero",
	}
	for k, v := range m {
		v1 := camelCase(k)
		if v != v1 {
			t.Fatalf("camel case %s failed: %s", k, v1)
		}
	}
}

func TestSplitData(t *testing.T) {
	var lines = []string{
		"test {",
		"service1 = 1",
		"} test2 {} test3 {",
		"service2 = 2",
		"}\n",
	}

	data := strings.Join(lines, "# comment 1\r\n # comment 2 \r\n")
	result := splitData(data)
	if len(lines) != len(result) {
		t.Errorf("split data error, lines change:%d -> %d", len(lines), len(result))
	}
	for i, line := range result {
		if line != strings.TrimSpace(lines[i]) {
			t.Errorf("split data error, line:%d, %s -> %s", i, lines[i], line)
		}
	}

	modules := splitModules(result)
	n := strings.Count(data, "}")
	m := 0
	for _, line := range modules {
		if line == "}" {
			m += 1
		}
		t.Logf("%v", line)
	}
	if n != m {
		t.Errorf("split modules failed:%d -> %d", n, m)
	}
}

type Case struct {
	lines  []string
	module Module
}

var case1 = Case{
	lines: []string{
		"test {",
		"service1 = 1",
		"service2:input1 = 2",
		"service3:input1[] = 3",
		"service4:input1[output1]=4",
		"service5:[ output1 ]=5",
		"service6:[]=6",
		"}",
	},
	module: Module{
		Name: "test",
		Services: []Service{
			{1, "service1", "proto.test.Service1", "proto.test.Service1.Response"},
			{2, "service2", "proto.test.Input1", "proto.test.Input1.Response"},
			{3, "service3", "proto.test.Input1", ""},
			{4, "service4", "proto.test.Input1", "proto.test.Output1"},
			{5, "service5", "proto.test.Service5", "proto.test.Output1"},
			{6, "service6", "proto.test.Service6", ""},
		},
	},
}

var case2 = Case{
	lines: []string{
		"test1 {",
		"service11 = 11",
		"service12:input1 = 12",
		"service13:input1[] = 13",
		"service14:input1[output1]=14",
		"service15:[ output1 ]=15",
		"service16:[]=16",
		"}",
	},
	module: Module{
		Name: "test1",
		Services: []Service{
			{11, "service11", "proto.test1.Service11", "proto.test1.Service11.Response"},
			{12, "service12", "proto.test1.Input1", "proto.test1.Input1.Response"},
			{13, "service13", "proto.test1.Input1", ""},
			{14, "service14", "proto.test1.Input1", "proto.test1.Output1"},
			{15, "service15", "proto.test1.Service15", "proto.test1.Output1"},
			{16, "service16", "proto.test1.Service16", ""},
		},
	},
}

var cases = []Case{case1, case2}

func checkService(s1, s2 Service) bool {
	if s1.Id != s2.Id || s1.Name != s2.Name || s1.Input != s2.Input || s1.Output != s2.Output {
		fmt.Printf("service:%+v -> %+v\n", s1, s2)
		return false
	}
	return true
}

func checkModule(m1, m2 Module) bool {
	if m1.Name != m2.Name {
		return false
	}
	if len(m1.Services) != len(m2.Services) {
		return false
	}
	for i := 0; i < len(m1.Services); i++ {
		if !checkService(m1.Services[i], m2.Services[i]) {
			return false
		}
	}
	return true
}

func TestParseModule(t *testing.T) {
	for _, c := range cases {
		lines := c.lines
		module := c.module
		data := strings.Join(lines, "# test \r\n")
		m, err := ParseData(data)
		if err != nil {
			t.Fatal(err)
		}
		// t.Errorf("%+v", m)
		if !checkModule(m[0], module) {
			t.Fatalf("parse module failed, %+v ------------> %+v", module, m)
		}
	}
}

func TestParseFile(t *testing.T) {
	tmpfile, err := ioutil.TempFile(os.TempDir(), "proto")
	if err != nil {
		t.Fatal(err)
	}
	var all []string
	for _, c := range cases {
		all = append(all, strings.Join(c.lines, "# test\n"))
	}

	data := strings.Join(all, "\n")
	if _, err := tmpfile.Write([]byte(data)); err != nil {
		t.Fatal(err)
	}
	filePath := tmpfile.Name()
	tmpfile.Close()
	m, err := ParseFile(filePath)
	if err != nil {
		t.Fatal(err)
	}
	if len(m) != len(cases) {
		t.Fatalf("modules less than expected:%d -> %d", len(cases), len(m))
	}
	for i, c := range cases {
		if !checkModule(c.module, m[i]) {
			t.Fatalf("parse module failed, %+v ------------> %+v", c.module, m[i])
		}
	}
	os.Remove(filePath)
}
