import logging
import json
import time
from src.query import ChunkWithoutEmbeddings
from pgvector import Vector
from google.genai.chats import Part
from google.genai.types import UserContent
from voyageai import Client as VoyageClient
from typing import Callable, List
from langchain_voyageai import VoyageAIEmbeddings
from langchain_experimental.text_splitter import SemanticChunker
from langchain_core.documents import Document as LangchainDocument
from config import VOYAGE_API_KEY, gemini_client, GEMINI_LITE

assert GEMINI_LITE is not None, "Gemini Lite is None"
assert VOYAGE_API_KEY is not None, "Voyage API Key is None"

SYSTEM_INSTRUCTION="""I will provide you an Arabic chunk taken from a larger document.

Please generate a self-contained, succinct Arabic title that captures the main idea of the chunk, and situates it within the overall document for the purposes of improving search retrieval of the chunk."""

RESPONSE_SCHEMA = {
    "type": "OBJECT",
    "properties": {
        "title": {"type": "STRING"}
    }
}

RESPONSE_MIME_TYPE = "application/json"

logger = logging.getLogger(__name__)

vo_client = VoyageClient(
    api_key=VOYAGE_API_KEY.get_secret_value()
)

vo_embed = VoyageAIEmbeddings(
    model="voyage-3.5-lite",
    batch_size=25,
    show_progress_bar=True,
    truncation=False,
    api_key=VOYAGE_API_KEY,
)

semantic_chunker = SemanticChunker(
    embeddings=vo_embed,
    buffer_size=1,
    add_start_index=True,
    breakpoint_threshold_type="percentile",
    breakpoint_threshold_amount=90,
    number_of_chunks=None,
    sentence_split_regex = r"(?<!\.\.)(?<=[.؟!])\s+",
    min_chunk_size=500
)

def voyage_token_counter(texts: List[str]) -> int:
    return vo_client.count_tokens(
        texts=texts,
        model="voyage-3.5"
    )

def recursive_semantic_splitter(
    chunker: SemanticChunker,
    document: str,
    token_limit: int,
    token_counter: Callable[[List[str]], int]
) -> List[LangchainDocument]:
    final_chunks: List[LangchainDocument] = []
    initial_chunks = chunker.create_documents(
        texts=[document]
    )
    logger.info(msg=f"Split document into {len(initial_chunks)} chunks")
    for i, chunk in enumerate(initial_chunks):
        token_count = token_counter(
            [chunk.page_content]
        )
        if token_count > token_limit:
            logger.info(msg=f"Recursively splitting chunk {i+1} into smaller chunks\ntoken_count: {token_count}, token_limit: {token_limit}")
            child_chunks = recursive_semantic_splitter(
                chunker=chunker,
                document=chunk.page_content,
                token_limit=token_limit,
                token_counter=token_counter
            )
            final_chunks.extend(child_chunks)
        else:
            logger.info(msg=f"Appending chunk {i+1} without recursive splitting\ntoken_count: {token_count}, token_limit: {token_limit}")
            final_chunks.append(chunk)
    logger.info(msg=f"Final list has {len(final_chunks)} chunks")
    return final_chunks

def find_chunk_titles(chunks: List[LangchainDocument]) -> List[LangchainDocument]:
    logger.info(msg=f"Sending {len(chunks)} to Gemini...")
    for i, chunk in enumerate(chunks):
        prompt = UserContent(
            parts= [
                Part(text=f"Chunk #{i+1}:\n\n{chunk.page_content}"),
            ]
        )

        res = gemini_client.models.generate_content(
        model=GEMINI_LITE,
        contents=prompt,
        config={
            "response_mime_type": RESPONSE_MIME_TYPE,
            "response_schema": RESPONSE_SCHEMA,
            "system_instruction": SYSTEM_INSTRUCTION,
            }
        )
        logger.info(msg=f"Found title for Chunk #{i+1}")

        response_dict = json.loads(str(res.text))
        chunks[i].metadata["title"] = response_dict["title"]
    return chunks

def embed_chunks(chunks: List[ChunkWithoutEmbeddings]) -> List[Vector]:
    texts: List[str] = [ch.embedded_chunk for ch in chunks]
    logger.info(msg=f"Sending {len(texts)} chunks to Voyage to embed...")
    start_time = time.time()
    embedding_obj = vo_client.embed(
        texts=texts,
        model="voyage-3.5",
        input_type="document",
        truncation=False,
        output_dimension=1024,
        output_dtype="float",
    )

    logger.info(msg=f"Returned {len(embedding_obj.embeddings)} embeddings in {time.time() - start_time}")
    vectors: List[Vector] = []
    for list_of_float in embedding_obj.embeddings:
        vectors.append(Vector(list_of_float))
    return vectors
