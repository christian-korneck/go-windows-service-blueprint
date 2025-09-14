@echo off
setlocal

sc start svctest1

exit /B %errorlevel%