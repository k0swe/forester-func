package kellog

import (
	adifpb "github.com/k0swe/adif-json-protobuf/go"
	"reflect"
	"testing"
)

func Test_mergeQso(t *testing.T) {
	type args struct {
		base     *adifpb.Qso
		backfill *adifpb.Qso
	}
	tests := []struct {
		name string
		args args
		want *adifpb.Qso
	}{
		{
			name: "Simple",
			args: args{
				base:     &adifpb.Qso{Band: "20m", Freq: 14.050},
				backfill: &adifpb.Qso{Band: "20m", Mode: "CW"},
			},
			want: &adifpb.Qso{Band: "20m", Freq: 14.050, Mode: "CW"},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_ = mergeQso(tt.args.base, tt.args.backfill)
			if !reflect.DeepEqual(tt.args.base, tt.want) {
				t.Errorf("adifToProto() got = %v, want %v", tt.args.base, tt.want)
			}
		})
	}
}
