-- +goose Up
-- +goose StatementBegin
select create_distributed_table('messages', 'chat_id')
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
select undistribute_table('message')
-- +goose StatementEnd
