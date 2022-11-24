package util

import (
	"os"
	"testing"
)

func TestGetEnvOrDefault(t *testing.T) {
	type args struct {
		envName      string
		defaultValue string
	}
	tests := []struct {
		name string
		args args
		want string
	}{
		{
			name: "Test with existing env",
			args: args{
				envName:      "good",
				defaultValue: "default",
			},
			want: "good",
		},
		{
			name: "Test with not existing env",
			args: args{
				envName:      "",
				defaultValue: "default",
			},
			want: "default",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if len(tt.args.envName) != 0 {
				os.Setenv(tt.args.envName, tt.args.envName)
				defer os.Unsetenv(tt.args.envName)
			}
			if got := GetEnvOrDefault(tt.args.envName, tt.args.defaultValue); got != tt.want {
				t.Errorf("GetEnvOrDefault() = %v, want %v", got, tt.want)
			}
		})
	}
}
