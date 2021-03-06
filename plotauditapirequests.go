package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"math"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strings"
	"time"

	"github.com/montanaflynn/stats"
	"github.com/spf13/cobra"
)

const graphTemplate = `set terminal png size 1200, 1200
set output 'apirequestscount-%v-distribution.png'
#set xdata time
#set timefmt "%%Y%%m%%d%%H%%M%%S"
set xtics rotate
set autoscale
set ylabel "api watch requests distribution (sampling size=%v)"
set title "WATCH requests %v"
set grid
set boxwidth 0.5
set style fill solid

`

const kaAuditGraphTemplate = `set terminal png size 1200, 1200
set output 'kaaudit-%v.png'
set xdata time
set timefmt "%%Y%%m%%d%%H%%M%%S"
set xtics rotate
set autoscale
set ylabel "api WATCH requests"
set title "watch requests %v"
set grid

`

const releaseGraphTemplate = `set terminal png size 2500, 1500
set output 'kaaudit-%v-%v.png'
set xdata time
set timefmt "%%Y%%m%%d%%H%%M%%S"
set xtics rotate
set autoscale
set ylabel "api WATCH requests"
set title "watch requests %v (%v)"
set key right bottom
set key outside
set rmargin 70
set grid

plot \
`

func releaseGraph(operator, suffix, title string, graphs []string) string {
	return fmt.Sprintf(
		releaseGraphTemplate,
		operator,
		suffix,
		operator,
		title,
	) + strings.Join(graphs, ", \\\n") + "\n"
}

const ChartJSCode = `var myChart = new Chart("%v", {
	type: "line",
	data: {
		datasets: %v,
	},
	options: {
		scales: {
			xAxes: [{
				type:'time',
				distribution:'series',
				time:{format:'YYYYMMDDHHmmss',unit:'day'},
				scaleLabel:{display:true,labelString:'Date'}
			}],
			yAxes: [{
				scaleLabel: {display: true, labelString: 'watch API requests'}
			}]
		},
		responsive:true,
		maintainAspectRatio:false,
		legend:{position:'right'},
		title:{display:true,text:'%v'},
	},
});`

const chartJSJobHTML = `<html>
<head>
  <title>4.10 report</title>
  <script src="https://cdnjs.cloudflare.com/ajax/libs/Chart.js/2.4.0/Chart.bundle.js"></script>
  <style type="text/css">
    .chart-container {
	    width: 1300px;
	    height: 800px;
			background-color: khaki;
			margin: 20px 0px;
		}
  </style>
</head>
<body>
	<center>
  %v
	</center>

  <script>

	%v

	</script>
</body>
</html>
`

func renderChartJSHtmlFile(datasets []dataSet, operator, htmlfile string) error {
	bytes, err := json.Marshal(datasets)
	if err != nil {
		return fmt.Errorf("unable to marshall datasets for %v: %v", operator, err)
	}

	f, err := os.Create(htmlfile)
	if err != nil {
		return fmt.Errorf("unable to create %v: %v", htmlfile, err)
	}
	defer f.Close()

	chartCode := fmt.Sprintf(ChartJSCode, operator, string(bytes), operator)
	canvas := fmt.Sprintf("<div class=\"chart-container\"><canvas id=\"%v\" width=\"200\" height=\"200\"></canvas></div>", operator)

	f.WriteString(fmt.Sprintf(chartJSJobHTML+"\n", canvas, chartCode))
	return nil
}

func renderMultipleChartJSHtmlFile(datasets map[string][]dataSet, htmlfile string) error {
	operators := []string{}
	for operator := range datasets {
		operators = append(operators, operator)
	}
	sort.Strings(operators)
	canvases := []string{}
	chartcodes := []string{}
	for _, operator := range operators {
		canvases = append(canvases, fmt.Sprintf("<div class=\"chart-container\"><canvas id=\"%v\" width=\"200\" height=\"200\"></canvas></div>", operator))
		bytes, err := json.Marshal(datasets[operator])
		if err != nil {
			return fmt.Errorf("unable to marshall datasets for %v: %v", operator, err)
		}

		chartcodes = append(chartcodes, fmt.Sprintf(ChartJSCode, operator, string(bytes), operator))
	}

	f, err := os.Create(htmlfile)
	if err != nil {
		return fmt.Errorf("unable to create %v: %v", htmlfile, err)
	}
	defer f.Close()

	f.WriteString(fmt.Sprintf(chartJSJobHTML+"\n", strings.Join(canvases, "\n"), strings.Join(chartcodes, "\n")))

	return nil
}

