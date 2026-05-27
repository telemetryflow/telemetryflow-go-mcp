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

func TestNewResource(t *testing.T) {
	uri, _ := vo.NewResourceURI("file:///test.txt")
	r, err := entities.NewResource(uri, "Test")
	require.NoError(t, err)
	assert.Equal(t, "file:///test.txt", r.URI().String())
	assert.Equal(t, "Test", r.Name())
	assert.False(t, r.IsTemplate())
	assert.False(t, r.CreatedAt().IsZero())
	assert.False(t, r.UpdatedAt().IsZero())
}

func TestNewResourceTemplate(t *testing.T) {
	r, err := entities.NewResourceTemplate("file:///{path}", "File", "desc")
	require.NoError(t, err)
	assert.True(t, r.IsTemplate())
	assert.Equal(t, "file:///{path}", r.URITemplate())
	assert.Equal(t, "desc", r.Description())
}

func TestResource_SetName(t *testing.T) {
	uri, _ := vo.NewResourceURI("file:///test")
	r, _ := entities.NewResource(uri, "Old")
	r.SetName("New")
	assert.Equal(t, "New", r.Name())
}

func TestResource_SetDescription(t *testing.T) {
	uri, _ := vo.NewResourceURI("file:///test")
	r, _ := entities.NewResource(uri, "Test")
	r.SetDescription("updated")
	assert.Equal(t, "updated", r.Description())
}

func TestResource_SetMimeType(t *testing.T) {
	uri, _ := vo.NewResourceURI("file:///test")
	r, _ := entities.NewResource(uri, "Test")
	mt, _ := vo.NewMimeType("application/json")
	r.SetMimeType(mt)
	assert.Equal(t, "application/json", r.MimeType().String())
}

func TestResource_SetAnnotations(t *testing.T) {
	uri, _ := vo.NewResourceURI("file:///test")
	r, _ := entities.NewResource(uri, "Test")
	ann := &entities.ResourceAnnotations{Audience: []string{"user"}, Priority: 0.5}
	r.SetAnnotations(ann)
	assert.Equal(t, ann, r.Annotations())
}

func TestResource_SetReader(t *testing.T) {
	uri, _ := vo.NewResourceURI("file:///test")
	r, _ := entities.NewResource(uri, "Test")
	assert.Nil(t, r.Reader())
	reader := func(u string) (*entities.ResourceContent, error) {
		return &entities.ResourceContent{URI: u, Text: "data"}, nil
	}
	r.SetReader(reader)
	assert.NotNil(t, r.Reader())
}

func TestResource_Metadata(t *testing.T) {
	uri, _ := vo.NewResourceURI("file:///test")
	r, _ := entities.NewResource(uri, "Test")
	r.SetMetadata("key", "val")
	assert.Equal(t, "val", r.Metadata()["key"])
}

func TestResource_Read_NoReader(t *testing.T) {
	uri, _ := vo.NewResourceURI("file:///test")
	r, _ := entities.NewResource(uri, "Test")
	content, err := r.Read()
	require.NoError(t, err)
	assert.Equal(t, "file:///test", content.URI)
}

func TestResource_Read_WithReader(t *testing.T) {
	uri, _ := vo.NewResourceURI("file:///test")
	r, _ := entities.NewResource(uri, "Test")
	r.SetReader(func(u string) (*entities.ResourceContent, error) {
		return &entities.ResourceContent{URI: u, Text: "hello"}, nil
	})
	content, err := r.Read()
	require.NoError(t, err)
	assert.Equal(t, "hello", content.Text)
}

func TestResource_ToMCPResource(t *testing.T) {
	uri, _ := vo.NewResourceURI("file:///test")
	r, _ := entities.NewResource(uri, "Test")
	r.SetDescription("desc")
	mt, _ := vo.NewMimeType("text/plain")
	r.SetMimeType(mt)

	mcp := r.ToMCPResource()
	assert.Equal(t, "Test", mcp["name"])
	assert.Equal(t, "file:///test", mcp["uri"])
	assert.Equal(t, "desc", mcp["description"])
	assert.Equal(t, "text/plain", mcp["mimeType"])
}

func TestResource_ToMCPResource_Template(t *testing.T) {
	r, _ := entities.NewResourceTemplate("file:///{path}", "T", "")
	mcp := r.ToMCPResource()
	assert.Equal(t, "file:///{path}", mcp["uriTemplate"])
	_, hasURI := mcp["uri"]
	assert.False(t, hasURI)
}

func TestResource_ToMCPResource_WithAnnotations(t *testing.T) {
	uri, _ := vo.NewResourceURI("file:///test")
	r, _ := entities.NewResource(uri, "Test")
	r.SetAnnotations(&entities.ResourceAnnotations{
		Audience: []string{"user"},
		Priority: 0.9,
	})
	mcp := r.ToMCPResource()
	assert.Contains(t, mcp, "annotations")
	ann := mcp["annotations"].(*entities.ResourceAnnotations)
	assert.Equal(t, 0.9, ann.Priority)
}

func TestResource_ToJSON(t *testing.T) {
	uri, _ := vo.NewResourceURI("file:///test")
	r, _ := entities.NewResource(uri, "Test")
	data, err := r.ToJSON()
	require.NoError(t, err)
	assert.True(t, len(data) > 0)

	var m map[string]interface{}
	require.NoError(t, json.Unmarshal(data, &m))
	assert.Equal(t, "Test", m["name"])
}

func TestResourceList(t *testing.T) {
	rl := entities.NewResourceList()
	assert.True(t, rl.IsEmpty())
	assert.Equal(t, 0, rl.Count())

	uri, _ := vo.NewResourceURI("file:///a")
	r, _ := entities.NewResource(uri, "A")
	rl.Add(r)
	assert.False(t, rl.IsEmpty())
	assert.Equal(t, 1, rl.Count())
}

func TestResourceList_ToMCPResourceList(t *testing.T) {
	rl := entities.NewResourceList()
	uri, _ := vo.NewResourceURI("file:///a")
	r, _ := entities.NewResource(uri, "A")
	rl.Add(r)
	rl.NextCursor = "cursor123"

	mcp := rl.ToMCPResourceList()
	assert.Contains(t, mcp, "resources")
	assert.Contains(t, mcp, "nextCursor")

	rl2 := entities.NewResourceList()
	mcp2 := rl2.ToMCPResourceList()
	_, hasCursor := mcp2["nextCursor"]
	assert.False(t, hasCursor)
}

func TestResourceAnnotations_JSON(t *testing.T) {
	ann := &entities.ResourceAnnotations{
		Audience:    []string{"user", "assistant"},
		Priority:    0.8,
		Description: "Test annotation",
	}
	data, err := json.Marshal(ann)
	require.NoError(t, err)
	assert.Contains(t, string(data), "user")
	assert.Contains(t, string(data), "0.8")
}

func TestResourceContent_JSON(t *testing.T) {
	rc := entities.ResourceContent{URI: "file:///test", Text: "hello", MimeType: "text/plain"}
	data, err := json.Marshal(rc)
	require.NoError(t, err)
	assert.Contains(t, string(data), "hello")
}

func TestResource_UpdatedAt(t *testing.T) {
	uri, _ := vo.NewResourceURI("file:///test")
	r, _ := entities.NewResource(uri, "Test")
	before := r.UpdatedAt()
	time.Sleep(time.Millisecond)
	r.SetName("Updated")
	assert.True(t, r.UpdatedAt().After(before))
}
