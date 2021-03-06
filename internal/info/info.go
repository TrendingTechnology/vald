//
// Copyright (C) 2019-2021 vdaas.org vald team <vald@vdaas.org>
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//    https://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.
//

// Package info provides build-time info
package info

import (
	"fmt"
	"reflect"
	"runtime"
	"sort"
	"strconv"
	"strings"
	"sync"

	"github.com/vdaas/vald/internal/log"
)

// Detail represents environment information of system and stacktrace information.
type Detail struct {
	Version           string       `json:"vald_version,omitempty" yaml:"vald_version,omitempty"`
	ServerName        string       `json:"server_name,omitempty" yaml:"server_name,omitempty"`
	GitCommit         string       `json:"git_commit,omitempty" yaml:"git_commit,omitempty"`
	BuildTime         string       `json:"build_time,omitempty" yaml:"build_time,omitempty"`
	GoVersion         string       `json:"go_version,omitempty" yaml:"go_version,omitempty"`
	GoOS              string       `json:"go_os,omitempty" yaml:"go_os,omitempty"`
	GoArch            string       `json:"go_arch,omitempty" yaml:"go_arch,omitempty"`
	CGOEnabled        string       `json:"cgo_enabled,omitempty" yaml:"cgo_enabled,omitempty"`
	NGTVersion        string       `json:"ngt_version,omitempty" yaml:"ngt_version,omitempty"`
	BuildCPUInfoFlags []string     `json:"build_cpu_info_flags,omitempty" yaml:"build_cpu_info_flags,omitempty"`
	StackTrace        []StackTrace `json:"stack_trace,omitempty" yaml:"stack_trace,omitempty"`
	PrepOnce          sync.Once    `json:"-" yaml:"-"`
}

// StackTrace represents stacktrace information about url, function name, file, line ..etc.
type StackTrace struct {
	URL      string `json:"url,omitempty" yaml:"url,omitempty"`
	FuncName string `json:"function_name,omitempty" yaml:"func_name,omitempty"`
	File     string `json:"file,omitempty" yaml:"file,omitempty"`
	Line     int    `json:"line,omitempty" yaml:"line,omitempty"`
}

var (
	Version      = "v0.0.1"
	GitCommit    = "master"
	Organization = "vdaas"
	Repository   = "vald"
	BuildTime    = ""

	GoVersion  string
	GoOS       string
	GoArch     string
	CGOEnabled string

	NGTVersion string

	BuildCPUInfoFlags string

	reps = strings.NewReplacer("_", " ", ",omitempty", "")

	once sync.Once

	detail Detail
)

// String calls String method of global detail object.
func String() string {
	return detail.String()
}

// Get calls Get method of global detail object.
func Get() Detail {
	return detail.Get()
}

// String returns summary of Detail object.
func (d Detail) String() string {
	if len(d.StackTrace) == 0 {
		d = d.Get()
	}
	d.Version = log.Bold(d.Version)
	maxlen, l := 0, 0
	rt, rv := reflect.TypeOf(d), reflect.ValueOf(d)
	info := make(map[string]string, rt.NumField())
	for i := 0; i < rt.NumField(); i++ {
		v := rv.Field(i).Interface()
		value, ok := v.(string)
		if !ok {
			sts, ok := v.([]StackTrace)
			if ok {
				tag := reps.Replace(rt.Field(i).Tag.Get("json"))
				l = len(tag) + 2
				if maxlen < l {
					maxlen = l
				}
				urlMaxLen := 0
				for _, st := range sts {
					ul := len(st.URL)
					if urlMaxLen < ul {
						urlMaxLen = ul
					}
				}
				urlFormat := fmt.Sprintf("%%-%ds\t%%s", urlMaxLen)
				for i, st := range sts {
					info[fmt.Sprintf("%s-%d", tag, i)] = fmt.Sprintf(urlFormat, st.URL, st.FuncName)
				}
			} else {
				strs, ok := v.([]string)
				if ok {
					tag := reps.Replace(rt.Field(i).Tag.Get("json"))
					l = len(tag)
					if maxlen < l {
						maxlen = l
					}
					info[tag] = fmt.Sprintf("%v", strs)
				}
			}
			continue
		}
		tag := reps.Replace(rt.Field(i).Tag.Get("json"))
		l = len(tag)
		if maxlen < l {
			maxlen = l
		}
		info[tag] = value
	}

	infoFormat := fmt.Sprintf("%%-%ds ->\t%%s", maxlen)
	strs := make([]string, 0, rt.NumField())
	for tag, value := range info {
		if len(tag) != 0 && len(value) != 0 {
			strs = append(strs, fmt.Sprintf(infoFormat, tag, value))
		}
	}
	sort.Strings(strs)
	return "\n" + strings.Join(strs, "\n")
}

