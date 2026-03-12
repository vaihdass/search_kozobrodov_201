"""Булев парсер запросов (рекурсивный спуск).

Приоритет операторов (от высшего к низшему): NOT, AND, OR.
Скобки позволяют переопределить порядок вычислений.
Каждое слово запроса лемматизируется перед поиском в индексе.
"""

import re

import pymorphy3

TOKEN_RE = re.compile(r"AND|OR|NOT|[()]|[а-яёА-ЯЁ]+", re.IGNORECASE)
morph = pymorphy3.MorphAnalyzer()


def lemmatize(word):
    return morph.parse(word.lower())[0].normal_form


def search(query, index, all_docs):
    """Выполняет булев запрос, возвращает set(doc_id)."""
    tokens = TOKEN_RE.findall(query)
    pos = [0]  # список для мутации из замыканий

    def current():
        return tokens[pos[0]] if pos[0] < len(tokens) else None

    def consume(expected):
        if current() != expected:
            raise ValueError(f"ожидался '{expected}', получен '{current()}'")
        pos[0] += 1

    def or_expr():
        result = and_expr()
        while current() and current().upper() == "OR":
            pos[0] += 1
            result |= and_expr()
        return result

    def and_expr():
        result = not_expr()
        while current() and current().upper() == "AND":
            pos[0] += 1
            result &= not_expr()
        return result

    def not_expr():
        if current() and current().upper() == "NOT":
            pos[0] += 1
            return all_docs - not_expr()
        return atom()

    def atom():
        tok = current()
        if tok == "(":
            pos[0] += 1
            result = or_expr()
            consume(")")
            return result
        if tok is None:
            raise ValueError("неожиданный конец запроса")
        pos[0] += 1
        return index.get(lemmatize(tok), set())

    if not tokens:
        return set()

    result = or_expr()
    if pos[0] < len(tokens):
        raise ValueError(f"неожиданный токен: '{tokens[pos[0]]}'")
    return result
