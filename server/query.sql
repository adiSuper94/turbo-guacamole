-- name: InsertMessage :one
INSERT INTO messages (id, chat_room_id ,sender_id, body, created_at, modified_at)
  VALUES (@id, @chat_room_id, @sender_id, @body, @created_at, @modified_at) RETURNING *;

-- name: InsertUser :one
INSERT INTO users (username,  created_at, modified_at)
  VALUES (@username, @created_at, @modified_at) RETURNING *;

-- name: InsertChatRoom :one
INSERT INTO chat_rooms (name, created_at, modified_at) VALUES (@name, @created_at, @modified_at) returning *;

-- name: InsertMember :one
INSERT INTO members (chat_room_id, user_id) VALUES (@chat_room_id, @user_id) RETURNING *;

-- name: GetChatRoomMembers :many
SELECT members.*, users.username, chat_rooms.name as chat_room_name FROM members
  INNER JOIN users on users.id = member.user_id  INNER JOIN chat_rooms on chat_rooms.id = members.chat_room_id
  WHERE chat_rooms.id = @chat_room_id;

-- name: InsertMessageDelivery :one
INSERT INTO message_deliveries (message_id, chat_room_id, recipient_id, delivered)
  VALUES (@message_id, @chat_room_id, @recipient_id, @delivered) RETURNING *;

-- name: GetUserByUsername :one
SELECT * FROM users WHERE username = @username;
