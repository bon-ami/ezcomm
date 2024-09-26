@echo off
setlocal enabledelayedexpansion

set paramsIn=
:loopIn
if "%~1"=="" goto endloopIn
set "paramsIn=!paramsIn! %1"
shift
goto loopIn

:endloopIn
call sample_flow_common.bat out %paramsIn%
