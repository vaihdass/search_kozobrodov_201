import json
import re
import ssl
from pathlib import Path

from bs4 import BeautifulSoup
import pymorphy3
import nltk

# Обход проблемы с SSL-сертификатами при загрузке данных nltk
ssl._create_default_https_context = ssl._create_unverified_context
nltk.download("stopwords", quiet=True)
from nltk.corpus import stopwords

BASE_DIR = Path(__file__).resolve().parent
PAGES_DIR = BASE_DIR / "../pages"

STOP_WORDS = set(stopwords.words("russian"))
SKIP_POS = {"PREP", "CONJ", "PRCL", "INTJ", "NUMR"}

morph = pymorphy3.MorphAnalyzer()


def extract_text(html):
    """Извлекает чистый текст рецепта из HTML."""
    soup = BeautifulSoup(html, "lxml")

    for tag in soup.find_all({"script", "style", "nav", "header", "footer", "aside"}):
        tag.decompose()
    for el in soup.find_all(class_=re.compile(
        r"menu-main|similar-recipe|banner|author-card-short|post-category|rating|comments"
    )):
        el.decompose()
    for el in soup.find_all(id="comments"):
        el.decompose()

    return (soup.find("article", class_="recipe") or soup).get_text(separator=" ")


def tokenize_and_lemmatize(text):
    """Возвращает set лемм из текста."""
    lemmas = set()
    for word in re.findall(r"[а-яёА-ЯЁ]+", text):
        t = word.lower()
        parsed = morph.parse(t)[0]
        if t not in STOP_WORDS and parsed.tag.POS not in SKIP_POS:
            lemmas.add(parsed.normal_form)
    return lemmas


def main():
    html_files = sorted(PAGES_DIR.glob("*.html"))
    print(f"Found {len(html_files)} HTML files")

    # Инвертированный индекс: лемма -> множество doc_id
    index = {}
    for path in html_files:
        doc_id = path.stem
        html = path.read_text(encoding="utf-8", errors="replace")
        for lemma in tokenize_and_lemmatize(extract_text(html)):
            index.setdefault(lemma, set()).add(doc_id)

    # set -> sorted list для JSON
    out = {k: sorted(v) for k, v in sorted(index.items())}
    (BASE_DIR / "inverted_index.json").write_text(
        json.dumps(out, ensure_ascii=False, separators=(",", ":")) + "\n",
        encoding="utf-8",
    )
    print(f"Index built: {len(out)} lemmas")


if __name__ == "__main__":
    main()
