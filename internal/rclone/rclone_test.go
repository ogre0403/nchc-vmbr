package rclone

import "testing"

func TestBuildS3Fs(t *testing.T) {
	cfg := S3Config{Endpoint: "s3.example.local:9000", AccessKey: "AKIA", SecretKey: "SECRET"}
	got := BuildS3Fs(cfg)
	if got == "" {
		t.Fatalf("expected non-empty fs string")
	}
	if !contains(got, "provider=Other") {
		t.Fatalf("expected provider=Other in fs, got %s", got)
	}
	if !contains(got, "endpoint='s3.example.local:9000'") {
		t.Fatalf("endpoint missing or not quoted: %s", got)
	}
	if !contains(got, "access_key_id=AKIA") {
		t.Fatalf("access key missing: %s", got)
	}
	if !contains(got, "secret_access_key=SECRET") {
		t.Fatalf("secret key missing: %s", got)
	}
}

// small helper to avoid importing strings in test for minimal footprint
func contains(s, sub string) bool {
	for i := 0; i+len(sub) <= len(s); i++ {
		if s[i:i+len(sub)] == sub {
			return true
		}
	}
	return false
}

func TestObjectExists(t *testing.T) {
	orig := rpc
	defer func() { rpc = orig }()

	cfg := S3Config{Endpoint: "e", AccessKey: "a", SecretKey: "s", Bucket: "b"}

	// Case: object exists
	rpc = func(ep, body string) (string, int) {
		if ep == "operations/stat" {
			return `{"item":{"Size":123}}`, 200
		}
		return "", 500
	}
	ok, err := ObjectExists(cfg, "some/path")
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if !ok {
		t.Fatalf("expected object to exist")
	}

	// Case: not found
	rpc = func(ep, body string) (string, int) {
		if ep == "operations/stat" {
			return "object not found", 404
		}
		return "", 500
	}
	ok, err = ObjectExists(cfg, "notfound")
	if err != nil {
		t.Fatalf("expected no error for not-found, got %v", err)
	}
	if ok {
		t.Fatalf("expected object to be absent")
	}

	// Case: RPC error
	rpc = func(ep, body string) (string, int) {
		return "internal failure", 500
	}
	ok, err = ObjectExists(cfg, "bad")
	if err == nil {
		t.Fatalf("expected error for RPC failure")
	}
	if ok {
		t.Fatalf("did not expect object to exist on RPC failure")
	}
}
