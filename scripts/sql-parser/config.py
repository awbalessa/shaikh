from pydantic import SecretStr
from dotenv import load_dotenv
from pathlib import Path
from google import genai
import os
import logging
import sqlite3
import psycopg2

logging.basicConfig(
    level=logging.INFO,
    format="%(asctime)s - %(message)s"
)

load_dotenv()

GCP_PROJECT = os.getenv("GCP_PROJECT")
GCP_REGION = os.getenv("GCP_REGION")

gemini_client = genai.Client(
    vertexai=True,
    project=GCP_PROJECT,
    location=GCP_REGION,
)

GEMINI_LITE = os.getenv("GEMINI_LITE")
VOYAGE_API_KEY = SecretStr(str(os.getenv("VOYAGE_API_KEY")))

GRANULARITY = "ayah"
CONTENT_TYPE = "tafsir"
SOURCE = "Tafsir Ibn Kathir"

BASE_DIR = Path(__file__).resolve().parent
ASSETS_DIR = BASE_DIR / "assets"
OUTPUT_DIR = BASE_DIR / "output"
SQLITE_PATH = ASSETS_DIR / "ar-tafsir-ibn-kathir.db"
POSTGRES_DSN = "host=localhost dbname=shaikh user=azizalessa port=5432"
CHUNK_TOKEN_LIMIT = 2500

SQLITE_CURSOR = sqlite3.connect(
    database=SQLITE_PATH
).cursor()

POSTGRES_CURSOR = psycopg2.connect(
    dsn=POSTGRES_DSN
).cursor()
