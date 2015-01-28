package protolist

import (
	"fmt"
	"os"
	"regexp"
	"strconv"
	"strings"
)

// exported struct
type Service struct {
	Id     int
	Name   string
	Input  string
	Output string
}

type Module struct {
	Name     string
	Services []Service
}

var ProtoPrefixEscape = "."
var ProtoPrefix = "proto"
var ProtoResponse = ".Response"

// exported interfaces
func ParseFile(path string) ([]Module, error) {
	var file *os.File
	var err error
	if file, err = os.Open(path); err != nil {
		return nil, err
	}

	var sz int64
	if sz, err = file.Seek(0, 2); err != nil {
		return nil, err
	}

	if _, err = file.Seek(0, 0); err != nil {
		return nil, err
	}

	data := make([]byte, sz)
	if _, err = file.Read(data); err != nil {
		return nil, err
	}
	return ParseData(string(data))
}

func ParseData(data string) ([]Module, error) {
	a := splitModules(splitData(data))
	mcount := 0
	for _, l := range a {
		if l == MODULE_END {
			mcount += 1
		}
	}

	var d decodeState
	d.init(a)

	modules := make([]Module, mcount)
	for i := 0; i < mcount; i++ {
		module, err := d.nextModule()
		if err != nil {
			return nil, err
		}
		modules[i] = *module
	}

	if !d.eof() {
		return nil, fmt.Errorf("unexpected tail:%s", d.lines[d.off])
	}

	if err := normalizeModules(modules); err != nil {
		return nil, err
	}
	return modules, nil
}

// internal implemention
// split data line by line and trim comments
const (
	COMMENT      = "#"
	NEWLINE      = "\r\n"
	MODULE_START = "{"
	MODULE_END   = "}"
)

// camel case
func camelCase(src string) string {
	a := strings.Split(src, "_")
	for i, v := range a {
		a[i] = strings.Title(v)
	}
	return strings.Join(a, "")
}

func addPrefix(str, prefix string) string {
	if str[:1] == ProtoPrefixEscape {
		return str[1:]
	} else {
		return strings.Join([]string{ProtoPrefix, prefix, camelCase(str)}, ".")
	}
}

func normalizeModule(module *Module) {
	for i, _ := range module.Services {
		service := &module.Services[i]
		if service.Input == "" {
			service.Input = addPrefix(service.Name, module.Name)
		} else {
			service.Input = addPrefix(service.Input, module.Name)
		}
		if service.Output == "" {
			service.Output = service.Input + ProtoResponse
		} else if service.Output == "[]" {
			service.Output = ""
		} else {
			service.Output = addPrefix(service.Output[1:len(service.Output)-1], module.Name)
		}
	}
}

func normalizeModules(modules []Module) error {
	moduleMap := make(map[string]bool)
	idMap := make(map[int]bool)
	serviceMap := make(map[string]bool)
	for i, _ := range modules {
		module := &(modules[i])
		if moduleMap[module.Name] {
			return fmt.Errorf("repeated module name:%s", module.Name)
		}
		moduleMap[module.Name] = true

		for _, service := range module.Services {
			if idMap[service.Id] || serviceMap[service.Name] {
				return fmt.Errorf("repeated service:(%s:%d)", service.Name, service.Id)
			}
			idMap[service.Id] = true
			serviceMap[service.Name] = true
		}
		normalizeModule(module)
	}
	return nil
}

// append if s is not empty
func appendString(a []string, s string) []string {
	if s = strings.TrimSpace(s); len(s) > 0 {
		a = append(a, s)
	}
	return a
}

// strip all comment
func trimComment(line string) string {
	if n := strings.Index(line, COMMENT); n != -1 {
		line = line[0:n]
	}
	return line
}

func splitData(data string) []string {
	var a []string
	for {
		pos := strings.IndexAny(data, NEWLINE)
		if pos == -1 {
			a = appendString(a, trimComment(data))
			break
		}

		if pos > 0 {
			a = appendString(a, trimComment(data[0:pos]))
		}
		data = data[pos+1:]
	}
	return a
}

// keep "{" and "}" as a seprate line
func splitLines(lines []string, sep string) []string {
	var a []string
	for _, line := range lines {
		data := line
		for {
			n := strings.Index(data, sep)
			if n == -1 {
				a = appendString(a, data)
				break
			}
			if n > 0 {
				a = appendString(a, data[0:n])
			}
			a = appendString(a, data[n:n+1])
			if n < len(data)-1 {
				data = data[n+1:]
			} else {
				break
			}
		}
	}
	return a
}

func splitModules(lines []string) []string {
	a := splitLines(lines, MODULE_START)
	return splitLines(a, MODULE_END)
}

func isWordElem(c rune) bool {
	if c >= 'a' && c <= 'z' {
		return true
	}
	if c >= 'A' && c <= 'Z' {
		return true
	}
	if c >= '0' && c <= '9' {
		return true
	}
	return false
}

func isSpace(c rune) bool {
	return c == ' ' || c == '\t'
}

func checkWord(data string) bool {
	data = strings.TrimSpace(data)
	for _, c := range data {
		if !isWordElem(c) {
			return false
		}
	}
	return true
}

// name = id
// name:input = id
// name:input[] = id
// name:input[output] = id
// name:[output] = id
// name:[] = id
var serviceRegex = regexp.MustCompile("^\\s*([a-zA-Z0-9\\._]+)\\s*(?:\\:\\s*([a-zA-Z0-9\\._]*)\\s*(\\[\\s*[a-zA-Z0-9\\._]*\\s*\\])?)?\\s*=\\s*([0-9]+)\\s*$")

func parseService(data string) (s Service, err error) {
	sections := serviceRegex.FindStringSubmatch(data)
	if sections == nil || len(sections) != 5 {
		err = fmt.Errorf("invalid service(%s)", data)
		return
	}
	var id uint64
	if id, err = strconv.ParseUint(sections[4], 0, 32); err != nil {
		err = fmt.Errorf("invalid service id(%s) in line %s", sections[3], data)
		return
	}
	s.Id = int(id)
	s.Name = sections[1]
	s.Input = sections[2]
	// normalize output
	// "" or "[]" or "[output]"
	s.Output = strings.Map(func(r rune) rune {
		if isSpace(r) {
			return -1
		}
		return r
	}, sections[3])
	return
}

type decodeState struct {
	lines []string
	off   int
	state int
}

func (d *decodeState) init(lines []string) {
	d.lines = lines
	d.off = 0
}

func (d *decodeState) scanLine(line string) int {
	for i := d.off; i < len(d.lines); i++ {
		if d.lines[i] == line {
			return i
		}
	}
	return -1
}

func (d *decodeState) nextModule() (m *Module, err error) {
	if d.off >= len(d.lines) {
		err = fmt.Errorf("eof")
		return
	}
	mstart := d.scanLine(MODULE_START)
	mend := d.scanLine(MODULE_END)
	name := d.lines[d.off]
	if mstart >= mend || mstart != d.off+1 || !checkWord(name) {
		err = fmt.Errorf("illegal module struct:%s", name)
		return
	}

	m = new(Module)
	m.Name = name
	m.Services = make([]Service, mend-mstart-1)
	var service Service
	for i := mstart + 1; i < mend; i++ {
		service, err = parseService(d.lines[i])
		if err != nil {
			return
		}
		m.Services[i-mstart-1] = service
	}
	d.off = mend + 1
	return
}

func (d *decodeState) eof() bool {
	return d.off == len(d.lines)
}
