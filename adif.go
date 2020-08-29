package kellog

import (
	"github.com/Matir/adifparser"
	"github.com/golang/protobuf/ptypes"
	"github.com/golang/protobuf/ptypes/timestamp"
	"github.com/xylo04/kellog-qrz-sync/generated/adifpb"
	"io"
	"log"
	"math"
	"regexp"
	"strconv"
	"strings"
	"time"
)

func adifToJson(adifString string) ([]*adifpb.Qso, error) {
	reader := adifparser.NewADIFReader(strings.NewReader(adifString))
	qsos := make([]*adifpb.Qso, reader.RecordCount())
	record, err := reader.ReadRecord()
	for err == nil {
		qsos = append(qsos, recordToQso(record))
		record, err = reader.ReadRecord()
	}
	if err != io.EOF {
		return nil, err
	}
	return qsos, nil
}

func recordToQso(record adifparser.ADIFRecord) *adifpb.Qso {
	qso := new(adifpb.Qso)
	qso.Band, _ = record.GetValue("band")
	qso.BandRx, _ = record.GetValue("band_rx")
	qso.Comment, _ = record.GetValue("comment")
	qso.DistanceKm = getUint32(record, "distance")
	qso.Freq = getFloat64(record, "freq")
	qso.FreqRx = getFloat64(record, "freq_rx")
	qso.Mode, _ = record.GetValue("mode")
	qso.Notes, _ = record.GetValue("notes")
	qso.PublicKey, _ = record.GetValue("public_key")
	qso.Complete, _ = record.GetValue("qso_complete")
	qso.TimeOn = getTimestamp(record, "qso_date", "time_on")
	qso.TimeOff = getTimestamp(record, "qso_date_off", "time_off")
	qso.Random = getBool(record, "random")
	qso.RstReceived, _ = record.GetValue("rst_rcvd")
	qso.RstSent, _ = record.GetValue("rst_sent")
	qso.Submode, _ = record.GetValue("submode")
	qso.Swl = getBool(record, "swl")

	// TODO: where to put this?
	_, _ = record.GetValue("app_qrzlog_logid")

	qso.ContactedStation = new(adifpb.Station)
	qso.ContactedStation.Address, _ = record.GetValue("address")
	qso.ContactedStation.Age = getUint32(record, "age")
	qso.ContactedStation.StationCall, _ = record.GetValue("call")
	qso.ContactedStation.County, _ = record.GetValue("cnty")
	qso.ContactedStation.Continent, _ = record.GetValue("cont")
	qso.ContactedStation.OpCall, _ = record.GetValue("contacted_op")
	qso.ContactedStation.Country, _ = record.GetValue("country")
	qso.ContactedStation.CqZone = getUint32(record, "cqz")
	qso.ContactedStation.DarcDok, _ = record.GetValue("darc_dok")
	qso.ContactedStation.Dxcc = getUint32(record, "dxcc")
	qso.ContactedStation.Email, _ = record.GetValue("email")
	qso.ContactedStation.OwnerCall, _ = record.GetValue("eq_call")
	qso.ContactedStation.Fists = getUint32(record, "fists")
	qso.ContactedStation.FistsCc = getUint32(record, "fists_cc")
	qso.ContactedStation.GridSquare, _ = record.GetValue("gridsquare")
	qso.ContactedStation.Iota, _ = record.GetValue("iota")
	qso.ContactedStation.IotaIslandId = getUint32(record, "iota_island_id")
	qso.ContactedStation.ItuZone = getUint32(record, "ituz")
	qso.ContactedStation.Latitude = getLatLon(record, "lat")
	qso.ContactedStation.Longitude = getLatLon(record, "lon")
	qso.ContactedStation.OpName, _ = record.GetValue("name")
	qso.ContactedStation.Pfx, _ = record.GetValue("pfx")
	qso.ContactedStation.QslVia, _ = record.GetValue("qsl_via")
	qso.ContactedStation.City, _ = record.GetValue("qth")
	qso.ContactedStation.Region, _ = record.GetValue("region")
	qso.ContactedStation.Rig, _ = record.GetValue("rig")
	qso.ContactedStation.Power = getFloat64(record, "rx_pwr")
	qso.ContactedStation.Sig, _ = record.GetValue("sig")
	qso.ContactedStation.SigInfo, _ = record.GetValue("sig_info")
	qso.ContactedStation.SilentKey = getBool(record, "silent_key")
	qso.ContactedStation.Skcc, _ = record.GetValue("skcc")
	qso.ContactedStation.SotaRef, _ = record.GetValue("sota_ref")
	qso.ContactedStation.State, _ = record.GetValue("state")
	qso.ContactedStation.TenTen = getUint32(record, "ten_ten")
	qso.ContactedStation.Uksmg = getUint32(record, "uksmg")
	qso.ContactedStation.UsacaCounties, _ = record.GetValue("usaca_counties")
	qso.ContactedStation.VuccGrids, _ = record.GetValue("vucc_grids")
	qso.ContactedStation.Web, _ = record.GetValue("web")
	return qso
}

func getLatLon(record adifparser.ADIFRecord, field string) float64 {
	st, _ := record.GetValue(field)
	r := regexp.MustCompile(`([NESW])(\d+) ([\d.]+)`)
	groups := r.FindStringSubmatch(st)
	cardinal := groups[1]
	degrees, _ := strconv.ParseFloat(groups[2], 64)
	minutes, _ := strconv.ParseFloat(groups[3], 64)
	retval := degrees + (minutes / 60.0)
	if cardinal == "S" || cardinal == "W" {
		retval *= -1
	}
	// 4 decimal places is enough; https://xkcd.com/2170/
	retval = math.Round(retval*10000) / 10000
	return retval
}

func getBool(record adifparser.ADIFRecord, field string) bool {
	st, _ := record.GetValue(field)
	return st == "Y"
}

func getFloat64(record adifparser.ADIFRecord, field string) float64 {
	st, _ := record.GetValue(field)
	fl, _ := strconv.ParseFloat(st, 64)
	return fl
}

func getUint32(record adifparser.ADIFRecord, field string) uint32 {
	s, _ := record.GetValue(field)
	i64, _ := strconv.ParseUint(s, 10, 32)
	return uint32(i64)
}

func getTimestamp(record adifparser.ADIFRecord, dateField string, timeField string) *timestamp.Timestamp {
	dateStr, _ := record.GetValue(dateField)
	timeStr, _ := record.GetValue(timeField)
	t, err := time.Parse("20060102 1504", dateStr+" "+timeStr)
	if err != nil {
		log.Print(err)
	}
	ts, err := ptypes.TimestampProto(t)
	if err != nil {
		log.Print(err)
	}
	return ts
}