// Get returns parased Detail object.
func (d Detail) Get() Detail {
	d.prepare()
	valdRepo := fmt.Sprintf("github.com/%s/%s", Organization, Repository)
	defaultURL := fmt.Sprintf("https://%s/tree/%s", valdRepo, d.GitCommit)

	d.StackTrace = make([]StackTrace, 0, 10)
	for i := 3; ; i++ {
		pc, file, line, ok := runtime.Caller(i)
		if !ok {
			break
		}
		funcName := runtime.FuncForPC(pc).Name()
		if funcName == "runtime.main" {
			break
		}
		url := defaultURL
		switch {
		case strings.HasPrefix(file, runtime.GOROOT()+"/src"):
			url = fmt.Sprintf("https://github.com/golang/go/blob/%s%s#L%d", d.GoVersion, strings.TrimPrefix(file, runtime.GOROOT()), line)
		case strings.Contains(file, "go/pkg/mod/"):
			url = "https:/"
			for _, path := range strings.Split(strings.SplitN(file, "go/pkg/mod/", 2)[1], "/") {
				if strings.Contains(path, "@") {
					sv := strings.SplitN(path, "@", 2)
					if strings.Count(sv[1], "-") > 2 {
						path = sv[0] + "/blob/master"
					} else {
						path = sv[0] + "/blob/" + sv[1]
					}
				}
				url += "/" + path
			}
			url += "#L" + strconv.Itoa(line)
		case strings.Contains(file, "go/src/") && strings.Contains(file, valdRepo):
			url = strings.Replace(strings.SplitN(file, "go/src/", 2)[1]+"#L"+strconv.Itoa(line), valdRepo, "https://"+valdRepo+"/blob/"+d.GitCommit, -1)
		}
		d.StackTrace = append(d.StackTrace, StackTrace{
			FuncName: funcName,
			File:     file,
			Line:     line,
			URL:      url,
		})
	}
	return d
}

func (d *Detail) prepare() {
	d.PrepOnce.Do(func() {
		if len(d.GitCommit) == 0 {
			d.GitCommit = "master"
		}
		if len(Version) == 0 && len(d.Version) == 0 {
			d.Version = GitCommit
		}
		if len(d.BuildTime) == 0 {
			d.BuildTime = BuildTime
		}
		if len(d.GoVersion) == 0 {
			d.GoVersion = runtime.Version()
		}
		if len(d.GoOS) == 0 {
			d.GoOS = runtime.GOOS
		}
		if len(d.GoArch) == 0 {
			d.GoArch = runtime.GOARCH
		}
		if len(d.CGOEnabled) == 0 && len(CGOEnabled) != 0 {
			d.CGOEnabled = CGOEnabled
		}
		if len(d.NGTVersion) == 0 && len(NGTVersion) != 0 {
			d.NGTVersion = NGTVersion
		}
		if len(d.BuildCPUInfoFlags) == 0 && len(BuildCPUInfoFlags) != 0 {
			d.BuildCPUInfoFlags = strings.Split(strings.TrimSpace(BuildCPUInfoFlags), " ")
		}
	})
}

// Init initializes Detail object only once.
func Init(name string) {
	once.Do(func() {
		detail = Detail{
			Version:           Version,
			ServerName:        name,
			GitCommit:         GitCommit,
			BuildTime:         BuildTime,
			GoVersion:         GoVersion,
			GoOS:              GoOS,
			GoArch:            GoArch,
			CGOEnabled:        CGOEnabled,
			NGTVersion:        NGTVersion,
			BuildCPUInfoFlags: strings.Split(strings.TrimSpace(BuildCPUInfoFlags), " "),
		}
		detail.prepare()
	})
}

func (s StackTrace) String() string {
	return fmt.Sprintf("URL: %s\tFile: %s\tLine: #%d\tFuncName: %s", s.URL, s.File, s.Line, s.FuncName)
}
