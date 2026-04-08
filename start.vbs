Set WshShell = CreateObject("WScript.Shell")
WshShell.CurrentDirectory = "C:\KeyboardSwitcher"
WshShell.Run "pythonw.exe main.py", 0, False