type filter func(string) bool

func listJobsForRelease(dir string, fnc filter) ([]string, error) {
	files, err := ioutil.ReadDir(dir)
	if err != nil {
		return nil, err
	}

	jobs := []string{}
	for _, f := range files {
		if strings.HasPrefix(f.Name(), "periodic") || strings.HasPrefix(f.Name(), "release") {
			jobs = append(jobs, filepath.Join(dir, f.Name()))
		}
	}
	return jobs, nil
}

func listKnownOperatorsForRelease(dir string, fnc filter) ([]string, error) {
	jobs, err := listJobsForRelease(dir, fnc)
	if err != nil {
		return nil, err
	}
	operators := []string{}
	operatorsMaps := map[string]struct{}{}
	for _, job := range jobs {
		if !fnc(job) {
			continue
		}
		files, err := ioutil.ReadDir(job)
		if err != nil {
			return nil, err
		}

		for _, f := range files {
			if strings.HasPrefix(f.Name(), "kaaudit-") && strings.HasSuffix(f.Name(), ".g") {
				operator := strings.TrimRight(strings.TrimLeft(f.Name(), "kaaudit-"), ".g")
				operatorsMaps[operator] = struct{}{}
			}
		}

	}
	for operator := range operatorsMaps {
		operators = append(operators, operator)
	}
	return operators, nil
}

func listJobIdDirsFromJobDir(dir string) ([]string, error) {
	files, err := ioutil.ReadDir(dir)
	if err != nil {
		return nil, err
	}

	jobids := []string{}
	for _, f := range files {
		match, err := regexp.MatchString("^([0-9]+)$", f.Name())
		if err != nil {
			return nil, err
		}
		if !match {
			continue
		}
		jobids = append(jobids, filepath.Join(dir, f.Name()))
	}
	return jobids, nil
}

func listFilesForJobIDs(dir, targetFile string) ([]string, error) {
	dirs, err := listJobIdDirsFromJobDir(dir)
	if err != nil {
		return nil, fmt.Errorf("unable to list jobids in %v: %v", dir, err)
	}

	files := []string{}
	for _, jobiddir := range dirs {
		if _, err := os.Stat(filepath.Join(jobiddir, targetFile)); errors.Is(err, os.ErrNotExist) {
			continue
		}
		files = append(files, filepath.Join(jobiddir, targetFile))
	}
	return files, nil
}

type Request struct {
	NodeName string `json:"nodeName"`
	Operator string `json:"operator"`
	Count    int    `json:"count"`
	Hour     int    `json:"hour"`
}

type Requests struct {
	CreationTimestamp  int       `json:"creationTimestamp"`
	WatchRequestCounts []Request `json:"watchRequestCounts"`
}

type ResponseStatus struct {
	Metadata interface{} `json:"metadata"`
	Status   string      `json:"status"`
	Message  string      `json:"message"`
	Code     int         `json:"code"`
}

type AuditLog struct {
	AuditID                  string         `json:"auditID"`
	RequestURI               string         `json:"requestURI"`
	Username                 string         `json:"username"`
	StageTimestamp           string         `json:"stageTimestamp"`
	RequestReceivedTimestamp string         `json:"requestReceivedTimestamp"`
	ResponseStatus           ResponseStatus `json:"responseStatus"`
}

type OpenshiftE2ETest struct {
	Total     int    `json:"total"`
	Timestamp int    `json:"timestamp"`
	Passed    bool   `json:"passed"`
	Result    string `json:"result"`
	Reivsion  string `json:"revision"`
}

