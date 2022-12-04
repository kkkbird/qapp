package qhttp

import (
	"sync"
	"testing"
	"time"

	"github.com/go-redis/redis/v8"
	"github.com/spf13/viper"
	"github.com/stretchr/testify/suite"
)

type ClientTestSuite struct {
	suite.Suite
	redisClient *redis.Client
}

func (s *ClientTestSuite) SetupSuite() {
	viper.SetDefault("redis.url", "192.168.1.233:30790")
	viper.SetDefault("redis.password", "12345678")

	s.redisClient = redis.NewClient(&redis.Options{
		Addr:     viper.GetString("redis.url"),
		Password: viper.GetString("redis.password"),
		DB:       0,
	})
}

func (s *ClientTestSuite) TestHTTPGet() {
	// l := NewLimit(redis_rate.NewLimiter(s.redisClient), "test", 1, time.Second, 0)

	resp, err := Get("https://m.baidu.com")

	if !s.NoError(err) {
		return
	}

	s.T().Logf("http status: %d", resp.StatusCode)
}

func (s *ClientTestSuite) runNCall(l *Limit, n int) []error {
	errs := make([]error, n)
	var wg sync.WaitGroup
	for i := 0; i < n; i++ {
		wg.Add(1)
		go func(j int) {
			defer wg.Done()
			resp, err := Get("https://m.baidu.com", WithLimit(l))

			if err != nil {
				s.T().Logf("get err:" + err.Error())
				errs[j] = err
				return
			}

			errs[j] = nil
			s.T().Logf("http status: %d", resp.StatusCode)
		}(i)
	}
	wg.Wait()
	return errs
}

func (s *ClientTestSuite) errorCount(errs []error) int {
	c := 0
	for _, e := range errs {
		if e != nil {
			c++
		}
	}
	return c
}

func (s *ClientTestSuite) TestHTTPGetWithLimitAlwaysBlock() {
	l := NewLimit(s.redisClient, "alwayswait", 10, time.Second, 1, LimitAlwaysBlock)

	errs := s.runNCall(l, 4)

	s.Equal(0, s.errorCount(errs))
}

func (s *ClientTestSuite) TestHTTPGetWithLimitNoBlock() {
	l := NewLimit(s.redisClient, "alwayscancel", 10, time.Second, 1, LimitNoBlock)

	errs := s.runNCall(l, 4)

	s.Equal(3, s.errorCount(errs))
}

func (s *ClientTestSuite) TestHTTPGetWithLimit() {
	l := NewLimit(s.redisClient, "withlimit", 2, time.Second, 1, time.Second)

	errs := s.runNCall(l, 4)

	s.Equal(2, s.errorCount(errs))
}

func (s *ClientTestSuite) TearDownSuite() {
	s.redisClient.Close()
}

func TestClient(t *testing.T) {
	suite.Run(t, new(ClientTestSuite))
}
