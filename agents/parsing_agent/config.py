import os
import logging
from pathlib import Path
from dotenv import load_dotenv

load_dotenv()
GEMINI = os.getenv("GEMINI")
GCP_REGION = os.getenv("GCP_REGION")
GCP_PROJECT = os.getenv("GCP_PROJECT")
BASE_DIR = Path(__file__).resolve().parent
FUNCTIONS_DIR = BASE_DIR / "functions"
ASSETS_DIR = BASE_DIR / "assets"
OUTPUT_DIR = BASE_DIR / "output"
IMAGE_DPI = 600
IMAGE_FMT = "png"
TESSERACT_CONFIG = "--oem 3 --psm 6 -l ara"
SYSTEM_INSTRUCTION = """You are an Arabic linguistics expert specializing in reviewing and correcting OCR (Optical Character Recognition) results from classical Arabic books.

    You will receive:
	1.	A scanned image of a book page (image).
	2.	The OCR-generated text from that page (text).

    Your task is to fully review and correct the OCR-extracted text:
	•	Correct spelling, grammatical, and diacritical errors.
	•	Maintain paragraph formatting and headings.
	•	Ensure that proper names, dates, and punctuation are accurate.
	•	Refer to the original image in cases of uncertainty or clear distortion in the extracted text.

    Output the corrected text in well-formed Arabic, as if it were manually transcribed from a high-quality printed version.
    Output the text in Markdown format, and apply visual separation between different sections of text."""

logging.basicConfig(
    level=logging.INFO,
    format="%(asctime)s - %(levelname)s - %(message)s"
)

assert GEMINI is not None, "Gemini is missing"
assert GCP_PROJECT is not None, "GCP Project is missing"
assert GCP_REGION is not None, "GCP Region is missing"
