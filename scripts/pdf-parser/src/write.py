import logging
from pathlib import Path

logger = logging.getLogger(__name__)

def append_page_to_file(page_num: int, page: str, output_file: Path):
    logger.info(msg=f"Writing Page #{page_num} to {output_file.name}...")
    with open(output_file, "a", encoding="utf-8") as f:
        f.write(f'# Page {page_num}\n\n')
        f.write(page.strip() + "\n\n")
    logger.info(msg=f"Successfully wrote page {page_num}")
