// Package id implements the `id` applet.
package id

import (
	"fmt"
	"os/user"
	"strings"

	"github.com/Real-Fruit-Snacks/topsail/internal/applet"
	"github.com/Real-Fruit-Snacks/topsail/internal/ioutil"
)

func init() {
	applet.Register(applet.Applet{
		Name:  "id",
		Help:  "print user and group identity",
		Usage: usage,
		Main:  Main,
	})
}

const usage = `Usage: id [OPTION]... [USER]
Print real and effective user/group IDs of USER (or current user).

Options:
  -u, --user      print only the effective user ID
  -g, --group     print only the effective group ID
  -G, --groups    print all group IDs
  -n, --name      print names instead of numbers (with -u/-g/-G)
`

// Main is the applet entry point.
func Main(argv []string) int {
	args := argv[1:]
	var u, g, gs, names bool

	stop := false
	for !stop && len(args) > 0 {
		a := args[0]
		switch {
		case a == "--":
			args = args[1:]
			stop = true
		case a == "-u", a == "--user":
			u = true
			args = args[1:]
		case a == "-g", a == "--group":
			g = true
			args = args[1:]
		case a == "-G", a == "--groups":
			gs = true
			args = args[1:]
		case a == "-n", a == "--name":
			names = true
			args = args[1:]
		case strings.HasPrefix(a, "-") && len(a) > 1 && a != "-":
			for _, c := range a[1:] {
				switch c {
				case 'u':
					u = true
				case 'g':
					g = true
				case 'G':
					gs = true
				case 'n':
					names = true
				default:
					ioutil.Errf("id: invalid option -- '%c'", c)
					return 2
				}
			}
			args = args[1:]
		default:
			stop = true
		}
	}

	var who *user.User
	var err error
	if len(args) > 0 {
		who, err = user.Lookup(args[0])
	} else {
		who, err = user.Current()
	}
	if err != nil {
		ioutil.Errf("id: %v", err)
		return 1
	}

	switch {
	case u && names:
		_, _ = fmt.Fprintln(ioutil.Stdout, who.Username)
	case u:
		_, _ = fmt.Fprintln(ioutil.Stdout, who.Uid)
	case g && names:
		gr, err := user.LookupGroupId(who.Gid)
		if err == nil {
			_, _ = fmt.Fprintln(ioutil.Stdout, gr.Name)
		} else {
			_, _ = fmt.Fprintln(ioutil.Stdout, who.Gid)
		}
	case g:
		_, _ = fmt.Fprintln(ioutil.Stdout, who.Gid)
	case gs:
		gids, err := who.GroupIds()
		if err != nil {
			ioutil.Errf("id: %v", err)
			return 1
		}
		for i, gid := range gids {
			if i > 0 {
				_, _ = ioutil.Stdout.Write([]byte(" "))
			}
			if names {
				if gr, err := user.LookupGroupId(gid); err == nil {
					_, _ = ioutil.Stdout.Write([]byte(gr.Name))
					continue
				}
			}
			_, _ = ioutil.Stdout.Write([]byte(gid))
		}
		_, _ = ioutil.Stdout.Write([]byte("\n"))
	default:
		groupName := who.Gid
		if gr, err := user.LookupGroupId(who.Gid); err == nil {
			groupName = gr.Name
		}
		_, _ = fmt.Fprintf(ioutil.Stdout, "uid=%s(%s) gid=%s(%s)\n",
			who.Uid, who.Username, who.Gid, groupName)
	}
	return 0
}
