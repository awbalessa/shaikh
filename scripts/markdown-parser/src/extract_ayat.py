import regex as re
from regex import Match
import logging
from typing import List, Dict
from pathlib import Path

# from google import genai
# from google.genai.types import Part, UserContent, GenerateContentConfig
# from config import GCP_PROJECT, GCP_REGION, GEMINI_LITE, SYSTEM_INSTRUCTION

logger = logging.getLogger(__name__)

verse_marker_match = re.compile(
    pattern=r"""
    ^\[\d+\]:\s[\*\#\s]*سورة\s.*الآ(?:ية|يتان|يات)
    \s*:?\s*
    [\uFD3F]?\s*
    (?P<ayahs>[\p{Nd}\s\-–ـ،,]+)
    [\uFD3E]?\s*
    """,
    flags=re.VERBOSE
)

def test_verse_marker_sequence(list: List[List[int]]):
    prev: List[int] = []
    for l in list:
        if l == list[-1]:
            return
        if prev == []:
            prev = l
            continue
        if l == prev:
            continue
        if max(prev) > max(l):
            raise Exception(f"List not ordered properly:\ncurrent: {l}\nprevious: {prev}")
        if min(l) not in prev:
            if min(l) != max(prev) + 1:
                raise Exception(f"List not ordered properly:\ncurrent: {l}\nprevious: {prev}")
        else:
            print(f"Current list: {l}")
            print(f"Previous list: {prev}")
        prev = l

def tafsir_starts(file: Path) -> List[Match[str]]:
    matches: List[Match[str]] = []
    with open(file=file, mode="r", encoding="utf-8") as f:
        for line in f:
            match = verse_marker_match.search(line.strip())
            if match:
                matches.append(match)
    return matches

def arabic_to_english_numerals(text: str):
    arabic_nums = '٠١٢٣٤٥٦٧٨٩'
    english_nums = '0123456789'
    translation_table = str.maketrans(arabic_nums, english_nums)
    return text.translate(translation_table)

def normalize_numbers(ayah_str: str) -> List[int]:
    ayah_str = arabic_to_english_numerals(ayah_str)

    # Remove any unwanted characters (e.g. spaces around dashes/commas)
    ayah_str = re.sub(r"\s*[\-–ـ]\s*", "-", ayah_str)
    ayah_str = re.sub(r"\s*[,،]\s*", ",", ayah_str)

    # Split by comma
    parts = ayah_str.split(",")
    results: List[int] = []

    for part in parts:
        if "-" in part:
            start, end = sorted(map(int, part.split("-")))
            results.extend(range(start, end + 1))
        elif part.strip().isdigit():
            results.append(int(part.strip()))

    return sorted(set(results))

def extract_intro(surah_file: Path, output_dir: Path):
    logger.info(msg=f"Extracting intro for {surah_file.name}...")
    with open(file=surah_file, mode="r", encoding="utf-8") as f:
        intro_lines: List[str] = []
        for line in f:
            stripped = line.strip()
            if verse_marker_match.search(stripped):
                break
            if stripped == "":
                continue
            intro_lines.append(stripped)
    if len(intro_lines) <= 1:
        logger.info(msg=f"No intro lines found for {surah_file.name}")
        return
    logger.info(msg=f"Writing intro to {output_dir.name}...")
    file_name = output_dir / "intro.md"
    output_dir.mkdir(parents=True, exist_ok=True)
    with open(file=file_name, mode="w", encoding="utf-8") as f:
        for line in intro_lines:
            f.write(line)
            f.write(f"\n")
    logger.info(msg=f"Done writing!")

def extract_ayah_range(surah_file: Path, output_dir: Path):
    group_to_nums: Dict[int, List[int]] = {}
    group_to_range: Dict[int, List[str]] = {}
    logger.info(msg=f"Reading {surah_file.name}...")
    with open(file=surah_file, mode="r", encoding="utf-8") as f:
        boundary: bool = False
        group_num: int = 0
        previous_group: int = 0
        parse = False
        ayah_nums: List[int] = []
        for line in f:
            stripped = line.strip()
            if parse == True:
                next_match = verse_marker_match.search(stripped)
                if not next_match:
                    if stripped != "":
                        group_to_range[group_num].append(stripped)
                        if boundary:
                            group_to_range[previous_group].append(stripped)
                else:
                    if boundary:
                        boundary = False
                    next_group = normalize_numbers(
                        ayah_str=next_match.group("ayahs")
                    )
                    next_ayah = min(next_group)
                    if next_ayah not in ayah_nums:
                        if not boundary:
                            previous_group = group_num
                            boundary = True
                        group_num += 1
                        ayah_nums = next_group
                        if group_num not in group_to_nums:
                            group_to_nums[group_num] = []
                        group_to_nums[group_num] = ayah_nums
                        if group_num not in group_to_range:
                            group_to_range[group_num] = []
                        group_to_range[previous_group].append(stripped)
            else:
                match = verse_marker_match.search(stripped)
                if not match:
                    continue
                group_num += 1
                ayah_nums = normalize_numbers(
                    ayah_str=match.group("ayahs")
                )
                if group_num not in group_to_nums:
                    group_to_nums[group_num] = []
                group_to_nums[group_num] = ayah_nums
                if group_num not in group_to_range:
                    group_to_range[group_num] = []
                group_to_range[group_num].append(stripped)
                parse = True
    logger.info(msg=f"Done reading!")
    output_dir.mkdir(parents=True, exist_ok=True)
    logger.info(msg=f"Captured {len(group_to_nums)} groups")
    for group in group_to_nums:
        if group == max(group_to_nums.keys()):
            print(f"Current group: {group}")
            print(f"Previous group: {group-1}")
            if group-1 in group_to_nums:
                if min(group_to_nums[group]) < max(group_to_nums[group-1]):
                    logger.info(msg=f"Skipping residue: {group} -> {group_to_nums[group]}")
                    break
        if len(group_to_nums[group]) > 1:
            file_name = f"{group_to_nums[group][0]}-{group_to_nums[group][-1]}.md"
        else:
            file_name = f"{group_to_nums[group][0]}.md"
        file_path = output_dir / file_name
        logger.info(msg=f"Writing to {file_name}...")
        with open(file=file_path, mode="w", encoding="utf-8") as f:
            for line in group_to_range[group]:
                f.write(line)
                f.write(f"\n\n")
        logger.info(msg=f"Done writing!")





# def get_gemini_response(description: str, surah_file: Path) -> str:
#     client = genai.Client(
#         vertexai=True,
#         project=GCP_PROJECT,
#         location=GCP_REGION
#     )

#     with open(file=surah_file, mode="r", encoding="utf-8") as f:
#         surah_contents = f.read()

#         print(surah_contents)
#         contents = UserContent(
#             parts = [
#                 Part(text=description),
#                 Part(text=surah_contents),
#             ]
#         )
#         config = GenerateContentConfig(
#             system_instruction = SYSTEM_INSTRUCTION
#         )

#         response = client.models.generate_content(
#             model=str(GEMINI_LITE),
#             contents=contents,
#             config=config,
#         )

#     logger.info(msg=f"Successfully received response")
#     return str(response.text)
