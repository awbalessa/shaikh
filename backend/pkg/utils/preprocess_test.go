package utils

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestCleanAndFilterStopwords(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "empty string",
			input:    "",
			expected: "",
		},
		{
			name:     "only stopwords",
			input:    "غير وسوى فغير",
			expected: "",
		},
		{
			name:     "stopwords and normal words with diacritics",
			input:    "غَيْرَ نُورُ وسُوَى الحَقِّ",
			expected: "نور الحق",
		},
		{
			name:     "with tatweel and symbols",
			input:    "غــير،؟ الحــــق * وسوى%",
			expected: "الحق",
		},
		{
			name:     "preserves digits and Arabic letters",
			input:    "١٢٣ غير محمد ٤٥٦",
			expected: "١٢٣ محمد ٤٥٦",
		},
		{
			name:     "repeated stopwords and valid words",
			input:    "غير غير الحق غير",
			expected: "الحق",
		},
		{
			name:     "extra spaces and diacritics",
			input:    "   فَاطِمَةُ   تُحِبُّ  التُّفَّاحَ   ",
			expected: "فاطمة تحب التفاح",
		},
		{
			name:     "lam-alef ligature removed properly",
			input:    "ﻻ وسوى ﻷ ﻹ جميل",
			expected: "جميل",
		},
		{
			name:     "non-Arabic letters are stripped",
			input:    "hello وسوى world محمد",
			expected: "محمد",
		},
		{
			name:     "non-diacritic stopwords in between",
			input:    "النور غير الحق",
			expected: "النور الحق",
		},
		{
			name:     "tatweel only",
			input:    "ــ ـــ ـــ",
			expected: "",
		},
		{
			name:     "ligature variants mixed in",
			input:    "ﻷ غير اللَّهُ ﻻ جميل",
			expected: "الله جميل",
		},
		{
			name:     "diacritics only",
			input:    "َ ً ُ ٌ ِ ٍ ْ ٰ",
			expected: "",
		},
		{
			name:     "arabic numbers with letters and stopwords",
			input:    "١٢٣ وسوى بيت ٤٥٦",
			expected: "١٢٣ بيت ٤٥٦",
		},
		{
			name:     "mix of ligatures and tatweel",
			input:    "ﻻــﻷــﻹ جميل",
			expected: "جميل",
		},
		{
			name:     "only non-arabic symbols",
			input:    "!@#$%^&*()_+-=<>?/\\|~`[]{}",
			expected: "",
		},
		{
			name:     "arabic with english and symbols",
			input:    "غير test وسوى جميل!",
			expected: "جميل",
		},
		{
			name:     "mixed whitespace characters",
			input:    "غير\tوسوى\nمحمد\r\n\tالحق",
			expected: "محمد الحق",
		},
		{
			name:     "stopword merged into word",
			input:    "وسوىالحق غيرالنور",
			expected: "وسوىالحق غيرالنور", // Should remain, not tokenized as stopword
		},
		{
			name:     "mixed punctuation and arabic",
			input:    "،؟! النور؛: غير.",
			expected: "النور",
		},
		{
			name:     "words with tatweel padding",
			input:    "مــحــمــد وســوى",
			expected: "محمد",
		},
		{
			name:     "empty after cleaning",
			input:    "123 *&^%$#@!",
			expected: "123", // Digits preserved
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			result, err := CleanAndFilterStopwords(tc.input)
			require.NoError(t, err)
			assert.Equal(t, tc.expected, result)
		})
	}
}