func data2datFile(data map[int]int, targetFile string) error {
	keys := make([]int, 0, len(data))
	for k := range data {
		keys = append(keys, k)
	}
	sort.Ints(keys)

	f, err := os.Create(targetFile)
	if err != nil {
		return fmt.Errorf("unable to create %v: %v", targetFile, err)
	}
	defer f.Close()

	// fmt.Printf("file: %v\n", targetFile)
	for _, k := range keys {
		// fmt.Printf("%v: %v\n", k, data[k])
		f.WriteString(fmt.Sprintf("%v %v\n", k, data[k]))
	}
	return nil
}

func timeData2datFile(data map[time.Time]int, targetFile string) error {
	tsKeys := []time.Time{}
	for key := range data {
		tsKeys = append(tsKeys, key)
	}
	sort.Slice(tsKeys, func(i, j int) bool {
		return tsKeys[i].Before(tsKeys[j])
	})

	f, err := os.Create(targetFile)
	if err != nil {
		return fmt.Errorf("unable to create %v: %v", targetFile, err)
	}
	defer f.Close()

	fmt.Printf("file: %v\n", targetFile)
	for _, k := range tsKeys {
		// layout := "2006-01-02T15:04:05.000000Z"
		// fmt.Printf("%d%02d%02d%02d%02d%02d %v\n", k.Year(), k.Month(), k.Day(), k.Hour(), k.Minute(), k.Second(), data[k])
		f.WriteString(fmt.Sprintf("%d%02d%02d%02d%02d%02d %v\n", k.Year(), k.Month(), k.Day(), k.Hour(), k.Minute(), k.Second(), data[k]))
		// break
	}
	return nil
}

type dataItem struct {
	X string `json:"x"`
	Y int    `json:"y"`
}

func timeData2JSONFile(data map[time.Time]int, targetFile string) error {
	tsKeys := []time.Time{}
	for key := range data {
		tsKeys = append(tsKeys, key)
	}
	sort.Slice(tsKeys, func(i, j int) bool {
		return tsKeys[i].Before(tsKeys[j])
	})

	f, err := os.Create(targetFile)
	if err != nil {
		return fmt.Errorf("unable to create %v: %v", targetFile, err)
	}
	defer f.Close()

	fmt.Printf("file: %v\n", targetFile)
	dataItems := []dataItem{}
	for _, k := range tsKeys {
		dataItems = append(dataItems, dataItem{
			X: fmt.Sprintf("%d%02d%02d%02d%02d%02d", k.Year(), k.Month(), k.Day(), k.Hour(), k.Minute(), k.Second()),
			Y: data[k],
		})
	}

	bytes, err := json.Marshal(dataItems)
	if err != nil {
		return fmt.Errorf("unable to marshall data: %v", err)
	}

	f.WriteString(string(bytes))
	return nil
}

type dataSet struct {
	Label       string     `json:"label"`
	Fill        bool       `json:"fill"`
	Data        []dataItem `json:"data"`
	BorderColor string     `json:"borderColor"`
}

var colors = []string{
	"rgb(255,140,0)",
	"rgb(154,205,50)",
	"rgb(0,191,255)",
	"rgb(0,0,128)",
	"rgb(186,85,211)",
	"rgb(139,69,19)",
	"rgb(0,100,0)",
}

func timeData2ChartJSDataset(data map[time.Time]int, title string) dataSet {
	tsKeys := []time.Time{}
	for key := range data {
		tsKeys = append(tsKeys, key)
	}
	sort.Slice(tsKeys, func(i, j int) bool {
		return tsKeys[i].Before(tsKeys[j])
	})

	items := []dataItem{}
	for _, k := range tsKeys {
		items = append(items, dataItem{
			X: fmt.Sprintf("%d%02d%02d%02d%02d%02d", k.Year(), k.Month(), k.Day(), k.Hour(), k.Minute(), k.Second()),
			Y: data[k],
		})
	}

	return dataSet{
		Label:       title,
		Data:        items,
		BorderColor: "rgb(255, 99, 132)",
	}
}

