package main

import (
	"fmt"
	"log"
	"time"

	"github.com/gocql/gocql"
)

// User is a simple model mapped to a Cassandra table.
type User struct {
	ID   gocql.UUID
	Name string
}

// Cassandra is a NoSQL database designed for scalability and high availability.
func main() {
	log.SetFlags(log.LstdFlags | log.Lshortfile)

	// Step 1: connect to Cassandra (system keyspace first)
	session := mustCreateSession("", 9042)
	defer session.Close()
	log.Println("connected to Cassandra (system keyspace)")

	// Step 2: create keyspace if not exists
	if err := createKeyspace(session); err != nil {
		log.Fatalf("create keyspace: %v", err)
	}
	log.Println("ensured keyspace playground exists")

	// Step 3: connect to our playground keyspace
	session.Close()
	session = mustCreateSession("playground", 9042)
	defer session.Close()
	log.Println("connected to keyspace playground")

	// Step 4: create table if not exists
	if err := createUsersTable(session); err != nil {
		log.Fatalf("create users table: %v", err)
	}
	log.Println("ensured table users exists")

	// Step 5: insert a user
	u := User{
		ID:   gocql.TimeUUID(),
		Name: "Cristi",
	}

	if err := insertUser(session, u); err != nil {
		log.Fatalf("insert user: %v", err)
	}
	log.Printf("inserted user: id=%s name=%s\n", u.ID.String(), u.Name)

	// Step 6: read all users back
	users, err := listUsers(session)
	if err != nil {
		log.Fatalf("list users: %v", err)
	}

	fmt.Println("Users in Cassandra:")
	for _, user := range users {
		fmt.Printf("  id=%s name=%s\n", user.ID.String(), user.Name)
	}
}

// mustCreateSession tries to create a session with simple retry logic.
func mustCreateSession(keyspace string, port int) *gocql.Session {
	cluster := gocql.NewCluster("127.0.0.1")
	cluster.Port = port
	if keyspace != "" {
		cluster.Keyspace = keyspace
	}
	// Consistency and timeout settings
	// Consistency level Quorum ensures that a majority of replicas respond to a read or write operation.
	// Timeout is set to 5 seconds to avoid hanging indefinitely.
	cluster.Consistency = gocql.Quorum
	cluster.Timeout = 5 * time.Second

	// Session creation with retries
	// Pointer for session because gocql.CreateSession returns a pointer.
	var session *gocql.Session
	var err error

	for i := 0; i < 10; i++ {
		session, err = cluster.CreateSession()
		if err == nil {
			return session
		}
		log.Printf("CreateSession failed (attempt %d): %v; retrying...", i+1, err)
		time.Sleep(5 * time.Second)
	}

	log.Fatalf("unable to connect to Cassandra after retries: %v", err)
	return nil
}

func createKeyspace(session *gocql.Session) error {
	const cql = `
CREATE KEYSPACE IF NOT EXISTS playground
WITH replication = {
  'class': 'SimpleStrategy',
  'replication_factor': 1
};`
	return session.Query(cql).Exec()
}

func createUsersTable(session *gocql.Session) error {
	const cql = `
CREATE TABLE IF NOT EXISTS users (
  id   uuid PRIMARY KEY,
  name text
);`
	return session.Query(cql).Exec()
}

func insertUser(session *gocql.Session, u User) error {
	const cql = `INSERT INTO users (id, name) VALUES (?, ?);`
	return session.Query(cql, u.ID, u.Name).Exec()
}

func listUsers(session *gocql.Session) ([]User, error) {
	const cql = `SELECT id, name FROM users;`

	iter := session.Query(cql).Iter()
	// Defer closing the iterator to free resources
	defer iter.Close()

	var users []User
	var id gocql.UUID
	var name string

	// iter.Scan populates the variables with the next row's data
	for iter.Scan(&id, &name) {
		users = append(users, User{ID: id, Name: name})
	}
	if err := iter.Close(); err != nil {
		return nil, err
	}

	// return users slice and nil error
	// nil means no error occurred
	return users, nil
}
