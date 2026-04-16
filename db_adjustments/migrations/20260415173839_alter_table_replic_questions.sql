-- +goose Up
ALTER TABLE repli_questions
    ADD COLUMN is_a_complete_question boolean NOT NULL DEFAULT false,
    ADD COLUMN refined_question text;

-- +goose Down
ALTER TABLE repli_questions
    DROP COLUMN IF EXISTS refined_question,
    DROP COLUMN IF EXISTS is_a_complete_question;