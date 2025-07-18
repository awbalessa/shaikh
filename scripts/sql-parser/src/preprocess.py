from tashaphyne.stopwords import STOPWORDS
import camel_tools.utils.dediac as camel
import regex as re
from typing import List, Tuple, Any
from bs4 import BeautifulSoup
from regex.regex import Match

not_arabic_char = re.compile(
    pattern = r"[^\p{IsArabic}]+"
)

inline_annotations = re.compile(
    pattern=r"\s*\[\[.*?\]\]\s*"
)
div_blocks = re.compile(
    pattern=r"[\s\n]*</?div.*?>[\s\n]*"
)
opening_paragraph_blocks = re.compile(
    pattern=r"[\s\n]*<p>[\s\n]*"
)
closing_paragraph_blocks = re.compile(
    pattern=r"[\s\n]*</?p>[\n\s]*"
)
ayat_span_block = re.compile(
    pattern=
    r"""[\s\n]*
    <span[^>]*>
    [\s\n]*
    (?P<ayat_text>.*?)
    [\s\n]*
    </span>
    [\s\n]*
    (?:\s*\n*\[(?P<ayat_num>.*?)\])?""",
    flags=re.VERBOSE
)
ayat_block = re.compile(
    pattern=r"\uFD3F(?P<ayat_text>.*?)\uFD3E"
)
heading_block = re.compile(
    pattern=
    r"""
    [\s\n]*<h\d>
    \s*
    (?P<heading>.*?)
    \s*
    </h\d>[\s\n]*
    """,
    flags=re.VERBOSE
)

def strip_line_diacritics(text: str) -> str:
    final: List[str] = []
    for part in re.split(pattern=r"\*\*", string=text):
        match = ayat_block.search(part)
        if match:
            final.append(part)
            continue
        final.append(strip_diacritics(part))
    return "".join(final)

def replace_ayah_span(match: Match[str]) -> str:
    ayat_text = match.group("ayat_text").strip()
    ayat_num = match.group("ayat_num")
    if ayat_num is None:
        return f" **{ayat_text}** "
    else:
        ayat_num = ayat_num.strip()
        ayat_num = strip_diacritics(ayat_num)
    return f" **{ayat_text}** [{ayat_num}]"

def preprocess(text: str) -> str:
    text = re.sub(inline_annotations, ' ', text)
    text = re.sub(div_blocks, f'\n\n' , text)
    text = re.sub(opening_paragraph_blocks, f'\n\n', text)
    text = re.sub(closing_paragraph_blocks, f'\n\n', text)
    text = re.sub(heading_block, '', text)
    text = ayat_span_block.sub(replace_ayah_span, text)
    text = strip_line_diacritics(text)
    text = text.strip()
    return text

def strip_diacritics(text: str) -> str:
    return camel.dediac_ar(text)

def remove_punctuation(text: str) -> str:
    return re.sub(not_arabic_char, ' ', text)

def filter_stopwords(text: str) -> str:
    words: List[str] = text.split(" ")
    filtered: List[str] = []
    for word in words:
        stripped = word.strip()
        if stripped in STOPWORDS:
            continue
        filtered.append(stripped)
    return " ".join(filtered)

def tokenize_for_bm25(text: str) -> str:
    return filter_stopwords(
        remove_punctuation(
            strip_diacritics(text)
        )
    )