func plotApiRequestsCount(dir string) error {
	files, err := listFilesForJobIDs(dir, "requests.json")
	if err != nil {
		return err
	}
	fileslen := len(files)

	maxapirequestcounts := map[string]map[int]int{}
	frequency := map[string]map[int]int{}
	// sampling size s=10
	sampling := 5

	var errs []error
	for idx, file := range files {
		fmt.Printf("Processing %v (%v/%v)\n", file, idx+1, fileslen)
		// check the file is a JSON file, otherwise ignore
		if err := func() error {
			jsonFile, err := os.Open(file)
			if err != nil {
				return fmt.Errorf("unable to open %v: %v", file, err)
			}
			defer jsonFile.Close()
			byteValue, err := ioutil.ReadAll(jsonFile)
			if err != nil {
				return fmt.Errorf("unable to read %v: %v", file, err)
			}

			var result Requests
			if err := json.Unmarshal([]byte(byteValue), &result); err != nil {
				return fmt.Errorf("unable to unmarshal %v: %v", file, err)
			}

			// Check if the requets are empty
			if result.CreationTimestamp == 0 {
				return fmt.Errorf("empty creationTimestamp %v, skipping", file)
			}

			for _, item := range result.WatchRequestCounts {
				parts := strings.Split(item.Operator, ":")
				operator := parts[len(parts)-1]
				if _, exists := maxapirequestcounts[operator]; !exists {
					maxapirequestcounts[operator] = map[int]int{}
					frequency[operator] = map[int]int{}
				}
				maxapirequestcounts[operator][result.CreationTimestamp] = item.Count
				bucket := int(math.Ceil(float64(item.Count/sampling))) * sampling
				if _, exists := frequency[operator][bucket]; !exists {
					frequency[operator][bucket] = 0
				}
				frequency[operator][bucket]++
			}
			return nil
		}(); err != nil {
			fmt.Printf("err: %v\n", err)
			errs = append(errs, err)
			continue
		}
		// if idx > 5 {
		// 	break
		// }
	}

	for operator := range maxapirequestcounts {
		data2datFile(maxapirequestcounts[operator], filepath.Join(dir, fmt.Sprintf("kaapirequestcounts-%v-max.dat", operator)))
		data2datFile(frequency[operator], filepath.Join(dir, fmt.Sprintf("kaapirequestcounts-%v-max-distribution.dat", operator)))
	}

	for operator := range frequency {
		if err := func() error {
			graphs := []string{}
			graphs = append(graphs, fmt.Sprintf("\"%v\" using 1:2 title \"apirequestscount distribution size=%v\" with boxes", filepath.Join(dir, fmt.Sprintf("kaapirequestcounts-%v-max-distribution.dat", operator)), sampling))
			graphfile := filepath.Join(dir, fmt.Sprintf("apirequestscount-%v-distribution.g", operator))
			f, err := os.Create(graphfile)
			if err != nil {
				return fmt.Errorf("unable to create %v: %v", graphfile, err)

			}
			defer f.Close()
			f.WriteString(fmt.Sprintf(graphTemplate+"plot "+strings.Join(graphs, ", \\\n")+"\n", operator, sampling, operator))
			// fmt.Printf("graphfile: %v\n", graphfile)
			return nil
		}(); err != nil {
			errs = append(errs, err)
		}
	}

	if len(errs) > 0 {
		var errStr []string
		for _, err := range errs {
			errStr = append(errStr, err.Error())
		}
		return fmt.Errorf("%v", strings.Join(errStr, ";"))
	}

	return nil
}

