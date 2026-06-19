@echo off
REM Launches `quail serve` and opens the local UI in the default browser.
REM Double-click from Explorer, or pin to Start menu / taskbar.
REM Works regardless of where the archive was extracted — %~dp0 resolves
REM to this .bat's own directory.
start "" /B "%~dp0quail.exe" serve
REM Give the server a moment to bind before opening the browser.
timeout /t 2 /nobreak >NUL
start "" "http://127.0.0.1:8765/"
