package beater

import (
	"fmt"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/elastic/beats/libbeat/beat"
	"github.com/elastic/beats/libbeat/common"
	"github.com/elastic/beats/libbeat/logp"

	"github.com/FracKenA/op5beat/config"
	"github.com/vbatoufflet/go-livestatus"
)

// Op5beat configuration.
type Op5beat struct {
	done   chan struct{}
	config config.Config
	client beat.Client
}

// New creates an instance of op5beat.
func New(b *beat.Beat, cfg *common.Config) (beat.Beater, error) {
	c := config.DefaultConfig
	if err := cfg.Unpack(&c); err != nil {
		return nil, fmt.Errorf("Error reading config file: %v", err)
	}

	bt := &Op5beat{
		done:   make(chan struct{}),
		config: c,
	}
	return bt, nil
}

// Run starts op5beat.
func (bt *Op5beat) Run(b *beat.Beat) error {
	logp.Info("op5beat is running! Hit CTRL-C to stop it.")


	if len(bt.config.Op5host) < 1 {
		return fmt.Errorf("Error: Invalid op5host config \"%s\"", bt.config.Op5host)
	}
	if len(bt.config.Query) < 1 {
		return fmt.Errorf("Error: Invalid query config \"%s\"", bt.config.Query)
	}
	if len(bt.config.Columns) < 1 {
		return fmt.Errorf("Error: Invalid columns config \"%s\"", bt.config.Columns)
	}
	if bt.config.Metrics == true {
		var pd bool = false
		for _, v := range bt.config.Columns {
			if v == "perf_data" {
				pd = true
			}
		}
		if pd == false {
			return fmt.Errorf("Error: Metrics require searching for the perf_data column.")
		}
	}

	logp.Info("------Config-------")
	logp.Info("Host: %s", bt.config.Op5host)
	logp.Info("Query: %s", bt.config.Query)
	logp.Info("Columns: %s", bt.config.Columns)
	logp.Info("Filter: %s", bt.config.Filter)
	logp.Info("Metrics: %t", bt.config.Metrics)
	logp.Info("--------------")

	var err error
	bt.client, err = b.Publisher.Connect()
	if err != nil {
		return err
	}

	ticker := time.NewTicker(bt.config.Period)
	for {
		select {
		case <-bt.done:
			return nil
		case <-ticker.C:
		}

		err := bt.lsQuery(bt.config.Op5host)
		if err != nil {
			logp.Warn("Error executing query: %s", err)
			return err
		}
//		event := beat.Event{
//			Timestamp: time.Now(),
//			Fields: common.MapStr{
//				"type":    b.Info.Name,
//				"counter": counter,
//			},
//		}
//		bt.client.Publish(event)
//		logp.Info("Event sent")
//		counter++
	}
}

// Stop stops op5beat.
func (bt *Op5beat) Stop() {
	bt.client.Close()
	close(bt.done)
}

func (bt *Op5beat) lsQuery(lshost string) error {

	start := time.Now()
	count := 30
	then := start.Add(time.Duration(-count) * time.Second).Unix()
	timeFilter := fmt.Sprintf("last_check > %d", then)

	l := livestatus.NewClient(bt.config.Op5connect, bt.config.Op5host)
	q := livestatus.NewQuery(bt.config.Query)
	q.Columns(bt.config.Columns...)
	q.Filter(timeFilter)

	if len(bt.config.Filter) > 0 {
		for _, f := range bt.config.Filter {
			if strings.HasPrefix(f, "And") {
				and, _ := strconv.Atoi(strings.TrimPrefix(f, "And: "))
				q.And(and)
			} else if strings.HasPrefix(f, "Or") {
				or, _ := strconv.Atoi(strings.TrimPrefix(f, "Or: "))
				q.Or(or)
			} else {
				//				if len(f) > 1 {
				q.Filter(f)
				//				}
			}
		}
	}

	resp, err := l.Exec(q)
	if err != nil {
		return err
	}

	numRecords := 0

	for _, r := range resp.Records {

		event := beat.Event{
			Timestamp:    time.Now(),
                        Fields:       common.MapStr{},
		}

		var colData map[string]string
		colData = make(map[string]string)

		for _, c := range bt.config.Columns {
			var data interface{}
			data, err = GetCorrectDataType(r, c)
			if err != nil {
				logp.Warn("Problem parsing response field '%s': %s", c, err)
			}
			if c == "custom_variables" {
				event.Fields[c] = data
                        } else if strData, ok := data.(string); ok {
				colData[c] = strData
				event.Fields[c] = strData
			} else {
				strData := fmt.Sprint(data)
				colData[c] = strData
				event.Fields[c] = strData
			}
		}

		if bt.config.Metrics {
			var allow bool = true
			if len(bt.config.MetricsAllow) > 0 {
				allow = false
				for _, a := range bt.config.MetricsAllow {
					if a == colData["display_name"] {
						logp.Info("Allowing metric: %s", a)
						allow = true
					}
				}
			}
			if len(bt.config.MetricsBlock) > 0 {
				for _, a := range bt.config.MetricsBlock {
					if a == colData["display_name"] {
						logp.Info("Blocking metric: %s", a)
						allow = false
					}
				}
			}
			if allow {
				serviceMap := common.MapStr{}
				var perf_data string
				perf_data = colData["perf_data"]
				if len(perf_data) > 0 {
					var perfDataSplit []string

					perfDataSplit = strings.Split(perf_data, " ")
					for _, perfObj := range perfDataSplit {
						var perfObjSplit []string
						var dataSplit []string
						if len(perfObj) > 0 {
							perfObjSplit = strings.Split(perfObj, "=")
							if len(perfObjSplit) == 2 {
								item := perfObjSplit[0]
								data := perfObjSplit[1]
								if len(data) > 0 {
									var num string
									if strings.Contains(data, ";") {
										dataSplit = strings.Split(data, ";")
										dsLen := len(dataSplit)
										if dsLen >= 1 {
											if len(dataSplit[0]) > 0 {
												re := regexp.MustCompile("[0-9\\.]+")
												num = re.FindString(dataSplit[0])
											}
										}
									} else {
										re := regexp.MustCompile("[0-9\\.]+")
										num = re.FindString(data)
									}
									mItem := common.MapStr{
										item: num,
									}
									serviceMap = common.MapStrUnion(serviceMap, mItem)
								}
							}
						}
					}
					event.Fields["metrics"] = serviceMap
				}
			}
		}
		bt.client.Publish(event)
		numRecords++
	}
	elapsed := time.Since(start)
	logp.Info("%v events submitted in %s.", numRecords, elapsed)
	return nil
}
