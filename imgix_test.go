package imgix

import (
	"net/url"
	"testing"

	"github.com/stretchr/testify/assert"
)

func testClient() Client {
	return NewClient("prod.imgix.net", "stag.imgix.net", "dev.imgix.net")
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
	u := c.URL("/jax.jpg")
	assert.Equal(t, "https://prod.imgix.net/jax.jpg", u.String())
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
