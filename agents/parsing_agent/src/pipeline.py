from pathlib import Path
from src.gemini import get_gemini_response
from src.ocr import images_to_text, pdf_to_images, preprocess_images
from src.write import append_page_to_file


def run_pipeline(pdf: Path, firstPage: int, lastPage: int, output_file: Path):
    images = pdf_to_images(pdf, firstPage, lastPage)
    processed_images = preprocess_images(images)
    text_list = images_to_text(processed_images)
    for i, _ in enumerate(text_list):
        page = get_gemini_response(text_list[i], processed_images[i], firstPage+i)
        append_page_to_file(page_num=i+firstPage, page=page, output_file=output_file)
