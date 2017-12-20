// Copyright 2014 Google Inc. All Rights Reserved.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//      http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package requester

import (
	"fmt"
	"io"
	"sort"
	"strings"
	"time"
	"os"
)

const (
	barChar = "∎"
)

type report struct {
	avgTotal float64
	fastest  float64
	slowest  float64
	average  float64
	rps      float64
	size 	int
	conn 	int
	avgConn   float64
	avgDNS    float64
	avgReq    float64
	avgRes    float64
	avgDelay  float64
	connLats  []float64
	dnsLats   []float64
	reqLats   []float64
	resLats   []float64
	delayLats []float64

	results chan *result
	total   time.Duration

	errorDist      map[string]int
	statusCodeDist map[int]int
	lats           []float64
	sizeTotal      int64

	output string

	w io.Writer
}

func newReport(w io.Writer, size int, conn int, results chan *result, output string, total time.Duration) *report {
	return &report{
		output:         output,
		results:        results,
		total:          total,
		size: 			size,
		conn: 			conn,
		statusCodeDist: make(map[int]int),
		errorDist:      make(map[string]int),
		w:              w,
	}
}

func (r *report) finalize(f *os.File) {
	for res := range r.results {
		if res.err != nil {
			r.errorDist[res.err.Error()]++
		} else {
			r.lats = append(r.lats, res.duration.Seconds())
			r.avgTotal += res.duration.Seconds()
			r.avgConn += res.connDuration.Seconds()
			r.avgDelay += res.delayDuration.Seconds()
			r.avgDNS += res.dnsDuration.Seconds()
			r.avgReq += res.reqDuration.Seconds()
			r.avgRes += res.resDuration.Seconds()
			r.connLats = append(r.connLats, res.connDuration.Seconds())
			r.dnsLats = append(r.dnsLats, res.dnsDuration.Seconds())
			r.reqLats = append(r.reqLats, res.reqDuration.Seconds())
			r.delayLats = append(r.delayLats, res.delayDuration.Seconds())
			r.resLats = append(r.resLats, res.resDuration.Seconds())
			r.statusCodeDist[res.statusCode]++
			if res.contentLength > 0 {
				r.sizeTotal += res.contentLength
			}
		}
	}
	r.rps = float64(len(r.lats)) / r.total.Seconds()
	r.average = r.avgTotal / float64(len(r.lats))
	r.avgConn = r.avgConn / float64(len(r.lats))
	r.avgDelay = r.avgDelay / float64(len(r.lats))
	r.avgDNS = r.avgDNS / float64(len(r.lats))
	r.avgReq = r.avgReq / float64(len(r.lats))
	r.avgRes = r.avgRes / float64(len(r.lats))
	r.print(f)
}

func (r *report) printCSV() {
	r.printf("response-time,DNS+dialup,DNS,Request-write,Response-delay,Response-read\n")
	for i, val := range r.lats {
		r.printf("%4.4f,%4.4f,%4.4f,%4.4f,%4.4f,%4.4f\n",
			val, r.connLats[i], r.dnsLats[i], r.reqLats[i], r.delayLats[i], r.resLats[i])
	}
}

func (r *report) printoutputtocsv(f *os.File){
	//f1, err := os.OpenFile("Output.csv", os.O_APPEND|os.O_WRONLY, 0600)
	// if err != nil {
	//     panic(err)
	// }
	
	// if _, err = f1.WriteString(fmt.Sprintf("%d, %d, %.4f\n", r.size, r.conn, r.rps)); err != nil {
	//     panic(err)
	// }
	_,_ = f.WriteString(fmt.Sprintf("%d, %d, %.4f\n", r.size, r.conn, r.rps))
	// if err != nil {
	//      panic(err)
	// }
	fmt.Printf("conc: %d done\n", r.conn)
	//f1.Close()
}

func (r *report) print(f *os.File){
	if r.output == "csv" {
		r.printCSV()
		return
	}

	if len(r.lats) > 0 {
		r.printoutputtocsv(f)
	}
	if len(r.errorDist) > 0 {
		r.printErrors()
	}
	r.printf("\n")
}

// printSection prints details for http-trace fields
func (r *report) printSection(tag string, avg float64, lats []float64) {
	sort.Float64s(lats)
	fastest, slowest := lats[0], lats[len(lats)-1]
	r.printf("\n  %s:\t", tag)
	r.printf(" %4.4f secs, %4.4f secs, %4.4f secs", avg, fastest, slowest)
}

// printLatencies prints percentile latencies.
func (r *report) printLatencies() {
	pctls := []int{10, 25, 50, 75, 90, 95, 99}
	data := make([]float64, len(pctls))
	j := 0
	for i := 0; i < len(r.lats) && j < len(pctls); i++ {
		current := i * 100 / len(r.lats)
		if current >= pctls[j] {
			data[j] = r.lats[i]
			j++
		}
	}
	r.printf("\nLatency distribution:\n")
	for i := 0; i < len(pctls); i++ {
		if data[i] > 0 {
			r.printf("  %v%% in %4.4f secs\n", pctls[i], data[i])
		}
	}
}

func (r *report) printHistogram() {
	bc := 10
	buckets := make([]float64, bc+1)
	counts := make([]int, bc+1)
	bs := (r.slowest - r.fastest) / float64(bc)
	for i := 0; i < bc; i++ {
		buckets[i] = r.fastest + bs*float64(i)
	}
	buckets[bc] = r.slowest
	var bi int
	var max int
	for i := 0; i < len(r.lats); {
		if r.lats[i] <= buckets[bi] {
			i++
			counts[bi]++
			if max < counts[bi] {
				max = counts[bi]
			}
		} else if bi < len(buckets)-1 {
			bi++
		}
	}
	r.printf("\nResponse time histogram:\n")
	for i := 0; i < len(buckets); i++ {
		// Normalize bar lengths.
		var barLen int
		if max > 0 {
			barLen = (counts[i]*40 + max/2) / max
		}
		r.printf("  %4.3f [%v]\t|%v\n", buckets[i], counts[i], strings.Repeat(barChar, barLen))
	}
}

// printStatusCodes prints status code distribution.
func (r *report) printStatusCodes() {
	r.printf("\n\nStatus code distribution:\n")
	for code, num := range r.statusCodeDist {
		r.printf("  [%d]\t%d responses\n", code, num)
	}
}

func (r *report) printErrors() {
	r.printf("\nError distribution:\n")
	for err, num := range r.errorDist {
		r.printf("  [%d]\t%s\n", num, err)
	}
}

func (r *report) printf(s string, v ...interface{}) {
	fmt.Fprintf(r.w, s, v...)
}
