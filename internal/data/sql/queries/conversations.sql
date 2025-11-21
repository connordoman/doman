-- name: CreateConversation :one
INSERT INTO conversations (id, model, service)
VALUES (?, ?, ?)
RETURNING *;

-- name: GetConversation :one
SELECT * FROM conversations
WHERE id = ? LIMIT 1;

-- name: UpdateConversationTimestamp :exec
UPDATE conversations
SET updated_at = CURRENT_TIMESTAMP
WHERE id = ?;

-- name: ListConversations :many
SELECT * FROM conversations
ORDER BY updated_at DESC
LIMIT ? OFFSET ?;

-- name: CreateMessage :one
INSERT INTO messages (conversation_id, role, content)
VALUES (?, ?, ?)
RETURNING *;

-- name: GetMessagesByConversationId :many
SELECT * FROM messages
WHERE conversation_id = ?
ORDER BY created_at ASC;

-- name: DeleteConversation :exec
DELETE FROM conversations
WHERE id = ?;

