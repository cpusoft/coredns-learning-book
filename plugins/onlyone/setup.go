package onlyone

import (
	"errors"
	"fmt"
	"math/rand"
	"strings"

	"github.com/coredns/caddy"
	"github.com/coredns/coredns/core/dnsserver"
	"github.com/coredns/coredns/plugin"
	clog "github.com/coredns/coredns/plugin/pkg/log"
	_ "github.com/cpusoft/goutil/conf"
	_ "github.com/cpusoft/goutil/logs"
	"github.com/miekg/dns"
)

var log = clog.NewWithPlugin("onlyone")

func init() {
	log.Info("init()")
	caddy.RegisterPlugin("onlyone", caddy.Plugin{
		ServerType: "dns",
		Action:     setup,
	})
}

func setup(c *caddy.Controller) error {
	log.Info("setup()")
	t, err := parse(c)
	if err != nil {
		return plugin.Error("onlyone", err)
	}
	log.Info("setup(): t:", t)

	dnsserver.GetConfig(c).AddPlugin(func(next plugin.Handler) plugin.Handler {
		log.Info("setup(): AddPlugin t:", t)
		t.Next = next
		return t
	})

	return nil
}

func parse(c *caddy.Controller) (*onlyone, error) {
	log.Info("parse():")
	o := &onlyone{types: typeMap{dns.TypeA: true, dns.TypeAAAA: true},
		pick: rand.Intn}

	found := false
	for c.Next() {
		// onlyone should just be in the server block once.
		if found {
			log.Info("parse(): Next and found")
			return nil, plugin.ErrOnce
		}
		found = true

		// parse the zone list, normalizing each to a FQDN, and
		// using the zones from the server block if none are given.
		args := c.RemainingArgs()
		log.Info("parse(): args:", args)
		if len(args) == 0 {
			o.zones = make([]string, len(c.ServerBlockKeys))
			copy(o.zones, c.ServerBlockKeys)
		}
		for _, str := range args {
			log.Info("parse(): range args, str:", str)
			o.zones = append(o.zones, plugin.Host(str).Normalize())
		}
		log.Info("parse(): o.zones:", o.zones)
		log.Info("parse(): c.Val():", c.Val())
		for c.NextBlock() {
			switch c.Val() {
			case "types":
				args := c.RemainingArgs()
				if len(args) == 0 {
					return nil, errors.New(
						"at least one type must be listed")
				}
				o.types = make(typeMap, len(args))
				log.Info("parse(): o.types:", o.types)
				for _, a := range args {
					t, ok := dns.StringToType[strings.ToUpper(a)]
					log.Info("parse(): range args, a:", a)
					if !ok {
						return nil,
							fmt.Errorf("invalid type %q",
								a)
					}
					o.types[t] = true
				}
			default:
				return nil, fmt.Errorf("invalid option %q", c.Val())
			}
		}
	}
	return o, nil
}
