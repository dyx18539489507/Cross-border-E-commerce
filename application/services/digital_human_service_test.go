package services

import "testing"

func TestResolveDigitalHumanVideoURLFromRespData(t *testing.T) {
	result := &getResultData{
		RespData: `{"data":{"video_url":"https://example.com/video.mp4"}}`,
	}

	videoURL := resolveDigitalHumanVideoURL(result)
	if videoURL != "https://example.com/video.mp4" {
		t.Fatalf("expected video url from resp_data, got %q", videoURL)
	}
}

func TestDescribeDigitalHumanTaskFailure(t *testing.T) {
	result := &getResultData{
		RespData: `{"code":50430,"message":"API Concurrent Limit"}`,
	}

	message := describeDigitalHumanTaskFailure(result)
	if message != "code=50430: API Concurrent Limit" {
		t.Fatalf("unexpected failure message: %q", message)
	}
}

func TestDescribeDigitalHumanTaskFailureWithNestedJSON(t *testing.T) {
	result := &getResultData{
		RespData: "\"{\\\"code\\\":50514,\\\"message\\\":\\\"Pre Audio Risk Not Pass\\\"}\"",
	}

	message := describeDigitalHumanTaskFailure(result)
	if message != "code=50514: Pre Audio Risk Not Pass" {
		t.Fatalf("unexpected nested failure message: %q", message)
	}
}
