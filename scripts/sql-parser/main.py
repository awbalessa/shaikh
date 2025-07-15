import logging
from config import CHUNK_TOKEN_LIMIT, OUTPUT_DIR
from typing import List
from src.query import get_documents_by_keys
from src.chunk import find_chunk_titles, recursive_semantic_splitter, semantic_chunker, voyage_token_counter
from langchain_core.documents import Document as LangchainDocument

logger = logging.getLogger(__name__)

output_file = OUTPUT_DIR / "surah_starts"

tafsir_documents = get_documents_by_keys(
    keys=[
        (1, 1)
    ]
)

logger.info(
    msg=f"""Fetched documents by keys:
        Surahs: {sorted({row.surah for row in tafsir_documents})}
        Ayahs: {sorted({row.ayah for row in tafsir_documents})}
    """
)

docs: List[str] = [row.document for row in tafsir_documents]

for i, doc in enumerate(docs):
    chunks = recursive_semantic_splitter(
        chunker=semantic_chunker,
        document=doc,
        token_limit=CHUNK_TOKEN_LIMIT,
        token_counter=voyage_token_counter
    )
    nil = find_chunk_titles(chunks)
    for j, chunk in enumerate(chunks):
        with open(file=output_file, mode="a", encoding="utf-8") as f:
            f.write(f"""
                ===DOCUMENT {i+1} CHUNK {j+1}===
                Tokens: {voyage_token_counter([chunk.page_content])}\n
                {chunk.page_content}\n
            """)
