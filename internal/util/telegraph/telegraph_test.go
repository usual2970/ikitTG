package telegraph

import (
	"reflect"
	"testing"

	"gitlab.com/toby3d/telegraph"
)

func TestTelegraph_getAccount(t *testing.T) {
	type fields struct {
		conf *Config
	}
	tests := []struct {
		name    string
		fields  fields
		want    *telegraph.Account
		wantErr bool
	}{
		{
			name: "default",
			fields: fields{
				conf: &Config{
					ShortName:  defaultShortName,
					AuthorName: defaultAuthorName,
				},
			},
			want: &telegraph.Account{
				ShortName:  defaultShortName,
				AuthorName: defaultAuthorName,
			},
		},	
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tr := &Telegraph{
				conf: tt.fields.conf,
			}
			got, err := tr.getAccount()
			if (err != nil) != tt.wantErr {
				t.Errorf("Telegraph.getAccount() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("Telegraph.getAccount() = %v, want %v", got, tt.want)
			}
		})
	}
}
