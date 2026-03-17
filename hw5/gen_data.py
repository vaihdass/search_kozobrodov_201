"""Генерация JSON-данных для Go поисковой системы из hw4 TF-IDF файлов."""

import json
import re
from pathlib import Path

from bs4 import BeautifulSoup

BASE_DIR = Path(__file__).resolve().parent
PAGES_DIR = BASE_DIR / "../pages"
LEMMAS_DIR = BASE_DIR / "../hw4/lemmas"
DATA_DIR = BASE_DIR / "data"

INDEX_FILE = BASE_DIR / "../hw1/index.txt"


# Переиспользовано из задания 4
def extract_text(html):
    soup = BeautifulSoup(html, "lxml")
    for tag in soup.find_all({"script", "style", "nav", "header", "footer", "aside"}):
        tag.decompose()
    for el in soup.find_all(class_=re.compile(
        r"menu-main|similar-recipe|banner|author-card-short|about-daria|post-category|rating|comments"
    )):
        el.decompose()
    for el in soup.find_all(id="comments"):
        el.decompose()
    return (soup.find("article", class_="recipe") or soup).get_text(separator=" ")


def build_index_and_idf():
    """Парсит hw4/lemmas/*.txt -> index.json (tfidf > 0) и idf.json."""
    index = {}
    idf = {}

    for path in sorted(LEMMAS_DIR.glob("*.txt")):
        doc_id = path.stem
        vec = {}
        for line in path.read_text(encoding="utf-8").strip().split("\n"):
            parts = line.split()
            lemma, idf_val, tfidf_val = parts[0], float(parts[1]), float(parts[2])
            idf[lemma] = idf_val
            if tfidf_val > 0:
                vec[lemma] = tfidf_val
        index[doc_id] = vec

    return index, idf


def load_urls():
    """Загружает маппинг doc_id -> URL из hw1/index.txt."""
    urls = {}
    for line in INDEX_FILE.read_text(encoding="utf-8").strip().split("\n"):
        doc_id, url = line.split(maxsplit=1)
        urls[doc_id] = url
    return urls


def build_docs():
    """Извлекает заголовки, тексты и URL из HTML страниц."""
    urls = load_urls()
    docs = {}
    for path in sorted(PAGES_DIR.glob("*.html")):
        doc_id = path.stem
        html = path.read_text(encoding="utf-8", errors="replace")
        soup = BeautifulSoup(html, "lxml")

        title = soup.title.string.strip() if soup.title and soup.title.string else doc_id
        # Убираем суффиксы вроде " | Меню недели"
        title = re.split(r"\s*[|–-]\s*Меню недели", title)[0].strip()

        text = extract_text(html)
        text = re.sub(r"\s+", " ", text).strip()

        docs[doc_id] = {
            "title": title,
            "text": text,
            "url": urls.get(doc_id, ""),
        }

    return docs


def write_json(path, data):
    path.write_text(
        json.dumps(data, ensure_ascii=False, separators=(",", ":")),
        encoding="utf-8",
    )
    print(f"  {path.name}: {path.stat().st_size // 1024}K")


def main():
    DATA_DIR.mkdir(exist_ok=True)

    print("Building index and IDF...")
    index, idf = build_index_and_idf()
    write_json(DATA_DIR / "index.json", index)
    write_json(DATA_DIR / "idf.json", idf)

    print("Building docs (titles + texts)...")
    docs = build_docs()
    write_json(DATA_DIR / "docs.json", docs)

    print(f"Done: {len(index)} docs, {len(idf)} lemmas")


if __name__ == "__main__":
    main()
