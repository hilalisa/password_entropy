# Password Entropy

Really, a character N-gram entropy modeller

It has been noted that (good) passwords have high entropy,
and we should be able to use that fact to find (good) passwords in code (where they shoudn't be).

To train:

-  Get some (source) code to train on, and train on it.

The following trains on the 1.7.3 Go distribution code, after removing some crypto files, as well as test files.

The resulting model can be found in the `data` directory.

```bash
find /usr/local/Cellar/go/1.7.3/libexec/src/ | grep "\.go" | grep -v "crypto" | grep -v "_test" | xargs cat > /tmp/go_text
bin/password_entropy -train -in /tmp/go_text -model data/go-3.tsv -ngram_size 3
```

To predict:

- Use the model to predict on some source code, for example,
the source for this program, which has some high-entropy
passwords in it.

```bash
cat src/password_entropy/password_entropy.go|  bin/password_entropy -predict -model data/go-3.tsv  | sort -g | head
-16.285189	-960.826143	59	 magic_password := "PXKXoyThngGrjCgBLuf2ivrpFFNKA9UgBHrxpLaW"
-15.055465	-286.053838	19	 log_prob_total += lp
-15.024271	-405.655320	27	 stupid_password := "letmein"
-14.990013	-1019.320886	68	 fmt.Sscanf(scanner.Text(), "%v\t%s\t%f", &ngram_size, &ngram, &count)
-14.989606	-44.968818	3	 "os"
-14.885426	-401.906513	27	 lp, _ := models.LogProb(key)
-14.643432	-673.597876	46	 logProb = math.Log2(1.0) - math.Log2(lastTotal)
-14.577021	-335.271482	23	type byEntropyDesc []line
-14.487198	-101.410389	7	 if !ok {
-14.487198	-101.410389	7	 if !ok {
```

Columns are, for each line: average log probability (take negative for entropy), total
log probability, number of ngrams, and the line.

The `Sccanf` line reminds me that format strings always look line line noise, and now we have the science to prove it!
