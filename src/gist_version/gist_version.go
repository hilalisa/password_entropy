package main

import (
	"bufio"
	"fmt"
	"math"
	"os"
	"regexp"
	"sort"
)

type line struct {
	s       string
	entropy float64
}

func main() {

	var lines []line

	compressRegexp := regexp.MustCompile(`\s+`)

	sc := bufio.NewScanner(os.Stdin)
	for sc.Scan() {
		text := compressRegexp.ReplaceAllString(sc.Text(), " ")
		lines = append(lines, line{
			s:       text,
			entropy: shannonEntropy(text),
		})
	}

	sort.Sort(byEntropyDesc(lines))

	for _, l := range lines {
		fmt.Println(l.entropy, l.s)
	}

}

type byEntropyDesc []line

func (a byEntropyDesc) Len() int           { return len(a) }
func (a byEntropyDesc) Swap(i, j int)      { a[i], a[j] = a[j], a[i] }
func (a byEntropyDesc) Less(i, j int) bool { return a[i].entropy > a[j].entropy }

func shannonEntropy(s string) float64 {
	runes := map[rune]int{}
	for _, c := range s {
		runes[c]++
	}
	var result float64
	var log2 = math.Log(2)
	for _, ct := range runes {
		frequency := float64(ct) / float64(len(runes))
		result -= frequency * (math.Log(frequency) / log2)
	}
	return result
}
