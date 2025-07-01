-- name: CreateUser :one
INSERT INTO users (email, password_hash)
VALUES ($1, $2)
RETURNING *;

-- name: GetUserByEmail :one
SELECT * FROM users 
WHERE email = $1 LIMIT 1;

-- name: GetUserById :one
SELECT * FROM users 
WHERE id = $1 LIMIT 1;

-- name: CreateUserIfNotExists :one
INSERT INTO users (email, password_hash)
VALUES ($1, $2)
ON CONFLICT (email) DO NOTHING
RETURNING *;

-- name: CreateDocument :one
INSERT INTO documents (
    user_id, file_name, mime_type, file_data, file_size, file_hash
) VALUES ($1, $2, $3, $4, $5, $6)
ON CONFLICT (user_id, file_hash) DO NOTHING
RETURNING *;

-- name: GetDocumentBase64 :one
SELECT 
    id,
    file_name,
    mime_type,
    encode(file_data, 'base64') AS file_data_base64,
    status
FROM documents
WHERE id = $1 AND user_id = $2;

-- name: ListDocuments :many
-- @param user_id - UUID пользователя
-- @param search_query - поиск по file_name
-- @param mime_filter - фильтр по MIME-типу
-- @param status_filter - фильтр по статусу
-- @param result_limit - лимит
-- @param result_offset - смещение
SELECT 
    id,
    file_name,
    mime_type,
    file_size,
    status,
    chunkr_task_id,
    created_at
FROM documents
WHERE 
    user_id = @user_id AND
    file_name LIKE '%' || @search_query || '%' AND
    mime_type LIKE CASE 
        WHEN @mime_filter = '' THEN '%'
        ELSE @mime_filter 
    END AND
    status = CASE
        WHEN @status_filter::text = '' THEN status
        ELSE @status_filter::document_status
    END
ORDER BY created_at DESC
LIMIT @result_limit OFFSET @result_offset;

-- name: UpdateDocumentStatus :exec
UPDATE documents
SET 
    status = $2,
    updated_at = NOW()
WHERE id = $1;

-- name: GetPendingDocuments :many
SELECT *
FROM documents
WHERE status = 'pending'
ORDER BY created_at ASC
LIMIT $1;

-- name: GetDocumentsByStatus :many
SELECT *
FROM documents
WHERE status = $1
ORDER BY created_at ASC
LIMIT $2;

-- name: LockDocumentForProcessing :one
UPDATE documents
SET 
    status = 'processing',
    updated_at = NOW()
WHERE id = $1 AND status = 'pending'
RETURNING *;

-- name: ResetStuckDocuments :exec
UPDATE documents
SET status = 'pending'
WHERE status = 'processing' 
AND updated_at < NOW() - INTERVAL '60 minutes';

-- name: SetChunkrTaskID :one
UPDATE documents
SET 
    chunkr_task_id = $2,
    updated_at = NOW()
WHERE id = $1
RETURNING *;

-- name: GetProcessingDocuments :many
SELECT *
FROM documents
WHERE status = 'processing'
ORDER BY created_at ASC
LIMIT $1;

-- name: LockDocumentForChecking :one
UPDATE documents
SET 
    status = 'checking',
    updated_at = NOW()
WHERE id = $1 AND status = 'processing'
RETURNING *;

-- name: SetChunkrResult :one
UPDATE documents
SET 
    chunkr_result = $2,
    status = 'completed',
    updated_at = NOW()
WHERE id = $1
RETURNING *;
