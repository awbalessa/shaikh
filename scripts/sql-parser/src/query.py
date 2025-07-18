import logging
from dataclasses import dataclass
import regex as re
from pgvector import Vector
from src.preprocess import preprocess
from config import CONTENT_TYPE, GRANULARITY, SQLITE_CURSOR, SOURCE, POSTGRES_CURSOR
from typing import List, Any, Tuple, Optional

logger = logging.getLogger(__name__)

class TafsirObj:
    surah_num: int
    ayah_nums: List[int]
    tafsir: str
    context_header: str

    def __init__(self, surah_num: int, ayah_nums: List[int], tafsir: str, context_header: str = ""):
        self.surah_num = surah_num
        self.ayah_nums = ayah_nums
        self.tafsir = tafsir
        self.context_header = context_header

class Document:
    id: int
    surah: int
    ayah: int
    document: str

    def __init__(self, id: int, surah: int, ayah: int, document: str):
        self.id = id
        self.surah = surah
        self.ayah = ayah
        self.document = document

@dataclass
class ChunkWithoutEmbeddings:
    seq_id: int
    raw_chunk: str
    tokenized_chunk: str
    chunk_title: str
    tokenized_chunk_title: str
    context_header: str
    embedded_chunk: str
    labels: List[int]
    has_parent: bool
    parent_id: Optional[int]
    surah: Optional[int]
    ayah: Optional[int]

@dataclass
class Chunk(ChunkWithoutEmbeddings):
    embedding: Vector

def find_ayah_keys(row: Any) -> List[Tuple[int, int]]:
    list_of_ayat: List[Tuple[int, int]] = []
    matches = re.findall(
        pattern=r"(\d+):(\d+)",
        string=row[0]
    )
    for surah, ayah in matches:
        surah = int(surah)
        ayah = int(ayah)
        list_of_ayat.append((surah, ayah))
    return list_of_ayat

def get_tafsirs(limit: int | None, offset: int | None) -> List[TafsirObj]:
    tafsir_objs: List[TafsirObj] = []
    if limit is not None and offset is not None:
        SQLITE_CURSOR.execute(f"""
            SELECT ayah_keys, text FROM tafsir
            WHERE TRIM(text) != ''
            LIMIT {limit}
            OFFSET {offset};
            """)
        results = SQLITE_CURSOR.fetchmany(limit)
    elif limit is None and offset is not None:
        SQLITE_CURSOR.execute(f"""
            SELECT ayah_keys, text FROM tafsir
            WHERE TRIM(text) != ''
            LIMIT {-1}
            OFFSET {offset};
            """)
        results = SQLITE_CURSOR.fetchall()
    elif limit is not None and offset is None:
        SQLITE_CURSOR.execute(f"""
            SELECT ayah_keys, text FROM tafsir
            WHERE TRIM(text) != ''
            LIMIT {limit};
            """)
        results = SQLITE_CURSOR.fetchmany(limit)
    else:
        SQLITE_CURSOR.execute(f"""
            SELECT ayah_keys, text FROM tafsir
            WHERE TRIM(text) != ''
            """)
        results = SQLITE_CURSOR.fetchall()
    logger.info(msg=f"Fetched {len(results)} results")
    for row in results:
        ayah_keys = find_ayah_keys(row)
        surah_num: int = ayah_keys[0][0]
        ayah_nums: List[int] = []
        tafsir = str(row[1])
        for key in ayah_keys:
            ayah_nums.append(key[1])
        tafsir_objs.append(TafsirObj(
            surah_num=surah_num,
            ayah_nums=ayah_nums,
            tafsir=tafsir,
        ))
        logger.info(msg=f"Pulled Tafsir for {surah_num}:{ayah_nums[0]}-{ayah_nums[-1]}")
    return tafsir_objs

def create_documents(tafsirs_objs: List[TafsirObj]):
    logger.info(msg=f"Inserting {len(tafsirs_objs)} tafsirs into documents table...")
    for obj in tafsirs_objs:
        for ayah in obj.ayah_nums:
            context_header = f"{SOURCE} for Ayah {obj.surah_num}:{ayah}"
            POSTGRES_CURSOR.execute(
                query="""
                INSERT INTO documents
                (granularity, content_type, source, context_header, document, surah, ayah)
                VALUES (%s, %s, %s, %s, %s, %s, %s)
                """,
                vars=(GRANULARITY, CONTENT_TYPE, SOURCE, context_header, preprocess(obj.tafsir), obj.surah_num, ayah)
            )
            logger.info(msg=f"Inserted {context_header}")
    POSTGRES_CURSOR.connection.commit()
    logger.info(msg=f"Committed to documents table!")

def get_documents_by_keys(keys: List[Tuple[int, int]]) -> List[Document]:
    docs: List[Document] = []
    for key in keys:
        [surah, ayah] = key
        POSTGRES_CURSOR.execute(
            query="""
            SELECT id, surah, ayah, document, context_header
            FROM documents
            WHERE surah = %s
                AND ayah = %s
            """,
            vars=(surah, ayah)
        )
        result = POSTGRES_CURSOR.fetchone()
        assert result is not None, "No result"
        docs.append(Document(
            id=result[0],
            surah=result[1],
            ayah=result[2],
            document=result[3],
        ))
    return docs

def create_chunks(chunks: List[Chunk]):
    logger.info(msg=f"Inserting {len(chunks)} chunks into chunks table...")
    for ch in chunks:
        POSTGRES_CURSOR.execute(
            query="""
            INSERT INTO chunks
            (sequence_id, granularity, content_type, source, raw_chunk, tokenized_chunk, chunk_title, tokenized_chunk_title, context_header, embedded_chunk, labels, embedding, has_parent, parent_id, surah, ayah)
            VALUES (%s, %s, %s, %s, %s, %s, %s, %s, %s, %s, %s, %s, %s, %s, %s, %s)
            """,
            vars=(ch.seq_id, GRANULARITY, CONTENT_TYPE, SOURCE, ch.raw_chunk, ch.tokenized_chunk, ch.chunk_title, ch.tokenized_chunk_title, ch.context_header, ch.embedded_chunk, ch.labels, ch.embedding, ch.has_parent, ch.parent_id, ch.surah, ch.ayah)
        )
        logger.info(msg=f"Inserted Chunk #{ch.seq_id} from {SOURCE} for {ch.surah}:{ch.ayah}")
    POSTGRES_CURSOR.connection.commit()
    logger.info(msg=f"Committed to chunks table!")
