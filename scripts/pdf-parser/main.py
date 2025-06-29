import argparse
from pathlib import Path

from src.pipeline import run_pipeline

def main(pdf_path: Path, firstPage: int, lastPage: int, output_file: Path):
    print(f"PDF Path: {pdf_path}")
    print(f"Pages: {firstPage}:{lastPage}")
    run_pipeline(pdf_path, firstPage, lastPage, output_file)


if __name__ == "__main__":
    parser = argparse.ArgumentParser(description="Process a range of PDF pages.")
    parser.add_argument("pdf", type=Path, help="Path to the input PDF file")
    parser.add_argument("--first-page", required=True, type=int, help="The first page to start parsing from")
    parser.add_argument("--last-page", required=True, type=int, help="The last page to parse")
    parser.add_argument("--output-file", required=True, type=Path, help="The output file file to write into")

    args = parser.parse_args()
    main(args.pdf, args.first_page, args.last_page, args.output_file)
