/*
 * Shea Odland, Von Castro, Eric Wedemire
 * CMPT315
 * Group Project: Codenames
 */

package main

import (
	"context"
	"log"

	"github.com/go-redis/redis/v8"
)

var ctx = context.Background()

func dbConnect() *redis.Client {
	redisDatabase := redis.NewClient(&redis.Options{
		Addr:     DBHOST + ":" + DBPORT,
		Password: "", // no password set
		DB:       0,  // use default DB
	})

	//test db connection
	ping, err := redisDatabase.Ping(ctx).Result()
	if err != nil {
		log.Println("Failure to open DB, ping result: " + ping)
		panic(err)
	}
	log.Println("Database successfully started")

	return redisDatabase
}
