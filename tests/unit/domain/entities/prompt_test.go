package entities_test

import (
	"encoding/json"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/telemetryflow/telemetryflow-go-mcp/internal/domain/entities"
	vo "github.com/telemetryflow/telemetryflow-go-mcp/internal/domain/valueobjects"
)

func createTestPrompt(t *testing.T, name string) *entities.Prompt {
	t.Helper()
	pn, err := vo.NewToolName(name)
	require.NoError(t, err)
	p, err := entities.NewPrompt(pn, "A test prompt")
	require.NoError(t, err)
	return p
}

func TestNewPrompt(t *testing.T) {
	pn, _ := vo.NewToolName("test_prompt")
	p, err := entities.NewPrompt(pn, "A description")
	require.NoError(t, err)
	assert.Equal(t, "test_prompt", p.Name().String())
	assert.Equal(t, "A description", p.Description())
	assert.Empty(t, p.Arguments())
	assert.False(t, p.CreatedAt().IsZero())
}

func TestPrompt_SetDescription(t *testing.T) {
	p := createTestPrompt(t, "p")
	p.SetDescription("new desc")
	assert.Equal(t, "new desc", p.Description())
}

func TestPrompt_AddArgument(t *testing.T) {
	p := createTestPrompt(t, "p")
	p.AddArgument(&entities.PromptArgument{Name: "code", Description: "The code", Required: true})
	assert.Len(t, p.Arguments(), 1)
	assert.Equal(t, "code", p.Arguments()[0].Name)
}

func TestPrompt_RemoveArgument(t *testing.T) {
	p := createTestPrompt(t, "p")
	p.AddArgument(&entities.PromptArgument{Name: "a"})
	p.AddArgument(&entities.PromptArgument{Name: "b"})
	p.RemoveArgument("a")
	assert.Len(t, p.Arguments(), 1)
	assert.Equal(t, "b", p.Arguments()[0].Name)
}

func TestPrompt_RemoveArgument_NotFound(t *testing.T) {
	p := createTestPrompt(t, "p")
	p.AddArgument(&entities.PromptArgument{Name: "a"})
	p.RemoveArgument("nonexistent")
	assert.Len(t, p.Arguments(), 1)
}

func TestPrompt_GetArgument(t *testing.T) {
	p := createTestPrompt(t, "p")
	p.AddArgument(&entities.PromptArgument{Name: "x"})
	found := p.GetArgument("x")
	assert.NotNil(t, found)
	assert.Equal(t, "x", found.Name)
	assert.Nil(t, p.GetArgument("missing"))
}

func TestPrompt_RequiredArguments(t *testing.T) {
	p := createTestPrompt(t, "p")
	p.AddArgument(&entities.PromptArgument{Name: "required", Required: true})
	p.AddArgument(&entities.PromptArgument{Name: "optional", Required: false})
	req := p.RequiredArguments()
	assert.Len(t, req, 1)
	assert.Equal(t, "required", req[0].Name)
}

func TestPrompt_SetGenerator(t *testing.T) {
	p := createTestPrompt(t, "p")
	assert.Nil(t, p.Generator())
	gen := func(args map[string]string) (*entities.PromptMessages, error) {
		return &entities.PromptMessages{Description: "gen"}, nil
	}
	p.SetGenerator(gen)
	assert.NotNil(t, p.Generator())
}

func TestPrompt_Generate_NoGenerator(t *testing.T) {
	p := createTestPrompt(t, "p")
	msgs, err := p.Generate(nil)
	require.NoError(t, err)
	assert.Equal(t, "A test prompt", msgs.Description)
	assert.Len(t, msgs.Messages, 1)
	assert.Equal(t, "user", msgs.Messages[0].Role)
}

func TestPrompt_Generate_WithGenerator(t *testing.T) {
	p := createTestPrompt(t, "p")
	p.SetGenerator(func(args map[string]string) (*entities.PromptMessages, error) {
		return &entities.PromptMessages{
			Description: "custom",
			Messages: []entities.PromptMessage{
				{Role: "assistant", Content: entities.PromptContent{Type: "text", Text: "response"}},
			},
		}, nil
	})
	msgs, err := p.Generate(map[string]string{"key": "val"})
	require.NoError(t, err)
	assert.Equal(t, "custom", msgs.Description)
	assert.Equal(t, "response", msgs.Messages[0].Content.Text)
}

