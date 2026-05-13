package main

import (
	"context"
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"os"
	"strconv"
	"strings"
	"text/tabwriter"

	"github.com/alramalho/namecheapcli/internal/namecheap"
)

func main() {
	if err := run(os.Args[1:]); err != nil {
		fmt.Fprintln(os.Stderr, "error:", err)
		os.Exit(1)
	}
}

func run(args []string) error {
	if len(args) > 0 && (args[0] == "help" || args[0] == "-h" || args[0] == "--help") {
		usage()
		return nil
	}
	global := flag.NewFlagSet("namecheapcli", flag.ContinueOnError)
	global.SetOutput(os.Stderr)
	sandbox := global.Bool("sandbox", false, "use Namecheap sandbox API")
	jsonOut := global.Bool("json", false, "print JSON output")
	if err := global.Parse(args); err != nil {
		if errors.Is(err, flag.ErrHelp) {
			return nil
		}
		return err
	}
	args = global.Args()
	if len(args) == 0 {
		usage()
		return nil
	}

	client, err := namecheap.NewClient(namecheap.ConfigFromEnv(*sandbox))
	if err != nil {
		return err
	}

	ctx := context.Background()
	switch args[0] {
	case "domains":
		return runDomains(ctx, client, args[1:], *jsonOut)
	case "dns":
		return runDNS(ctx, client, args[1:], *jsonOut)
	case "help", "-h", "--help":
		usage()
		return nil
	default:
		return fmt.Errorf("unknown command %q", args[0])
	}
}

func runDomains(ctx context.Context, client *namecheap.Client, args []string, jsonOut bool) error {
	if len(args) != 1 || args[0] != "list" {
		return errors.New("usage: namecheapcli domains list")
	}
	domains, err := client.Domains(ctx)
	if err != nil {
		return err
	}
	if jsonOut {
		return printJSON(domains)
	}
	w := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)
	fmt.Fprintln(w, "DOMAIN\tEXPIRES\tLOCKED\tAUTORENEW\tNAMECHEAP_DNS")
	for _, domain := range domains {
		fmt.Fprintf(w, "%s\t%s\t%v\t%v\t%v\n", domain.Name, domain.Expires, domain.IsLocked, domain.AutoRenew, domain.IsOurDNS)
	}
	return w.Flush()
}

func runDNS(ctx context.Context, client *namecheap.Client, args []string, jsonOut bool) error {
	if len(args) < 1 {
		return errors.New("usage: namecheapcli dns <list|add|delete|set> ...")
	}
	switch args[0] {
	case "list":
		if len(args) != 2 {
			return errors.New("usage: namecheapcli dns list example.com")
		}
		hosts, err := getHosts(ctx, client, args[1])
		if err != nil {
			return err
		}
		return printHosts(hosts, jsonOut)
	case "add":
		return changeHosts(ctx, client, args[1:], addHost)
	case "set":
		return changeHosts(ctx, client, args[1:], setHost)
	case "delete":
		return deleteHost(ctx, client, args[1:])
	default:
		return fmt.Errorf("unknown dns command %q", args[0])
	}
}

func getHosts(ctx context.Context, client *namecheap.Client, domain string) ([]namecheap.Host, error) {
	sld, tld, err := namecheap.SplitDomain(domain)
	if err != nil {
		return nil, err
	}
	return client.Hosts(ctx, sld, tld)
}

type hostMutation func([]namecheap.Host, namecheap.Host) ([]namecheap.Host, error)

func changeHosts(ctx context.Context, client *namecheap.Client, args []string, mutate hostMutation) error {
	pos, opts, err := parseMutationOptions(args)
	if err != nil {
		return err
	}
	if len(pos) != 4 {
		return errors.New("usage: namecheapcli dns add|set example.com TYPE HOST VALUE [--ttl 1800] [--mx-pref 10] [--dry-run]")
	}
	domain, recordType, hostName, address := pos[0], strings.ToUpper(pos[1]), pos[2], pos[3]
	host := namecheap.Host{Name: hostName, Type: recordType, Address: address, TTL: opts.ttl, MXPref: opts.mxPref}
	hosts, err := getHosts(ctx, client, domain)
	if err != nil {
		return err
	}
	updated, err := mutate(hosts, host)
	if err != nil {
		return err
	}
	if opts.dryRun {
		return printJSON(updated)
	}
	sld, tld, err := namecheap.SplitDomain(domain)
	if err != nil {
		return err
	}
	return client.SetHosts(ctx, sld, tld, updated)
}

