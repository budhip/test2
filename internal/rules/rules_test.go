package rules

import (
	"testing"

	"github.com/hyperjumptech/grule-rule-engine/pkg"
	"github.com/stretchr/testify/assert"

	"bitbucket.org/Amartha/go-megatron/internal/config"
)

func Test_newRule(t *testing.T) {
	th := newTestHelper(t)

	RuleLoaderVariable = th.ruleLoader
	defer func() {
		RuleLoaderVariable = &FileRuleLoader{}
	}()

	type args struct {
		cfg      *config.Configuration
		name     string
		fileName string
	}
	tests := []struct {
		name    string
		args    args
		doMock  func()
		wantErr bool
	}{
		{
			name: "success new rule",
			args: args{
				cfg:      th.defaultConfig,
				name:     "default",
				fileName: "test",
			},
			doMock: func() {
				th.ruleLoader.EXPECT().
					LoadRule("test", "test", Version).
					Return(pkg.NewBytesResource(th.defaultRule), nil)
			},
		},
		{
			name: "error load rule",
			args: args{
				cfg:      th.defaultConfig,
				name:     "default",
				fileName: "test",
			},
			doMock: func() {
				th.ruleLoader.EXPECT().
					LoadRule("test", "test", Version).
					Return(nil, assert.AnError)
			},
			wantErr: true,
		},
		{
			name: "error build rule",
			args: args{
				cfg:      th.defaultConfig,
				name:     "default",
				fileName: "test",
			},
			doMock: func() {
				th.ruleLoader.EXPECT().
					LoadRule("test", "test", Version).
					Return(pkg.NewBytesResource([]byte(`INVALID_RULE_HERE`)), nil)
			},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.doMock != nil {
				tt.doMock()
			}

			_, err := newRule(tt.args.cfg, tt.args.name, tt.args.fileName)
			assert.Equal(t, tt.wantErr, err != nil, err)
		})
	}
}
