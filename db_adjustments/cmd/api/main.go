package main

import (
	"context"
	"database/sql"
	"fmt"
	"log"
	"os"

	"github.com/joho/godotenv"
	"github.com/rolanveroncruz/astar_reviewer_go/db_adjustments/internal/db"
	"github.com/rolanveroncruz/astar_reviewer_go/db_adjustments/internal/repli_questions"

	"google.golang.org/genai"
)

func main() {
	if err := godotenv.Load(); err != nil {
		log.Fatal("Error loading .env file")
	}
	apiKey := os.Getenv("GEMINI_API_KEY")
	if apiKey == "" {
		log.Fatal("GEMINI_API_KEY is not set")
	}
	ctx := context.Background()
	client, err := genai.NewClient(ctx, &genai.ClientConfig{
		APIKey: apiKey,
	})
	if err != nil {
		log.Fatal(err)
	}

	result, err := client.Models.GenerateContent(
		ctx,
		"gemini-3-flash-preview",
		genai.Text("Explain how AI works in a few words"),
		nil,
	)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println(result.Text())
	connStr := os.Getenv("DB_URL")
	sqlDB, err := sql.Open("postgres", connStr)
	if err != nil {
		log.Fatal(err)
	}
	defer func(sqlDB *sql.DB) {
		_ = sqlDB.Close()
	}(sqlDB)

	if err := sqlDB.Ping(); err != nil {
		log.Fatal(err)
	}
	ctx = context.Background()
	queries := db.New(sqlDB)

	questions, err := repli_questions.GetAllRepliQuestionsTransformed(ctx, queries)
	if err != nil {
		log.Fatal(err)
	}

	for i, question := range questions {
		fmt.Print(formatQuestion(i, question))
		if i >= 10 {
			break
		}
	}
}

func formatQuestion(i int, question repli_questions.RepliQuestionDTO) string {
	choices := "   choices:("
	for j, ch := range question.Choices {
		choices += fmt.Sprintf(" %d, %s: %s ", j, ch.Letter, ch.Answer)
	}
	choices += ")"
	answer := question.CorrectChoiceAnswer
	return fmt.Sprintf("%d: %s %s %s\n", i, question.Question, choices, *answer)
}
