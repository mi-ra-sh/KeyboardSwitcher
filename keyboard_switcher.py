"""
KeyboardSwitcher — змінює розкладку вже набраного тексту (EN↔UA).
Перекодовує символи по фізичних клавішах: A↔Ф, S↔І, D↔В, і т.д.

Хоткей:
  Ctrl+` — конвертує весь набраний текст з буфера (від останнього Enter)

Працює глобально: браузери, термінали, IDE, месенджери.
"""

import keyboard
import time
import sys
import os
import threading
from collections import deque

# ===== Маппінг фізичних клавіш EN <-> UA =====

EN_CHARS = "`qwertyuiop[]\\asdfghjkl;'zxcvbnm,./~QWERTYUIOP{}|ASDFGHJKL:\"ZXCVBNM<>?"
UA_CHARS = "'йцукенгшщзхїґфівапролджєячсмитьбю.₴ЙЦУКЕНГШЩЗХЇҐФІВАПРОЛДЖЄЯЧСМИТЬБЮ,"

EN_NUMS_SHIFT = "!@#$%^&*()_+"
UA_NUMS_SHIFT = "!\"№;%:?*()_+"

_en_to_ua = {}
_ua_to_en = {}

for en, ua in zip(EN_CHARS, UA_CHARS):
    _en_to_ua[en] = ua
    _ua_to_en[ua] = en

for en, ua in zip(EN_NUMS_SHIFT, UA_NUMS_SHIFT):
    _en_to_ua[en] = ua
    _ua_to_en[ua] = en


def convert_text(text: str) -> str:
    """Конвертує текст між розкладками. Автоматично визначає напрямок."""
    if not text:
        return text

    ua_count = sum(1 for c in text if c.lower() in 'абвгґдеєжзиіїйклмнопрстуфхцчшщьюя')
    en_count = sum(1 for c in text if c.lower() in 'abcdefghijklmnopqrstuvwxyz')

    if ua_count > en_count:
        return ''.join(_ua_to_en.get(c, c) for c in text)
    elif en_count > ua_count:
        return ''.join(_en_to_ua.get(c, c) for c in text)
    else:
        converted = ''.join(_ua_to_en.get(c, c) for c in text)
        if converted != text:
            return converted
        return ''.join(_en_to_ua.get(c, c) for c in text)


# ===== Буфер набраних символів =====

class KeyBuffer:
    """Зберігає все набране з моменту останнього Enter."""

    def __init__(self, max_size: int = 2000):
        self.buffer = deque(maxlen=max_size)
        self.lock = threading.Lock()

    def add_char(self, char: str):
        with self.lock:
            self.buffer.append(char)

    def pop_char(self):
        with self.lock:
            if self.buffer:
                self.buffer.pop()

    def clear(self):
        with self.lock:
            self.buffer.clear()

    def get_text(self) -> str:
        """Повертає весь текст у буфері."""
        with self.lock:
            return ''.join(self.buffer)

    def set_text(self, text: str):
        """Замінює вміст буфера."""
        with self.lock:
            self.buffer.clear()
            self.buffer.extend(text)

    def length(self) -> int:
        with self.lock:
            return len(self.buffer)


# ===== Головна логіка =====

key_buffer = KeyBuffer()
_converting = False


def on_key_event(event: keyboard.KeyboardEvent):
    """Обробник натискань — буферизує всі набрані символи."""
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
                'f7', 'f8', 'f9', 'f10', 'f11', 'f12',
                'escape'):
        return

    # Enter — очищуємо буфер (текст "закомічений")
    if name == 'enter':
        key_buffer.clear()
        return

    # Tab — очищуємо (перехід фокусу)
    if name == 'tab':
        key_buffer.clear()
        return

    # Backspace — видаляємо останній символ з буфера
    if name == 'backspace':
        key_buffer.pop_char()
        return

    # Пробіл
    if name == 'space':
        key_buffer.add_char(' ')
        return

    # Звичайний символ
    if len(name) == 1:
        key_buffer.add_char(name)


def convert_buffer():
    """Ctrl+` — конвертує весь буфер (все набране з останнього Enter)."""
    global _converting

    text = key_buffer.get_text()
    if not text:
        return

    converted = convert_text(text)
    if converted == text:
        return

    text_len = len(text)

    _converting = True
    try:
        # Видаляємо весь набраний текст через backspace
        for _ in range(text_len):
            keyboard.send('backspace')
            time.sleep(0.005)

        time.sleep(0.03)

        # Друкуємо конвертований текст
        keyboard.write(converted, delay=0.005)

        # Оновлюємо буфер
        key_buffer.set_text(converted)
    finally:
        _converting = False


def main():
    print("KeyboardSwitcher v1.0")
    print("Ctrl+` — конвертувати набраний текст (EN↔UA)")
    print("Enter  — очистити буфер")
    print("Ctrl+C — вихід")
    print("-" * 40)

    keyboard.hook(on_key_event)
    keyboard.add_hotkey('ctrl+`', convert_buffer, suppress=True)

    print("Працюю...")

    try:
        keyboard.wait()
    except KeyboardInterrupt:
        print("\nВихід.")
        keyboard.unhook_all()
        sys.exit(0)


if __name__ == '__main__':
    main()
