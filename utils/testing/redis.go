// Copyright © 2016 The Things Network
// Use of this source code is governed by the MIT license that can be found in the LICENSE file.

package testing

import (
	"fmt"
	"os"

	redis "gopkg.in/redis.v4"
)

// GetRedisClient returns a redis client that can be used for testing
func GetRedisClient() *redis.Client {
	host := os.Getenv("REDIS_HOST")
	if host == "" {
		host = "localhost"
	}
	return redis.NewClient(&redis.Options{
		Addr:     fmt.Sprintf("%s:6379", host),
		Password: "", // no password set
		DB:       1,  // use default DB
	})
}
