import re
import ssl
from pathlib import Path

from bs4 import BeautifulSoup
import pymorphy3
import nltk

# Хак: обход проблемы с SSL-сертификатами при загрузке данных nltk
ssl._create_default_https_context = ssl._create_unverified_context
nltk.download("stopwords", quiet=True)
from nltk.corpus import stopwords

BASE_DIR = Path(__file__).resolve().parent

# Стоп-слова из nltk (предлоги, союзы, местоимения и т.д.)
STOP_WORDS = set(stopwords.words("russian"))

# Части речи pymorphy3, которые отфильтровываем:
# PREP - предлоги, CONJ - союзы, PRCL - частицы,
# INTJ - междометия, NUMR - числительные
SKIP_POS = {"PREP", "CONJ", "PRCL", "INTJ", "NUMR"}

morph = pymorphy3.MorphAnalyzer()


def extract_text(html: str) -> str:
    """Извлекает чистый текст рецепта из HTML.

    Удаляет мусорные теги и блоки (навигация, реклама, комментарии),
    берет текст из <article class="recipe">.
    """
    soup = BeautifulSoup(html, "lxml")

    # Удаляем служебные теги и мусорные блоки
    for tag in soup.find_all({"script", "style", "nav", "header", "footer", "aside"}):
        tag.decompose()
    for el in soup.find_all(class_=re.compile(
        r"menu-main|similar-recipe|banner|author-card-short|post-category|rating|comments"
    )):
        el.decompose()
    for el in soup.find_all(id="comments"):
        el.decompose()

    # Основной контент рецепта
    article = soup.find("article", class_="recipe")
    return (article or soup).get_text(separator=" ")


def tokenize(text: str) -> set[str]:
    """Токенизация: только кириллица, без стоп-слов и служебных частей речи.

    Regex отсекает числа, смешанные токены и обрывки разметки.
    """
    tokens = set()
    for word in re.findall(r"[а-яёА-ЯЁ]+", text):
        t = word.lower()
        if t not in STOP_WORDS and morph.parse(t)[0].tag.POS not in SKIP_POS:
            tokens.add(t)
    return tokens


def main():
    # Этап 1: извлекаем текст из HTML и собираем уникальные токены
    html_files = sorted((BASE_DIR / "../pages").glob("*.html"))
    print(f"Found {len(html_files)} HTML files")

    all_tokens: set[str] = set()
    for path in html_files:
        html = path.read_text(encoding="utf-8", errors="replace")
        all_tokens.update(tokenize(extract_text(html)))

    sorted_tokens = sorted(all_tokens)
    print(f"Total unique tokens: {len(sorted_tokens)}")

    # Этап 2: записываем токены в tokens.txt
    (BASE_DIR / "tokens.txt").write_text(
        "\n".join(sorted_tokens) + "\n", encoding="utf-8"
    )

    # Этап 3: лемматизация - группируем токены по нормальной форме (pymorphy3)
    lemmas: dict[str, set[str]] = {}
    for token in sorted_tokens:
        lemma = morph.parse(token)[0].normal_form
        lemmas.setdefault(lemma, set()).add(token)

    # Записываем в формате: <лемма> <токен1> <токен2> ... <токенN>
    lines = [
        f"{lemma} {' '.join(sorted(lemmas[lemma]))}"
        for lemma in sorted(lemmas)
    ]
    (BASE_DIR / "lemmas.txt").write_text(
        "\n".join(lines) + "\n", encoding="utf-8"
    )

    print(f"Total unique lemmas: {len(lemmas)}")


if __name__ == "__main__":
    main()
