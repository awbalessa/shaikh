-- +goose Up
-- +goose StatementBegin
UPDATE public.users
SET is_admin = true
WHERE email = 'azizalessa321@gmail.com';
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
UPDATE public.users
SET is_admin = false
WHERE email = 'azizalessa321@gmail.com';
-- +goose StatementEnd
