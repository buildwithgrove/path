package authenticator

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"syscall"
	"testing"
	"time"

	"github.com/ory/dockertest/v3"
	"github.com/ory/dockertest/v3/docker"
	"github.com/pokt-network/poktroll/pkg/polylog/polyzero"
	"github.com/redis/go-redis/v9"
	"github.com/stretchr/testify/require"

	reqCtx "github.com/buildwithgrove/path/request/context"
	"github.com/buildwithgrove/path/user"
)

var redisAddr string

func TestMain(m *testing.M) {
	// Initialize the ephemeral redis docker container
	pool, resource, redisHostAndPort := setupRedisDocker()
	redisAddr = redisHostAndPort

	// Run auth rate limit test
	exitCode := m.Run()

	// Cleanup the ephemeral postgres docker container
	cleanupRedisDocker(pool, resource)
	os.Exit(exitCode)
}

func Test_authenticate(t *testing.T) {
	tests := []struct {
		name           string
		userApp        user.UserApp
		expectedResult *invalidResp
		requests       int
	}{
		{
			name: "should allow request within rate limit",
			userApp: user.UserApp{
				ID:                  "user_app_1",
				RateLimitThroughput: 30,
			},
			expectedResult: nil,
			requests:       30,
		},
		{
			name: "should block request exceeding rate limit",
			userApp: user.UserApp{
				ID:                  "user_app_2",
				RateLimitThroughput: 30,
			},
			expectedResult: &throughputLimitExceeded,
			requests:       40,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			c := require.New(t)

			authenticator := newRateLimitAuthenticator(redisAddr, polyzero.NewLogger())

			results := make(chan *invalidResp, test.requests)
			for i := 0; i < test.requests; i++ {
				go func() {
					result := authenticator.authenticate(context.Background(), reqCtx.HTTPDetails{}, test.userApp)
					results <- result
				}()
			}

			var limitedResp *invalidResp
			for i := 0; i < test.requests; i++ {
				result := <-results
				if result != nil {
					limitedResp = result
					break
				}
			}

			c.Equal(test.expectedResult, limitedResp)
		})
	}
}

func Test_authThroughputLimit(t *testing.T) {
	tests := []struct {
		name           string
		userApp        user.UserApp
		expectedResult *invalidResp
		requests       int
	}{
		{
			name: "should allow request within rate limit",
			userApp: user.UserApp{
				ID:                  "user_app_3",
				RateLimitThroughput: 30,
			},
			expectedResult: nil,
			requests:       30,
		},
		{
			name: "should block request exceeding rate limit",
			userApp: user.UserApp{
				ID:                  "user_app_4",
				RateLimitThroughput: 30,
			},
			expectedResult: &throughputLimitExceeded,
			requests:       40,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			c := require.New(t)

			authenticator := newRateLimitAuthenticator(redisAddr, polyzero.NewLogger())

			results := make(chan *invalidResp, test.requests)
			for i := 0; i < test.requests; i++ {
				go func() {
					result := authenticator.authThroughputLimit(context.Background(), test.userApp)
					results <- result
				}()
			}

			var limitedResp *invalidResp
			for i := 0; i < test.requests; i++ {
				result := <-results
				if result != nil {
					limitedResp = result
					break
				}
			}

			c.Equal(test.expectedResult, limitedResp)
		})
	}
}

/* -------------------- Dockertest Ephemeral Redis Container Setup -------------------- */

const (
	redisContainerName = "redis"
	redisContainerRepo = "redis"
	redisContainerTag  = "latest"
	redisPort          = "6379/tcp"
	redisTimeout       = 10
)

func setupRedisDocker() (*dockertest.Pool, *dockertest.Resource, string) {
	opts := dockertest.RunOptions{
		Name:       redisContainerName,
		Repository: redisContainerRepo,
		Tag:        redisContainerTag,
	}

	pool, err := dockertest.NewPool("")
	if err != nil {
		fmt.Printf("Could not construct pool: %s", err)
		os.Exit(1)
	}
	resource, err := pool.RunWithOptions(&opts, func(config *docker.HostConfig) {
		config.AutoRemove = true
		config.RestartPolicy = docker.RestartPolicy{Name: "no"}
	})
	if err != nil {
		fmt.Printf("Could not start resource: %s", err)
		os.Exit(1)
	}

	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt, syscall.SIGTERM)
	go func() {
		for sig := range c {
			fmt.Printf("exit signal %d received\n", sig)
			if err := pool.Purge(resource); err != nil {
				fmt.Printf("could not purge resource: %s", err)
			}
		}
	}()

	if err := resource.Expire(redisTimeout); err != nil {
		fmt.Printf("[ERROR] Failed to set expiration on docker container: %v", err)
		os.Exit(1)
	}

	hostAndPort := resource.GetHostPort(redisPort)

	poolRetryChan := make(chan struct{}, 1)
	retryConnectFn := func() error {
		rdb := redis.NewClient(&redis.Options{
			Addr: hostAndPort,
		})
		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()

		_, err := rdb.Ping(ctx).Result()
		if err != nil {
			return fmt.Errorf("unable to connect to redis: %v", err)
		}
		poolRetryChan <- struct{}{}
		return nil
	}
	if err = pool.Retry(retryConnectFn); err != nil {
		fmt.Printf("could not connect to docker: %s", err)
		os.Exit(1)
	}

	<-poolRetryChan

	return pool, resource, hostAndPort
}

func cleanupRedisDocker(pool *dockertest.Pool, resource *dockertest.Resource) {
	if err := pool.Purge(resource); err != nil {
		fmt.Printf("could not purge resource: %s", err)
	}
}
