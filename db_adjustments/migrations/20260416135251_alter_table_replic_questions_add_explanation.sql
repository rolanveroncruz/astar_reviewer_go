-- +goose Up
ALTER TABLE repli_questions
    ADD COLUMN ans_explanation text;

-- +goose Down
ALTER TABLE repli_questions
    DROP COLUMN IF EXISTS ans_explanation;
