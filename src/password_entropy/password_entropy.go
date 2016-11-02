package main

import (
	"bufio"
	"flag"
	"fmt"
	"io"
	"math"
	"os"
	"regexp"
)

type CharacterNGramModel struct {
	Size   int
	Counts map[string]float64
	Total  float64
}

func (model *CharacterNGramModel) Update(line string) {
	for _, key := range sliding([]rune(line), model.Size) {
		model.Counts[string(key)] += 1.0
		model.Total += 1.0
	}
}

func (model CharacterNGramModel) Count(key string) (count float64, found bool) {
	count, ok := model.Counts[key]
	if !ok {
		count = 1.0
	}
	return
}

type CharacterNGramModels map[int]*CharacterNGramModel

// LogProb returns the best matching log probability for a key given
// a set of models
func (models CharacterNGramModels) LogProb(key string) (logProb float64, found bool) {
	if len(models) == 0 || len(key) == 0 {
		logProb = math.Inf(-1)
		return
	}
	var lastTotal float64
	// else...
	// find the table that is the same size as the key
	// looking at increasing shorter suffixes
	// e.g. abc, bc, c
	runes := []rune(key)
	for i := 0; i < len(runes); i += 1 {
		key := string(runes[i:])
		model, foundModel := models[len(key)]
		if foundModel {
			count, f := model.Counts[key]
			if f {
				found = true
				logProb = math.Log2(count) - math.Log2(model.Total)
				// fmt.Printf("Found model with count;  count is %f; logProb is %f\n", count, logProb)
				return
			}
			lastTotal = model.Total
		}
	}
	// found it nowhere ... use last Total, and '1 count'
	found = false
	logProb = math.Log2(1.0) - math.Log2(lastTotal)
	// fmt.Printf("Didn't model with count; logProb is %f\n", logProb)
	return
}

// Dump dumps a set of ngram models to *one* file
func (models CharacterNGramModels) Dump(f io.Writer) {
	for sz, model := range models {
		for key, value := range model.Counts {
			outs := fmt.Sprintf("%v\t%s\t%f\n", sz, key, value)
			f.Write([]byte(outs))
		}
	}
}

func (models CharacterNGramModels) NgramSize() (ngram_size int) {
	for sz, _ := range models {
		if sz > ngram_size {
			ngram_size = sz
		}
	}
	return
}

func (models CharacterNGramModels) Predict(inf io.Reader, outf io.Writer) {
	// fmt.Println("Calling Predict")
	scanner := bufio.NewScanner(inf)
	for scanner.Scan() {
		text := processLine(scanner.Text())
		keys := sliding([]rune(text), models.NgramSize())
		if len(keys) == 0 {
			outf.Write([]byte(fmt.Sprintf("%f\t%s\n", 0.0, text))) // must be a better way
		} else {
			var log_prob_total float64
			for _, key := range keys {
				lp, _ := models.LogProb(key)
				//  fmt.Printf("LogProb for %s is %f\n", key, lp)
				log_prob_total += lp
			}
			log_prob_average := log_prob_total / float64(len(keys))
			outf.Write([]byte(fmt.Sprintf("%f\t%f\t%v\t%s\n", log_prob_average, log_prob_total, len(keys), text))) // TODO: must be a better way
		}
	}
}

// Read reads in a set of models (already initialized) from a file
func (models CharacterNGramModels) Read(f io.Reader) {
	// fmt.Println("Reading model")
	scanner := bufio.NewScanner(f)
	for scanner.Scan() {
		var ngram_size int
		var ngram string
		var count float64
		fmt.Sscanf(scanner.Text(), "%v\t%s\t%f", &ngram_size, &ngram, &count)
		// fmt.Printf("Read <%s> count: %v, size: %v\n", ngram, count, ngram_size)
		model, ok := models[ngram_size]
		if ok {
			model.Total += count
			model.Counts[ngram] = count
		} else {
			// fmt.Printf("Could not find model <%s> count: %v, size: %v\n", ngram, count, ngram_size)
		}
	}
}

