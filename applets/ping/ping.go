// Package ping implements a `ping`-like applet using TCP probes.
//
// Pure-Go ICMP requires raw sockets and elevated privileges; this build
// instead does a TCP connect to a probe port (default 80) and reports
// connect latency. The output is intentionally close to ping's format
// for muscle-memory parity.
package ping

import (
	"fmt"
	"net"
	"strconv"
	"strings"
	"time"

	"github.com/Real-Fruit-Snacks/topsail/internal/applet"
	"github.com/Real-Fruit-Snacks/topsail/internal/ioutil"
)

func init() {
	applet.Register(applet.Applet{
		Name:  "ping",
		Help:  "probe a host's reachability over TCP",
		Usage: usage,
		Main:  Main,
	})
}

const usage = `Usage: ping [OPTION]... HOST
Probe HOST by attempting a TCP connect to a port (default 80) and
reporting connect latency. This is NOT ICMP — pure-Go ICMP requires
raw sockets and elevated privileges, which we deliberately do not
require. Use 'tcpping' as a mental model.

Options:
  -c COUNT            stop after COUNT probes (default 4)
  -p PORT             probe TCP port (default 80)
  -W SECONDS          per-probe timeout (default 5)
  -i SECONDS          interval between probes (default 1)
`

// Main is the applet entry point.
func Main(argv []string) int {
	args := argv[1:]
	count := 4
	port := 80
	timeout := 5 * time.Second
	interval := 1 * time.Second

	stop := false
	for !stop && len(args) > 0 {
		a := args[0]
		switch {
		case a == "--":
			args = args[1:]
			stop = true
		case a == "-c":
			n, err := intArg(args, "c")
			if err != nil {
				return 2
			}
			count = n
			args = args[2:]
		case a == "-p":
			n, err := intArg(args, "p")
			if err != nil {
				return 2
			}
			port = n
			args = args[2:]
		case a == "-W":
			n, err := intArg(args, "W")
			if err != nil {
				return 2
			}
			timeout = time.Duration(n) * time.Second
			args = args[2:]
		case a == "-i":
			n, err := intArgAllowZero(args, "i")
			if err != nil {
				return 2
			}
			interval = time.Duration(n) * time.Second
			args = args[2:]
		case strings.HasPrefix(a, "-") && len(a) > 1 && a != "-":
			ioutil.Errf("ping: unknown option: %s", a)
			return 2
		default:
			stop = true
		}
	}

	if len(args) == 0 {
		ioutil.Errf("ping: missing host")
		return 2
	}
	host := args[0]
	addr := net.JoinHostPort(host, strconv.Itoa(port))

	_, _ = fmt.Fprintf(ioutil.Stdout, "PING %s (%s) via TCP/%d\n", host, addr, port)

	var sent, received int
	var totalLatency time.Duration
	for i := 0; i < count; i++ {
		start := time.Now()
		conn, err := net.DialTimeout("tcp", addr, timeout)
		latency := time.Since(start)
		sent++
		if err != nil {
			_, _ = fmt.Fprintf(ioutil.Stdout, "Request %d to %s: %v (after %v)\n", i+1, addr, err, latency.Round(time.Millisecond))
		} else {
			_ = conn.Close()
			received++
			totalLatency += latency
			_, _ = fmt.Fprintf(ioutil.Stdout, "Reply from %s: seq=%d time=%v\n", addr, i+1, latency.Round(time.Millisecond))
		}
		if i+1 < count {
			time.Sleep(interval)
		}
	}

	loss := 0
	if sent > 0 {
		loss = 100 * (sent - received) / sent
	}
	avg := time.Duration(0)
	if received > 0 {
		avg = totalLatency / time.Duration(received)
	}
	_, _ = fmt.Fprintf(ioutil.Stdout, "\n--- %s ping statistics ---\n", host)
	_, _ = fmt.Fprintf(ioutil.Stdout, "%d probes sent, %d responses, %d%% loss, avg %v\n",
		sent, received, loss, avg.Round(time.Millisecond))
	if received == 0 {
		return 1
	}
	return 0
}

func intArg(args []string, flag string) (int, error) {
	if len(args) < 2 {
		ioutil.Errf("ping: option requires an argument -- '%s'", flag)
		return 0, errMissing
	}
	n, err := strconv.Atoi(args[1])
	if err != nil || n <= 0 {
		ioutil.Errf("ping: invalid -%s value: %s", flag, args[1])
		return 0, errInvalid
	}
	return n, nil
}

func intArgAllowZero(args []string, flag string) (int, error) {
	if len(args) < 2 {
		ioutil.Errf("ping: option requires an argument -- '%s'", flag)
		return 0, errMissing
	}
	n, err := strconv.Atoi(args[1])
	if err != nil || n < 0 {
		ioutil.Errf("ping: invalid -%s value: %s", flag, args[1])
		return 0, errInvalid
	}
	return n, nil
}

var (
	errMissing = stringErr("missing argument")
	errInvalid = stringErr("invalid value")
)

type stringErr string

func (e stringErr) Error() string { return string(e) }
