import logging
from typing import List
from src.models import surah_dict
from src.extract_ayat import extract_intro, normalize_numbers, tafsir_starts, test_verse_marker_sequence
from src.extract_ayat import extract_ayah_range
from src.extract_surahs import extract_surah_range
from config import ASSETS_DIR, OUTPUT_DIR

logger = logging.getLogger(__name__)

SURAH_DIR = OUTPUT_DIR / "surahs"
IBN_KATHIR = ASSETS_DIR / "tafsir-ibn-kathir.md"
AYAT_DIR = OUTPUT_DIR / "ayat"

for num in surah_dict:
    matches = tafsir_starts(
        file=SURAH_DIR / f"{num}.md"
    )
    normal_list: List[List[int]] = []
    for m in matches:
        normal = normalize_numbers(m.group("ayahs"))
        normal_list.append(normal)
        print(normal)
        if normal == []:
            print(m.group("ayahs"))
            print(m.string)
            exit()

    try:
        test_verse_marker_sequence(normal_list)
    except Exception as e:
        print(f"Surah: {num}")
        print(e)
        exit()

    extract_intro(
        surah_file=SURAH_DIR / f"{num}.md",
        output_dir=AYAT_DIR / f"{num}"
    )
    extract_ayah_range(
        surah_file=SURAH_DIR / f"{num}.md",
        output_dir=AYAT_DIR / f"{num}"
    )


# extract_surah_range(
#     input_path=IBN_KATHIR,
#     range=surah_dict[114],
#     output_path=SURAH_DIR / "114.md"
# )
