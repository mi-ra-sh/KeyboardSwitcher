"""
KeyboardSwitcher — головний файл.
Запускає tray іконку + keyboard hook.
"""

import sys
import os
import threading
import keyboard

from keyboard_switcher import key_buffer, on_key_event, convert_last_word
from tray import create_tray, create_icon_image


def main():
    # Реєструємо глобальний keyboard hook
    keyboard.hook(on_key_event)
    keyboard.add_hotkey('ctrl+`', convert_last_word, suppress=True)

    def on_quit():
        keyboard.unhook_all()
        os._exit(0)

    # Створюємо tray іконку
    icon = create_tray(on_quit=on_quit)

    # Запускаємо tray (блокуючий виклик)
    icon.run()


if __name__ == '__main__':
    main()
