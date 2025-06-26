import gc
import time
import logging
from pathlib import Path
from src.gemini import get_gemini_response
from src.ocr import image_to_text, pdf_to_image, preprocess_image
from src.write import append_page_to_file

logger = logging.getLogger(__name__)

def run_pipeline(pdf: Path, firstPage: int, lastPage: int, output_file: Path):
    for page_num in range(firstPage, lastPage + 1):
        start_time = time.time()
        try:
            image = pdf_to_image(pdf, page_num=page_num)
            processed_image = preprocess_image(image)
            text = image_to_text(processed_image)
            page = get_gemini_response(text, processed_image, page_num)
            append_page_to_file(page_num=page_num, page=page, output_file=output_file)
        except Exception as e:
            logger.error(msg=f"Failed on page {page_num}: {e}")
            raise e

        del image, processed_image, text, page
        gc.collect()

        logger.info(msg=f"Page {page_num} done in {time.time() - start_time:.2f}s")
