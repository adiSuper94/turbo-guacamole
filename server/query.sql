-- name: InsertMessage :one
INSERT INTO messages (id, chat_room_id ,sender_id, body, created_at, modified_at)
  VALUES (@id, @chat_room_id, @sender_id, @body, @created_at, @modified_at) RETURNING *;

-- name: InsertUser :one
INSERT INTO users (username,  created_at, modified_at)
  VALUES (@username, @created_at, @modified_at) RETURNING *;

-- name: InsertChatRoom :one
INSERT INTO chat_rooms (name, created_at, modified_at) VALUES (@name, @created_at, @modified_at) returning *;

-- name: InsertMember :one
INSERT INTO members (chat_room_id, username) VALUES (@chat_room_id, @username) RETURNING *;

-- name: GetChatRoomMembers :many
SELECT members.*,  chat_rooms.name as chat_room_name FROM members
  INNER JOIN chat_rooms on chat_rooms.id = members.chat_room_id
  WHERE chat_rooms.id = @chat_room_id;

-- name: InsertMessageDelivery :one
INSERT INTO message_deliveries (message_id, chat_room_id, recipient_id, delivered)
  VALUES (@message_id, @chat_room_id, @recipient_id, @delivered) RETURNING *;

-- name: GetUserByUsername :one
SELECT * FROM users WHERE username = @username;

-- name: GetChatRoomDetailsByUsername :many
SELECT chat_rooms.* FROM members
  INNER JOIN chat_rooms on chat_rooms.id = members.chat_room_id
  WHERE members.username = @user_name;

-- name: GetChatRoomById :one
SELECT * FROM chat_rooms WHERE id = @id;

-- name: GetDMs :many
SELECT private_chats.chat_room_id, members.username FROM (SELECT members.chat_room_id AS chat_room_id FROM members
  GROUP BY members.chat_room_id HAVING count(members.username) = 2) AS private_chats
  INNER JOIN members ON members.chat_room_id = private_chats.chat_room_id WHERE members.username != @username;

-- name: GetMessagesByChatRoomId :many
SELECT chat_room_id, id, body, sender_id  FROM messages WHERE  messages.chat_room_id = @chat_room_id;
