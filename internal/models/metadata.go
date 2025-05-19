package models

// Word-level metadata — used for phrase-based embeddings
type WordLvlMetadata struct {
	SurahNumber int    `json:"surah"`  // 1–114
	AyahNumber  int    `json:"ayah"`   // Local ayah number
	Phrase      string `json:"phrase"` // e.g., "رحمة الله"
}

// Ayah-level metadata — used for entire ayah embeddings
type AyahLvlMetadata struct {
	SurahNumber int `json:"surah"`
	AyahNumber  int `json:"ayah"`
}

// Surah-level metadata — used for surah summaries, themes
type SurahLvlMetadata struct {
	SurahNumber int    `json:"surah"`
	SurahName   string `json:"surah_name"` // optional, useful for display
}

// Quran-level metadata — full corpus summaries or thematic embeddings
type QuranLvlMetadata struct {
	Note string `json:"note"` // e.g., "entire Quran summary vector"
}
