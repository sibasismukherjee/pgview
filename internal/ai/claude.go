package ai

import (
	"bytes"
	"fmt"
	"os/exec"
	"strings"
)

const systemPrompt = `You are a PostgreSQL expert assistant embedded in a CLI database tool.
When given a schema and a user request, return ONLY a valid SQL query — no markdown fences,
no explanation, no preamble. End the query with a semicolon.
If the user asks to tune or fix a query, return the improved query only.`

// AskClaude sends a schema context + user prompt to the claude CLI
// and returns the raw SQL response.
func AskClaude(schemaContext, userPrompt string) (string, error) {
	if _, err := exec.LookPath("claude"); err != nil {
		return "", fmt.Errorf("'claude' CLI not found in PATH — install Claude Code to use \\ai")
	}

	fullPrompt := fmt.Sprintf(`%s

Database schema:
%s

Request: %s`, systemPrompt, schemaContext, userPrompt)

	cmd := exec.Command("claude", "-p", fullPrompt)
	var out, errBuf bytes.Buffer
	cmd.Stdout = &out
	cmd.Stderr = &errBuf

	if err := cmd.Run(); err != nil {
		return "", fmt.Errorf("claude error: %v\n%s", err, strings.TrimSpace(errBuf.String()))
	}

	return cleanSQL(strings.TrimSpace(out.String())), nil
}

// TuneQuery asks Claude to improve an existing SQL query given a hint.
func TuneQuery(schemaContext, existingSQL, hint string) (string, error) {
	prompt := fmt.Sprintf("Improve or fix this SQL query based on the following request: %s\n\nCurrent query:\n%s",
		hint, existingSQL)
	return AskClaude(schemaContext, prompt)
}

// cleanSQL strips markdown code fences that claude may include despite instructions.
func cleanSQL(s string) string {
	lines := strings.Split(s, "\n")
	var out []string
	for _, line := range lines {
		if strings.HasPrefix(strings.TrimSpace(line), "```") {
			continue
		}
		out = append(out, line)
	}
	return strings.TrimSpace(strings.Join(out, "\n"))
}
