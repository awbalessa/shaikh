import os
import logging
from dotenv import load_dotenv
from pathlib import Path

load_dotenv()
GEMINI = os.getenv("GEMINI")
GEMINI_LITE = os.getenv("GEMINI_LITE")
GCP_REGION = os.getenv("GCP_REGION")
GCP_PROJECT = os.getenv("GCP_PROJECT")
BASE_DIR = Path(__file__).resolve().parent
ASSETS_DIR = BASE_DIR / "assets"
OUTPUT_DIR = BASE_DIR / "output"

SYSTEM_INSTRUCTION = """You will be given a document that contains the full Tafsir of a single surah, where each line is marked with its line number in square brackets (e.g. [1], [2], etc). Use the line numbers to determine where each ayah's tafsir begins and ends.

Your goal is to extract a JSON object like this:

{
  "ayah-tafsir": {
    "1": { "start": 12, "end": 44 },
    "2": { "start": 45, "end": 76 },
    "3": { "start": 77, "end": 85 }
    // Continue until the final ayah
  }
}

This means that the tafsir of Ayah 1 begins at line 12 and ends at line 44, inclusive.

Guidelines:
- You are a **JSON-only API**. Output only valid JSON — no comments, no explanations, and no markdown formatting.
- Each ayah must be assigned a line range, and **every line in the document must be covered** by at least one ayah's tafsir.
- If a single tafsir section covers **multiple ayat together**, assign the same `start` and `end` values to each ayah number.
- The ayah numbers must appear in sequential order in the JSON output."""

logging.basicConfig(
    level=logging.INFO,
    format="%(asctime)s - %(message)s"
)

assert GEMINI is not None, "Gemini is missing"
assert GEMINI_LITE is not None, "Gemini Lite is missing"
assert GCP_PROJECT is not None, "GCP Project is missing"
assert GCP_REGION is not None, "GCP Region is missing"
