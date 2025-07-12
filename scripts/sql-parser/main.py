from src.query import create_documents, get_tafsirs

tafsirs = get_tafsirs(
    limit=None,
    offset=101,
)

create_documents(tafsirs_objs=tafsirs)
