@echo off
setlocal

sc create svctest1 binPath= "%~dp0\svctest1.exe" start= auto
sc start svctest1

exit /B %errorlevel%

