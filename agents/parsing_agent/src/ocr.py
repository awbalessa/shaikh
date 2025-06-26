import logging
import cv2
import numpy as np
import pytesseract
from config import IMAGE_DPI, IMAGE_FMT, TESSERACT_CONFIG
from pdf2image import convert_from_path
from PIL import Image as ImageModule
from PIL.Image import Image as ImageClass
from pathlib import Path

logger = logging.getLogger(__name__)

def pdf_to_images(pdf: Path, firstPage: int, lastPage: int):
    for i, page in enumerate(convert_from_path(
        pdf_path=pdf,
        dpi=IMAGE_DPI,
        fmt=IMAGE_FMT,
        first_page=firstPage,
        last_page=lastPage,
    )):
        logger.info(msg=f"Converting page {firstPage + i} from {pdf.name}...")
        yield page
        logger.info("Successfully converted")

def preprocess_image(image: ImageClass) -> ImageClass:
    logger.info(msg=f"Preprocessing image...")
    img = cv2.cvtColor(np.array(image), cv2.COLOR_RGB2GRAY)
    img = cv2.GaussianBlur(img, (5, 5), 0)
    img = cv2.adaptiveThreshold(
        img, 255,
        cv2.ADAPTIVE_THRESH_GAUSSIAN_C,
        cv2.THRESH_BINARY,
        11, 2
    )
    processed = ImageModule.fromarray(img)
    logger.info(msg=f"Successfully preprocessed")
    return processed

def image_to_text(image: ImageClass) -> str:
    logger.info(msg=f"Converting preprocessed image to text...")
    output = pytesseract.image_to_string(
        image=image,
        config=TESSERACT_CONFIG,
    )
    logger.info(msg=f"Successfully converted image to text")
    return str(output)
