package models

import (
	"encoding/json"
	"fmt"
)

// Chunk metadata
type ChunkMetadata struct {
	ParentIeD
}

// Word-level metadata — used for phrase-based embeddings
type WordLvlMetadata struct {
	SurahName   string `json:"surah_name"`
	SurahNumber int    `json:"surah_number"` // 1–114
	AyahNumber  int    `json:"ayah_number"`  // Local ayah number
	Phrase      string `json:"phrase"`       // e.g., "رحمة الله"
}

// Ayah-level metadata — used for ayah-based embeddings
type AyahLvlMetadata struct {
	SurahName   string `json:"surah_name"`
	SurahNumber int    `json:"surah_number"` // 1–114
	Ayah        string `json:"ayah"`
	AyahNumber  int    `json:"ayah_number"` // Local ayah number
}

// Surah-level metadata — used for surah summaries, themes
type SurahLvlMetadata struct {
	SurahName   string `json:"surah_name"`
	SurahNumber int    `json:"surah_number"` // 1–114
}

// Quran-level metadata — full corpus summaries or thematic embeddings
type QuranLvlMetadata struct {
	Note string `json:"note"` // e.g., "entire Quran summary vector"
}

type Metadata interface {
	Level() string
	Describe() string
}

func (w WordLvlMetadata) Level() string  { return "word" }
func (w AyahLvlMetadata) Level() string  { return "ayah" }
func (w SurahLvlMetadata) Level() string { return "surah" }
func (w QuranLvlMetadata) Level() string { return "quran" }

func (m WordLvlMetadata) Describe() string {
	ref := fmt.Sprintf("%d:%d", m.SurahNumber, m.AyahNumber)
	return fmt.Sprintf("%s %s | phrase: %s", m.SurahName, ref, m.Phrase)
}

func (m AyahLvlMetadata) Describe() string {
	ref := fmt.Sprintf("%d:%d", m.SurahNumber, m.AyahNumber)
	return fmt.Sprintf("%s %s", m.SurahName, ref)
}

func (m SurahLvlMetadata) Describe() string {
	return fmt.Sprintf("Surah: %s (%d)", m.SurahName, m.SurahNumber)
}

func (m QuranLvlMetadata) Describe() string {
	return fmt.Sprintf("Note: %s", m.Note)
}

func ExtractMetadata(data []byte) (Metadata, error) {
	// Attempt to unmarshal into WordLvlMetadata
	var wordMeta WordLvlMetadata
	if err := json.Unmarshal(data, &wordMeta); err == nil && wordMeta.Phrase != "" {
		return wordMeta, nil
	}

	// Attempt to unmarshal into AyahLvlMetadata
	var ayahMeta AyahLvlMetadata
	if err := json.Unmarshal(data, &ayahMeta); err == nil && ayahMeta.AyahNumber != 0 {
		return ayahMeta, nil
	}

	// Attempt to unmarshal into SurahLvlMetadata
	var surahMeta SurahLvlMetadata
	if err := json.Unmarshal(data, &surahMeta); err == nil && surahMeta.SurahName != "" {
		return surahMeta, nil
	}

	// Attempt to unmarshal into QuranLvlMetadata
	var quranMeta QuranLvlMetadata
	if err := json.Unmarshal(data, &quranMeta); err == nil && quranMeta.Note != "" {
		return quranMeta, nil
	}

	return nil, fmt.Errorf("Unrecognized metadata format")
}
