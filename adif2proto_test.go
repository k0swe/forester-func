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
				adifString: `QRZLogbook download for k0swe<eoh>
<band:3>40m<band_rx:3>20m<freq:5>7.282<freq_rx:6>14.282<eor>
<band:3>10m<band_rx:3>10m<freq:6>28.282<freq_rx:6>28.282<eor>
`,
				createTime: createTime,
			},
			want: &adifpb.Adif{
				Header: standardHeader,
				Qsos: []*adifpb.Qso{
					{
						Band:             "40m",
						BandRx:           "20m",
						Freq:             7.282,
						FreqRx:           14.282,
						LoggingStation:   &adifpb.Station{},
						ContactedStation: &adifpb.Station{},
						Propagation:      &adifpb.Propagation{},
					}, {
						Band:             "10m",
						BandRx:           "10m",
						Freq:             28.282,
						FreqRx:           28.282,
						LoggingStation:   &adifpb.Station{},
						ContactedStation: &adifpb.Station{},
						Propagation:      &adifpb.Propagation{},
					},
				},
			},
			wantErr: false,
		},
		{
			name: "App-defined",
			args: args{
				adifString: `QRZLogbook download for k0swe<eoh>
<app_qrzlog_logid:9>488692380<app_qrzlog_qsldate:8>20200406<app_qrzlog_status:1>C<eor>
<APP_MOMNPOP_HEIGHT:2>72<app_yourlog_weight:3>187<eor>
`,
				createTime: createTime,
			},
			want: &adifpb.Adif{
				Header: standardHeader,
				Qsos: []*adifpb.Qso{
					{
						AppDefined: map[string]string{
							"app_qrzlog_logid":   "488692380",
							"app_qrzlog_qsldate": "20200406",
							"app_qrzlog_status":  "C",
						},
						LoggingStation:   &adifpb.Station{},
						ContactedStation: &adifpb.Station{},
						Propagation:      &adifpb.Propagation{},
					}, {
						AppDefined: map[string]string{
							"app_momnpop_height": "72",
							"app_yourlog_weight": "187",
						},
						LoggingStation:   &adifpb.Station{},
						ContactedStation: &adifpb.Station{},
						Propagation:      &adifpb.Propagation{},
					},
				},
			},
			wantErr: false,
		},
		{
			name: "Stations",
			args: args{
				adifString: `QRZLogbook download for k0swe<eoh>
<call:4>KK9A<cnty:12>NC, Cabarrus<cont:2>NA<country:13>United States<cqz:1>5<dxcc:3>291
<email:13>john@kk9a.com<gridsquare:6>EM95re<lat:11>N035 12.004<lon:11>W080 31.464
<my_city:11>Westminster<my_cnty:13>CO, Jefferson<my_country:13>United States<my_cq_zone:1>4
<my_gridsquare:6>DM79lv<my_itu_zone:1>7<my_lat:11>N039 54.615<my_lon:11>W105 03.416
<my_name:20>Christopher C Keller<my_state:2>CO<name:12>JOHN P BAYNE<qsl_via:6>WD9DZV
<qth:7>MIDLAND<state:2>NC<station_callsign:5>K0SWE<tx_pwr:3>100<eor>
`,
				createTime: createTime,
			},
			want: &adifpb.Adif{
				Header: standardHeader,
				Qsos: []*adifpb.Qso{
					{
						LoggingStation: &adifpb.Station{
							City:        "Westminster",
							Country:     "United States",
							County:      "CO, Jefferson",
							CqZone:      4,
							GridSquare:  "DM79lv",
							ItuZone:     7,
							Latitude:    39.9103,
							Longitude:   -105.0569,
							OpName:      "Christopher C Keller",
							Power:       100,
							State:       "CO",
							StationCall: "K0SWE",
						},
						ContactedStation: &adifpb.Station{
							City:        "MIDLAND",
							Continent:   "NA",
							Country:     "United States",
							County:      "NC, Cabarrus",
							CqZone:      5,
							Dxcc:        291,
							Email:       "john@kk9a.com",
							GridSquare:  "EM95re",
							Latitude:    35.2001,
							Longitude:   -80.5244,
							OpName:      "JOHN P BAYNE",
							QslVia:      "WD9DZV",
							State:       "NC",
							StationCall: "KK9A",
						},
						Propagation: &adifpb.Propagation{},
					},
				},
			},
			wantErr: false,
		},
		{
			name: "Awards",
			args: args{
				adifString: `QRZLogbook download for k0swe<eoh>
<AWARD_SUBMITTED:61>ADIF_CENTURY_BASIC,ADIF_CENTURY_SILVER,ADIF_SPECTRUM_100-160m
<AWARD_GRANTED:41>ADIF_CENTURY_BASIC,ADIF_SPECTRUM_100-160m<eor>
`,
				createTime: createTime,
			},
			want: &adifpb.Adif{
				Header: standardHeader,
				Qsos: []*adifpb.Qso{
					{
						AwardSubmitted:   []string{"ADIF_CENTURY_BASIC", "ADIF_CENTURY_SILVER", "ADIF_SPECTRUM_100-160m"},
						AwardGranted:     []string{"ADIF_CENTURY_BASIC", "ADIF_SPECTRUM_100-160m"},
						LoggingStation:   &adifpb.Station{},
						ContactedStation: &adifpb.Station{},
						Propagation:      &adifpb.Propagation{},
					},
				},
			},
			wantErr: false,
		},
		{
			name: "Credit",
			args: args{
				adifString: `QRZLogbook download for k0swe<eoh>
<CREDIT_SUBMITTED:28>IOTA,WAS:LOTW&CARD,DXCC:CARD
<CREDIT_GRANTED:14>IOTA,DXCC:CARD<eor>
`,
				createTime: createTime,
			},
			want: &adifpb.Adif{
				Header: standardHeader,
				Qsos: []*adifpb.Qso{
					{
						CreditSubmitted: []*adifpb.Credit{
							{Credit: "IOTA"},
							{Credit: "WAS", QslMedium: "LOTW&CARD"},
							{Credit: "DXCC", QslMedium: "CARD"},
						},
						CreditGranted: []*adifpb.Credit{
							{Credit: "IOTA"},
							{Credit: "DXCC", QslMedium: "CARD"},
						},
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
