package kellog

import (
	"github.com/Matir/adifparser"
	"github.com/golang/protobuf/ptypes"
	"github.com/golang/protobuf/ptypes/timestamp"
	"github.com/xylo04/kellog-qrz-sync/generated/adifpb"
	"io"
	"log"
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
	distS, _ := record.GetValue("distance")
	dist64, _ := strconv.ParseUint(distS, 10, 32)
	qso.DistanceKm = uint32(dist64)
	freq, _ := record.GetValue("freq")
	qso.Freq, _ = strconv.ParseFloat(freq, 64)
	freqRx, _ := record.GetValue("freq_rx")
	qso.FreqRx, _ = strconv.ParseFloat(freqRx, 64)
	qso.Mode, _ = record.GetValue("mode")
	qso.Notes, _ = record.GetValue("notes")
	qso.PublicKey, _ = record.GetValue("public_key")
	qso.Complete, _ = record.GetValue("qso_complete")
	qso.TimeOn = getTimestamp(record, "qso_date", "time_on")
	qso.TimeOff = getTimestamp(record, "qso_date_off", "time_off")
	random, _ := record.GetValue("random")
	qso.Random = random == "Y"
	qso.RstReceived, _ = record.GetValue("rst_rcvd")
	qso.RstSent, _ = record.GetValue("rst_sent")
	qso.Submode, _ = record.GetValue("submode")
	swl, _ := record.GetValue("swl")
	qso.Swl = swl == "Y"

	qso.ContactedStation = new(adifpb.Station)
	qso.ContactedStation.OpCall, _ = record.GetValue("call")
	return qso
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