func addHost(hosts []namecheap.Host, host namecheap.Host) ([]namecheap.Host, error) {
	return append(hosts, host), nil
}

func setHost(hosts []namecheap.Host, host namecheap.Host) ([]namecheap.Host, error) {
	replaced := false
	for i := range hosts {
		if sameRecordIdentity(hosts[i], host) {
			hosts[i].Address = host.Address
			hosts[i].TTL = host.TTL
			hosts[i].MXPref = host.MXPref
			replaced = true
		}
	}
	if !replaced {
		hosts = append(hosts, host)
	}
	return hosts, nil
}

func deleteHost(ctx context.Context, client *namecheap.Client, args []string) error {
	pos, opts, err := parseMutationOptions(args)
	if err != nil {
		return err
	}
	if len(pos) != 3 {
		return errors.New("usage: namecheapcli dns delete example.com TYPE HOST [--dry-run]")
	}
	domain := pos[0]
	target := namecheap.Host{Type: strings.ToUpper(pos[1]), Name: pos[2]}
	hosts, err := getHosts(ctx, client, domain)
	if err != nil {
		return err
	}
	updated := hosts[:0]
	removed := 0
	for _, host := range hosts {
		if sameRecordIdentity(host, target) {
			removed++
			continue
		}
		updated = append(updated, host)
	}
	if removed == 0 {
		return errors.New("no matching records found")
	}
	if opts.dryRun {
		return printJSON(updated)
	}
	sld, tld, err := namecheap.SplitDomain(domain)
	if err != nil {
		return err
	}
	return client.SetHosts(ctx, sld, tld, updated)
}

func sameRecordIdentity(a, b namecheap.Host) bool {
	return strings.EqualFold(a.Type, b.Type) && a.Name == b.Name
}

type mutationOptions struct {
	ttl    string
	mxPref string
	dryRun bool
}

func parseMutationOptions(args []string) ([]string, mutationOptions, error) {
	opts := mutationOptions{ttl: "1800"}
	pos := make([]string, 0, len(args))
	for i := 0; i < len(args); i++ {
		arg := args[i]
		switch {
		case arg == "--dry-run":
			opts.dryRun = true
		case arg == "--ttl":
			i++
			if i >= len(args) {
				return nil, opts, errors.New("--ttl requires a value")
			}
			opts.ttl = args[i]
		case strings.HasPrefix(arg, "--ttl="):
			opts.ttl = strings.TrimPrefix(arg, "--ttl=")
		case arg == "--mx-pref":
			i++
			if i >= len(args) {
				return nil, opts, errors.New("--mx-pref requires a value")
			}
			opts.mxPref = args[i]
		case strings.HasPrefix(arg, "--mx-pref="):
			opts.mxPref = strings.TrimPrefix(arg, "--mx-pref=")
		case strings.HasPrefix(arg, "-"):
			return nil, opts, fmt.Errorf("unknown option %q", arg)
		default:
			pos = append(pos, arg)
		}
	}
	return pos, opts, nil
}

func printHosts(hosts []namecheap.Host, jsonOut bool) error {
	if jsonOut {
		return printJSON(hosts)
	}
	w := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)
	fmt.Fprintln(w, "#\tTYPE\tHOST\tVALUE\tMX\tTTL")
	for i, host := range hosts {
		fmt.Fprintf(w, "%s\t%s\t%s\t%s\t%s\t%s\n", strconv.Itoa(i+1), host.Type, host.Name, host.Address, host.MXPref, host.TTL)
	}
	return w.Flush()
}

func printJSON(v any) error {
	enc := json.NewEncoder(os.Stdout)
	enc.SetIndent("", "  ")
	return enc.Encode(v)
}

func usage() {
	fmt.Fprintln(os.Stderr, `Usage:
  namecheapcli [--sandbox] [--json] domains list
  namecheapcli [--sandbox] [--json] dns list example.com
  namecheapcli [--sandbox] dns add example.com A @ 203.0.113.10 --ttl 1800 --dry-run
  namecheapcli [--sandbox] dns set example.com CNAME www example.com --ttl 1800
  namecheapcli [--sandbox] dns delete example.com TXT _verify --dry-run

Environment:
  NAMECHEAP_API_USER
  NAMECHEAP_API_KEY
  NAMECHEAP_USERNAME
  NAMECHEAP_CLIENT_IP
  NAMECHEAP_ENDPOINT       optional override

Config file:
  ~/.namecheapcli         optional KEY=VALUE file; env vars override it

Note: Namecheap setHosts replaces the full DNS host set. Mutation commands first fetch
the current records, modify them locally, then submit the complete resulting set.`)
}
