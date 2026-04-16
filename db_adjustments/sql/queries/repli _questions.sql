-- name: ListRepliQuestionsWithChoices :many
SELECT
    rq.id AS question_id,
    rq.from_question_id,
    rq.level_of_difficulty,
    rq.question,
    rq.correct_choice_id,
    rq.is_an_acceptable_question,
    rq.refined_question,

    rc.id AS choice_id,
    rc.letter AS choice_letter,
    rc.answer AS choice_answer,

    cc.letter AS correct_choice_letter,
    cc.answer AS correct_choice_answer
FROM repli_questions rq
         JOIN repli_choices rc
              ON rc.repli_question_id = rq.id
         LEFT JOIN repli_choices cc
                   ON cc.id = rq.correct_choice_id
ORDER BY rq.id, rc.letter;


-- name: UpdateRepliQuestionAcceptance :exec
UPDATE repli_questions
SET
    is_an_acceptable_question = $2,
    refined_question = $3,
    ans_explanation = $4
WHERE id = $1;