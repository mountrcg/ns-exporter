package main

import (
	"context"
	"flag"
	"fmt"
	"html"
	"log"
	"os"
	"regexp"
	"strconv"
	"sync"

	influxdb2 "github.com/influxdata/influxdb-client-go/v2"
	"github.com/influxdata/influxdb-client-go/v2/api/write"
	"github.com/peterbourgon/ff/v3"
)

var wg sync.WaitGroup
var wgInflux sync.WaitGroup

func main() {
	fs := flag.NewFlagSet("ns-exporter", flag.ContinueOnError)
	var (
		mongoUri     = fs.String("mongo-uri", "", "Mongo-db uri to download from")
		mongoDb      = fs.String("mongo-db", "", "Mongo-db database name")
		nsUri        = fs.String("ns-uri", "", "Nightscout server url to download from")
		nsToken      = fs.String("ns-token", "", "Nigthscout server API Authorization Token")
		limit        = fs.Int64("limit", 0, "number of records to read")
		skip         = fs.Int64("skip", 0, "number of records to skip")
		influxUri    = fs.String("influx-uri", "", "InfluxDb uri to download from")
		influxToken  = fs.String("influx-token", "", "InfluxDb access token")
		influxOrg    = fs.String("influx-org", "ns", "InfluxDb organization to use")
		influxBucket = fs.String("influx-bucket", "ns", "InfluxDb bucket to use")
	)
	if err := ff.Parse(fs, os.Args[1:], ff.WithEnvVarPrefix("NS_EXPORTER")); err != nil {
		fmt.Fprintf(os.Stderr, "error: %v\n", err)
		os.Exit(1)
	}

	ctx := context.Background()

	deviceStatuses := make(chan NsEntry)
	treatments := make(chan NsTreatment)
	influx := make(chan write.Point)

	if *mongoUri != "" && *mongoDb != "" {
		mongo := NewMongoClient(*mongoUri, *mongoDb, ctx)
		defer mongo.Close(ctx)
		processClient(mongo, deviceStatuses, treatments, *limit, *skip, ctx)
	}

	if *nsUri != "" && *nsToken != "" {
		ns := NewNSClient(*nsUri, *nsToken)
		processClient(ns, deviceStatuses, treatments, *limit, *skip, ctx)
	}

	var wgTransform = &sync.WaitGroup{}
	wgTransform.Add(2)

	go parseDeviceStatuses(wgTransform, influx, deviceStatuses)
	go parseTreatments(wgTransform, influx, treatments)

	go func() {
		wgInflux.Add(1)
		defer wgInflux.Done()
		var count = 0
		writeAPI := influxdb2.NewClient(*influxUri, *influxToken).WriteAPIBlocking(*influxOrg, *influxBucket)

		for point := range influx {

			if len(point.FieldList()) == 0 && len(point.TagList()) == 0 {

				fmt.Println("empty point for time: ", point.Time(), " of type: ", point.Name())
				continue
			}

			err := writeAPI.WritePoint(ctx, &point)
			count++
			if err != nil {

				fmt.Println("error writing: ", point.Time(), ", name: ", point.Name())
				log.Fatal(err)
			}
		}

		fmt.Println("total writen: ", count)

	}()

	wg.Wait()
	close(deviceStatuses)
	close(treatments)
	wgTransform.Wait()
	close(influx)
	wgInflux.Wait()
}

func processClient(client IExporter, deviceStatuses chan NsEntry, treatments chan NsTreatment, limit int64, skip int64, ctx context.Context) {
	wg.Add(2)
	go client.LoadDeviceStatuses(deviceStatuses, limit, skip, ctx)
	go client.LoadTreatments(treatments, limit, skip, ctx)
}

