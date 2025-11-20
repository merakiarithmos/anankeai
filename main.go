package main

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"os"
	"strconv"
	"strings"

	"github.com/joho/godotenv"
	"github.com/tmc/langchaingo/llms"
	"github.com/tmc/langchaingo/llms/ollama"
	"github.com/tmc/langchaingo/outputparser"
)

type Movie struct {
	Title    string `json:"title"`
	Director string `json:"director"`
	Year     int    `json:"year"`
}

func readOrCreateFile(filename string) ([]byte, error) {
	fileData, err := os.ReadFile(filename)

	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			log.Printf("%s not found creating it.\n", filename)

			file, err := os.Create(filename)
			if err != nil {
				return nil, err

			}
			file.Close()
			fmt.Printf("created empty file: %s\n", filename)
		} else {
			return nil, err
		}
	}
	return fileData, nil

}

func callLocalModel(ctx context.Context, prompt string) (string, error) {
	model := "mistral:7b"

	if v := os.Getenv("OLLAMA_MODEL_NAME"); v != "" {
		model = v
	}

	log.Println("Extracted model name")

	llm, err := ollama.New(ollama.WithModel(model))
	if err != nil {
		log.Fatalf("failure to create model, err: %v", err)
	}
	var msgs []llms.MessageContent

	// system message defines the available tools
	msgs = append(msgs, llms.TextParts(llms.ChatMessageTypeSystem, "You are a helpful assistant."))
	msgs = append(msgs, llms.TextParts(llms.ChatMessageTypeHuman, prompt))

	log.Println("Calling model")
	completion, err := llm.GenerateContent(ctx, msgs)
	if err != nil {
		log.Fatalf("failure to call model, err: %v", err)
	}
	llmResponse := completion.Choices[0].Content

	return llmResponse, nil
}

func (m *Movie) UnmarshalJSON(b []byte) error {
	// decode to a temporary map
	var raw map[string]interface{}
	if err := json.Unmarshal(b, &raw); err != nil {
		return err
	}

	// required string fields
	if v, ok := raw["title"].(string); ok {
		m.Title = v
	}

	if v, ok := raw["director"].(string); ok {
		m.Director = v
	}

	// handle year which might be float, string, int, etc
	switch v := raw["year"].(type) {
	case float64:
		m.Year = int(v)
	case string:
		i, err := strconv.Atoi(v)
		if err != nil {
			return fmt.Errorf("year is string but not numeric: %v", v)
		}
		m.Year = i
	case int:
		m.Year = v
	default:
		return fmt.Errorf("unexpected type for year: %T", v)
	}
	return nil
}

func main() {
	godotenv.Load()

	ctx := context.Background()
	modelResponse, err := callLocalModel(ctx, "What is the capital of France?")
	if err != nil {
		panic(err)
	}
	fmt.Println(modelResponse)

	fields := []outputparser.ResponseSchema{
		{
			Name:        "title",
			Description: "The movie title",
		},
		{
			Name:        "director",
			Description: "The movie title",
		},
		{
			Name:        "year",
			Description: "The movie title",
		},
	}

	parser := outputparser.NewStructured(fields)
	formatInstructions := parser.GetFormatInstructions()

	prompt := fmt.Sprintf(
		`
	Return ONLY a valid JSON array.
	No explanation.
	No markdown.
	No code fences.
	No backticks.
	No text before or after.

	Generate 20 random movies. 

	You MUST return a JSON array of 20 items, where each item matches this schema:
	%s

	Return ONLY the JSON array, no text before or after.
	`, formatInstructions)

	model := "mistral:7b"

	if v := os.Getenv("OLLAMA_MODEL_NAME"); v != "" {
		model = v
	}

	log.Println("Extracted model name")

	llm, err := ollama.New(ollama.WithModel(model))
	if err != nil {
		panic(err)
	}
	resp, err := llms.GenerateFromSinglePrompt(ctx, llm, prompt)
	fmt.Println(resp)
	if err != nil {
		panic(err)
	}
	raw, err := parser.Parse(resp)
	if err != nil {
		panic(err)
	}
	fmt.Println(raw)

	jsonBytes, _ := json.Marshal(raw)
	var movies []Movie

	clean := cleanJSON(string(jsonBytes))

	err = json.Unmarshal([]byte(clean), &movies)
	if err != nil {
		log.Fatal(err)
	}

	for _, m := range movies {
		fmt.Printf("%s (%d), Dir: %s\n", m.Title, m.Year, m.Director)
	}

	writeMoviesToFile(movies, "movies.json")
}

func cleanJSON(raw string) string {
	fmt.Println("printing raw string")
	fmt.Println(raw)
	raw = strings.TrimSpace(raw)

	raw = strings.TrimPrefix(raw, "```json")
	raw = strings.TrimPrefix(raw, "```")
	raw = strings.TrimSuffix(raw, "```")
	return strings.TrimSpace(raw)
}

func writeMoviesToFile(movies []Movie, filename string) error {
	out, err := json.MarshalIndent(movies, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(filename, out, 0644)
}