// NewModels creates a set of ngram models up to size `ngram_size`
func NewModels(ngram_size int) (models CharacterNGramModels) {
	models = make(map[int]*CharacterNGramModel)
	for key := 1; key <= ngram_size; key++ {
		models[key] = new(CharacterNGramModel)
		models[key].Size = key
		models[key].Counts = make(map[string]float64)
	}
	return
}

func processLine(s string) string {
	compressRegexp := regexp.MustCompile(`\s+`) // should be param?
	text := compressRegexp.ReplaceAllString(s, " ")
	return text
}

// Train trains a set of ngram models from a file. Models must be
// initialized. returns the number of example lines used
func (models CharacterNGramModels) Train(f io.Reader) (exampleCount int) {
	// read the data
	if len(models) == 0 {
		return
	}
	sc := bufio.NewScanner(f)
	for sc.Scan() {
		text := processLine(sc.Text())
		exampleCount += 1
		for _, model := range models {
			model.Update(text)
		}
	}
	return
}

type line struct {
	s       string
	entropy float64
}

type byEntropyDesc []line

func (a byEntropyDesc) Len() int           { return len(a) }
func (a byEntropyDesc) Swap(i, j int)      { a[i], a[j] = a[j], a[i] }
func (a byEntropyDesc) Less(i, j int) bool { return a[i].entropy < a[j].entropy }

func log_prob(runes []rune, counts map[string]int, log_total float64) float64 {
	for i := 0; i < len(runes); i += 1 {
		key := string(runes[i:])
		count, ok := counts[key]
		if ok {
			lp := math.Log(float64(count)) - log_total
			return lp
		}
	}
	return math.Log(1.0) - log_total
}

func sliding(s []rune, length int) (windows []string) {
	for i := 0; i+length <= len(s); i += 1 {
		windows = append(windows, string(s[i:i+length]))
	}
	return
}

var ngram_size = flag.Int("ngram_size", 3, "Ngram size")
var train = flag.Bool("train", false, "Set if you want to train")
var predict = flag.Bool("predict", false, "Set if you want to predict")
var infile = flag.String("in", "", "Set for reading")
var outfile = flag.String("out", "", "Set for output")
var modelfile = flag.String("model", "", "Set for model file (dump or read)")

func main() {
	flag.Parse()
	ok := (*train || *predict)
	if !ok {
		fmt.Println("Either train or predict must be specified")
		return
	}

	models := NewModels(*ngram_size)

	inf := os.Stdin
	if *infile != "" {
		f2, err := os.Open(*infile)
		if err != nil {
			fmt.Printf("error reading %s\n", *infile)
			return
		}
		inf = f2
		defer inf.Close()
	}

	if *train {

		modf := os.Stdout
		if *modelfile != "" {
			f2, err := os.Create(*modelfile)
			if err != nil {
				fmt.Printf("error creating %s\n", *modelfile)
				return
			}
			modf = f2
			defer modf.Close()
		}
		models.Train(inf)
		models.Dump(modf)
		return
	}

	if *predict {
		modf := os.Stdin
		outf := os.Stdout
		if *modelfile != "" {
			f2, err := os.Open(*modelfile)
			if err != nil {
				fmt.Printf("error reading %s\n", *modelfile)
				return
			}
			modf = f2
		}
		if *outfile != "" {
			f2, err := os.Create(*outfile)
			if err != nil {
				fmt.Printf("error creating %s\n", *outfile)
				return
			}
			outf = f2
			defer outf.Close()
		}

		models.Read(modf)
		models.Predict(inf, outf)
		return
	}

	// Just setting this here so I can use this this source as an example..
	magic_password := "PXKXoyThngGrjCgBLuf2ivrpFFNKA9UgBHrxpLaW"
	if magic_password == "" {
		fmt.Println("this would be surprising to see")
	}
	standard_password := "escalate-latrine-footed-ping"
	if standard_password == "" {
		fmt.Println("this would be surprising to see")
	}
	stupid_password := "letmein"
	if stupid_password == "" {
		fmt.Println("this would be surprising to see")
	}

}