func plotOpenshiftE2ETests(dir string) error {
	files, err := listFilesForJobIDs(dir, "openshift-e2e-tests.json")
	if err != nil {
		return err
	}
	fileslen := len(files)
	testsTotal := map[int]int{}

	var errs []error
	for idx, file := range files {
		fmt.Printf("Processing %v (%v/%v)\n", file, idx+1, fileslen)
		// check the file is a JSON file, otherwise ignore
		if err := func() error {
			jsonFile, err := os.Open(file)
			if err != nil {
				return fmt.Errorf("unable to open %v: %v", file, err)
			}
			defer jsonFile.Close()
			byteValue, err := ioutil.ReadAll(jsonFile)
			if err != nil {
				return fmt.Errorf("unable to read %v: %v", file, err)
			}

			var result OpenshiftE2ETest
			if err := json.Unmarshal([]byte(byteValue), &result); err != nil {
				return fmt.Errorf("unable to unmarshal %v: %v", file, err)
			}

			// TODO(jchaloup): extract the timestamp into openshift-e2e-tests.json as well
			// meantime, use requests.json
			requestsFile := filepath.Join(filepath.Dir(file), "requests.json")
			if _, err := os.Stat(requestsFile); errors.Is(err, os.ErrNotExist) {
				return nil
			}
			jsonFile, err = os.Open(requestsFile)
			if err != nil {
				return fmt.Errorf("unable to open %v: %v", file, err)
			}
			defer jsonFile.Close()
			byteValue, err = ioutil.ReadAll(jsonFile)
			if err != nil {
				return fmt.Errorf("unable to read %v: %v", file, err)
			}

			var resultRequests Requests
			if err := json.Unmarshal([]byte(byteValue), &resultRequests); err != nil {
				return fmt.Errorf("unable to unmarshal %v: %v", file, err)
			}

			// Check if the requets are empty
			if resultRequests.CreationTimestamp == 0 {
				return fmt.Errorf("empty creationTimestamp %v, skipping", file)
			}

			fmt.Printf("%v: %v\n", resultRequests.CreationTimestamp, result.Total)
			testsTotal[resultRequests.CreationTimestamp] = result.Total

			return nil
		}(); err != nil {
			fmt.Printf("err: %v\n", err)
			errs = append(errs, err)
			continue
		}
		// if idx > 5 {
		// 	break
		// }
	}

	data2datFile(testsTotal, filepath.Join(dir, "teststotal.dat"))

	return nil
}

func operatorFromUsername(username string) string {
	parts := strings.Split(username, ":")
	return parts[len(parts)-1]
}

