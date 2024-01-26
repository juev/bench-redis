package main

import (
	"log"
	"os"
	"strconv"
	"strings"
	"sync"
	"sync/atomic"
	"time"

	"github.com/go-redis/redis/v7"
	"github.com/joho/godotenv"
	"github.com/sourcegraph/conc/pool"
)

func main() {
	redisClient := createRedisClient()
	setKeysToRedis(redisClient, keys)
	if failedValues := readKeysWithGoRoutines(redisClient, keys); failedValues > 0 {
		log.Fatalf("errors on read values from redis: %d", failedValues)
	}

	log.Println("...Tests completed without errors...")
}

func createRedisClient() *redis.ClusterClient {
	if err := godotenv.Load(); err != nil {
		log.Fatalf("failed read env file: %s", err)
	}

	redisClient := redis.NewClusterClient(&redis.ClusterOptions{
		Username:     os.Getenv("REDIS_CLUSTER_USERNAME"),
		Password:     os.Getenv("REDIS_CLUSTER_PASSWORD"),
		Addrs:        strings.Split(os.Getenv("REDIS_CLUSTER_HOSTS"), ","),
		DialTimeout:  5 * time.Second,
		ReadTimeout:  1 * time.Second,
		WriteTimeout: 1 * time.Second,
		PoolTimeout:  5 * time.Second,
		IdleTimeout:  1 * time.Minute,
	})

	if _, err := redisClient.Ping().Result(); err != nil {
		log.Fatalf("failed init redis cluster: %s", err)
	}

	return redisClient
}
func setKeysToRedis(client *redis.ClusterClient, deliveries []string) {
	const deliveryCancelCacheDuration = 5 * time.Minute
	_, err := client.Pipelined(func(pipe redis.Pipeliner) error {
		for _, delivery := range deliveries {
			pipe.Set(
				delivery,
				true,
				deliveryCancelCacheDuration,
			)
		}
		return nil
	})
	if err != nil {
		log.Fatalf("cannot set keys to redis: %s", err)
	}
}

func readKeysWithPipe(client *redis.ClusterClient, deliveries []string) (failedValues int) {
	values, err := client.Pipelined(func(pipe redis.Pipeliner) error {
		for _, deliveryID := range deliveries {
			pipe.Get(deliveryID)
		}
		return nil
	})

	if err != nil {
		log.Fatalf("cannot read from redis: %s", err)
	}

	var deliveryIsCancellable = make(map[string]bool)

	for i, value := range values {
		strCmd, ok := value.(*redis.StringCmd)
		if !ok {
			failedValues++
			continue
		}
		deliveryID := deliveries[i]
		deliveryIsCancellable[deliveryID], _ = strconv.ParseBool(strCmd.Val())
	}

	return failedValues
}

func readKeys(client *redis.ClusterClient, deliveries []string) (failedValues int) {
	var deliveryIsCancellable = make(map[string]bool)

	for _, deliveryID := range deliveries {
		value := client.Get(deliveryID)
		if value == nil {
			failedValues++
			continue
		}
		val, err := strconv.ParseBool(value.Val())
		if err != nil {
			log.Fatalf("cannot parse bool: %s", err)
		}
		deliveryIsCancellable[deliveryID] = val
	}

	return failedValues
}

func readKeysWithGoRoutines(client *redis.ClusterClient, deliveries []string) (failedValues int) {
	var deliveryIsCancellable = make(map[string]bool)

	failed := atomic.Int32{}
	type V struct {
		deliveryID string
		value      bool
	}
	var res = make(chan V)
	go func() {
		wg := sync.WaitGroup{}
		for _, deliveryID := range deliveries {
			wg.Add(1)
			go func(deliveryID string) {
				defer wg.Done()
				value := client.Get(deliveryID)
				if value == nil {
					failed.Add(1)
					return
				}
				val, err := strconv.ParseBool(value.Val())
				if err != nil {
					log.Fatalf("cannot parse bool: %s", err)
				}
				res <- V{
					deliveryID: deliveryID,
					value:      val,
				}
			}(deliveryID)
		}
		wg.Wait()
		close(res)
	}()

	for v := range res {
		deliveryIsCancellable[v.deliveryID] = v.value
	}

	return failedValues
}

func readKeysWithConc(client *redis.ClusterClient, deliveries []string) (failedValues int) {
	var deliveryIsCancellable = make(map[string]bool)

	failed := atomic.Int32{}
	type V struct {
		deliveryID string
		value      bool
	}
	var res = make(chan V)
	var cn = pool.New().WithMaxGoroutines(10)
	go func() {
		for _, deliveryID := range deliveries {
			cn.Go(func() {
				value := client.Get(deliveryID)
				if value == nil {
					failed.Add(1)
					return
				}
				val, err := strconv.ParseBool(value.Val())
				if err != nil {
					log.Fatalf("cannot parse bool: %s", err)
				}
				res <- V{
					deliveryID: deliveryID,
					value:      val,
				}
			})
		}
		cn.Wait()
		close(res)
	}()

	for v := range res {
		deliveryIsCancellable[v.deliveryID] = v.value
	}

	return failedValues
}
