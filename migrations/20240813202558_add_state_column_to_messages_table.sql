-- +goose Up
-- +goose StatementBegin
alter table messages
add state integer not null default 1;
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
alter table messages
drop column state;
-- +goose StatementEnd
