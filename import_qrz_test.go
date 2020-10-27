package kellog

import (
	"github.com/golang/protobuf/proto"
	adifpb "github.com/k0swe/adif-json-protobuf/go"
	"testing"
)

func Test_mergeQso(t *testing.T) {
	type args struct {
		base     *adifpb.Qso
		backfill *adifpb.Qso
	}
	type returnValues struct {
		diff    bool
		wantQso *adifpb.Qso
	}
	tests := []struct {
		name string
		args args
		want returnValues
	}{
		{
			name: "Simple",
			args: args{
				base:     &adifpb.Qso{Band: "20m", Freq: 14.050},
				backfill: &adifpb.Qso{Band: "20m", Mode: "CW"},
			},
			want: returnValues{
				diff:    true,
				wantQso: &adifpb.Qso{Band: "20m", Freq: 14.050, Mode: "CW"},
			},
		},
		{
			name: "Simple no diff",
			args: args{
				base: &adifpb.Qso{Band: "20m", Freq: 14.050, Mode: "CW"},
				// Backfill freq is different, but it isn't used
				backfill: &adifpb.Qso{Band: "20m", Freq: 14.051},
			},
			want: returnValues{
				diff:    false,
				wantQso: &adifpb.Qso{Band: "20m", Freq: 14.050, Mode: "CW"},
			},
		},
		{
			name: "WsjtxPlusQrzcom",
			args: args{
				base: &adifpb.Qso{
					Band: "40m",
					Freq: 7.07595,
					Mode: "FT8",
					ContactedStation: &adifpb.Station{
						StationCall: "K9IJ",
						GridSquare:  "EN52",
						OpName:      "Johnny",
					},
				},
				backfill: &adifpb.Qso{
					Band: "40m",
					Freq: 7.07595,
					Mode: "FT8",
					ContactedStation: &adifpb.Station{
						StationCall: "K9IJ",
						GridSquare:  "EN52",
						OpName:      "JOHN F RICE",
						City:        "LAKE ZURICH",
						State:       "IL",
						Country:     "United States",
						Dxcc:        291,
					},
				},
			},
			want: returnValues{
				diff: true,
				wantQso: &adifpb.Qso{
					Band: "40m",
					Freq: 7.07595,
					Mode: "FT8",
					ContactedStation: &adifpb.Station{
						StationCall: "K9IJ",
						GridSquare:  "EN52",
						OpName:      "Johnny",
						City:        "LAKE ZURICH",
						State:       "IL",
						Country:     "United States",
						Dxcc:        291,
					},
				},
			},
		},
		{
			name: "QrzcomPlusQrzcom",
			args: args{
				base: &adifpb.Qso{
					Band: "40m",
					Freq: 7.07595,
					Mode: "FT8",
					ContactedStation: &adifpb.Station{
						StationCall: "K9IJ",
						GridSquare:  "EN52",
						OpName:      "JOHN F RICE",
						City:        "LAKE ZURICH",
						State:       "IL",
						Country:     "United States",
						Dxcc:        291,
					},
				},
				backfill: &adifpb.Qso{
					Band: "40m",
					Freq: 7.07595,
					Mode: "FT8",
					ContactedStation: &adifpb.Station{
						StationCall: "K9IJ",
						GridSquare:  "EN52",
						OpName:      "JOHN F RICE",
						City:        "LAKE ZURICH",
						State:       "IL",
						Country:     "United States",
						Dxcc:        291,
					},
				},
			},
			want: returnValues{
				diff: false,
				wantQso: &adifpb.Qso{
					Band: "40m",
					Freq: 7.07595,
					Mode: "FT8",
					ContactedStation: &adifpb.Station{
						StationCall: "K9IJ",
						GridSquare:  "EN52",
						OpName:      "JOHN F RICE",
						City:        "LAKE ZURICH",
						State:       "IL",
						Country:     "United States",
						Dxcc:        291,
					},
				},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			diff := mergeQso(tt.args.base, tt.args.backfill)
			if diff != tt.want.diff {
				t.Errorf("mergeQso() diff got = %v, want %v", diff, tt.want.diff)
			}
			if !proto.Equal(tt.args.base, tt.want.wantQso) {
				t.Errorf("mergeQso() qso got = %v, want %v", tt.args.base, tt.want.wantQso)
			}
		})
	}
}
