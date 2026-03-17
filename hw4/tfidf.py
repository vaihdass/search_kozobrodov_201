import math
import re
import ssl
from collections import Counter
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


# Переиспользовал код из задания 3
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


def get_tokens(text):
    """Возвращает список токенов с повторами (для подсчета TF)."""
    tokens = []
    for word in re.findall(r"[а-яёА-ЯЁ]+", text):
        t = word.lower()
        if t not in STOP_WORDS and morph.parse(t)[0].tag.POS not in SKIP_POS:
            tokens.append(t)
    return tokens


def lemmatize(token):
    return morph.parse(token)[0].normal_form


def compute_idf(doc_freqs, total_docs):
    """IDF = log(N / df) для каждого термина."""
    return {term: math.log(total_docs / df) for term, df in doc_freqs.items()}


def write_tfidf(path, term_counts, total_terms, idf):
    """Записывает файл в формате: <термин> <idf> <tf-idf>"""
    lines = []
    for term in sorted(term_counts):
        tf = term_counts[term] / total_terms
        tfidf = tf * idf[term]
        lines.append(f"{term} {idf[term]:.6f} {tfidf:.6f}")
    path.write_text("\n".join(lines) + "\n", encoding="utf-8")


def main():
    html_files = sorted(PAGES_DIR.glob("*.html"))
    total_docs = len(html_files)
    print(f"Found {total_docs} HTML files")

    # Этап 1: парсим все документы, собираем токены и леммы с повторами
    docs_tokens = {}   # doc_id -> Counter(token -> count)
    docs_lemmas = {}   # doc_id -> Counter(lemma -> count)
    docs_total = {}    # doc_id -> общее кол-во токенов

    for path in html_files:
        doc_id = path.stem
        html = path.read_text(encoding="utf-8", errors="replace")
        tokens = get_tokens(extract_text(html))

        docs_tokens[doc_id] = Counter(tokens)
        docs_lemmas[doc_id] = Counter(lemmatize(t) for t in tokens)
        docs_total[doc_id] = len(tokens)

    # Этап 2: считаем document frequency (в скольких документах встречается термин)
    token_df = Counter()
    lemma_df = Counter()
    for doc_id in docs_tokens:
        for token in docs_tokens[doc_id]:
            token_df[token] += 1
        for lemma in docs_lemmas[doc_id]:
            lemma_df[lemma] += 1

    # Этап 3: считаем IDF
    token_idf = compute_idf(token_df, total_docs)
    lemma_idf = compute_idf(lemma_df, total_docs)

    # Этап 4: записываем TF-IDF файлы
    tokens_dir = BASE_DIR / "tokens"
    lemmas_dir = BASE_DIR / "lemmas"
    tokens_dir.mkdir(exist_ok=True)
    lemmas_dir.mkdir(exist_ok=True)

    for doc_id in sorted(docs_tokens):
        total = docs_total[doc_id]
        write_tfidf(tokens_dir / f"{doc_id}.txt", docs_tokens[doc_id], total, token_idf)
        write_tfidf(lemmas_dir / f"{doc_id}.txt", docs_lemmas[doc_id], total, lemma_idf)

    print(f"Written: tokens/ ({len(docs_tokens)} files)")
    print(f"Written: lemmas/ ({len(docs_lemmas)} files)")


if __name__ == "__main__":
    main()
