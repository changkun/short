// Copyright 2020 Changkun Ou. All rights reserved.
// Use of this source code is governed by a GPLv3
// license that can be found in the LICENSE file.

package main

import (
	"flag"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"

	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
	"gopkg.in/yaml.v2"
)

const (
	home    = "https://changkun.de"
	home404 = "https://changkun.de/404.html"
)

func main() {
	r := gin.Default()
	r.NoRoute(func(c *gin.Context) {
		c.Redirect(http.StatusPermanentRedirect, home404)
	})
	r.GET("/s/:short", handleShortLinks)
	logrus.Infof("changkun.de/s/ is running at: http://%s", conf.Addr)
	r.Run(conf.Addr)
}

type config struct {
	Addr  string            `json:"addr"`
	Mode  string            `json:"mode"`
	Short map[string]string `json:"short"`
}

var conf *config

const help = `
short is a webservice that provides a short url service for changkun.de.
Usage:
`

func init() {
	c := flag.String("config", "", "path to config file")
	usage := func() {
		fmt.Fprintf(os.Stderr, fmt.Sprintf(help))
		flag.PrintDefaults()
	}
	flag.Usage = usage
	flag.Parse()
	f := *c
	if len(f) == 0 {
		usage()
		os.Exit(1)
	}

	y, err := ioutil.ReadFile(f)
	if err != nil {
		fmt.Fprintf(os.Stderr, "failed to read configuration file: %v", err)
		os.Exit(1)
	}

	conf = &config{}
	err = yaml.Unmarshal(y, conf)
	if err != nil {
		fmt.Fprintf(os.Stderr, "failed to parse configuration file: %v", err)
		os.Exit(1)
	}

	stat.links = map[string]*metric{}
	for k := range conf.Short {
		stat.init(k)
	}
	stat.start()
	gin.SetMode(conf.Mode)
}

func handleShortLinks(c *gin.Context) {
	s := c.Param("short")
	v, ok := conf.Short[s]
	if !ok {
		c.Redirect(http.StatusTemporaryRedirect, home404)
	}
	stat.inc(s, c.ClientIP())
	c.Redirect(http.StatusTemporaryRedirect, v)
}
