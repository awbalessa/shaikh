package utils

import (
	_ "embed"
	"encoding/json"
	"fmt"
	"strings"
	"sync"
	"unicode"
)

type Stopwords map[string]struct{}

//go:embed stopwords.json
var stopwordsJSON []byte

var (
	stopwords     Stopwords
	stopwordsOnce sync.Once
	stopwordsErr  error
)

func CleanAndFilterStopwords(s string) (string, error) {
	filtered, err := filterStopwords(
		cleanArabicString(s),
	)
	if err != nil {
		return "", fmt.Errorf("failed to preprocess string: %w", err)
	}

	return filtered, nil
}

func loadStopwords() (Stopwords, error) {
	var raw map[string]json.RawMessage
	if err := json.Unmarshal(stopwordsJSON, &raw); err != nil {
		return nil, fmt.Errorf("failed to unmarshal into stopwords map: %w", err)
	}

	stop := make(Stopwords, len(raw))
	for k := range raw {
		stop[k] = struct{}{}
	}

	return stop, nil
}

func getStopwords() (Stopwords, error) {
	stopwordsOnce.Do(func() {
		stopwords, stopwordsErr = loadStopwords()
	})
	return stopwords, stopwordsErr
}

func filterStopwords(str string) (string, error) {
	stopwords, err := getStopwords()
	if err != nil {
		return "", fmt.Errorf("failed to get stopwords: %w", err)
	}

	words := strings.Fields(str)
	filtered := make([]string, 0, len(words))
	for _, word := range words {
		if _, ok := stopwords[word]; ok {
			continue
		}
		filtered = append(filtered, word)
	}

	joined := strings.Join(filtered, " ")
	return joined, nil
}

func dediac(s string) string {
	var b strings.Builder
	for _, r := range s {
		if !isDiac(r) {
			b.WriteRune(r)
		}
	}
	return b.String()
}

func isDiac(r rune) bool {
	switch r {
	case '\u064B', // FATHATAN
		'\u064C', // DAMMATAN
		'\u064D', // KASRATAN
		'\u064E', // FATHA
		'\u064F', // DAMMA
		'\u0650', // KASRA
		'\u0651', // SHADDA
		'\u0652', // SUKUN
		'\u0670', // MINI ALEF
		'\u0653', // MADDA ABOVE
		'\u0654', // HAMZA ABOVE
		'\u0655', // HAMZA BELOW
		'\u0640': // TATWEEL ← Add this
		return true
	}
	return false
}

func removeNonArabicAlnum(s string) string {
	var b strings.Builder
	for _, r := range s {
		if isArabicLetter(r) || unicode.IsDigit(r) || unicode.IsSpace(r) {
			b.WriteRune(r)
		}
	}
	return b.String()
}

func isArabicLetter(r rune) bool {
	return (r >= 0x0621 && r <= 0x064A) || // Arabic letters
		(r == 0x0671) // ALEF_WASLA
}

func cleanArabicString(s string) string {
	noTashkeel := dediac(s)
	clean := removeNonArabicAlnum(noTashkeel)
	return clean
}
