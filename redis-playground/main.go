package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"time"

	// Redis Go client v9
	// Redis is an in-memory data structure store, used as a database, cache, and message broker.
	"github.com/redis/go-redis/v9"
)

// User is a simple model stored as JSON in Redis.
type User struct {
	ID   int64  `json:"id"`
	Name string `json:"name"`
}

func main() {
	// Flags to include date, time, and file info in logs.
	// Lshortfile adds final file name element and line number.
	// LstdFlags is the standard date and time format.
	log.SetFlags(log.LstdFlags | log.Lshortfile)

	ctx := context.Background()

	// Connect to Redis on localhost:6379 (from docker-compose).
	rdb := redis.NewClient(&redis.Options{
		Addr: "localhost:6379",
	})
	defer func() {
		if err := rdb.Close(); err != nil {
			log.Printf("close redis: %v\n", err)
		}
	}()

	// Quick ping to verify connection.
	if err := ping(ctx, rdb); err != nil {
		log.Fatalf("redis ping failed: %v", err)
	}
	log.Println("connected to Redis")

	// Insert a new user.
	user, err := insertUser(ctx, rdb, "Cristi")
	if err != nil {
		log.Fatalf("insert user: %v", err)
	}
	log.Printf("inserted user: id=%d name=%s\n", user.ID, user.Name)

	// List all users.
	users, err := listUsers(ctx, rdb)
	if err != nil {
		log.Fatalf("list users: %v", err)
	}

	fmt.Println("Users in Redis:")
	for _, u := range users {
		fmt.Printf("  id=%d name=%s\n", u.ID, u.Name)
	}
}

func ping(ctx context.Context, rdb *redis.Client) error {
	ctx, cancel := context.WithTimeout(ctx, 3*time.Second)
	defer cancel()

	res, err := rdb.Ping(ctx).Result()
	if err != nil {
		return err
	}
	log.Printf("PING result: %s\n", res)
	return nil
}

// insertUser generates a new ID and stores the user as JSON at key "user:<id>".
func insertUser(ctx context.Context, rdb *redis.Client, name string) (User, error) {
	// Generate a new incremental ID.
	id, err := rdb.Incr(ctx, "user:next-id").Result()
	if err != nil {
		return User{}, fmt.Errorf("incr user:next-id: %w", err)
	}

	u := User{
		ID:   id,
		Name: name,
	}

	data, err := json.Marshal(u)
	if err != nil {
		return User{}, fmt.Errorf("marshal user: %w", err)
	}

	key := fmt.Sprintf("user:%d", id)

	if err := rdb.Set(ctx, key, data, 0).Err(); err != nil {
		return User{}, fmt.Errorf("set %s: %w", key, err)
	}

	return u, nil
}

// listUsers scans "user:*" keys (except the counter) and decodes them.
func listUsers(ctx context.Context, rdb *redis.Client) ([]User, error) {
	var (
		cursor uint64
		users  []User
	)

	for {
		keys, nextCursor, err := rdb.Scan(ctx, cursor, "user:*", 50).Result()
		if err != nil {
			return nil, fmt.Errorf("scan: %w", err)
		}

		for _, key := range keys {
			// Skip the counter key if it matches the pattern for some reason.
			if key == "user:next-id" {
				continue
			}

			data, err := rdb.Get(ctx, key).Bytes()
			if err != nil {
				return nil, fmt.Errorf("get %s: %w", key, err)
			}

			var u User
			// Unmarshal JSON data into User struct.
			if err := json.Unmarshal(data, &u); err != nil {
				return nil, fmt.Errorf("unmarshal %s: %w", key, err)
			}

			users = append(users, u)
		}

		cursor = nextCursor
		if cursor == 0 {
			break
		}
	}

	return users, nil
}
