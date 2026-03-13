// Package builtin provides built-in tool implementations
package builtin

import (
	"context"
	"fmt"
	"strings"

	"github.com/gemone/libcode/internal/tool/framework"
)

// QuestionTool asks users questions during execution
type QuestionTool struct{}

// NewQuestionTool creates a new question tool
func NewQuestionTool() *QuestionTool {
	return &QuestionTool{}
}

// ID returns the tool identifier
func (t *QuestionTool) ID() string {
	return "question"
}

// Description returns the tool description
func (t *QuestionTool) Description() string {
	return `Use this tool when you need to ask the user questions during execution. This allows you to:
1. Gather user preferences or requirements
2. Clarify ambiguous instructions
3. Get decisions on implementation choices as you work
4. Offer choices to the user about what direction to take.

Usage notes:
- When custom is enabled (default), a "Type your own answer" option is added automatically; don't include "Other" or catch-all options
- Answers are returned as arrays of labels; set multiple: true to allow selecting more than one
- If you recommend a specific option, make that the first option in the list and add "(Recommended)" at the end of the label`
}

// Parameters returns the parameter schema
func (t *QuestionTool) Parameters() *framework.Schema {
	return framework.NewSchema().
		AddProperty("questions", framework.Property{
			Type:        "array",
			Description: "Questions to ask",
			Required:    true,
		})
}

// Execute asks questions and returns user answers
func (t *QuestionTool) Execute(ctx context.Context, params map[string]any, execCtx *framework.ExecutionContext) (*framework.Result, error) {
	questionsRaw, ok := params["questions"].([]any)
	if !ok || len(questionsRaw) == 0 {
		return nil, fmt.Errorf("questions parameter is required and must be a non-empty array")
	}

	// Parse questions
	var questions []*framework.QuestionRequest
	for i, qRaw := range questionsRaw {
		qMap, ok := qRaw.(map[string]any)
		if !ok {
			return nil, fmt.Errorf("question at index %d is not an object", i)
		}

		question, _ := qMap["question"].(string)
		header, _ := qMap["header"].(string)
		multiple, _ := qMap["multiple"].(bool)

		optionsRaw, hasOptions := qMap["options"].([]any)
		if !hasOptions {
			return nil, fmt.Errorf("question at index %d missing options", i)
		}

		var options []*framework.QuestionOption
		for j, optRaw := range optionsRaw {
			optMap, ok := optRaw.(map[string]any)
			if !ok {
				return nil, fmt.Errorf("option %d in question %d is not an object", j, i)
			}

			label, _ := optMap["label"].(string)
			description, _ := optMap["description"].(string)

			options = append(options, &framework.QuestionOption{
				Label:       label,
				Description: description,
			})
		}

		questions = append(questions, &framework.QuestionRequest{
			Question: question,
			Header:   header,
			Options:  options,
			Multiple: multiple,
		})
	}

	// Check if question asker is available
	if execCtx.QuestionAsker == nil {
		return nil, fmt.Errorf("question asking is not available in this context")
	}

	// Ask all questions and collect answers
	allAnswers := make([][]string, len(questions))
	for i, question := range questions {
		answers, err := execCtx.QuestionAsker.Ask(question)
		if err != nil {
			return nil, fmt.Errorf("failed to ask question %d: %w", i, err)
		}
		allAnswers[i] = answers
	}

	// Format answers for display
	var formattedPairs []string
	for i, q := range questions {
		answers := allAnswers[i]
		formattedAnswer := formatAnswer(answers)
		formattedPairs = append(formattedPairs, fmt.Sprintf(`"%s"="%s"`, q.Question, formattedAnswer))
	}
	formatted := strings.Join(formattedPairs, ", ")

	return &framework.Result{
		Title:  fmt.Sprintf("Asked %d question%s", len(questions), pluralize(len(questions))),
		Output: fmt.Sprintf("User has answered your questions: %s. You can now continue with the user's answers in mind.", formatted),
		Metadata: map[string]any{
			"answers": allAnswers,
		},
	}, nil
}

// formatAnswer formats an answer array for display
func formatAnswer(answers []string) string {
	if len(answers) == 0 {
		return "Unanswered"
	}
	return strings.Join(answers, ", ")
}

// pluralize returns a plural suffix if needed
func pluralize(count int) string {
	if count == 1 {
		return ""
	}
	return "s"
}
