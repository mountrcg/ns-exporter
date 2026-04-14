package main

import "time"

type NsEntry struct {
	OpenAps struct {
		Enacted struct {
			Temp             string    `json:"temp" bson:"temp"`
			Bg               float64   `json:"bg" bson:"bg"`
			Tick             float64   `json:"tick" bson:"tick"`
			EventualBG       float64   `json:"eventualBG" bson:"eventualBG"`
			TargetBG         float64   `json:"targetBG" bson:"target_bg"`
			InsulinReq       float64   `json:"insulinReq" bson:"insulinReq"`
			DeliverAt        time.Time `json:"deliverAt" bson:"deliverAt"`
			SensitivityRatio float64   `json:"sensitivityRatio" bson:"sensitivityRatio"`
			Tdd              float64   `json:"TDD" bson:"TDD"`
			DuraISFratio     float64   `json:"dura_ISFratio" bson:"dura_ISFratio"`
			BgISFratio       float64   `json:"bg_ISFratio" bson:"bg_ISFratio"`
			DeltaISFratio    float64   `json:"delta_ISFratio" bson:"delta_ISFratio"`
			PpISFratio       float64   `json:"pp_ISFratio" bson:"pp_ISFratio"`
			AcceISFratio     float64   `json:"acce_ISFratio" bson:"acce_ISFratio"`
			AutoISFratio     float64   `json:"auto_ISFratio" bson:"auto_ISFratio"`
			PredBGs          struct {
				IOB []float64 `json:"IOB"`
				ZT  []float64 `json:"ZT"`
				COB []float64 `json:"COB"`
				UAM []float64 `json:"UAM"`
			} `json:"predBGs"`
			COB       float64   `json:"COB"`
			IOB       float64   `json:"IOB"`
			Reason    string    `json:"reason"`
			Units     float64   `json:"units"`
			Rate      float64   `json:"rate"`
			Duration  int       `json:"duration"`
			Timestamp time.Time `json:"timestamp"`
		} `json:"enacted,omitempty" bson:"enacted,omitempty"`
		IOB struct {
			IOB      float64   `json:"iob" bson:"iob"`
			BasalIOB float64   `json:"basaliob" bson:"basaliob"`
			Activity float64   `json:"activity" bson:"activity"`
			Time     time.Time `json:"time" bson:"time"`
		} `json:"iob" bson:"iob"`
	} `json:"openaps" bson:"openaps"`
	Pump struct {
		Clock     time.Time `json:"clock"`
		Reservoir float64   `json:"reservoir"`
		Status    struct {
			Status    string `json:"status"`
			Timestamp int64  `json:"-" bson:"-"`
		} `json:"status"`
		Extended struct {
			Version               string  `json:"Version"`
			ActiveProfile         string  `json:"ActiveProfile"`
			TempBasalAbsoluteRate float64 `json:"TempBasalAbsoluteRate"`
			TempBasalPercent      int     `json:"TempBasalPercent"`
			TempBasalRemaining    int     `json:"TempBasalRemaining"`
		} `json:"extended"`
		Battery struct {
			Percent int `json:"percent"`
		} `json:"battery"`
	} `json:"pump"`
}

type NsTreatment struct {
	CreatedAt time.Time `json:"created_at"`
	EnteredBy string    `json:"enteredBy"`
	EventType string    `json:"eventType"`
	Carbs     int       `json:"carbs,omitempty"`
	Duration  int       `json:"duration,omitempty"`
	Insulin   float64   `json:"insulin,omitempty"`

	Notes        string  `json:"notes,omitempty"`
	Percent      int     `json:"percent,omitempty"`
	TargetTop    float64 `json:"targetTop,omitempty"`
	TargetBottom float64 `json:"targetBottom,omitempty"`
	Reason       string  `json:"reason,omitempty"`
	Rate         float64 `json:"rate,omitempty"`
	Units        string  `json:"units,omitempty"`
	Bolus        struct {
		IsSMB     bool      `json:"isSMB,omitempty"`
		Amount    float64   `json:"amount,omitempty"`
		Time      time.Time `json:"timestamp,omitempty"`
		BolusType string    `json:"_type,omitempty"`
	}
}
