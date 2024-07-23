-- +goose Up
-- +goose StatementBegin
create table messages
(
    message_id bigserial not null,
	chat_id text not null,
    sent_at timestamp not null,
    from_user_id text not null,
    to_user_id text not null,
    text text not null,
	primary key (message_id, chat_id)
)
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
drop table messages
-- +goose StatementEnd
