package imgix

import (
	"crypto/md5"
	"encoding/hex"
	"fmt"
	"hash/crc32"
	"net/url"
	"regexp"
	"strings"
	"unicode/utf8"
)

type ShardStrategy string

const (
	ShardStrategyCRC   = ShardStrategy(":crc")
	ShardStrategyCycle = ShardStrategy(":cycle")
)

var RegexpRemoveHTTPAndS = regexp.MustCompile("https?://")

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

func (c *Client) SignatureForPath(path string) string {
	return c.SignatureForPathAndParams(path, url.Values{})
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
	return c.PathWithParams(imgPath, url.Values{})
}

// `PathWithParams` will manually build a URL from a given path string and
// parameters passed in. Because of the differences in how net/url escapes
// path components, we need to manually build a URL as best we can.
//
// The behavior of this function is highly dependent upon its test suite.
func (c *Client) PathWithParams(imgPath string, params url.Values) string {
	u := url.URL{
		Scheme: c.Scheme(),
		Host:   c.Host(imgPath),
	}

	urlString := u.String()

	// If we are given a fully-qualified URL, escape it per the note located
	// near the `cgiEscape` function definition
	if RegexpRemoveHTTPAndS.MatchString(imgPath) {
		imgPath = cgiEscape(imgPath)
	}

	// Add a preceding slash if one does not exist:
	//     "users/1.png" -> "/users/1.png"
	if strings.Index(imgPath, "/") != 0 {
		imgPath = "/" + imgPath
	}

	urlString += imgPath

	// The signature in an imgix URL must always be the **last** parameter in a URL,
	// hence some of the gross string concatenation here. net/url will aggressively
	// alphabetize the URL parameters.
	signature := c.SignatureForPathAndParams(imgPath, params)
	parameterString := params.Encode()
	if signature != "" && len(params) > 0 {
		parameterString += "&" + signature
	} else if signature != "" && len(params) == 0 {
		parameterString = signature
	}

	// Only append the parameter string if it is not blank.
	if parameterString != "" {
		urlString += "?" + parameterString
	}

	return urlString
}

func (c *Client) crc32BasedIndexForPath(path string) int {
	crc := crc32.ChecksumIEEE([]byte(path))
	return int(crc) % len(c.hosts)
}

// This code is less than ideal, but it's the only way we've found out how to do it
// give Go's URL capabilities and escaping behavior.
//
// This method replicates the beavhior of Ruby's CGI::escape in Go.
//
// Here is that method:
//
//     def CGI::escape(string)
//       string.gsub(/([^ a-zA-Z0-9_.-]+)/) do
//         '%' + $1.unpack('H2' * $1.bytesize).join('%').upcase
//       end.tr(' ', '+')
//      end
//
// It replaces
//
// See:
//  - https://github.com/parkr/imgix-go/pull/1#issuecomment-109014369
//  - https://github.com/imgix/imgix-blueprint#securing-urls
func cgiEscape(s string) string {
	urlCharactersToEscape := regexp.MustCompile("([^ a-zA-Z0-9_.-])")
	return urlCharactersToEscape.ReplaceAllStringFunc(s, func(s string) string {
		rune, _ := utf8.DecodeLastRuneInString(s)
		return "%" + strings.ToUpper(fmt.Sprintf("%x", rune))
	})
}
