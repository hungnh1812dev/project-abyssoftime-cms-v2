package main

import "testing"

func TestResolveStorageProvider(t *testing.T) {
	tests := []struct {
		name    string
		env     string
		want    string
		wantErr bool
	}{
		{"empty defaults to cloudinary", "", "cloudinary", false},
		{"explicit cloudinary", "cloudinary", "cloudinary", false},
		{"explicit s3", "s3", "s3", false},
		{"unknown provider errors", "dropbox", "", true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := resolveStorageProvider(tt.env)
			if tt.wantErr {
				if err == nil {
					t.Fatalf("resolveStorageProvider(%q) error = nil, want error", tt.env)
				}
				return
			}
			if err != nil {
				t.Fatalf("resolveStorageProvider(%q) error = %v", tt.env, err)
			}
			if got != tt.want {
				t.Errorf("resolveStorageProvider(%q) = %q, want %q", tt.env, got, tt.want)
			}
		})
	}
}
