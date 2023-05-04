package notifier

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestMetadata_toMap(t *testing.T) {
	type fields struct {
		AppName      string
		InstanceName string
		Commit       string
		BuildDate    string
	}
	tests := []struct {
		name    string
		fields  fields
		want    map[string]string
		wantLen int
	}{
		{
			name: "all fields are empty",
			fields: fields{
				AppName:      "",
				InstanceName: "",
				Commit:       "",
				BuildDate:    "",
			},
			want:    map[string]string{},
			wantLen: 0,
		},
		{
			name: "app name is set and commit",
			fields: fields{
				AppName:      "test",
				InstanceName: "",
				Commit:       "123",
				BuildDate:    "",
			},
			want: map[string]string{
				"app_name": "test",
				"commit":   "123",
			},
			wantLen: 2,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			m := Metadata{
				AppName:      tt.fields.AppName,
				InstanceName: tt.fields.InstanceName,
				Commit:       tt.fields.Commit,
				BuildDate:    tt.fields.BuildDate,
			}

			got := m.toMap()

			assert.Equal(t, fmt.Sprintf("%#v", tt.want), fmt.Sprintf("%#v", got))
			assert.Equal(t, tt.wantLen, len(got))
		})
	}
}
