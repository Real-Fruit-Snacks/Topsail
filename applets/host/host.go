// Package host implements a minimal `host` applet: DNS lookups via the
// system resolver.
package host

import (
	"context"
	"fmt"
	"net"
	"strings"
	"time"

	"github.com/Real-Fruit-Snacks/topsail/internal/applet"
	"github.com/Real-Fruit-Snacks/topsail/internal/ioutil"
)

func init() {
	applet.Register(applet.Applet{
		Name:    "host",
		Aliases: []string{"nslookup"},
		Help:    "perform DNS lookups",
		Usage:   usage,
		Main:    Main,
	})
}

const usage = `Usage: host [OPTION]... NAME
Look up DNS records for NAME using the system resolver.

By default queries A/AAAA/CNAME (host's "all addresses" mode). With
-t TYPE, query only the requested record type.

Options:
  -t TYPE              record type: A, AAAA, MX, NS, TXT, CNAME, PTR
  -W TIMEOUT_SECONDS   per-query timeout (default 10)
`

// Main is the applet entry point.
func Main(argv []string) int {
	args := argv[1:]
	rtype := "ALL"
	timeout := 10 * time.Second

	stop := false
	for !stop && len(args) > 0 {
		a := args[0]
		switch {
		case a == "--":
			args = args[1:]
			stop = true
		case a == "-t":
			if len(args) < 2 {
				ioutil.Errf("host: option requires an argument -- 't'")
				return 2
			}
			rtype = strings.ToUpper(args[1])
			args = args[2:]
		case a == "-W":
			if len(args) < 2 {
				ioutil.Errf("host: option requires an argument -- 'W'")
				return 2
			}
			args = args[2:]
		case strings.HasPrefix(a, "-") && len(a) > 1 && a != "-":
			ioutil.Errf("host: unknown option: %s", a)
			return 2
		default:
			stop = true
		}
	}

	if len(args) == 0 {
		ioutil.Errf("host: missing operand")
		return 2
	}
	name := args[0]

	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()

	r := net.DefaultResolver
	rc := 0
	switch rtype {
	case "ALL":
		ips, err := r.LookupIPAddr(ctx, name)
		if err != nil {
			ioutil.Errf("host: %v", err)
			return 1
		}
		for _, ip := range ips {
			label := "has address"
			if ip.IP.To4() == nil {
				label = "has IPv6 address"
			}
			_, _ = fmt.Fprintf(ioutil.Stdout, "%s %s %s\n", name, label, ip.IP.String())
		}
	case "A":
		ips, err := r.LookupIP(ctx, "ip4", name)
		if err != nil {
			ioutil.Errf("host: %v", err)
			return 1
		}
		for _, ip := range ips {
			_, _ = fmt.Fprintf(ioutil.Stdout, "%s has address %s\n", name, ip.String())
		}
	case "AAAA":
		ips, err := r.LookupIP(ctx, "ip6", name)
		if err != nil {
			ioutil.Errf("host: %v", err)
			return 1
		}
		for _, ip := range ips {
			_, _ = fmt.Fprintf(ioutil.Stdout, "%s has IPv6 address %s\n", name, ip.String())
		}
	case "MX":
		mx, err := r.LookupMX(ctx, name)
		if err != nil {
			ioutil.Errf("host: %v", err)
			return 1
		}
		for _, m := range mx {
			_, _ = fmt.Fprintf(ioutil.Stdout, "%s mail is handled by %d %s\n", name, m.Pref, m.Host)
		}
	case "NS":
		ns, err := r.LookupNS(ctx, name)
		if err != nil {
			ioutil.Errf("host: %v", err)
			return 1
		}
		for _, n := range ns {
			_, _ = fmt.Fprintf(ioutil.Stdout, "%s name server %s\n", name, n.Host)
		}
	case "TXT":
		txt, err := r.LookupTXT(ctx, name)
		if err != nil {
			ioutil.Errf("host: %v", err)
			return 1
		}
		for _, t := range txt {
			_, _ = fmt.Fprintf(ioutil.Stdout, "%s descriptive text %q\n", name, t)
		}
	case "CNAME":
		c, err := r.LookupCNAME(ctx, name)
		if err != nil {
			ioutil.Errf("host: %v", err)
			return 1
		}
		_, _ = fmt.Fprintf(ioutil.Stdout, "%s is an alias for %s\n", name, c)
	case "PTR":
		names, err := r.LookupAddr(ctx, name)
		if err != nil {
			ioutil.Errf("host: %v", err)
			return 1
		}
		for _, n := range names {
			_, _ = fmt.Fprintf(ioutil.Stdout, "%s domain name pointer %s\n", name, n)
		}
	default:
		ioutil.Errf("host: unsupported record type: %s", rtype)
		return 2
	}
	return rc
}