func parseDeviceStatuses(group *sync.WaitGroup, influx chan write.Point, entries chan NsEntry) {
	defer group.Done()

	reg := regexp.MustCompile("Dev: (?P<dev>[-0-9.]+),.*ISF: (?:(?P<isf_nt>[-0-9.]+)/(?P<isf_bg>[-0-9.]+)+=)?(?P<isf>[-0-9.]+),.*CR: (?P<cr>[-0-9.]+)")

	var count = 0
	var lastbg = 0.0
	var lasttick float64 = 0

	for entry := range entries {

		point := influxdb2.NewPointWithMeasurement("openaps").
			AddField("iob", entry.OpenAps.IOB.IOB).
			AddField("basal_iob", entry.OpenAps.IOB.BasalIOB).
			AddField("activity", entry.OpenAps.IOB.Activity).
			SetTime(entry.OpenAps.IOB.Time)

		if entry.OpenAps.Enacted.Bg > 0 {

			var tick = entry.OpenAps.Enacted.Tick
			if lastbg == entry.OpenAps.Enacted.Bg &&
				lasttick == tick &&
				tick != 0.0 {
				// deduplication, because nightscout still allows duplicate records to be added
				fmt.Println("skipping duplicate bg record: ", entry.OpenAps.IOB.Time, ", bg: ", entry.OpenAps.Enacted.Bg, ", tick: ", tick)
				continue
			}

			lastbg = entry.OpenAps.Enacted.Bg
			lasttick = tick
			point.
				AddField("bg", entry.OpenAps.Enacted.Bg).
				AddField("tick", tick).
				AddField("eventual_bg", entry.OpenAps.Enacted.EventualBG).
				AddField("target_bg", entry.OpenAps.Enacted.TargetBG).
				AddField("insulin_req", entry.OpenAps.Enacted.InsulinReq).
				AddField("cob", entry.OpenAps.Enacted.COB).
				AddField("bolus", entry.OpenAps.Enacted.Units).
				AddField("tbs_rate", entry.OpenAps.Enacted.Rate).
				AddField("tbs_duration", entry.OpenAps.Enacted.Duration).
				AddField("sens", entry.OpenAps.Enacted.SensitivityRatio).
				AddField("tdd", entry.OpenAps.Enacted.Tdd).
				AddField("duraISF_ratio", entry.OpenAps.Enacted.DuraISFratio).
				AddField("bgISF_ratio", entry.OpenAps.Enacted.BgISFratio).
				AddField("deltaISF_ratio", entry.OpenAps.Enacted.DeltaISFratio).
				AddField("ppISF_ratio", entry.OpenAps.Enacted.PpISFratio).
				AddField("acceISF_ratio", entry.OpenAps.Enacted.AcceISFratio).
				AddField("autoISF_ratio", entry.OpenAps.Enacted.AutoISFratio)

			// if len(entry.OpenAps.Enacted.PredBGs.COB) > 0 {
			// 	point.AddField("pred_cob", entry.OpenAps.Enacted.PredBGs.COB[len(entry.OpenAps.Enacted.PredBGs.COB)-1])
			// }
			// if len(entry.OpenAps.Enacted.PredBGs.IOB) > 0 {
			// 	point.AddField("pred_iob", entry.OpenAps.Enacted.PredBGs.IOB[len(entry.OpenAps.Enacted.PredBGs.IOB)-1])
			// }
			// if len(entry.OpenAps.Enacted.PredBGs.UAM) > 0 {
			// 	point.AddField("pred_uam", entry.OpenAps.Enacted.PredBGs.UAM[len(entry.OpenAps.Enacted.PredBGs.UAM)-1])
			// }
			// if len(entry.OpenAps.Enacted.PredBGs.ZT) > 0 {
			// 	point.AddField("pred_zt", entry.OpenAps.Enacted.PredBGs.ZT[len(entry.OpenAps.Enacted.PredBGs.ZT)-1])
			// }

			// take 90min forecast not 4hrs
			if len(entry.OpenAps.Enacted.PredBGs.COB) > 0 {
				point.AddField("pred_cob", entry.OpenAps.Enacted.PredBGs.COB[18-1])
			}
			if len(entry.OpenAps.Enacted.PredBGs.IOB) > 0 {
				point.AddField("pred_iob", entry.OpenAps.Enacted.PredBGs.IOB[18-1])
			}
			if len(entry.OpenAps.Enacted.PredBGs.UAM) > 0 {
				point.AddField("pred_uam", entry.OpenAps.Enacted.PredBGs.UAM[18-1])
			}
			if len(entry.OpenAps.Enacted.PredBGs.ZT) > 0 {
				point.AddField("pred_zt", entry.OpenAps.Enacted.PredBGs.ZT[18-1])
			}

			if len(entry.OpenAps.Enacted.Reason) > 0 {
				matches := reg.FindStringSubmatch(entry.OpenAps.Enacted.Reason)
				names := reg.SubexpNames()
				for i, match := range matches {
					if i != 0 {
						if len(match) > 0 {
							if rvalue, err := strconv.ParseFloat(match, 32); err == nil {
								point.AddField(names[i], rvalue)
							}
						}
					}
				}

				point.AddField("reason", html.UnescapeString(entry.OpenAps.Enacted.Reason))
			}
		}

		count++
		influx <- *point

		fmt.Println("time: ", entry.OpenAps.IOB.Time, "iob:", entry.OpenAps.IOB.IOB, ", bg: ", entry.OpenAps.Enacted.Bg)
	}
	fmt.Println("total devicestatuses parsed: ", count)
}

