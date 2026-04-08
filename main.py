"""
KeyboardSwitcher — головний файл.
Запускає tray іконку + keyboard hook.
"""

import sys
import os
import keyboard

from keyboard_switcher import key_buffer, on_key_event, convert_buffer
from tray import create_tray


def main():
    keyboard.hook(on_key_event)
    keyboard.add_hotkey('ctrl+`', convert_buffer, suppress=True)

    def on_quit():
        keyboard.unhook_all()
        os._exit(0)

    icon = create_tray(on_quit=on_quit)
    icon.run()


if __name__ == '__main__':
    main()
