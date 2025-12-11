/*
Copyright 2015 The Kubernetes Authors.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package template

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"math/rand" // #nosec
	"net"
	"os"
	"regexp"
	"strings"
	text_template "text/template"
	"time"

	"bytetrade.io/web3os/bfl/internal/ingress/controllers/config"

	"github.com/pkg/errors"
	"k8s.io/klog/v2"
)

const (
	slash                   = "/"
	defBufferSize           = 65535
	writeIndentOnEmptyLines = true // backward-compatibility
)

const (
	stateCode = iota
	stateComment
)

// TemplateWriter is the interface to render a template
type TemplateWriter interface {
	Write(conf config.TemplateConfig) ([]byte, error)
}

// Template ...
type Template struct {
	tmpl *text_template.Template
	//fw   watch.FileWatcher
	bp *BufferPool
}

// NewTemplate returns a new Template instance or an
// error if the specified template file contains errors
func NewTemplate(file string) (*Template, error) {
	data, err := ioutil.ReadFile(file)
	if err != nil {
		return nil, errors.Wrapf(err, "unexpected error reading template %v", file)
	}

	tmpl, err := text_template.New("nginx.tmpl").Funcs(funcMap).Parse(string(data))
	if err != nil {
		return nil, err
	}

	return &Template{
		tmpl: tmpl,
		bp:   NewBufferPool(defBufferSize),
	}, nil
}

// 1. Removes carriage return symbol (\r)
// 2. Collapses multiple empty lines to single one
// 3. Re-indent
// (ATW: always returns nil)
func cleanConf(in *bytes.Buffer, out *bytes.Buffer) error {
	depth := 0
	lineStarted := false
	emptyLineWritten := false
	state := stateCode
	for {
		c, err := in.ReadByte()
		if err != nil {
			if err == io.EOF {
				return nil
			}
			return err // unreachable
		}

		needOutput := false
		nextDepth := depth
		nextLineStarted := lineStarted

		switch state {
		case stateCode:
			switch c {
			case '{':
				needOutput = true
				nextDepth = depth + 1
				nextLineStarted = true
			case '}':
				needOutput = true
				depth--
				nextDepth = depth
				nextLineStarted = true
			case ' ', '\t':
				needOutput = lineStarted
			case '\r':
			case '\n':
				needOutput = !(!lineStarted && emptyLineWritten)
				nextLineStarted = false
			case '#':
				needOutput = true
				nextLineStarted = true
				state = stateComment
			default:
				needOutput = true
				nextLineStarted = true
			}
		case stateComment:
			switch c {
			case '\r':
			case '\n':
				needOutput = true
				nextLineStarted = false
				state = stateCode
			default:
				needOutput = true
			}
		}

		if needOutput {
			if !lineStarted && (writeIndentOnEmptyLines || c != '\n') {
				for i := 0; i < depth; i++ {
					err = out.WriteByte('\t') // always nil
					if err != nil {
						return err
					}
				}
			}
			emptyLineWritten = !lineStarted
			err = out.WriteByte(c) // always nil
			if err != nil {
				return err
			}
		}

		depth = nextDepth
		lineStarted = nextLineStarted
	}
}

func (t *Template) Write(conf config.TemplateConfig) ([]byte, error) {
	tmplBuf := t.bp.Get()
	defer t.bp.Put(tmplBuf)

	outCmdBuf := t.bp.Get()
	defer t.bp.Put(outCmdBuf)

	if klog.V(3).Enabled() {
		b, err := json.Marshal(conf)
		if err != nil {
			klog.Errorf("unexpected error: %v", err)
		}
		klog.InfoS("NGINX", "configuration", string(b))
	}

	err := t.tmpl.Execute(tmplBuf, conf)
	if err != nil {
		return nil, err
	}

	// squeezes multiple adjacent empty lines to be single
	// spaced this is to avoid the use of regular expressions
	err = cleanConf(tmplBuf, outCmdBuf)
	if err != nil {
		return nil, err
	}

	return outCmdBuf.Bytes(), nil
}

var (
	funcMap = text_template.FuncMap{
		"empty": func(input interface{}) bool {
			check, ok := input.(string)
			if ok {
				return len(check) == 0
			}
			return true
		},
		"getenv":                os.Getenv,
		"contains":              strings.Contains,
		"split":                 strings.Split,
		"hasPrefix":             strings.HasPrefix,
		"hasSuffix":             strings.HasSuffix,
		"trimSpace":             strings.TrimSpace,
		"toUpper":               strings.ToUpper,
		"toLower":               strings.ToLower,
		"formatIP":              formatIP,
		"quote":                 quote,
		"isValidByteSize":       isValidByteSize,
		"buildForwardedFor":     buildForwardedFor,
		"buildHTTPListener":     buildHTTPListener,
		"buildHTTPSListener":    buildHTTPSListener,
		"buildServerName":       buildServerName,
		"buildNonAppServerName": buildNonAppServerName,
	}
)

// escapeLiteralDollar will replace the $ character with ${literal_dollar}
// which is made to work via the following configuration in the http section of
// the template:
//
//	geo $literal_dollar {
//	    default "$";
//	}
func escapeLiteralDollar(input interface{}) string {
	inputStr, ok := input.(string)
	if !ok {
		return ""
	}
	return strings.Replace(inputStr, `$`, `${literal_dollar}`, -1)
}

// formatIP will wrap IPv6 addresses in [] and return IPv4 addresses
// without modification. If the input cannot be parsed as an IP address
// it is returned without modification.
func formatIP(input string) string {
	ip := net.ParseIP(input)
	if ip == nil {
		return input
	}
	if v4 := ip.To4(); v4 != nil {
		return input
	}
	return fmt.Sprintf("[%s]", input)
}

func quote(input interface{}) string {
	var inputStr string
	switch input := input.(type) {
	case string:
		inputStr = input
	case fmt.Stringer:
		inputStr = input.String()
	case *string:
		inputStr = *input
	default:
		inputStr = fmt.Sprintf("%v", input)
	}
	return fmt.Sprintf("%q", inputStr)
}

var (
	denyPathSlugMap = map[string]string{}
)

// buildDenyVariable returns a nginx variable for a location in a
// server to be used in the whitelist check
// This method uses a unique id generator library to reduce the
// size of the string to be used as a variable in nginx to avoid
// issue with the size of the variable bucket size directive
func buildDenyVariable(a interface{}) string {
	l, ok := a.(string)
	if !ok {
		klog.Errorf("expected a 'string' type but %T was returned", a)
		return ""
	}

	if _, ok := denyPathSlugMap[l]; !ok {
		denyPathSlugMap[l] = randomString()
	}

	return fmt.Sprintf("$deny_%v", denyPathSlugMap[l])
}

// refer to http://nginx.org/en/docs/syntax.html
// Nginx differentiates between size and offset
// offset directives support gigabytes in addition
var nginxSizeRegex = regexp.MustCompile("^[0-9]+[kKmM]{0,1}$")
var nginxOffsetRegex = regexp.MustCompile("^[0-9]+[kKmMgG]{0,1}$")

// isValidByteSize validates size units valid in nginx
// http://nginx.org/en/docs/syntax.html
func isValidByteSize(input interface{}, isOffset bool) bool {
	s, ok := input.(string)
	if !ok {
		klog.Errorf("expected an 'string' type but %T was returned", input)
		return false
	}

	s = strings.TrimSpace(s)
	if s == "" {
		klog.V(2).Info("empty byte size, hence it will not be set")
		return false
	}

	if isOffset {
		return nginxOffsetRegex.MatchString(s)
	}

	return nginxSizeRegex.MatchString(s)
}

func buildForwardedFor(input interface{}) string {
	s, ok := input.(string)
	if !ok {
		klog.Errorf("expected a 'string' type but %T was returned", input)
		return ""
	}

	ffh := strings.Replace(s, "-", "_", -1)
	ffh = strings.ToLower(ffh)
	return fmt.Sprintf("$http_%v", ffh)
}

var letters = []rune("abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ")

func init() {
	rand.Seed(time.Now().UnixNano())
}

func randomString() string {
	b := make([]rune, 32)
	for i := range b {
		b[i] = letters[rand.Intn(len(letters))] // #nosec
	}

	return string(b)
}

func buildHTTPListener(t interface{}, s interface{}) string {
	var out []string

	tc, ok := t.(config.TemplateConfig)
	if !ok {
		klog.Errorf("expected a 'config.TemplateConfig' type but %T was returned", t)
		return ""
	}

	hostname, ok := s.(string)
	if !ok {
		klog.Errorf("expected a 'string' type but %T was returned", s)
		return ""
	}

	addrV4 := []string{""}
	if len(tc.Cfg.BindAddressIpv4) > 0 {
		addrV4 = tc.Cfg.BindAddressIpv4
	}

	co := commonListenOptions(tc, hostname)
	out = append(out, httpListener(addrV4, co, tc)...)

	if !tc.IsIPV6Enabled {
		return strings.Join(out, "\n")
	}

	// ipv6
	addrV6 := []string{"[::]"}
	if len(tc.Cfg.BindAddressIpv6) > 0 {
		addrV6 = tc.Cfg.BindAddressIpv6
	}

	out = append(out, httpListener(addrV6, co, tc)...)

	return strings.Join(out, "\n")
}

func buildHTTPSListener(p interface{}, t interface{}, s interface{}) string {
	var out []string

	tc, ok := t.(config.TemplateConfig)
	if !ok {
		klog.Errorf("expected a 'config.TemplateConfig' type but %T was returned", t)
		return ""
	}

	hostname, ok := s.(string)
	if !ok {
		klog.Errorf("expected a 'string' type but %T was returned", s)
		return ""
	}

	port, ok := p.(int)
	if !ok {
		klog.Errorf("expected a 'int' type but %T was returned", p)
		return ""
	}

	co := commonListenOptions(tc, hostname)

	addrV4 := []string{""}
	if len(tc.Cfg.BindAddressIpv4) > 0 {
		addrV4 = tc.Cfg.BindAddressIpv4
	}

	out = append(out, httpsListener(addrV4, co, tc, port)...)

	if !tc.IsIPV6Enabled {
		return strings.Join(out, "\n")
	}

	addrV6 := []string{"[::]"}
	if len(tc.Cfg.BindAddressIpv6) > 0 {
		addrV6 = tc.Cfg.BindAddressIpv6
	}

	out = append(out, httpsListener(addrV6, co, tc, port)...)

	return strings.Join(out, "\n")
}

func commonListenOptions(template config.TemplateConfig, hostname string) string {
	var out []string

	if hostname != "_" {
		return strings.Join(out, " ")
	}

	// setup options that are valid only once per port

	out = append(out, "default_server")

	if template.Cfg.ReusePort {
		out = append(out, "reuseport")
	}

	out = append(out, fmt.Sprintf("backlog=%v", template.BacklogSize))

	return strings.Join(out, " ")
}

func httpListener(addresses []string, co string, tc config.TemplateConfig) []string {
	out := make([]string, 0)

	fn := func(address string) []string {
		lo := []string{"listen"}

		if address == "" {
			lo = append(lo, fmt.Sprintf("%v", tc.ListenPorts.HTTP))
		} else {
			lo = append(lo, fmt.Sprintf("%v:%v", address, tc.ListenPorts.HTTP))
		}

		lo = append(lo, co)
		return lo
	}

	if len(addresses) > 0 {
		for _, address := range addresses {
			lo := fn(address)
			out = append(out, strings.Join(lo, " "))
		}
	} else {
		lo := fn("")
		out = append(out, strings.Join(lo, " "))
	}

	return out
}

func httpsListener(addresses []string, co string, tc config.TemplateConfig, port int) []string {
	out := make([]string, 0)

	fn := func(address string) []string {
		lo := []string{"listen"}

		if address == "" {
			lo = append(lo, fmt.Sprintf("%v", port))
		} else {
			lo = append(lo, fmt.Sprintf("%v:%v", address, port))
		}

		lo = append(lo, co)
		lo = append(lo, "ssl")

		if tc.Cfg.UseHTTP2 {
			lo = append(lo, "http2")
		}

		return lo
	}

	if len(addresses) > 0 {
		for _, address := range addresses {
			lo := fn(address)
			out = append(out, strings.Join(lo, " "))
		}
	} else {
		lo := fn("")
		out = append(out, strings.Join(lo, " "))
	}
	return out
}

// buildServerName ensures wildcard hostnames are valid
func buildServerName(hostname string) string {
	if !strings.HasPrefix(hostname, "*") {
		return hostname
	}

	hostname = strings.Replace(hostname, "*.", "", 1)
	parts := strings.Split(hostname, ".")

	return `~^(?<subdomain>[\w-]+)\.` + strings.Join(parts, "\\.") + `$`
}

func buildNonAppServerName(t any, name string) string {
	var out []string

	tc, ok := t.(config.TemplateConfig)
	if !ok {
		klog.Errorf("expected a 'config.TemplateConfig' type but %T was returned", t)
		return ""
	}

	if name == "" {
		klog.Error("got unexpected name string, could not be empty")
		return ""
	}

	out = append(out, fmt.Sprintf("%s.%s", name, tc.UserZone))

	if tc.IsEphemeralUser {
		out = append(out, fmt.Sprintf("%s-%s.%s", name, tc.UserName, tc.UserZone))
	}
	return strings.Join(out, " ")
}
