import camel_tools.utils.dediac as camel
from typing import List
import regex as re
from bs4 import BeautifulSoup
from regex.regex import Match
from config import CURSOR

inline_annotations = re.compile(
    pattern=r"\s\[\[.*?\]\]"
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
    \s*
    (?P<ayat_text>.*?)
    \s*
    </span>
    [\s\n]*
    (?:\s*\n*\[(?P<ayat_num>.*?)\])?""",
    flags=re.VERBOSE | re.DOTALL
)
ayat_block = re.compile(
    pattern=r"\uFD3F(?P<ayat_text>.*?)\uFD3E"
)
tuple_start = re.compile(
    pattern=r"^\('"
)
tuple_end = re.compile(
    pattern=r"',\)$"
)
bolded_block = re.compile(
    pattern=r"\*\*.*?\*\*"
)
heading_block = re.compile(
    pattern=r"[\s\n]*<"
)

def strip_diacritics(text: str) -> str:
    return camel.dediac_ar(text)

def strip_line_diacritics(text: str) -> str:
    final: List[str] = []
    for line in text.split("\n"):
        if bolded_block.search(line):
            final.append(line)
            continue
        final.append(strip_diacritics(line))
    return "\n".join(final).strip()

def replace_ayah_span(match: Match[str]) -> str:
    ayat_text = match.group("ayat_text").strip()
    ayat_num = match.group("ayat_num")
    if ayat_num is None:
        return f" {ayat_text} "
    else:
        ayat_num = ayat_num.strip()
        ayat_num = strip_diacritics(ayat_num)
    return f" {ayat_text} [{ayat_num}]"

def preprocess(text: str) -> str:
    text = re.sub(inline_annotations, '', text)
    text = re.sub(div_blocks, '', text)
    text = re.sub(opening_paragraph_blocks, '\n', text)
    text = re.sub(closing_paragraph_blocks, '', text)
    text = re.sub(tuple_start, '', text)
    text = re.sub(tuple_end, '', text)
    text = ayat_span_block.sub(replace_ayah_span, text)
    text = strip_line_diacritics(text)
    return text

CURSOR.execute("""
    SELECT text FROM tafsir
    WHERE TRIM(text) != ''
        AND (
            ayah_key = '1:1' OR
            ayah_key = '1:2' OR
            ayah_key = '1:3' OR
            ayah_key = '1:4' OR
            ayah_key = '1:5' OR
            ayah_key = '1:6' OR
            ayah_key = '1:7'
        )
    """)
results = CURSOR.fetchall()
for i ,row in enumerate(results):
    soup = BeautifulSoup(str(row), 'html.parser')
    pretty = soup.prettify()
    # print(pretty)
    clean = preprocess(
        text=str(pretty)
    )
    print(f"\n#{i+1}\n")
    print(clean)
