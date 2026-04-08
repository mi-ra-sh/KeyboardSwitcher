"""
KeyboardSwitcher — конвертує останнє набране слово між EN/UA розкладками.
Хоткей: Ctrl+` (налаштовується)
Працює глобально: браузери, термінали, IDE, месенджери.
"""

import keyboard
import time
import sys
import os
import threading
from collections import deque

# ===== Маппінг клавіш EN <-> UA =====
# Фізично ті самі клавіші, різні символи в залежності від розкладки

EN_CHARS = "`qwertyuiop[]\\asdfghjkl;'zxcvbnm,./~QWERTYUIOP{}|ASDFGHJKL:\"ZXCVBNM<>?"
UA_CHARS = "'йцукенгшщзхїґфівапролджєячсмитьбю.₴ЙЦУКЕНГШЩЗХЇҐФІВАПРОЛДЖЄЯЧСМИТЬБЮ,"

# Числовий ряд і спецсимволи (Shift варіанти)
EN_NUMS = "1234567890-="
UA_NUMS = "1234567890-="
EN_NUMS_SHIFT = "!@#$%^&*()_+"
UA_NUMS_SHIFT = "!\"№;%:?*()_+"

# Побудова таблиць перетворення
_en_to_ua = {}
_ua_to_en = {}

for en, ua in zip(EN_CHARS, UA_CHARS):
    _en_to_ua[en] = ua
    _ua_to_en[ua] = en

for en, ua in zip(EN_NUMS_SHIFT, UA_NUMS_SHIFT):
    _en_to_ua[en] = ua
    _ua_to_en[ua] = en


def convert_en_to_ua(text: str) -> str:
    """Конвертує текст набраний на EN розкладці в UA."""
    return ''.join(_en_to_ua.get(c, c) for c in text)


def convert_ua_to_en(text: str) -> str:
    """Конвертує текст набраний на UA розкладці в EN."""
    return ''.join(_ua_to_en.get(c, c) for c in text)


def is_en_chars(text: str) -> bool:
    """Перевіряє чи текст містить переважно EN символи."""
    en_count = sum(1 for c in text if c.lower() in 'abcdefghijklmnopqrstuvwxyz')
    return en_count > len(text) * 0.5


def is_ua_chars(text: str) -> bool:
    """Перевіряє чи текст містить переважно UA символи."""
    ua_count = sum(1 for c in text if c.lower() in 'абвгґдеєжзиіїйклмнопрстуфхцчшщьюя')
    return ua_count > len(text) * 0.5


# ===== Буфер набраних символів =====

class KeyBuffer:
    """Зберігає останнє набране слово."""

    def __init__(self, max_size: int = 200):
        self.buffer = deque(maxlen=max_size)
        self.lock = threading.Lock()

    def add_char(self, char: str):
        with self.lock:
            self.buffer.append(char)

    def clear(self):
        with self.lock:
            self.buffer.clear()

    def get_last_word(self) -> str:
        """Повертає останнє слово з буфера (до пробілу/переносу)."""
        with self.lock:
            text = ''.join(self.buffer)
        # Знаходимо останнє слово
        stripped = text.rstrip()
        if not stripped:
            return ''
        # Шукаємо початок останнього слова
        for i in range(len(stripped) - 1, -1, -1):
            if stripped[i] in ' \t\n\r':
                return stripped[i + 1:]
        return stripped

    def get_word_length(self) -> int:
        """Кількість символів в останньому слові."""
        return len(self.get_last_word())

    def remove_last_word(self):
        """Видаляє останнє слово з буфера."""
        with self.lock:
            text = ''.join(self.buffer)
            stripped = text.rstrip()
            if not stripped:
                return
            for i in range(len(stripped) - 1, -1, -1):
                if stripped[i] in ' \t\n\r':
                    self.buffer.clear()
                    self.buffer.extend(text[:i + 1])
                    return
            self.buffer.clear()

    def replace_last_word(self, new_word: str):
        """Замінює останнє слово в буфері."""
        with self.lock:
            text = ''.join(self.buffer)
            stripped = text.rstrip()
            if not stripped:
                return
            for i in range(len(stripped) - 1, -1, -1):
                if stripped[i] in ' \t\n\r':
                    self.buffer.clear()
                    self.buffer.extend(text[:i + 1] + new_word)
                    return
            self.buffer.clear()
            self.buffer.extend(new_word)


# ===== Головна логіка =====

key_buffer = KeyBuffer()
_converting = False  # Прапорець для ігнорування власного вводу


def on_key_event(event: keyboard.KeyboardEvent):
    """Обробник натискань клавіш."""
    global _converting

    if _converting:
        return

    if event.event_type != 'down':
        return

    name = event.name

    # Ігноруємо модифікатори та спеціальні клавіші
    if name in ('ctrl', 'alt', 'shift', 'left ctrl', 'right ctrl',
                'left alt', 'right alt', 'left shift', 'right shift',
                'left windows', 'right windows', 'caps lock',
                'num lock', 'scroll lock', 'print screen',
                'insert', 'delete', 'home', 'end', 'page up', 'page down',
                'up', 'down', 'left', 'right',
                'f1', 'f2', 'f3', 'f4', 'f5', 'f6',
                'f7', 'f8', 'f9', 'f10', 'f11', 'f12'):
        return

    # Enter, Tab — очищуємо буфер
    if name in ('enter', 'tab'):
        key_buffer.clear()
        return

    # Backspace — видаляємо останній символ
    if name == 'backspace':
        with key_buffer.lock:
            if key_buffer.buffer:
                key_buffer.buffer.pop()
        return

    # Пробіл — додаємо (розділювач слів)
    if name == 'space':
        key_buffer.add_char(' ')
        return

    # Звичайний символ
    if len(name) == 1:
        key_buffer.add_char(name)


def convert_last_word():
    """Конвертує останнє набране слово між розкладками."""
    global _converting

    word = key_buffer.get_last_word()
    if not word:
        return

    # Визначаємо напрямок конвертації
    if is_ua_chars(word):
        converted = convert_ua_to_en(word)
    elif is_en_chars(word):
        converted = convert_en_to_ua(word)
    else:
        # Спробуємо обидва напрямки — якщо є в маппінгу UA→EN
        converted = convert_ua_to_en(word)
        if converted == word:
            converted = convert_en_to_ua(word)
        if converted == word:
            return  # Нічого конвертувати

    word_len = len(word)

    _converting = True
    try:
        # Видаляємо старе слово через backspace
        for _ in range(word_len):
            keyboard.send('backspace')
            time.sleep(0.01)

        time.sleep(0.02)

        # Друкуємо нове слово
        keyboard.write(converted, delay=0.01)

        # Оновлюємо буфер
        key_buffer.replace_last_word(converted)
    finally:
        _converting = False


def main():
    print("KeyboardSwitcher v1.0")
    print(f"Хоткей: Ctrl+`")
    print("Натисніть Ctrl+` щоб конвертувати останнє слово між EN/UA")
    print("Ctrl+C для виходу")
    print("-" * 40)

    # Реєструємо глобальний хук
    keyboard.hook(on_key_event)

    # Реєструємо хоткей
    keyboard.add_hotkey('ctrl+`', convert_last_word, suppress=True)

    print("Працюю... (буфер активний)")

    try:
        keyboard.wait()
    except KeyboardInterrupt:
        print("\nВихід.")
        keyboard.unhook_all()
        sys.exit(0)


if __name__ == '__main__':
    main()
