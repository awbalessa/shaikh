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
        (1, 5)
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
        print(f"Title: {chunk.metadata['title']}\n")
        print(chunk.page_content)
        print(f"\n\n")

# Use Tasaphyne for ingestion phase. And implement your own light stemmer in Go. You have the code in front of you. It's not hard. The last step now before ingesting is using Tasaphyne to stem the chunk titles and chunks for bm25 search.
