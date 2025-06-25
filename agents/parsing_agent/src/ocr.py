import logging
import cv2
import numpy as np
import pytesseract
from config import IMAGE_DPI, IMAGE_FMT, TESSERACT_CONFIG
from pdf2image import convert_from_path
from PIL import Image as ImageModule
from PIL.Image import Image as ImageClass
from pathlib import Path
from typing import List

logger = logging.getLogger(__name__)

def pdf_to_images(pdf: Path, firstPage: int, lastPage: int) -> List[ImageClass]:
    logger.info(msg=f"Converting pages {firstPage}:{lastPage} from {pdf.name}...")
    images = convert_from_path(
        pdf_path=pdf,
        dpi=IMAGE_DPI,
        fmt=IMAGE_FMT,
        first_page=firstPage,
        last_page=lastPage
    )
    logger.info("Successfully converted")
    return images

def preprocess_images(images: List[ImageClass]) -> List[ImageClass]:
    logger.info(msg=f"Preprocessing {len(images)} images...")
    image_list: List[ImageClass] = []
    for image in images:
        img = cv2.cvtColor(np.array(image), cv2.COLOR_RGB2GRAY)
        img = cv2.GaussianBlur(img, (5, 5), 0)
        img = cv2.adaptiveThreshold(
            img, 255,
            cv2.ADAPTIVE_THRESH_GAUSSIAN_C,
            cv2.THRESH_BINARY,
            11, 2
        )
        image = ImageModule.fromarray(img)
        image_list.append(image)
    logger.info(msg=f"Successfully preprocessed")
    return image_list

def images_to_text(images: List[ImageClass]) -> List[str]:
    logger.info(msg=f"Converting {len(images)} preprocessed images to text...")
    output_list: List[str] = []
    for image in images:
        output = pytesseract.image_to_string(
            image=image,
            config=TESSERACT_CONFIG,
        )
        output_list.append(str(output))
    logger.info(msg=f"Successfully converted {len(images)} images to text")
    return output_list
