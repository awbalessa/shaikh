from pathlib import Path
import logging
import sqlite3
import psycopg2

logging.basicConfig(
    level=logging.INFO,
    format="%(asctime)s - %(message)s"
)

BASE_DIR = Path(__file__).resolve().parent
ASSETS_DIR = BASE_DIR / "assets"
SQLITE_PATH = ASSETS_DIR / "ar-tafsir-ibn-kathir.db"
GRANULARITY = "ayah"
CONTENT_TYPE = "tafsir"
SOURCE = "Tafsir Ibn Kathir"
POSTGRES_DSN = "host=localhost dbname=shaikh user=azizalessa port=5432"

SQLITE_CURSOR = sqlite3.connect(
    database=SQLITE_PATH
).cursor()

POSTGRES_CURSOR = psycopg2.connect(
    dsn=POSTGRES_DSN
).cursor()