func plotKAAuditRequests(dir string) error {
	files, err := listFilesForJobIDs(dir, "ka-audit-logs.json")
	if err != nil {
		return err
	}
	fileslen := len(files)

	layout := "2006-01-02T15:04:05.000000Z"

	bucketsMax60MinuteSequence := make(map[string]map[time.Time]int)
	var errs []error
	for idx, file := range files {
		fmt.Printf("Processing %v (%v/%v)\n", file, idx+1, fileslen)
		// check the file is a JSON file, otherwise ignore
		if err := func() error {
			jsonFile, err := os.Open(file)
			if err != nil {
				return fmt.Errorf("unable to open %v: %v", file, err)
			}
			defer jsonFile.Close()
			byteValue, err := ioutil.ReadAll(jsonFile)
			if err != nil {
				return fmt.Errorf("unable to read %v: %v", file, err)
			}

			var result []AuditLog
			if err := json.Unmarshal([]byte(byteValue), &result); err != nil {
				return fmt.Errorf("unable to unmarshal %v: %v", file, err)
			}
			if len(result) == 0 {
				return nil
			}

			t, err := time.Parse(layout, result[0].RequestReceivedTimestamp)
			if err != nil {
				return err
			}

			usernames := make(map[string]struct{})
			bucketsMinute := make(map[string]map[time.Time]int)
			minAuditTS := t
			var errs []error
			for _, item := range result {
				if item.ResponseStatus.Code != 200 {
					continue
				}
				username := operatorFromUsername(item.Username)
				usernames[username] = struct{}{}
				timeObj, err := time.Parse(layout, item.RequestReceivedTimestamp)
				if err != nil {
					errs = append(errs, err)
					continue
				}

				if _, exists := bucketsMinute[username]; !exists {
					bucketsMinute[username] = map[time.Time]int{}
				}

				timeMinutes := timeObj.Truncate(time.Minute)
				bucketsMinute[username][timeMinutes]++

				if minAuditTS.After(timeObj) {
					minAuditTS = timeObj
				}
			}

			for username := range bucketsMinute {
				if _, exists := bucketsMax60MinuteSequence[username]; !exists {
					bucketsMax60MinuteSequence[username] = map[time.Time]int{}
				}
				bucketsMax60MinuteSequence[username][minAuditTS] = 0
				// sort the times
				tsKeys := []time.Time{}
				for key := range bucketsMinute[username] {
					tsKeys = append(tsKeys, key)
				}
				tsKeysLen := len(tsKeys)
				sort.Slice(tsKeys, func(i, j int) bool {
					return tsKeys[i].Before(tsKeys[j])
				})

				for i := 0; i < tsKeysLen; i++ {
					j := i
					sum := 0
					for ; j < tsKeysLen && tsKeys[j].Sub(tsKeys[i]).Seconds() < (60*time.Minute).Seconds(); j++ {
						sum += bucketsMinute[username][tsKeys[j]]
					}
					if sum > bucketsMax60MinuteSequence[username][minAuditTS] {
						bucketsMax60MinuteSequence[username][minAuditTS] = sum
					}
					// if the current sequence has less than 60 minutes long interval, stop
					if tsKeys[tsKeysLen-1].Sub(tsKeys[i]).Seconds() < (60 * time.Minute).Seconds() {
						break
					}
				}
			}

			return nil
		}(); err != nil {
			fmt.Printf("err: %v\n", err)
			errs = append(errs, err)
			continue
		}
	}

	percentiles := []float64{50, 60, 70, 80, 90, 95, 99}
	datasets := make(map[string][]dataSet)
	for operator, data := range bucketsMax60MinuteSequence {
		timeData2datFile(data, filepath.Join(dir, fmt.Sprintf("kaaudit-%v-max-60minute-sequence.dat", operator)))

		// sort the times
		tsKeys := []time.Time{}
		for key := range bucketsMax60MinuteSequence[operator] {
			tsKeys = append(tsKeys, key)
		}
		tsKeysLen := len(tsKeys)
		sort.Slice(tsKeys, func(i, j int) bool {
			return tsKeys[i].Before(tsKeys[j])
		})

		datasets[operator] = []dataSet{}
		colorIdx := 0
		colorLen := len(colors)
		for _, percentile := range percentiles {
			percentileGrowing := map[time.Time]int{}
			samples := []float64{}
			for i := 0; i < tsKeysLen; i++ {
				samples = append(samples, float64(bucketsMax60MinuteSequence[operator][tsKeys[i]]))
				p, err := stats.Percentile(samples, percentile)
				if err != nil {
					errs = append(errs, fmt.Errorf("unable to compute %v-th percentile: %v", percentile, err))
					continue
				}
				percentileGrowing[tsKeys[i]] = int(math.Ceil(p))
			}
			timeData2datFile(percentileGrowing, filepath.Join(dir, fmt.Sprintf("kaaudit-%v-max-60minute-sequence-%v-percentile-growing.dat", operator, percentile)))
			timeData2JSONFile(percentileGrowing, filepath.Join(dir, fmt.Sprintf("kaaudit-%v-max-60minute-sequence-%v-percentile-growing.json", operator, percentile)))
			// build dataset for chartjs
			dataset := timeData2ChartJSDataset(percentileGrowing, fmt.Sprintf("%v-th percentile", percentile))
			dataset.BorderColor = colors[colorIdx]
			datasets[operator] = append(datasets[operator], dataset)
			colorIdx = (colorIdx + 1) % colorLen
		}
		err := renderChartJSHtmlFile(datasets[operator], operator, filepath.Join(dir, fmt.Sprintf("kaaudit-%v.html", operator)))
		if err != nil {
			errs = append(errs, fmt.Errorf("unable to render chart JS html page for %v: %v", operator, err))
			continue
		}
	}

	renderMultipleChartJSHtmlFile(datasets, filepath.Join(dir, "kaaudit-all.html"))

	for operator := range bucketsMax60MinuteSequence {
		if err := func() error {
			graphs := []string{}
			for _, percentile := range percentiles {
				graphs = append(graphs, fmt.Sprintf("\"%v\" using 1:2 title \"60 minute sequence max %v-p growing\" with linespoints", filepath.Join(dir, fmt.Sprintf("kaaudit-%v-max-60minute-sequence-%v-percentile-growing.dat", operator, percentile)), percentile))
			}
			graphs = append(graphs, fmt.Sprintf("\"%v\" using 1:2 title \"apirequestcounts CR hours max\" with linespoints", filepath.Join(dir, fmt.Sprintf("kaapirequestcounts-%v-max.dat", operator))))
			teststotalFile := filepath.Join(dir, "teststotal.dat")
			if _, err := os.Stat(teststotalFile); !errors.Is(err, os.ErrNotExist) {
				graphs = append(graphs, fmt.Sprintf("\"%v\" using 1:2 title \"tests total\" with linespoints", teststotalFile))
			}

			graphfile := filepath.Join(dir, fmt.Sprintf("kaaudit-%v.g", operator))
			f, err := os.Create(graphfile)
			if err != nil {
				return fmt.Errorf("unable to create %v: %v", graphfile, err)
			}
			defer f.Close()
			f.WriteString(fmt.Sprintf(kaAuditGraphTemplate+"plot "+strings.Join(graphs, ", \\\n")+"\n", operator, operator))
			// fmt.Printf("graphfile: %v\n", graphfile)
			return nil
		}(); err != nil {
			errs = append(errs, err)
		}
	}

	if len(errs) > 0 {
		var errStr []string
		for _, err := range errs {
			errStr = append(errStr, err.Error())
		}
		return fmt.Errorf("%v", strings.Join(errStr, ";"))
	}

	return nil
}

