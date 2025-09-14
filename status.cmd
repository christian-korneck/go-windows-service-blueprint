@echo off
setlocal

sc query svctest1

exit /B %errorlevel%