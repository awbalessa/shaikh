from pprint import pprint
import logging
from config import CHUNK_TOKEN_LIMIT
from typing import List
from src.labels import AyahLabelOffset, LabelContentTypeTafsir, LabelSourceTafsirIbnKathir, SurahLabelOffset
from src.preprocess import tokenize_for_bm25
from src.query import Chunk, ChunkWithoutEmbeddings, Document, create_chunks, get_documents_by_keys
from src.chunk import embed_chunks, find_chunk_titles, recursive_semantic_splitter, semantic_chunker, voyage_token_counter

logger = logging.getLogger(__name__)

def get_chunks_without_embedding(doc: Document) -> List[ChunkWithoutEmbeddings]:
    has_parent: bool = True
    parent_id: int = doc.id
    context_header: str = f"تفسير ابن كثير للآية {doc.surah}:{doc.ayah}"
    surah: int = doc.surah
    ayah: int = doc.ayah
    chunks = recursive_semantic_splitter(
        chunker=semantic_chunker,
        document=doc.document,
        token_limit=CHUNK_TOKEN_LIMIT,
        token_counter=voyage_token_counter
    )
    chunks = find_chunk_titles(chunks)
    ChunkObjects: List[ChunkWithoutEmbeddings] = []
    for i, chunk in enumerate(chunks):
        seq_id: int = i+1
        raw_chunk: str = chunk.page_content
        tokenized_chunk: str = tokenize_for_bm25(raw_chunk)
        chunk_title: str = chunk.metadata.get('title', '')
        if chunk_title == '':
            raise Exception("Chunk title is missing")
        tokenized_chunk_title: str = tokenize_for_bm25(chunk_title)
        embedded_chunk: str = f"{context_header} | {chunk_title}\n\n{raw_chunk}"
        labels: List[int] = [
            LabelContentTypeTafsir,
            LabelSourceTafsirIbnKathir,
            SurahLabelOffset + surah,
            AyahLabelOffset + ayah,
        ]
        ch: ChunkWithoutEmbeddings = ChunkWithoutEmbeddings(
            seq_id=seq_id,
            raw_chunk=raw_chunk,
            tokenized_chunk=tokenized_chunk,
            chunk_title=chunk_title,
            tokenized_chunk_title=tokenized_chunk_title,
            context_header=context_header,
            embedded_chunk=embedded_chunk,
            labels=labels,
            has_parent=has_parent,
            parent_id=parent_id,
            surah=surah,
            ayah=ayah,
        )
        ChunkObjects.append(ch)

    return ChunkObjects

def get_chunks_with_embedding(chunks: List[ChunkWithoutEmbeddings]) -> List[Chunk]:
    vectors = embed_chunks(chunks)
    final: List[Chunk] = []
    if len(vectors) != len(chunks):
        raise ValueError("Mismatch between chunks and corresponding vectors")
    for ch, vec in zip(chunks, vectors):
        new = Chunk(
            **ch.__dict__,
            embedding=vec
        )
        final.append(new)
    return final


tafsir_documents = get_documents_by_keys(
    keys=[
        (2, i) for i in range (1, 287)
    ]
)

logger.info(
    msg=f"""Fetched documents by keys:
        Surahs: {sorted({row.surah for row in tafsir_documents})}
        Ayahs: {sorted({row.ayah for row in tafsir_documents})}
    """
)

for doc in tafsir_documents:
    chunk_objs = get_chunks_without_embedding(doc)
    chunk_objs = get_chunks_with_embedding(chunk_objs)
    create_chunks(chunk_objs)
