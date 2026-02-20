"""
Шаг 1: сбор ссылок на рецепты.

Читает XML-сайтмапы сайта menunedeli.ru, берёт первые URLS_PER_SITEMAP
ссылок из каждого и записывает итоговый список в index.txt.
"""

import xml.etree.ElementTree as ET

import requests

# Три сайтмапа дают разнообразие рецептов из разных временных срезов
SITEMAPS = [
    "https://menunedeli.ru/recipe-sitemap.xml",
    "https://menunedeli.ru/recipe-sitemap2.xml",
    "https://menunedeli.ru/recipe-sitemap3.xml",
]
URLS_PER_SITEMAP = 60
INDEX_FILE = "index.txt"

# Пространство имён стандарта Sitemap Protocol
NS = {"sm": "http://www.sitemaps.org/schemas/sitemap/0.9"}

HEADERS = {
    "User-Agent": (
        "Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7) "
        "AppleWebKit/537.36 (KHTML, like Gecko) "
        "Chrome/120.0.0.0 Safari/537.36"
    )
}


def fetch_urls(sitemap_url: str) -> list[str]:
    """Скачивает сайтмап и возвращает список URL из тегов <loc>."""
    print(f"Получаем сайтмап: {sitemap_url}")
    resp = requests.get(sitemap_url, headers=HEADERS, timeout=15)
    resp.raise_for_status()
    root = ET.fromstring(resp.text)
    urls = [loc.text for loc in root.findall(".//sm:loc", NS) if loc.text]
    print(f"  Найдено URL: {len(urls)}, берём первые {URLS_PER_SITEMAP}")
    return urls[:URLS_PER_SITEMAP]


def main() -> None:
    # Собираем URL из всех сайтмапов в единый список
    all_urls: list[str] = []
    for sitemap_url in SITEMAPS:
        try:
            all_urls.extend(fetch_urls(sitemap_url))
        except Exception as e:
            print(f"  Ошибка при загрузке {sitemap_url}: {e}")

    # Записываем index.txt: порядковый номер (4 цифры) + URL
    with open(INDEX_FILE, "w", encoding="utf-8") as f:
        for i, url in enumerate(all_urls, start=1):
            f.write(f"{i:04d} {url}\n")

    print(f"\nГотово. Записано {len(all_urls)} ссылок в {INDEX_FILE}")


if __name__ == "__main__":
    main()
