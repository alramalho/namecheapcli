# namecheapcli

Small CLI wrapper for Namecheap domain and DNS record management.

## Configuration

Create Namecheap API credentials and whitelist your current public IPv4 in Namecheap.

For day-to-day use, run the setup command once:

```sh
namecheapcli configure
```

That writes a private global config file at `~/.namecheapcli`:

```sh
NAMECHEAP_API_USER=your-api-user
NAMECHEAP_API_KEY=your-api-key
NAMECHEAP_USERNAME=your-username
NAMECHEAP_CLIENT_IP=your-whitelisted-public-ip
```

You can also export the same values as environment variables:

```sh
export NAMECHEAP_API_USER="your-api-user"
export NAMECHEAP_API_KEY="your-api-key"
export NAMECHEAP_USERNAME="your-username"
export NAMECHEAP_CLIENT_IP="your-whitelisted-public-ip"
```

Environment variables override `~/.namecheapcli`, which is useful for CI or temporary credentials.
Use `--sandbox` for Namecheap's sandbox API.

## Usage

Install the CLI globally:

```sh
go install ./cmd/namecheapcli
```

After the repo is public, install it from anywhere with:

```sh
go install github.com/alramalho/namecheapcli/cmd/namecheapcli@latest
```

If your shell cannot find `namecheapcli` after install, add Go's binary directory to your `PATH`:

```sh
export PATH="$(go env GOPATH)/bin:$PATH"
```

For zsh, make that permanent:

```sh
printf '\nexport PATH="$(go env GOPATH)/bin:$PATH"\n' >> ~/.zshrc
source ~/.zshrc
```

You can verify the installed binary directly with:

```sh
"$(go env GOPATH)/bin/namecheapcli" help
```

Then use the installed binary from any directory:

```sh
namecheapcli configure
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
