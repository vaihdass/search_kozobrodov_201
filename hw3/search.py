import json
import sys
from pathlib import Path

import pymorphy3

from bool_parser import search

BASE_DIR = Path(__file__).resolve().parent
morph = pymorphy3.MorphAnalyzer()


def load_index(path):
    data = json.loads(path.read_text(encoding="utf-8"))
    index = {lemma: set(doc_ids) for lemma, doc_ids in data.items()}
    all_docs = {d for docs in index.values() for d in docs}
    return index, all_docs


def main():
    if len(sys.argv) < 2:
        print('Использование: python3 search.py "<запрос>"')
        print('Пример: python3 search.py "курица AND рис"')
        sys.exit(1)

    query = " ".join(sys.argv[1:])

    index_path = BASE_DIR / "inverted_index.json"
    if not index_path.exists():
        print("Ошибка: сначала запусти python3 index_builder.py")
        sys.exit(1)

    index, all_docs = load_index(index_path)

    try:
        results = search(query, index, all_docs)
    except ValueError as e:
        print(f"Ошибка: {e}")
        sys.exit(1)

    # Адекватное склонение слов "найден" и "документ" в зависимости от количества найденных результатов
    if results:
        n = len(results)
        doc = morph.parse("документ")[0].make_agree_with_number(n).word
        verb = "Найден" if n == 1 else "Найдено"
        print(f"{verb} {n} {doc}: {', '.join(sorted(results))}")
    else:
        print("Ничего не найдено")


if __name__ == "__main__":
    main()
