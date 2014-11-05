package kstest

import (
	log "github.com/cihub/seelog"
	app "github.com/nvcook42/morgoth/app/types"
	"github.com/nvcook42/morgoth/engine"
	metric "github.com/nvcook42/morgoth/metric/types"
	"github.com/nvcook42/morgoth/schedule"
	"math"
	"encoding/json"
	"sort"
	"time"
)

type fingerprint struct {
	Data  []float64
	Count uint
}

type KSTest struct {
	rotation     schedule.Rotation
	reader       engine.Reader
	writer       engine.Writer
	config       *KSTestConf
	fingerprints map[metric.MetricID][]fingerprint
}

func (self *KSTest) Initialize(app app.App, rotation schedule.Rotation) error {
	self.rotation = rotation
	self.reader = app.GetReader()
	self.writer = app.GetWriter()
	self.load()
	return nil
}

func (self *KSTest) Detect(metric metric.MetricID, start, stop time.Time) bool {
	fingerprints := self.fingerprints[metric]
	log.Debugf("KSTest.Detect Rotation: %s FP: %v", self.rotation.GetPrefix(), fingerprints)
	points := self.reader.GetData(&self.rotation, metric, start, stop)
	data := make([]float64, len(points))
	for i, point := range points {
		data[i] = point.Value
	}
	sort.Float64s(data)
	log.Debugf("Testing %v", data)

	minError := 0.0
	bestMatch := -1
	isMatch := false
	for i, fingerprint := range fingerprints {
		thresholdD := self.getThresholdD(len(fingerprint.Data), len(data))

		D := calcTestD(fingerprint.Data, data)
		log.Debug("D: ", D)
		if D < thresholdD {
			isMatch = true
		}
		e := (D - thresholdD) / thresholdD
		if bestMatch == -1 || e < minError {
			minError = e
			bestMatch = i
		}
	}

	anomalous := false
	if isMatch {
		anomalous = fingerprints[bestMatch].Count < self.config.NormalCount
		fingerprints[bestMatch].Count++
	} else {
		anomalous = true
		//We know its anomalous, now we need to update our fingerprints

		if len(fingerprints) == int(self.config.MaxFingerprints) {
			log.Debug("Reached MaxFingerprints")
			//TODO: Update bestMatch to learn new fingerprint
		} else {
			fingerprints = append(fingerprints, fingerprint{
				Data:  data,
				Count: 1,
			})
		}
	}

	self.fingerprints[metric] = fingerprints
	//go self.save()
	return anomalous
}

func (self *KSTest) getThresholdD(n, m int) float64 {
	c := 0.0
	switch self.config.Confidence {
	case 0: // 0.10
		c = 1.22
	case 1: // 0.05
		c = 1.36
	case 2: // 0.025
		c = 1.48
	case 3: // 0.01
		c = 1.63
	case 4: // 0.005
		c = 1.73
	case 5: // 0.001
		c = 1.95
	}
	return c * math.Sqrt(float64(n+m)/float64(n*m))
}

func calcTestD(f1, f2 []float64) float64 {
	D := 0.0
	n := float64(len(f1))
	m := float64(len(f2))
	cdf1 := 0.0
	cdf2 := 0.0
	j := 0
	for _, x1 := range f1 {
		cdf1 += 1 / n
		for j < int(m) && x1 >= f2[j] {
			j++
			cdf2 += 1 / m
		}
		if d := math.Abs(cdf1 - cdf2); d > D {
			D = d
		}
		if j == int(m) { //Optimization only
			break
		}
	}
	return D
}

func (self *KSTest) save(metric metric.MetricID) {

	data, err := json.Marshal(self.fingerprints[metric])
	if err != nil {
		log.Error("Could not save KSTest", err.Error())
	}
	self.writer.StoreDoc(self.rotation.GetPrefix() + "kstest." + string(metric), data)
}

func (self *KSTest) load() {

	data := self.reader.GetDoc(self.rotation.GetPrefix() + "kstest")
	if len(data) != 0 {
		err := json.Unmarshal(data, &self.fingerprints)
		if err != nil {
			log.Error("Could not load KSTest ", err.Error())
		}
	}
	if self.fingerprints == nil {
		self.fingerprints = make(map[metric.MetricID][]fingerprint)
	}
}