type Command struct {
	// Path to the directory with extracted job artifacts
	dataDir string
	// Process only requests.json files
	onlyAPIRequestsCount bool
	// Process only ka-audit-logs.json files
	onlyKAAudits bool
	// Process only openshift-e2e-tests.json files
	onlyOpenshiftTests bool
	// Plot percentiles through all jobs in a given release
	aggregateJobsInRelease bool
}

func myFilter(jobname string) bool {
	return strings.Contains(jobname, "aws") && strings.Contains(jobname, "upgrade")
	// return true
}

func (c *Command) Run() error {
	if c.aggregateJobsInRelease {
		operators, err := listKnownOperatorsForRelease(c.dataDir, myFilter)
		if err != nil {
			return err
		}

		variants := []struct {
			filter func(string) bool
			suffix string
			title  string
		}{
			// {
			// 	filter: func(jobname string) bool {
			// 		return strings.Contains(jobname, "aws") && strings.Contains(jobname, "upgrade")
			// 	},
			// 	suffix: "4.10-aws-upgrade",
			// 	title:  "60 minute sequence max 50-th percentile growing aws upgrade",
			// },
			// {
			// 	filter: func(jobname string) bool {
			// 		return strings.Contains(jobname, "aws") && !strings.Contains(jobname, "upgrade")
			// 	},
			// 	suffix: "4.10-aws",
			// 	title:  "60 minute sequence max 50-th percentile growing aws",
			// },
			{
				filter: func(jobname string) bool {
					return strings.Contains(jobname, "aws")
				},
				suffix: "4.10-aws-all",
				title:  "60 minute sequence max 50-th percentile growing aws all",
			},
			{
				filter: func(jobname string) bool {
					return strings.Contains(jobname, "azure")
				},
				suffix: "4.10-azure-all",
				title:  "60 minute sequence max 50-th percentile growing azure all",
			},
			{
				filter: func(jobname string) bool {
					return strings.Contains(jobname, "gcp")
				},
				suffix: "4.10-gcp-all",
				title:  "60 minute sequence max 50-th percentile growing gcp all",
			},
			{
				filter: func(jobname string) bool {
					return strings.Contains(jobname, "vsphere")
				},
				suffix: "4.10-vsphere-all",
				title:  "60 minute sequence max 50-th percentile growing vsphere all",
			},
			{
				filter: func(jobname string) bool {
					return strings.Contains(jobname, "openstack")
				},
				suffix: "4.10-openstack-all",
				title:  "60 minute sequence max 50-th percentile growing openstack all",
			},
			{
				filter: func(jobname string) bool {
					return strings.Contains(jobname, "metal")
				},
				suffix: "4.10-metal-all",
				title:  "60 minute sequence max 50-th percentile growing metal all",
			},
			{
				filter: func(jobname string) bool {
					return strings.Contains(jobname, "upgrade")
				},
				suffix: "4.10-upgrade-all",
				title:  "60 minute sequence max 50-th percentile growing upgrade all",
			},
		}

		for _, variant := range variants {
			fmt.Printf("Processing %q variant\n", variant.title)
			for _, operator := range operators {
				graphs := []string{}
				jobs, err := listJobsForRelease(c.dataDir, variant.filter)
				if err != nil {
					return err
				}
				for _, job := range jobs {
					if !variant.filter(job) {
						continue
					}
					dataFilePath := filepath.Join(job, fmt.Sprintf("kaaudit-%v-max-60minute-sequence-50-percentile-growing.dat", operator))
					if _, err := os.Stat(dataFilePath); errors.Is(err, os.ErrNotExist) {
						continue
					}
					graphs = append(graphs, fmt.Sprintf("\"%v\" using 1:2 title \"%v\" with linespoints", dataFilePath, filepath.Base(job)))
				}

				if len(graphs) == 0 {
					continue
				}
				graphfile := filepath.Join(c.dataDir, fmt.Sprintf("kaaudit-%v-max-60minute-sequence-50-percentile-growing-%v.g", operator, variant.suffix))
				f, err := os.Create(graphfile)
				if err != nil {
					return fmt.Errorf("unable to create %v: %v", graphfile, err)
				}
				defer f.Close()
				f.WriteString(releaseGraph(
					operator,
					variant.suffix,
					variant.title,
					graphs,
				))
			}
		}
		return nil
	}
	if c.onlyAPIRequestsCount {
		plotApiRequestsCount(c.dataDir)
	} else if c.onlyKAAudits {
		plotKAAuditRequests(c.dataDir)
	} else if c.onlyOpenshiftTests {
		plotOpenshiftE2ETests(c.dataDir)
	} else {
		plotApiRequestsCount(c.dataDir)
		plotKAAuditRequests(c.dataDir)
	}
	return nil
}