func TestPrompt_ValidateArguments(t *testing.T) {
	p := createTestPrompt(t, "p")
	p.AddArgument(&entities.PromptArgument{Name: "code", Required: true})

	err := p.ValidateArguments(map[string]string{"code": "x := 1"})
	assert.NoError(t, err)

	err = p.ValidateArguments(map[string]string{})
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "missing required argument: code")
}

func TestMissingArgumentError(t *testing.T) {
	e := &entities.MissingArgumentError{ArgumentName: "test"}
	assert.Equal(t, "missing required argument: test", e.Error())
}

func TestPrompt_Metadata(t *testing.T) {
	p := createTestPrompt(t, "p")
	p.SetMetadata("key", "val")
	assert.Equal(t, "val", p.Metadata()["key"])
}

func TestPrompt_Timestamps(t *testing.T) {
	p := createTestPrompt(t, "p")
	before := p.UpdatedAt()
	time.Sleep(time.Millisecond)
	p.SetDescription("changed")
	assert.True(t, p.UpdatedAt().After(before))
}

func TestPrompt_ToMCPPrompt(t *testing.T) {
	p := createTestPrompt(t, "my_prompt")
	p.AddArgument(&entities.PromptArgument{Name: "input", Description: "The input", Required: true})

	mcp := p.ToMCPPrompt()
	assert.Equal(t, "my_prompt", mcp["name"])
	assert.Equal(t, "A test prompt", mcp["description"])
	args, ok := mcp["arguments"].([]map[string]interface{})
	require.True(t, ok)
	assert.Len(t, args, 1)
	assert.Equal(t, "input", args[0]["name"])
}

func TestPrompt_ToMCPPrompt_NoDescription(t *testing.T) {
	pn, _ := vo.NewToolName("bare")
	p, _ := entities.NewPrompt(pn, "")
	mcp := p.ToMCPPrompt()
	assert.Equal(t, "bare", mcp["name"])
	_, hasDesc := mcp["description"]
	assert.False(t, hasDesc)
}

func TestPrompt_ToJSON(t *testing.T) {
	p := createTestPrompt(t, "json_prompt")
	data, err := p.ToJSON()
	require.NoError(t, err)
	assert.True(t, len(data) > 0)

	var m map[string]interface{}
	require.NoError(t, json.Unmarshal(data, &m))
	assert.Equal(t, "json_prompt", m["name"])
}

func TestPromptList(t *testing.T) {
	pl := entities.NewPromptList()
	assert.True(t, pl.IsEmpty())
	assert.Equal(t, 0, pl.Count())

	p := createTestPrompt(t, "p1")
	pl.Add(p)
	assert.False(t, pl.IsEmpty())
	assert.Equal(t, 1, pl.Count())
}

func TestPromptList_ToMCPPromptList(t *testing.T) {
	pl := entities.NewPromptList()
	p := createTestPrompt(t, "p1")
	pl.Add(p)
	pl.NextCursor = "cur"

	mcp := pl.ToMCPPromptList()
	assert.Contains(t, mcp, "prompts")
	assert.Contains(t, mcp, "nextCursor")

	pl2 := entities.NewPromptList()
	mcp2 := pl2.ToMCPPromptList()
	_, hasCursor := mcp2["nextCursor"]
	assert.False(t, hasCursor)
}

func TestPromptContent(t *testing.T) {
	pc := entities.PromptContent{Type: "text", Text: "hello"}
	data, err := json.Marshal(pc)
	require.NoError(t, err)
	assert.Contains(t, string(data), "hello")
}

func TestPromptMessage(t *testing.T) {
	pm := entities.PromptMessage{
		Role:    "user",
		Content: entities.PromptContent{Type: "text", Text: "hi"},
	}
	data, err := json.Marshal(pm)
	require.NoError(t, err)
	assert.Contains(t, string(data), "user")
}

func TestPromptMessages(t *testing.T) {
	pm := &entities.PromptMessages{
		Description: "test",
		Messages:    []entities.PromptMessage{{Role: "user", Content: entities.PromptContent{Type: "text", Text: "hi"}}},
	}
	data, err := json.Marshal(pm)
	require.NoError(t, err)
	assert.Contains(t, string(data), "test")
}

func TestPromptArgument(t *testing.T) {
	pa := entities.PromptArgument{Name: "x", Description: "desc", Required: true}
	data, err := json.Marshal(pa)
	require.NoError(t, err)
	assert.Contains(t, string(data), "x")
}
