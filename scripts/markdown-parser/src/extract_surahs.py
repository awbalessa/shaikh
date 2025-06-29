from src.models import surah_dict
from config import ASSETS_DIR, OUTPUT_DIR
from src.extract_surahs import extract_surah_range
import regex as re
import logging
from pathlib import Path
from typing import Tuple, List

logger = logging.getLogger(__name__)
page_number_match = re.compile(
    pattern=r"^#\sPage\s\d+$|^\p{Nd}+$"
)
header_match = re.compile(
    pattern=r"^#+\s"
)
bold_match = re.compile(
    pattern=r"^\*+$"
)
dash_match = re.compile(
    pattern=r"^---$"
)

def filter_presentational_lines(line: str) -> bool:
    if dash_match.search(line):
        return True
    if bold_match.search(line):
        return True
    if page_number_match.search(line):
        return True
    return False

def extract_surah_range(input_path: Path, range: Tuple[str, int, int], output_path: Path):
    surah, first_page, last_page = range
    extracted_lines: List[str] = []
    logger.info(msg=f"Extracting {surah} from ({first_page}, {last_page})...")
    with open(file=input_path, mode="r", encoding="utf-8") as f:
        parse = False
        for line in f:
            if parse == True:
                if line.strip() == f"# Page {last_page + 2}":
                   break
                if line.strip() == "":
                    continue
                if filter_presentational_lines(line.strip()):
                    continue
                extracted_lines.append(line.strip())
                continue
            else:
                if line.strip() != f"# Page {first_page}":
                    continue
                else:
                    parse = True
                    continue

    logger.info(msg=f"Writing to {output_path.name}...")
    with open(file=output_path, mode="w", encoding="utf-8") as f:
        for num, line in enumerate(extracted_lines):
            f.write(f"[{num+1}]: {line}")
            f.write("\n\n")

    logger.info(msg="Done!")


SURAH_DIR = OUTPUT_DIR / "surahs"
for num in surah_dict:
   extract_surah_range(
       input_path=ASSETS_DIR / "tafsir-ibn-kathir.md",
       range=surah_dict[num],
       output_path=SURAH_DIR / f"{num}.md"
   )
