package image

import (
	"encoding/json"
	"testing"
)

func TestVolcEngineImageRequestMarshalIncludesWatermarkFalse(t *testing.T) {
	req := VolcEngineImageRequest{
		Model:                     "doubao-seedream-4-0-250828",
		Prompt:                    "test prompt",
		SequentialImageGeneration: "disabled",
		Size:                      "1K",
		Watermark:                 false,
	}

	b, err := json.Marshal(req)
	if err != nil {
		t.Fatalf("marshal request failed: %v", err)
	}

	var payload map[string]any
	if err := json.Unmarshal(b, &payload); err != nil {
		t.Fatalf("unmarshal request failed: %v", err)
	}

	value, ok := payload["watermark"]
	if !ok {
		t.Fatalf("expected watermark field to exist, got payload=%s", string(b))
	}
	boolValue, ok := value.(bool)
	if !ok {
		t.Fatalf("expected watermark field to be bool, got %T", value)
	}
	if boolValue {
		t.Fatalf("expected watermark=false, got true")
	}
}
