// Package curl implements a minimal `curl` applet (also registered as
// `wget`) for HTTP(S) GETs.
//
// Full curl is enormous; this build covers the everyday case: fetch
// a URL and print the body, optionally to a file, optionally with
// redirect-following or HEAD-only requests.
package curl

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/Real-Fruit-Snacks/topsail/internal/applet"
	"github.com/Real-Fruit-Snacks/topsail/internal/ioutil"
)

func init() {
	applet.Register(applet.Applet{
		Name:    "curl",
		Aliases: []string{"wget"},
		Help:    "fetch a URL via HTTP(S)",
		Usage:   usage,
		Main:    Main,
	})
}

const usage = `Usage: curl [OPTION]... URL
Fetch URL and write the response body to stdout (or to FILE with -o).

Options:
  -o FILE, --output=FILE   write body to FILE (default: stdout)
  -s, --silent             don't show progress; errors still go to stderr
  -L, --location           follow HTTP redirects
  -I, --head               issue a HEAD request and print headers
  -X METHOD                use HTTP METHOD (default: GET, or HEAD with -I)
  -H "K: V"                add a header; may be repeated
  --max-time SECONDS       overall timeout (default: 30)
  --user-agent UA          set the User-Agent header

The 'wget' alias differs only by name; the option set is identical.
`

// Main is the applet entry point.
func Main(argv []string) int {
	args := argv[1:]
	var (
		output     string
		silent     bool
		follow     bool
		headOnly   bool
		method     string
		headers    []string
		userAgent  = "topsail-curl/1.0"
		maxSeconds = 30
	)

	stop := false
	for !stop && len(args) > 0 {
		a := args[0]
		switch {
		case a == "--":
			args = args[1:]
			stop = true
		case a == "-o":
			if len(args) < 2 {
				ioutil.Errf("curl: option requires an argument -- 'o'")
				return 2
			}
			output = args[1]
			args = args[2:]
		case strings.HasPrefix(a, "--output="):
			output = a[len("--output="):]
			args = args[1:]
		case a == "-s", a == "--silent":
			silent = true
			args = args[1:]
		case a == "-L", a == "--location":
			follow = true
			args = args[1:]
		case a == "-I", a == "--head":
			headOnly = true
			args = args[1:]
		case a == "-X":
			if len(args) < 2 {
				ioutil.Errf("curl: option requires an argument -- 'X'")
				return 2
			}
			method = args[1]
			args = args[2:]
		case a == "-H":
			if len(args) < 2 {
				ioutil.Errf("curl: option requires an argument -- 'H'")
				return 2
			}
			headers = append(headers, args[1])
			args = args[2:]
		case strings.HasPrefix(a, "--user-agent="):
			userAgent = a[len("--user-agent="):]
			args = args[1:]
		case strings.HasPrefix(a, "--max-time="):
			n, err := strconv.Atoi(a[len("--max-time="):])
			if err != nil || n < 0 {
				ioutil.Errf("curl: invalid --max-time")
				return 2
			}
			maxSeconds = n
			args = args[1:]
		case strings.HasPrefix(a, "-") && len(a) > 1 && a != "-":
			ioutil.Errf("curl: unknown option: %s", a)
			return 2
		default:
			stop = true
		}
	}

	if len(args) == 0 {
		ioutil.Errf("curl: no URL specified")
		return 2
	}
	if len(args) > 1 {
		ioutil.Errf("curl: only one URL is supported in this build")
		return 2
	}
	url := args[0]
	if method == "" {
		if headOnly {
			method = "HEAD"
		} else {
			method = "GET"
		}
	}
	_ = silent

	ctx, cancel := context.WithTimeout(context.Background(), time.Duration(maxSeconds)*time.Second)
	defer cancel()

	req, err := http.NewRequestWithContext(ctx, method, url, http.NoBody)
	if err != nil {
		ioutil.Errf("curl: %v", err)
		return 2
	}
	req.Header.Set("User-Agent", userAgent)
	for _, h := range headers {
		i := strings.Index(h, ":")
		if i < 0 {
			ioutil.Errf("curl: malformed header: %q", h)
			return 2
		}
		req.Header.Set(strings.TrimSpace(h[:i]), strings.TrimSpace(h[i+1:]))
	}
	client := &http.Client{}
	if !follow {
		client.CheckRedirect = func(_ *http.Request, _ []*http.Request) error {
			return http.ErrUseLastResponse
		}
	}
	resp, err := client.Do(req)
	if err != nil {
		ioutil.Errf("curl: %v", err)
		return 1
	}
	defer func() { _ = resp.Body.Close() }()

	if headOnly {
		_, _ = fmt.Fprintf(ioutil.Stdout, "HTTP/%d.%d %s\n", resp.ProtoMajor, resp.ProtoMinor, resp.Status)
		for k, vs := range resp.Header {
			for _, v := range vs {
				_, _ = fmt.Fprintf(ioutil.Stdout, "%s: %s\n", k, v)
			}
		}
		return 0
	}

	var w io.Writer
	w = ioutil.Stdout
	if output != "" {
		f, ferr := os.Create(output) //nolint:gosec // user-supplied path is the whole point
		if ferr != nil {
			ioutil.Errf("curl: %s: %v", output, ferr)
			return 1
		}
		defer func() { _ = f.Close() }()
		w = f
	}

	if _, err := io.Copy(w, resp.Body); err != nil {
		ioutil.Errf("curl: %v", err)
		return 1
	}
	if resp.StatusCode >= 400 {
		ioutil.Errf("curl: HTTP %d", resp.StatusCode)
		return 22 // curl's "HTTP error" exit code
	}
	return 0
}
