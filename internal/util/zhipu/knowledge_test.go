package zhipu

import (
	"reflect"
	"testing"
)

func Test_knowledge_Invoke(t *testing.T) {
	type fields struct {
		apiKey string
	}
	type args struct {
		content string
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		want    *KnowledgeInvokeResp
		wantErr bool
	}{
		{
			name: "1",
			fields: fields{
				apiKey: "7a909dd632f00f28a0d08efba999b3dc.50FHMHrklQIaaD16",
			},
			args: args{
				content: "如何下载app",
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			k := &knowledge{
				apiKey: tt.fields.apiKey,
			}
			got, err := k.Invoke(tt.args.content)
			if (err != nil) != tt.wantErr {
				t.Errorf("knowledge.Invoke() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("knowledge.Invoke() = %v, want %v", got, tt.want)
			}
		})
	}
}