func parseTreatments(group *sync.WaitGroup, influx chan write.Point, entries chan NsTreatment) {
	defer group.Done()

	var noted = map[string]bool{
		"Site Change":         true,
		"Insulin Change":      true,
		"Pump Battery Change": true,
		"Sensor Change":       true,
		"Sensor Start":        true,
		"Sensor Stop":         true,
		"BG Check":            true,
		"Exercise":            true,
		"Announcement":        true,
		"Question":            true,
		//"Note": true,
		"OpenAPS Offline": true,
		"D.A.D. Alert":    true,
		"Mbg":             true,
		//"Carb Correction": true,
		//"Bolus Wizard": true,
		//"Correction Bolus": true,
		//"Meal Bolus": true,
		//"Combo Bolus": true,
		//"Temporary Target": true,
		//"Temporary Target Cancel": true,
		"Profile Switch": true,
		//"Snack Bolus": true,
		//"Temp Basal": true,
		//"Temp Basal Start": true,
		//"Temp Basal End": true,
	}

	var count = 0
	for entry := range entries {

		point := influxdb2.NewPointWithMeasurement("treatments").
			SetTime(entry.CreatedAt)

		tagName := "type"
		if entry.Carbs > 0 {
			point.
				AddField("carbs", entry.Carbs).
				AddTag(tagName, "carbs")
		}
		if entry.Insulin > 0 {
			point.
				AddField("bolus", entry.Insulin).
				AddTag(tagName, "bolus").
				AddTag("smb", strconv.FormatBool(entry.IsSMB))
		}
		if entry.EventType == "Temp Basal" {
			point.
				AddField("duration", entry.Duration).
				AddField("percent", entry.Percent).
				AddField("rate", entry.Rate).
				AddTag(tagName, "tbs")
		} else if entry.EventType == "Temporary Target" {
			point.
				AddField("duration", entry.Duration).
				AddField("target_top", entry.TargetTop).
				AddField("target_bottom", entry.TargetBottom).
				AddField("units", entry.Units).
				AddField("reason", entry.Reason).
				AddTag(tagName, "tt")
		} else if len(entry.Notes) > 0 {
			point.AddField("notes", entry.Notes)
		} else if noted[entry.EventType] {
			point.AddField("notes", entry.EventType)
		}

		count++
		influx <- *point
		fmt.Println("time: ", point.Time(), ", type: ", entry.EventType)
	}

	fmt.Println("total treatments parsed: ", count)
}
