"""
Шаг 2: скачивание HTML-страниц по списку из index.txt.

Для каждой строки index.txt делает GET-запрос и сохраняет
сырой HTML-ответ в pages/NNNN.html (без картинок, CSS и JS).
"""

import os
import time

import requests

INDEX_FILE = "index.txt"
PAGES_DIR = "../pages"
DELAY = 0.5  # задержка между запросами, чтобы не задудосить сервер

HEADERS = {
    "User-Agent": (
        "Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7) "
        "AppleWebKit/537.36 (KHTML, like Gecko) "
        "Chrome/120.0.0.0 Safari/537.36"
    )
}


def main() -> None:
    os.makedirs(PAGES_DIR, exist_ok=True)

    # Читаем заранее подготовленный список ссылок
    with open(INDEX_FILE, encoding="utf-8") as f:
        entries = [line.strip().split(maxsplit=1) for line in f if line.strip()]

    print(f"Загружаем {len(entries)} страниц из {INDEX_FILE}\n")

    downloaded = 0
    for num_str, url in entries:
        filepath = os.path.join(PAGES_DIR, f"{num_str}.html")

        try:
            # Обычный HTTP GET, возвращает только HTML
            resp = requests.get(url, headers=HEADERS, timeout=15)
            resp.raise_for_status()
            with open(filepath, "w", encoding="utf-8") as f:
                f.write(resp.text)
            downloaded += 1
            print(f"[{num_str}] OK  {url}")
        except Exception as e:
            print(f"[{num_str}] ERR {url}: {e}")

        time.sleep(DELAY)

    print(f"\nГотово. Скачано {downloaded} из {len(entries)} страниц в ./{PAGES_DIR}/")


if __name__ == "__main__":
    main()
