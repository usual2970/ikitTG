package audio

import (
	"context"
	"reflect"
	"testing"
)

func TestAzure(t *testing.T) {
	type args struct {
		ctx  context.Context
		text string
	}
	tests := []struct {
		name    string
		args    args
		want    []byte
		wantErr bool
	}{
		{
			name: "test",
			args: args{
				ctx:  context.Background(),
				text: "hello",
			},
			want:    []byte("test"),
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := Azure(tt.args.ctx, tt.args.text)
			if (err != nil) != tt.wantErr {
				t.Errorf("Azure() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("Azure() = %v, want %v", got, tt.want)
			}
		})
	}
}
