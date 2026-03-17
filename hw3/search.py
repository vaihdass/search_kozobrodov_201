import json
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


def print_results(results):
    if results:
        n = len(results)
        doc = morph.parse("документ")[0].make_agree_with_number(n).word
        verb = "Найден" if n == 1 else "Найдено"
        print(f"{verb} {n} {doc}: {', '.join(sorted(results))}")
    else:
        print("Ничего не найдено")


def main():
    index_path = BASE_DIR / "inverted_index.json"
    if not index_path.exists():
        print("Ошибка: сначала запусти python3 index_builder.py")
        return

    index, all_docs = load_index(index_path)

    print('Булев поиск. Операторы: AND, OR, NOT, скобки.')
    print('Для выхода введите "exit" или нажмите Ctrl+C.\n')

    while True:
        try:
            query = input("Запрос: ").strip()
        except (KeyboardInterrupt, EOFError):
            print()
            break

        if not query:
            continue
        if query.lower() == "exit":
            break

        try:
            results = search(query, index, all_docs)
            print_results(results)
        except ValueError as e:
            print(f"Ошибка: {e}")

        print()


if __name__ == "__main__":
    main()
