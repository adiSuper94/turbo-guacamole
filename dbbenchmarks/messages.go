package main

import (
	"container/list"
	"context"
	"fmt"
	"log"
	"math/rand"
	"os"
	"strconv"
	"strings"
	"time"

	// "github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	pgxUUID "github.com/vgarvardt/pgx-google-uuid/v5"
)

type Member struct {
	chat_room_id string
	user_id      string
}

func CreateDBConnection(connectionCount int32) *pgxpool.Pool {
	pgxConfig, err := pgxpool.ParseConfig("postgres://adisuper:password@localhost:5432/turbo?sslmode=disable")
	if err != nil {
		panic(err)
	}
	pgxConfig.MaxConns = connectionCount

	pgxConfig.AfterConnect = func(ctx context.Context, conn *pgx.Conn) error {
		pgxUUID.Register(conn.TypeMap())
		return nil
	}
	conn, err := pgxpool.NewWithConfig(context.TODO(), pgxConfig)
	if err != nil {
		panic(err)
	}
	return conn
}

func main() {
	var connectionCount int32
	connectionCount = 16
	duration := 2 * 60 * time.Second
	conn := CreateDBConnection(connectionCount)
	defer conn.Close()
	populateUsers(conn, 1000)
	populateChatRooms(conn, 10000)
	fmt.Println("Populated users and chat rooms")
	membersRows, err := conn.Query(context.Background(), `SELECT * FROM members`)
	var members []Member
	for membersRows.Next() {
		var member Member
		err := membersRows.Scan(&member.chat_room_id, &member.user_id)
		if err != nil {
			log.Fatal("Could not scan row", err)
		}
		members = append(members, member)
	}
	c := make(chan time.Duration)
	for i := 0; i < int(connectionCount); i++ {
		go populateMessages(conn, members, duration, c)
	}
	timeTaken := make([]time.Duration, 1024*8)
	done := 0
	for done < int(connectionCount) {
		t := <-c
		if t == -1*time.Second {
			done++
		} else {
			timeTaken = append(timeTaken, t)
		}
	}
	f, err := os.Create("result.txt")
	if err != nil {
		log.Fatal(err)
	}
	// remember to close the file
	defer f.Close()
	for i, t := range timeTaken {
		_, err := f.WriteString(strconv.Itoa(i) + "," + t.String() + "\n")
		if err != nil {
			log.Fatal(err)
		}
	}
}

func populateMessages(conn *pgxpool.Pool, members []Member, duration time.Duration, c chan time.Duration) {
	fmt.Println("Starting to insert messages")
	start := time.Now()
	rowsInserted := 0
	for true {
		memeber := members[rand.Intn(len(members))]
		args := pgx.NamedArgs{
			"chat_room_id": memeber.chat_room_id,
			"sender_id":    memeber.user_id,
			"body":         randomString(200),
			"created_at":   time.Now(),
			"modified_at":  time.Now(),
		}
		query := `INSERT INTO messages (chat_room_id ,sender_id, body, created_at, modified_at) VALUES (@chat_room_id, @sender_id, @body, @created_at, @modified_at)`
		queryStart := time.Now()
		_, err := conn.Exec(context.Background(), query, args)
		if err != nil {
			fmt.Println(args)
			log.Fatal("Could not execute INSERT Query", err)
		}
		queryEnd := time.Now()
		c <- queryEnd.Sub(queryStart)
		rowsInserted++
		if time.Since(start) > duration {
			break
		}
	}
	log.Println("Inserted ", rowsInserted, " rows in ", duration, " seconds")
	c <- -1 * time.Second
}

func populateUsers(conn *pgxpool.Pool, userCount int) *list.List {
	timeTaken := list.New()
	for i := 0; i < userCount; i++ {
		username := randomString(64)
		email := randomString(128)
		args := pgx.NamedArgs{
			"username":    username,
			"email_id":    email,
			"created_at":  time.Now(),
			"modified_at": time.Now(),
		}
		start := time.Now()
		_, err := conn.Exec(context.Background(), `INSERT INTO users (username, email_id, created_at, modified_at) VALUES (@username, @email_id, @created_at, @modified_at)`, args)
		if err != nil {
			log.Fatal("Could not execute INSERT Query", err)
			log.Fatal(err)
		}
		end := time.Now()
		timeTaken.PushBack(end.Sub(start))
	}
	return timeTaken
}

func populateChatRooms(conn *pgxpool.Pool, chatRoomCount int) {
	userRows, err := conn.Query(context.Background(), `SELECT id FROM users`)
	if err != nil {
		log.Fatal("Could not execute SELECT Query", err)
	}
	var userIds []string
	for userRows.Next() {
		var id string
		err := userRows.Scan(&id)
		if err != nil {
			log.Fatal("Could not scan row", err)
		}
		userIds = append(userIds, id)
	}
	for i := 0; i < chatRoomCount; i++ {
		name := randomString(64)
		args := pgx.NamedArgs{
			"name":        name,
			"created_at":  time.Now(),
			"modified_at": time.Now(),
		}
		var id string
		err := conn.QueryRow(context.Background(), `INSERT INTO chat_rooms (name, created_at, modified_at) VALUES (@name, @created_at, @modified_at) returning id`, args).Scan(&id)
		if err != nil {
			log.Fatal("Could not execute INSERT Query", err)
			log.Fatal(err)
		}
		userId1 := userIds[rand.Intn(len(userIds))]
		userId2 := userIds[rand.Intn(len(userIds))]
		for userId1 == userId2 {
			userId2 = userIds[rand.Intn(len(userIds))]
		}
		chatRoomArgs := pgx.NamedArgs{
			"chat_room_id": id,
			"user_id":      userId1,
		}
		_, err = conn.Exec(context.Background(), `INSERT INTO members (chat_room_id, user_id) VALUES (@chat_room_id, @user_id)`, chatRoomArgs)
		chatRoomArgs = pgx.NamedArgs{
			"chat_room_id": id,
			"user_id":      userId2,
		}
		_, err = conn.Exec(context.Background(), `INSERT INTO members (chat_room_id, user_id) VALUES (@chat_room_id, @user_id)`, chatRoomArgs)
	}
}

const charset = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789 "

func randomString(n int) string {
	sb := strings.Builder{}
	v := rand.Intn(n-5) + 5
	sb.Grow(v)
	for i := 0; i < v; i++ {
		sb.WriteByte(charset[rand.Intn(len(charset))])
	}
	return sb.String()
}
