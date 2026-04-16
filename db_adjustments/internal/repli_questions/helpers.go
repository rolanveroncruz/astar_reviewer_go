package repli_questions

import (
	"context"

	"github.com/rolanveroncruz/astar_reviewer_go/db_adjustments/internal/db"
)

type RepliChoiceDTO struct {
	ID     int32  `json:"id"`
	Letter string `json:"letter"`
	Answer string `json:"answer"`
}

type RepliQuestionDTO struct {
	ID                     int32            `json:"id"`
	FromQuestionID         int32            `json:"from_question_id"`
	LevelOfDifficulty      int32            `json:"level_of_difficulty"`
	Question               string           `json:"question"`
	CorrectChoiceID        *int32           `json:"correct_choice_id,omitempty"`
	CorrectChoiceLetter    *string          `json:"correct_choice_letter,omitempty"`
	CorrectChoiceAnswer    *string          `json:"correct_choice_answer,omitempty"`
	IsAnAcceptableQuestion bool             `json:"is_an_acceptable_question"`
	RefinedQuestion        *string          `json:"refined_question,omitempty"`
	Choices                []RepliChoiceDTO `json:"choices"`
}

func GetAllRepliQuestionsTransformed(ctx context.Context, q *db.Queries) ([]RepliQuestionDTO, error) {
	rows, err := q.ListRepliQuestionsWithChoices(ctx)
	if err != nil {
		return nil, err
	}

	result := make([]RepliQuestionDTO, 0)
	indexByQuestionID := make(map[int32]int)

	for _, row := range rows {
		idx, exists := indexByQuestionID[row.QuestionID]
		if !exists {
			item := RepliQuestionDTO{
				ID:                     row.QuestionID,
				FromQuestionID:         row.FromQuestionID,
				LevelOfDifficulty:      row.LevelOfDifficulty,
				Question:               row.Question,
				IsAnAcceptableQuestion: row.IsAnAcceptableQuestion,
				Choices:                make([]RepliChoiceDTO, 0),
			}

			if row.CorrectChoiceID.Valid {
				v := row.CorrectChoiceID.Int32
				item.CorrectChoiceID = &v
			}

			if row.CorrectChoiceLetter.Valid {
				v := row.CorrectChoiceLetter.String
				item.CorrectChoiceLetter = &v
			}

			if row.CorrectChoiceAnswer.Valid {
				v := row.CorrectChoiceAnswer.String
				item.CorrectChoiceAnswer = &v
			}

			if row.RefinedQuestion.Valid {
				v := row.RefinedQuestion.String
				item.RefinedQuestion = &v
			}

			result = append(result, item)
			idx = len(result) - 1
			indexByQuestionID[row.QuestionID] = idx
		}

		result[idx].Choices = append(result[idx].Choices, RepliChoiceDTO{
			ID:     row.ChoiceID,
			Letter: row.ChoiceLetter,
			Answer: row.ChoiceAnswer,
		})
	}

	return result, nil
}
