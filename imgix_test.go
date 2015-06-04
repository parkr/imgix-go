package imgix

import (
	"net/url"
	"testing"

	"github.com/stretchr/testify/assert"
)

func testClient() Client {
	return NewClient("prod.imgix.net", "stag.imgix.net", "dev.imgix.net")
}

func testClientWithToken() Client {
	return NewClientWithToken("my-social-network.imgix.net", "FOO123bar")
}

func TestBasicClientPath(t *testing.T) {
	c := testClient()
	assert.Equal(t, "https://prod.imgix.net/1/users.jpg", c.Path("/1/users.jpg"))
}

func TestClientPathWithParams(t *testing.T) {
	c := testClient()
	params := url.Values{"w": []string{"200"}, "h": []string{"400"}}
	assert.Equal(t, "https://prod.imgix.net/1/users.jpg?h=400&w=200", c.PathWithParams("/1/users.jpg", params))
}

func TestClientScheme(t *testing.T) {
	c := testClient()
	c.secure = false
	assert.Equal(t, "http", c.Scheme())
	c.secure = true
	assert.Equal(t, "https", c.Scheme())
}

func TestClientURL(t *testing.T) {
	c := testClient()
	u := c.Path("/jax.jpg")
	assert.Equal(t, "https://prod.imgix.net/jax.jpg", u)
}

func TestClientPathWithSignature(t *testing.T) {
	c := testClientWithToken()
	u := c.Path("/users/1.png")
	assert.Equal(t, "https://my-social-network.imgix.net/users/1.png?s=6797c24146142d5b40bde3141fd3600c", u)
}

func TestClientPathWithSignatureAndParams(t *testing.T) {
	c := testClientWithToken()
	params := url.Values{"w": []string{"400"}, "h": []string{"300"}}
	assert.Equal(t, "https://my-social-network.imgix.net/users/1.png?h=300&w=400&s=1a4e48641614d1109c6a7af51be23d18", c.PathWithParams("/users/1.png", params))
}

func TestClientPathWithSignatureAndEmptyParams(t *testing.T) {
	c := testClientWithToken()
	params := url.Values{}
	assert.Equal(t, "https://my-social-network.imgix.net/users/1.png?s=6797c24146142d5b40bde3141fd3600c", c.PathWithParams("/users/1.png", params))
}

func TestClientFullyQualifiedUrlPath(t *testing.T) {
	c := testClientWithToken()
	assert.Equal(t, "https://my-social-network.imgix.net/http%3A%2F%2Favatars.com%2Fjohn-smith.png?s=493a52f008c91416351f8b33d4883135", c.Path("http://avatars.com/john-smith.png"))
}

func TestClientFullyQualifiedUrlPathWithParams(t *testing.T) {
	c := testClientWithToken()
	params := url.Values{"w": []string{"400"}, "h": []string{"300"}}
	assert.Equal(t, "https://my-social-network.imgix.net/http%3A%2F%2Favatars.com%2Fjohn-smith.png?h=300&w=400&s=a201fe1a3caef4944dcb40f6ce99e746", c.PathWithParams("http://avatars.com/john-smith.png", params))
}

func TestClientFallbackShardStrategy(t *testing.T) {
	c := testClient()
	assert.Equal(t, ShardStrategy(""), c.shardStrategy)
	assert.Equal(t, ShardStrategyCycle, c.ShardStrategy())
}

func TestClientHostUsingCRC(t *testing.T) {
	c := testClient()
	c.shardStrategy = ShardStrategyCRC
	assert.Equal(t, "prod.imgix.net", c.Host("/1/users.jpg"))
	assert.Equal(t, "dev.imgix.net", c.Host("/2/ellothere.png"))
}

func TestClientHostUsingCycle(t *testing.T) {
	c := testClient()
	c.shardStrategy = ShardStrategyCycle
	assert.Equal(t, "prod.imgix.net", c.Host("/1/users.jpg"))
	assert.Equal(t, "stag.imgix.net", c.Host("/1/users.jpg"))
	assert.Equal(t, "dev.imgix.net", c.Host("/1/users.jpg"))
	assert.Equal(t, "prod.imgix.net", c.Host("/1/users.jpg"))
}

func TestClientShardStrategyValidation(t *testing.T) {
	defer func() {
		if r := recover(); r != nil {
			e, ok := r.(error)
			assert.True(t, ok)
			assert.EqualError(t, e, "shard strategy 'hellothere' is not supported")
		}
	}()

	c := testClient()
	c.shardStrategy = ShardStrategy("hellothere")
	c.ShardStrategy()
}

func TestClientHostsCountValidation(t *testing.T) {
	defer func() {
		if r := recover(); r != nil {
			e, ok := r.(error)
			assert.True(t, ok)
			assert.EqualError(t, e, "hosts must be provided")
		}
	}()

	c := testClient()
	c.hosts = []string{}
	c.Hosts(1)
}
