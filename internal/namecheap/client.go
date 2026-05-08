package namecheap

import (
	"context"
	"encoding/xml"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"
)

const (
	ProductionEndpoint = "https://api.namecheap.com/xml.response"
	SandboxEndpoint    = "https://api.sandbox.namecheap.com/xml.response"
)

type Config struct {
	Endpoint string
	APIUser  string
	APIKey   string
	UserName string
	ClientIP string
}

type Client struct {
	config Config
	http   *http.Client
}

func NewClient(config Config) (*Client, error) {
	if config.Endpoint == "" {
		config.Endpoint = ProductionEndpoint
	}
	missing := missingConfig(config)
	if len(missing) > 0 {
		return nil, fmt.Errorf("missing required config: %s", strings.Join(missing, ", "))
	}
	return &Client{
		config: config,
		http: &http.Client{
			Timeout: 30 * time.Second,
		},
	}, nil
}

func missingConfig(config Config) []string {
	var missing []string
	if config.APIUser == "" {
		missing = append(missing, "api user")
	}
	if config.APIKey == "" {
		missing = append(missing, "api key")
	}
	if config.UserName == "" {
		missing = append(missing, "username")
	}
	if config.ClientIP == "" {
		missing = append(missing, "client ip")
	}
	return missing
}

type apiEnvelope struct {
	XMLName        xml.Name       `xml:"ApiResponse"`
	Status         string         `xml:"Status,attr"`
	Errors         []apiMessage   `xml:"Errors>Error"`
	Warnings       []apiMessage   `xml:"Warnings>Warning"`
	CommandResults commandResults `xml:"CommandResponse"`
}

type apiMessage struct {
	Number  string `xml:"Number,attr"`
	Message string `xml:",chardata"`
}

type commandResults struct {
	Domains []Domain `xml:"DomainGetListResult>Domain"`
	Hosts   []Host   `xml:"DomainDNSGetHostsResult>host"`
	SetDNS  *SetDNS  `xml:"DomainDNSSetHostsResult"`
}

type Domain struct {
	Name       string `xml:"Name,attr" json:"name"`
	User       string `xml:"User,attr" json:"user,omitempty"`
	Created    string `xml:"Created,attr" json:"created,omitempty"`
	Expires    string `xml:"Expires,attr" json:"expires,omitempty"`
	IsExpired  bool   `xml:"IsExpired,attr" json:"isExpired"`
	IsLocked   bool   `xml:"IsLocked,attr" json:"isLocked"`
	AutoRenew  bool   `xml:"AutoRenew,attr" json:"autoRenew"`
	WhoisGuard string `xml:"WhoisGuard,attr" json:"whoisGuard,omitempty"`
	IsOurDNS   bool   `xml:"IsOurDNS,attr" json:"isOurDNS"`
}

type Host struct {
	ID            string `xml:"HostId,attr" json:"id,omitempty"`
	Name          string `xml:"Name,attr" json:"name"`
	Type          string `xml:"Type,attr" json:"type"`
	Address       string `xml:"Address,attr" json:"address"`
	MXPref        string `xml:"MXPref,attr" json:"mxPref,omitempty"`
	TTL           string `xml:"TTL,attr" json:"ttl,omitempty"`
	AssociatedApp string `xml:"AssociatedAppTitle,attr" json:"associatedApp,omitempty"`
	FriendlyName  string `xml:"FriendlyName,attr" json:"friendlyName,omitempty"`
	IsActive      string `xml:"IsActive,attr" json:"isActive,omitempty"`
	IsDDNSEnabled string `xml:"IsDDNSEnabled,attr" json:"isDDNSEnabled,omitempty"`
}

type SetDNS struct {
	Domain  string `xml:"Domain,attr"`
	Updated bool   `xml:"IsSuccess,attr"`
}

func (c *Client) Domains(ctx context.Context) ([]Domain, error) {
	var envelope apiEnvelope
	if err := c.do(ctx, "namecheap.domains.getList", nil, &envelope); err != nil {
		return nil, err
	}
	return envelope.CommandResults.Domains, nil
}

func (c *Client) Hosts(ctx context.Context, sld, tld string) ([]Host, error) {
	params := url.Values{}
	params.Set("SLD", sld)
	params.Set("TLD", tld)
	var envelope apiEnvelope
	if err := c.do(ctx, "namecheap.domains.dns.getHosts", params, &envelope); err != nil {
		return nil, err
	}
	return envelope.CommandResults.Hosts, nil
}

func (c *Client) SetHosts(ctx context.Context, sld, tld string, hosts []Host) error {
	params := url.Values{}
	params.Set("SLD", sld)
	params.Set("TLD", tld)
	for i, host := range hosts {
		n := strconv.Itoa(i + 1)
		params.Set("HostName"+n, host.Name)
		params.Set("RecordType"+n, strings.ToUpper(host.Type))
		params.Set("Address"+n, host.Address)
		if host.MXPref != "" {
			params.Set("MXPref"+n, host.MXPref)
		}
		if host.TTL != "" {
			params.Set("TTL"+n, host.TTL)
		}
	}
	var envelope apiEnvelope
	if err := c.do(ctx, "namecheap.domains.dns.setHosts", params, &envelope); err != nil {
		return err
	}
	if envelope.CommandResults.SetDNS != nil && !envelope.CommandResults.SetDNS.Updated {
		return errors.New("namecheap returned unsuccessful DNS update")
	}
	return nil
}

func (c *Client) do(ctx context.Context, command string, params url.Values, out *apiEnvelope) error {
	if params == nil {
		params = url.Values{}
	}
	params.Set("ApiUser", c.config.APIUser)
	params.Set("ApiKey", c.config.APIKey)
	params.Set("UserName", c.config.UserName)
	params.Set("ClientIp", c.config.ClientIP)
	params.Set("Command", command)

	endpoint, err := url.Parse(c.config.Endpoint)
	if err != nil {
		return fmt.Errorf("invalid endpoint: %w", err)
	}
	endpoint.RawQuery = params.Encode()

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, endpoint.String(), nil)
	if err != nil {
		return err
	}
	resp, err := c.http.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return err
	}
	if resp.StatusCode < 200 || resp.StatusCode > 299 {
		return fmt.Errorf("namecheap http %d: %s", resp.StatusCode, strings.TrimSpace(string(body)))
	}
	if err := xml.Unmarshal(body, out); err != nil {
		return fmt.Errorf("decode namecheap response: %w", err)
	}
	if strings.EqualFold(out.Status, "ERROR") || len(out.Errors) > 0 {
		return apiError(out.Errors)
	}
	return nil
}

func apiError(messages []apiMessage) error {
	if len(messages) == 0 {
		return errors.New("namecheap returned an error")
	}
	parts := make([]string, 0, len(messages))
	for _, msg := range messages {
		if msg.Number == "" {
			parts = append(parts, strings.TrimSpace(msg.Message))
			continue
		}
		parts = append(parts, fmt.Sprintf("%s: %s", msg.Number, strings.TrimSpace(msg.Message)))
	}
	return errors.New(strings.Join(parts, "; "))
}

func SplitDomain(domain string) (string, string, error) {
	parts := strings.Split(strings.TrimSpace(domain), ".")
	if len(parts) < 2 || parts[0] == "" || parts[len(parts)-1] == "" {
		return "", "", fmt.Errorf("invalid domain %q", domain)
	}
	return strings.Join(parts[:len(parts)-1], "."), parts[len(parts)-1], nil
}
