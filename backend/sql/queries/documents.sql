-- name: GetDocumentByKey :one
SELECT id, source, document FROM documents
WHERE surah = $1 AND ayah = $2;

-- name: GetDocumentByID :one
SELECT source, document, surah, ayah FROM documents
WHERE id = $1;
