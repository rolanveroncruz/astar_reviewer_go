package main

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"os"

	_ "github.com/jackc/pgx/v5/stdlib"
	"google.golang.org/genai"

	"github.com/joho/godotenv"
	"github.com/rolanveroncruz/astar_reviewer_go/db_adjustments/internal/db"
	"github.com/rolanveroncruz/astar_reviewer_go/db_adjustments/internal/repli_questions"
)

type AIResponse struct {
	IsComplete        bool   `json:"is_complete"`
	IsClear           bool   `json:"is_clear"`
	IsUnambiguous     bool   `json:"is_unambiguous"`
	ImprovedQuestion  string `json:"improved_question"`
	AnswerExplanation string `json:"answer_explanation"`
}

func main() {
	if err := godotenv.Load(); err != nil {
		log.Fatal("Error loading .env file")
	}
	apiKey := os.Getenv("GEMINI_API_KEY")
	if apiKey == "" {
		log.Fatal("GEMINI_API_KEY is not set")
	}
	connStr := os.Getenv("DATABASE_URL")
	if connStr == "" {
		log.Fatal("DATABASE_URL is not set")
	}
	fmt.Println(fmt.Sprintf("Connection string is: %s", connStr))

	ctx := context.Background()
	client, err := genai.NewClient(ctx, &genai.ClientConfig{
		APIKey: apiKey,
	})
	if err != nil {
		log.Fatal(err)
	}

	sqlDB, err := sql.Open("pgx", connStr)
	if err != nil {
		log.Fatal(fmt.Sprintf("Error opening postgres: %s", err))
	}
	defer func(sqlDB *sql.DB) {
		_ = sqlDB.Close()
	}(sqlDB)

	if err := sqlDB.Ping(); err != nil {
		log.Fatal(fmt.Sprintf("Error in Ping(): %s", err))
	}
	ctx = context.Background()
	queries := db.New(sqlDB)

	questions, err := repli_questions.GetAllRepliQuestionsTransformed(ctx, queries)
	if err != nil {
		log.Fatal(err)
	}

	for i, question := range questions {
		resp, pQErr := processQuestion(ctx, client, question)
		if pQErr != nil {
			fmt.Printf("Error processing question %d: %s\n", i, pQErr)
		}
		fmt.Println(fmt.Sprintf("Question %d: %s\n\n Answer: %s\n\n is_complete: %v\n is_unambiguous:%v\n "+
			"is_clear:%v\n %s\n\n Explanation:%s\n",
			i,
			question.Question,
			*question.CorrectChoiceLetter+":"+*question.CorrectChoiceAnswer,
			resp.IsComplete,
			resp.IsUnambiguous,
			resp.IsClear,
			resp.ImprovedQuestion,
			resp.AnswerExplanation,
		))
		fmt.Println("----------------------------------------")

		isAcceptable := resp.IsComplete && resp.IsUnambiguous && resp.IsClear
		err := queries.UpdateRepliQuestionAcceptance(ctx, db.UpdateRepliQuestionAcceptanceParams{
			ID:                     question.ID,
			IsAnAcceptableQuestion: isAcceptable,
			RefinedQuestion:        sql.NullString{String: resp.ImprovedQuestion, Valid: true},
			AnsExplanation:         sql.NullString{String: resp.AnswerExplanation, Valid: true},
		})
		if err != nil {
			log.Println(fmt.Sprintf("UpdateRepliQuestionAcceptance() error: %s", err))
		}
	}
}

func formatQuestion(question repli_questions.RepliQuestionDTO) string {
	choices := " Choices are:\n"
	for _, ch := range question.Choices {
		choices += fmt.Sprintf("%s: %s\n", ch.Letter, ch.Answer)
	}
	answer := "The correct answer is:" + *question.CorrectChoiceLetter + ".: " + *question.CorrectChoiceAnswer
	return fmt.Sprintf("%s\n\n\n %s\n\n\n %s\n\n\n\n", question.Question, choices, answer)
}

func processQuestion(ctx context.Context, client *genai.Client, question repli_questions.RepliQuestionDTO) (*AIResponse, error) {
	intro := "We're administering a sample exam of multiple choice questions. Questions were scraped  from a" +
		"pdf file, and could have some errors. I will give you a question and your task is to say if" +
		"it is a complete, clear, and unambiguous question. If it fails in one of the criteria, provide a better " +
		"phrasing of the question without changing the multiple choice answers, nor the actual answer." +
		"Also give an explanation of why the correct answer is the best or only correct answer."
	prompt := fmt.Sprintf("%s\n The question is: %s", intro, formatQuestion(question))
	model := "gemini-3.1-pro-preview"

	content, err := client.Models.GenerateContent(
		ctx,
		model,
		genai.Text(prompt),
		&genai.GenerateContentConfig{
			ResponseMIMEType: "application/json",
			ResponseSchema: &genai.Schema{
				Type: "object",
				Properties: map[string]*genai.Schema{
					"is_complete":        {Type: "boolean"},
					"is_clear":           {Type: "boolean"},
					"is_unambiguous":     {Type: "boolean"},
					"improved_question":  {Type: "string"},
					"answer_explanation": {Type: "string"},
				},
				Required: []string{"is_complete", "is_clear", "is_unambiguous", "improved_question", "answer_explanation"},
			},
		},
	)
	if err != nil {
		return nil, err
	}
	var resp AIResponse
	if err := json.Unmarshal([]byte(content.Text()), &resp); err != nil {
		return nil, err
	}
	return &resp, nil

}
