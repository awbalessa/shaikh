import regex as re
import logging
from typing import List
from pathlib import Path
# from google import genai
# from google.genai.types import Part, UserContent, GenerateContentConfig
# from config import GCP_PROJECT, GCP_REGION, GEMINI_LITE, SYSTEM_INSTRUCTION

logger = logging.getLogger(__name__)

verse_marker_match = re.compile(
    pattern=r"^\[\d+\]:\s[\*\#\s]*سورة\s\p{Script=Arabic}+\b.*الآ(?:ية|يتان|يات)"
)

def arabic_to_english_numerals(text: str):
    arabic_nums = '٠١٢٣٤٥٦٧٨٩'
    english_nums = '0123456789'
    translation_table = str.maketrans(arabic_nums, english_nums)
    return text.translate(translation_table)

def extract_ayah_range(file: Path) -> List[str]:
    matches: List[str] = []
    with open(file=file, mode="r", encoding="utf-8") as f:
        for line in f:
            if verse_marker_match.search(line.strip()):
                matches.append(line.strip())
    return matches


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
