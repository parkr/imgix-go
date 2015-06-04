package imgix

import (
	"crypto/md5"
	"encoding/hex"
	"fmt"
	"hash/crc32"
	"net/url"
	"regexp"
)

type ShardStrategy string

const (
	ShardStrategyCRC   = ShardStrategy(":crc")
	ShardStrategyCycle = ShardStrategy(":cycle")
)

var RegexpRemoveHTTPAndS = regexp.MustCompile("http(s)://")

func NewClient(hosts ...string) Client {
	return Client{hosts: hosts, secure: true}
}

func NewClientWithToken(host string, token string) Client {
	return Client{hosts: []string{host}, secure: true, token: token}
}

type Client struct {
	hosts         []string
	token         string
	secure        bool
	shardStrategy ShardStrategy

	// For use with ShardStrategyCycle
	cycleIndex int
}

func (c *Client) ShardStrategy() ShardStrategy {
	switch c.shardStrategy {
	case ShardStrategyCRC, ShardStrategyCycle:
		return c.shardStrategy
	case "":
		c.shardStrategy = ShardStrategyCycle
		return c.shardStrategy
	default:
		panic(fmt.Errorf("shard strategy '%s' is not supported", c.shardStrategy))
	}
}

func (c *Client) Secure() bool {
	return c.secure
}

func (c *Client) Hosts(index int) string {
	if len(c.hosts) == 0 {
		panic(fmt.Errorf("hosts must be provided"))
	}
	return c.hosts[index]
}

func (c *Client) Scheme() string {
	if c.Secure() {
		return "https"
	} else {
		return "http"
	}
}

func (c *Client) Host(path string) string {
	var host string
	switch c.ShardStrategy() {
	case ShardStrategyCRC:
		host = c.Hosts(c.crc32BasedIndexForPath(path))
	case ShardStrategyCycle:
		host = c.Hosts(c.cycleIndex)
		c.cycleIndex = (c.cycleIndex + 1) % len(c.hosts)
	}

	return RegexpRemoveHTTPAndS.ReplaceAllString(host, "")
}

func (c *Client) URL(imgPath string) url.URL {
	return url.URL{
		Scheme:   c.Scheme(),
		Host:     c.Host(imgPath),
		Path:     imgPath,
		RawQuery: c.SignatureForPath(imgPath),
	}
}

func (c *Client) SignatureForPath(path string) string {
	if c.token == "" {
		return ""
	}

	hasher := md5.New()
	hasher.Write([]byte(c.token + path))
	return "s=" + hex.EncodeToString(hasher.Sum(nil))
}

func (c *Client) SignatureForPathAndParams(path string, params url.Values) string {
	if c.token == "" {
		return ""
	}

	hasher := md5.New()
	hasher.Write([]byte(c.token + path))

	// Do not mix in the parameters into the signature hash if no parameters
	// have been given
	if len(params) != 0 {
		hasher.Write([]byte("?" + params.Encode()))
	}

	return "s=" + hex.EncodeToString(hasher.Sum(nil))
}

func (c *Client) Path(imgPath string) string {
	url := c.URL(imgPath)
	return url.String()
}

func (c *Client) PathWithParams(imgPath string, params url.Values) string {
	u := url.URL{
		Scheme:   c.Scheme(),
		Host:     c.Host(imgPath),
		Path:     imgPath,
		RawQuery: params.Encode(),
	}

	signature := c.SignatureForPathAndParams(imgPath, params)
	if signature != "" && len(params) > 0 {
		u.RawQuery = u.RawQuery + "&" + signature
	} else if signature != "" && len(params) == 0 {
		u.RawQuery = signature
	}

	return u.String()
}

func (c *Client) crc32BasedIndexForPath(path string) int {
	crc := crc32.ChecksumIEEE([]byte(path))
	return int(crc) % len(c.hosts)
}
