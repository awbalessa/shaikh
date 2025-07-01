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

SYSTEM_INSTRUCTION = """"""

logging.basicConfig(
    level=logging.INFO,
    format="%(asctime)s - %(message)s"
)

assert GEMINI is not None, "Gemini is missing"
assert GEMINI_LITE is not None, "Gemini Lite is missing"
assert GCP_PROJECT is not None, "GCP Project is missing"
assert GCP_REGION is not None, "GCP Region is missing"
