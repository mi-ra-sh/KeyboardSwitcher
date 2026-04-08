"""
Системний трей для KeyboardSwitcher.
"""

import pystray
from PIL import Image, ImageDraw, ImageFont
import sys


def create_icon_image(text="KB", bg_color="#2196F3", fg_color="white"):
    """Створює іконку для трею."""
    size = 64
    img = Image.new('RGBA', (size, size), (0, 0, 0, 0))
    draw = ImageDraw.Draw(img)
    draw.rounded_rectangle([2, 2, size - 2, size - 2], radius=10, fill=bg_color)

    try:
        font = ImageFont.truetype("segoeui.ttf", 24)
    except (OSError, IOError):
        font = ImageFont.load_default()

    bbox = draw.textbbox((0, 0), text, font=font)
    text_w = bbox[2] - bbox[0]
    text_h = bbox[3] - bbox[1]
    x = (size - text_w) // 2
    y = (size - text_h) // 2 - 2
    draw.text((x, y), text, fill=fg_color, font=font)

    return img


def create_tray(on_quit=None):
    """Створює системний трей."""

    def quit_action(icon, item):
        icon.stop()
        if on_quit:
            on_quit()

    menu = pystray.Menu(
        pystray.MenuItem("KeyboardSwitcher v1.0", None, enabled=False),
        pystray.Menu.SEPARATOR,
        pystray.MenuItem("Ctrl+` — конвертувати текст", None, enabled=False),
        pystray.Menu.SEPARATOR,
        pystray.MenuItem("Вихід", quit_action),
    )

    icon = pystray.Icon(
        "KeyboardSwitcher",
        create_icon_image(),
        "KeyboardSwitcher — Ctrl+`",
        menu
    )

    return icon


if __name__ == '__main__':
    icon = create_tray(on_quit=lambda: sys.exit(0))
    icon.run()
