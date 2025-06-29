from src.extract_ayat import tafsir_starts
from config import OUTPUT_DIR

SURAH_DIR = OUTPUT_DIR / "surahs"
matches = tafsir_starts(
    file=SURAH_DIR / "2.md"
)

for m in matches:
    print(m)
    print(f"\n")
