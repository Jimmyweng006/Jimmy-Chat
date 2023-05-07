-- name: CreateMessage :one
INSERT INTO messages (
  room_id,
  reply_message_id,
  sender_id,
  message_text
) VALUES (
  $1, $2, $3, $4
) RETURNING *;