var command = &Command{}

func main() {
	cmd := &cobra.Command{
		Use:   "plotauditapirequests",
		Short: "KA audit logs plotter",
		Run: func(cmd *cobra.Command, args []string) {
			if err := command.Run(); err != nil {
				fmt.Fprintf(os.Stderr, "%v\n", err)
				os.Exit(1)
			}
		},
	}

	flags := cmd.Flags()
	flags.StringVar(&command.dataDir, "datadir", command.dataDir, "Path to job with extracted artifacts")
	flags.BoolVar(&command.onlyAPIRequestsCount, "only-apirequestscount", command.onlyAPIRequestsCount, "Process only requests.json files")
	flags.BoolVar(&command.onlyKAAudits, "only-kaauditlogs", command.onlyKAAudits, "Process only ka-audit-logs.json files")
	flags.BoolVar(&command.onlyOpenshiftTests, "only-openshift-tests", command.onlyOpenshiftTests, "Process only openshift-e2e-tests.json files")
	flags.BoolVar(&command.aggregateJobsInRelease, "aggregate-jobs-in-release", command.aggregateJobsInRelease, "Plot percentiles through all jobs in a given release")

	if err := cmd.Execute(); err != nil {
		fmt.Fprintf(os.Stderr, "%v\n", err)
		os.Exit(1)
	}
}
