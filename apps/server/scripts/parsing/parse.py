import os
import cv2
import pytesseract
import numpy as np
from io import BytesIO
from typing import List
from PIL import Image as ImageModule
from PIL.Image import Image as ImageClass
from pdf2image import convert_from_path
from pathlib import Path
from dotenv import load_dotenv
from google import genai
from google.genai import types
from google.genai.types import Part, UserContent

CWD: Path = Path.cwd()
ENV_PATH = CWD / "apps/server/.env"
PDF_PATH = CWD / "assets/tafsir/Quran Tafsir Ibn Kathir.pdf"
INPUT_DIR = CWD / "assets/pages"
OUTPUT_TEXT: List[str] = []
load_dotenv(ENV_PATH)
GCP_PROJECT = os.getenv("GCP_PROJECT")
GCP_REGION = os.getenv("GCP_REGION")
GENERATION_MODEL = os.getenv("GENERATION_MODEL")

assert GENERATION_MODEL is not None, "Model env var is missing"

def pil_image_to_bytes(image: ImageClass) -> bytes:
        buffer = BytesIO()
        image.save(buffer, format="PNG")
        return buffer.getvalue()

# def pdf_to_images(pdf_path: Path, startPage: int, endPage: int) -> List[Image]
# def process_images(image_dir: Path) -> List[Image]:
# def save_images(dir: Path, name_to_img: Dict[str, Image])
# def images_to_text(images: List[Image], pages_text: List[str])

client = genai.Client(
   vertexai=True,
   project=GCP_PROJECT,
   location=GCP_REGION
)

gemini_config = types.GenerateContentConfig(
    system_instruction = """You are an Arabic linguistics expert specializing in reviewing and correcting OCR (Optical Character Recognition) results from classical Arabic books.

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
)

images = convert_from_path(pdf_path=PDF_PATH, dpi=600, first_page=5, last_page=5)
# for i, image in enumerate(images):
#     image.save(fp=INPUT_DIR / f"page_{i+1:04}", format="PNG")
print("PDF converted to images")

for i, image in enumerate(images):
    img = cv2.cvtColor(np.array(image), cv2.COLOR_RGB2GRAY)
    img = cv2.GaussianBlur(img, (5, 5), 0)
    img = cv2.adaptiveThreshold(
        img, 255,
        cv2.ADAPTIVE_THRESH_GAUSSIAN_C,
        cv2.THRESH_BINARY,
        11, 2
    )
    print("Image processed")

    pil_image = ImageModule.fromarray(img)

    config = '--oem 3 --psm 6 -l ara'
    ocr_output = pytesseract.image_to_string(image=pil_image, config=config)
    print("Image converted to OCR text")

    contents = UserContent(
        parts = [
            Part(text="Here's the OCR output for this page:"),
            Part(text=str(ocr_output)),
            Part(text="Here's the image:"),
            Part.from_bytes(
                data=pil_image_to_bytes(pil_image),
                mime_type="image/png",
            ),
        ]
    )

    print("Sending to Gemini...")
    response = client.models.generate_content(
        model=GENERATION_MODEL,
        contents=contents,
        config=gemini_config
    )

    print("OCR output:")
    print(ocr_output)
    print("\n")
    print("Gemini output:")
    print(response.text)
