package kellog

import (
	"github.com/golang/protobuf/ptypes"
	adifpb "github.com/k0swe/adif-json-protobuf/go"
	"reflect"
	"testing"
	"time"
)

func Test_adifToProto(t *testing.T) {
	createTime := time.Now()
	createStamp, _ := ptypes.TimestampProto(createTime)
	standardHeader := &adifpb.Header{
		AdifVersion:      "3.1.1",
		CreatedTimestamp: createStamp,
		ProgramId:        "kellog-func",
		ProgramVersion:   "0.0.1",
	}
	type args struct {
		adifString string
		createTime time.Time
	}
	tests := []struct {
		name    string
		args    args
		want    *adifpb.Adif
		wantErr bool
	}{
		{
			name: "Top Level",
			args: args{
				adifString: `<band:3>40m`,
				createTime: createTime,
			},
			want: &adifpb.Adif{
				Header: standardHeader,
				Qsos: []*adifpb.Qso{
					{
						Band:             "40m",
						LoggingStation:   &adifpb.Station{},
						ContactedStation: &adifpb.Station{},
						Propagation:      &adifpb.Propagation{},
					},
				},
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := adifToProto(tt.args.adifString, tt.args.createTime)
			if (err != nil) != tt.wantErr {
				t.Errorf("adifToProto() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("adifToProto() got = %v, want %v", got, tt.want)
			}
		})
	}
}
