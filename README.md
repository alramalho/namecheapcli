# namecheapcli

Small CLI wrapper for Namecheap domain and DNS record management.

## Configuration

Create Namecheap API credentials and whitelist your current public IPv4 in Namecheap. Then export:

```sh
export NAMECHEAP_API_USER="your-api-user"
export NAMECHEAP_API_KEY="your-api-key"
export NAMECHEAP_USERNAME="your-username"
export NAMECHEAP_CLIENT_IP="your-whitelisted-public-ip"
```

Use `--sandbox` for Namecheap's sandbox API.

## Usage

Install the CLI:

```sh
go install ./cmd/namecheapcli
```

Then use the installed binary:

```sh
namecheapcli --sandbox domains list
namecheapcli --sandbox dns list example.com
namecheapcli --sandbox dns add example.com A @ 203.0.113.10 --ttl 1800 --dry-run
namecheapcli --sandbox dns set example.com CNAME www example.com --ttl 1800 --dry-run
namecheapcli --sandbox dns delete example.com TXT _verify --dry-run
```

`dns add`, `dns set`, and `dns delete` are implemented as read-modify-write operations:

1. fetch existing records with `namecheap.domains.dns.getHosts`
2. modify the record list locally
3. submit the complete list with `namecheap.domains.dns.setHosts`

This matters because Namecheap's `setHosts` replaces the full host-record set. Any existing record omitted from the update request is deleted.
