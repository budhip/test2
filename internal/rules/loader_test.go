package rules

import (
	"testing"

	"github.com/hyperjumptech/grule-rule-engine/pkg"
	"github.com/stretchr/testify/assert"
)

func TestFileRuleLoader_LoadRule(t *testing.T) {
	type args struct {
		name string
		env  string
		in2  string
	}
	tests := []struct {
		name    string
		args    args
		want    pkg.Resource
		wantErr bool
	}{
		{
			name: "success load rule",
			args: args{
				name: "test",
				env:  "test",
				in2:  "",
			},
			want:    pkg.NewFileResource("./rules/test/test"),
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			l := FileRuleLoader{}
			got, err := l.LoadRule(tt.args.name, tt.args.env, tt.args.in2)
			assert.Equal(t, tt.wantErr, err != nil, err)
			assert.Equal(t, tt.want, got)
		})
	}
}
