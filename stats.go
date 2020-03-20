// Copyright 2020 Changkun Ou. All rights reserved.
// Use of this source code is governed by a GPLv3
// license that can be found in the LICENSE file.

package main

import (
	"fmt"
	"os"
	"sync"
	"sync/atomic"
	"time"

	"github.com/sirupsen/logrus"
)

const (
	statsFile   = "./stats/%s.csv"
	statsFormat = "%v,%d,%d\n" // time,pv,uv
	period      = time.Minute * 10
)

var stat stats

type metric struct {
	// pv is incrementally counted
	pv uint64
	// uv is counted per period, not a global under standing,
	// it tries to avoid multiple access at a time, not aim for actual uv.
	uv  uint64
	ips *sync.Map // for O(1) access
}

// stats is a naive pv/uv statistics
type stats struct {
	links map[string]*metric // !!read-only
}

// init should only be called in a init function
func (s *stats) init(link string) {
	s.links[link] = &metric{pv: 0, uv: 0, ips: new(sync.Map)}
}

// inc increase a pv and a uv
func (s *stats) inc(link string, ip string) {
	v := s.links[link]
	atomic.AddUint64(&v.pv, 1)
	_, loaded := v.ips.LoadOrStore(ip, 0)
	if !loaded {
		atomic.AddUint64(&v.uv, 1)
	}
	logrus.Infof("record: ip %v link %v", ip, link)
}

// start dumps stats per ten period.
func (s *stats) start() {
	go func() {
		for {
			s.save()
			s.clear()
			time.Sleep(period)
		}
	}()
}

// save save current data
func (s *stats) save() {
	for link, v := range s.links {
		f, err := os.OpenFile(fmt.Sprintf(statsFile, link),
			os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
		if err != nil {
			logrus.Errorf("cannot open stats: %s.csv, err: %v", link, err)
			continue
		}
		l := fmt.Sprintf(statsFormat, time.Now().UTC(), v.pv, v.uv)
		if _, err := f.WriteString(l); err != nil {
			logrus.Errorf("cannot open stats: %s.csv", link)
			f.Close()
			continue
		}
		f.Close()
	}
}

func (s *stats) clear() {
	for _, v := range s.links {
		atomic.StoreUint64(&v.pv, 0)
		atomic.StoreUint64(&v.uv, 0)
		v.ips.Range(func(k, _ interface{}) bool {
			v.ips.Delete(k)
			return true
		})
	}
}
