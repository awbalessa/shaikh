from pathlib import Path
import logging
import sqlite3

logging.basicConfig(
    level=logging.INFO,
    format="%(asctime)s - %(message)s"
)

BASE_DIR = Path(__file__).resolve().parent
ASSETS_DIR = BASE_DIR / "assets"
SQLITE_PATH = ASSETS_DIR / "ar-tafsir-ibn-kathir.db"

CURSOR = sqlite3.connect(
    database=SQLITE_PATH
).cursor()
